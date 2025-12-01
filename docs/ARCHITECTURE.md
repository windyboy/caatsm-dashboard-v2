# CAATSM Dashboard Architecture

This document describes how the CAATSM Dashboard backend is organized today. It focuses on practical guidance: which packages own which responsibilities, how data flows through the system, and what you should touch when adding new features.

---

## 1. Overview

- **Goal:** Collect aviation telegrams, validate them, persist them, index them for search, and serve them to operators in real time.
- **Stack:** Go backend (Echo, pgx), PostgreSQL/TimescaleDB, Meilisearch, Valkey 9 (or Redis 7+), NATS 2.12.2 JetStream, SvelteKit 2.x frontend with Svelte 5.
- **Architecture Style:** Clean Architecture with three concentric layers. Domain entities and validation live inside the Application layer rather than a standalone `internal/domain` package.

```
Delivery ──► Application ──► Infrastructure
```

Dependencies only point inward: the Delivery layer depends on Application services; Infrastructure implements Application ports.

---

## 2. Layer Responsibilities

### Delivery (`internal/delivery/`)

- HTTP and WebSocket handlers built with Echo.
- Request parsing, lightweight input validation (range checks, rate limiting).
- Response formatting, error to JSON translation, CSV streaming.
- No direct imports from infrastructure packages.

### Application (`internal/app/`)

- Houses domain entities (`Telegram`, `SearchFilters`, `TimeWindow`, events) plus validation logic (90-day window cap, whitelist of sortable fields, enum validation). These entities live directly in `internal/app/` rather than a separate `internal/domain/` package.
- Provides services: `DashboardService`, `SearchService`, `StatsService`, `ExportService`, `RealtimeService`.
- Defines port interfaces in `ports.go` for repository/search/cache/event/websocket abstractions.
- Exposes a dependency container (`Container`) that wires config, logger, ports, and services. Health checks are implemented as `Container.HealthCheck()` in `internal/app/app.go`.
- Emits domain events consumed by infrastructure adapters.

### Infrastructure (`internal/infrastructure/`)

- Concrete implementations of Application ports:
  - `persistence/` – PostgreSQL repository via pgx and TimescaleDB.
  - `search/` – Meilisearch client for full-text and autocomplete.
  - `cache/` – Valkey 9 (or Redis 7+) cache for stats and query responses.
  - `streaming/` – NATS 2.12.2 JetStream consumer for ingestion.
  - `event/` – Pub/Sub bridge for domain events.
  - `ws/` – WebSocket hub with backpressure and per-IP limits.
- Handles connection lifecycle, retries, and adapter-specific metrics.

---

## 3. Data Flow

### Ingestion Path

1. NATS JetStream receives raw telegram messages.
2. `cmd/sync` worker pulls messages via the streaming adapter.
3. Application services validate and normalize telegrams (enforcing domain rules).
4. Repository stores telegrams in PostgreSQL/TimescaleDB.
5. Search adapter updates Meilisearch indices.
6. Event adapter publishes a `TelegramProcessed` domain event.
7. Realtime service pushes updates to the WebSocket hub.

### Query / Dashboard Path

1. Client issues REST requests or subscribes over WebSocket.
2. Delivery layer authenticates, rate-limits, and normalizes parameters.
3. Application services coordinate cache lookups, repository queries, and search index calls.
4. Validation ensures requests remain within the 90-day window and use supported sort keys.
5. Responses are returned as JSON or streamed CSV; realtime updates broadcast via WebSocket.

---

## 4. Application Services and Ports

| Service            | Purpose                                        | Ports Consumed                                |
| ------------------ | ---------------------------------------------- | --------------------------------------------- |
| `DashboardService` | Aggregated metrics for dashboard widgets       | `Repository`, `Cache`                         |
| `SearchService`    | Telegram search, autocomplete, CSV prep        | `Repository`, `SearchIndex`, `Cache`          |
| `StatsService`     | Priority/type aggregations, cached snapshots   | `Repository`, `Cache`                         |
| `ExportService`    | Streaming CSV export (chunked to avoid OOM)    | `Repository`                                  |
| `RealtimeService`  | Broadcasts domain events to WebSocket clients  | `EventPublisher`, `WebSocketHubPort`          |

**Health Checks:** Health check functionality is implemented in `internal/app/app.go` as `Container.HealthCheck()`, which verifies connectivity to PostgreSQL, Meilisearch, and Redis. The HTTP handler in `internal/delivery/http/handlers.go` exposes this via the `/api/health` endpoint.

All ports live in `internal/app/ports.go`. Add new ports there whenever you need to integrate an external system.

---

## 5. Runtime Components

- **HTTP Server (`internal/server/`):** Configures Echo, middleware (logging, correlation IDs, rate limiting), and wires handlers.
- **WebSocket Hub (`internal/infrastructure/ws/`):** Manages clients, enforces per-IP and global caps, applies backpressure by dropping slow consumers.
- **Sync Worker (`cmd/sync`):** Consumes NATS messages, invokes Application services, and acknowledges messages.
- **Observability (`internal/observability/`):** Provides structured logging (Zap) and Prometheus metrics. Tracing configuration keys exist, but OpenTelemetry exporters are not wired yet.

---

## 6. Configuration Model

- Config structs live in `config/config.go`; loader merges defaults, TOML files, and environment variables (prefix `CAATSM_`).
- `AppConfig.Validate()` enforces guardrails: TLS, non-default secrets, enabled auth in production, tracing enabled flag, etc.
- Production deployments must set `CAATSM_ENVIRONMENT=production` and satisfy validation (TLS on, no default credentials, at least one auth mechanism).

---

## 7. Observability

- **Logging:** JSON or console (based on config) with correlation IDs injected via middleware.
- **Metrics:** Prometheus metrics at `/metrics`, including HTTP request histograms, ingestion counters, WebSocket gauges (optional).
- **Tracing:** Configuration flags (`tracing.enabled`, `tracing.otlp_endpoint`) exist, but the exporter/instrumentation is not currently implemented. Keep `tracing.enabled=false` until tracing support is added.

---

## 8. Resilience Features

- 90-day time window guard prevents expensive queries.
- Sort field whitelist avoids SQL injection and misindexed queries.
- Streaming CSV export sends chunks (~1k rows) to keep memory usage bounded.
- WebSocket hub drops slow consumers and limits client buffers.
- Rate limiter defaults to 10 requests/sec per client.
- Graceful shutdown path closes HTTP server, drains WebSocket hub, and tears down NATS/PostgreSQL/Redis connections.

---

## 9. Working with the Architecture

### Adding a Feature

1. Start in `internal/app/`: update or introduce entities, validation, events, and service methods.
2. Define new port interfaces in `ports.go` if an external dependency is required.
3. Implement adapters under `internal/infrastructure/<category>/` that satisfy the ports.
4. Wire adapters into the container (see `internal/app/app.go` and `internal/server/server.go`).
5. Update Delivery handlers (`internal/delivery/http` or `internal/delivery/ws`) to expose the new functionality.
6. Add unit/integration tests across the layers.
7. Document configuration changes in `env.example`, `docs/configuration.md`, and the README.

### Extending Infrastructure

- Swap implementations by providing different adapters that satisfy the same port.
- Inject the chosen adapter through configuration and container wiring; no changes needed in the Delivery layer.

---

## 10. Summary

- The Delivery layer handles transport concerns only.
- The Application layer owns business logic, domain entities, validation, and orchestrates ports.
- The Infrastructure layer implements ports for persistence, search, cache, streaming, events, and WebSockets.
- Observability currently covers logging and metrics; tracing is planned but not yet active.
- Respect the dependency direction (outer layers depend on inner layers) to keep the codebase testable and maintainable.

Stay within these boundaries when implementing new features to maintain clarity and reliability across the system.