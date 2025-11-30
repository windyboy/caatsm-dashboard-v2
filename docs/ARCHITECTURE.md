# CAATSM Dashboard Architecture

This document explains how the CAATSM Dashboard backend is structured and how the major pieces work together. The intent is to keep things simple: understand the layers, know how data moves, and see where to add new behavior without breaking the design.

---

## 1. Big Picture

- **Goal:** Collect aviation telegrams, store them safely, index them for search, and stream them to the dashboard UI in real time.
- **Stack:** Go (backend), NATS (stream), PostgreSQL + TimescaleDB (storage), Meilisearch (search), Valkey/Redis (cache), SvelteKit (frontend).
- **Style:** Clean Architecture with four layers. Each outer layer depends only on the next inner layer.

---

## 2. Layer Map

```
Delivery ──► Application ──► Domain ──► Infrastructure
```

| Layer | Directory | What it does | Depends on |
| --- | --- | --- | --- |
| Delivery | `internal/delivery/` | HTTP + WebSocket handlers, request parsing, response formatting | Application services |
| Application | `internal/app/` | Business workflows, services, ports, dependency container | Domain + port interfaces |
| Domain | `internal/domain/` | Entities, validation, business rules, domain events | Only Go stdlib |
| Infrastructure | `internal/infrastructure/` | Concrete adapters (Postgres, Meilisearch, Valkey, NATS, WS hub) | Application ports, Domain types |

**Rules:**

- Delivery never calls infrastructure directly.
- Infrastructure implements interfaces defined in `internal/app/ports`.
- Domain stays pure Go with no third-party or infrastructure imports.

---

## 3. Responsibilities by Layer

### Delivery (`internal/delivery/`)
- Sets up HTTP routes and WebSocket endpoints with Echo.
- Validates input quickly (time range ≤ 90 days, safe sort fields).
- Applies rate limiting (10 requests/sec default) and CORS.
- Translates errors into consistent JSON responses.

### Application (`internal/app/`)
- Houses services such as `DashboardService`, `SearchService`, `StatsService`, `ExportService`, `RealtimeService`.
- Exposes a `Container` that wires config, logging, ports, and services.
- Orchestrates workflows (example: search → validate → query repository → enrich → cache).
- Publishes domain events to infrastructure listeners.

### Domain (`internal/domain/`)
- Defines core objects: `Telegram`, `SearchFilters`, events like `TelegramProcessed`.
- Enforces constraints: timestamp ordering, priority enums, 90-day window, safe filter combinations.
- Provides reusable validators and error types (e.g., `ErrInvalidTimeRange`).
- Emits domain events so that infrastructure can react without tight coupling.

### Infrastructure (`internal/infrastructure/`)
- **Persistence:** Direct PostgreSQL implementation of `ports.Repository` using pgx/Timescale.
- **Search:** Meilisearch adapter implementing `ports.SearchIndex`.
- **Cache:** Valkey/Redis adapter implementing `ports.Cache`.
- **Streaming:** NATS consumer producing domain events.
- **Events:** Pub/Sub bridge for cross-component notifications.
- **WebSocket hub:** Manages clients, backpressure, disconnects slow consumers.
- **Resilience utilities:** Circuit breakers, retry helpers, connection health checks.

---

## 4. Data Flow

### Message Ingestion
1. NATS JetStream receives raw telegram.
2. Sync worker (under `internal/sync/`) pulls the message.
3. Worker uses application services to validate the telegram (domain layer).
4. Repository stores the telegram in PostgreSQL.
5. Search indexer pushes the telegram into Meilisearch.
6. Event bus broadcasts `TelegramProcessed`.
7. Realtime service pushes updates to WebSocket hub.

### Query / Dashboard Request
1. Client calls REST API or opens a WebSocket.
2. Delivery layer validates parameters and rate limits.
3. Application service loads data:
   - Reads cache when possible.
   - Hits repository for fresh aggregates.
   - Fan-outs to search service when needed.
4. Domain layer ensures the request follows business rules.
5. Infrastructure returns data from Postgres/Meilisearch/Redis.
6. Delivery serializes response or streams via WebSocket.
7. Metrics and traces capture latency, errors, and call span.

---

## 5. Core Services and Ports

| Service | Key Functions | Ports used |
| --- | --- | --- |
| DashboardService | Load dashboard metrics in parallel | Repository, Cache |
| SearchService | Build queries, call search index, persist history | Repository, SearchIndex |
| StatsService | Aggregate totals, priority/type breakdowns | Repository, Cache |
| ExportService | Stream CSV in chunks (1000 rows) to avoid OOM | Repository |
| RealtimeService | Fan-out domain events to WebSocket hub | EventPublisher, WebSocketHubPort |
| HealthService | Probe dependencies and report degraded state | Repository, Cache, SearchIndex, StreamConsumer |

All ports live in `internal/app/ports/`. If a new external dependency is needed, define a new port interface there.

---

## 6. Runtime Components

- **HTTP server:** Configured in `internal/server/`, starts Echo, middleware, and routes.
- **WebSocket hub:** Maintains client registry, enforces per-IP and global connection caps.
- **Sync worker:** Lives under `cmd/sync`; consumes NATS messages and calls app services.
- **Observability:** `internal/observability/` injects logger (Zap), metrics (Prometheus), tracing (OpenTelemetry).
- **Config loader:** `config/` package parses TOML + env, validates production guardrails (TLS, non-default secrets, auth rules).

---

## 7. Configuration Highlights

- **Defaults:** Safe development values baked in.
- **Overrides:** `config/config.local.toml` + `.env.local`.
- **Production checks:** On startup we fail fast if TLS disabled, API keys are defaults, or auth is off.
- **Secrets:** Pull from env or secret managers (AWS Secrets Manager, Vault). Infrastructure layer provides helper adapters.

---

## 8. Resilience and Safety

- **Circuit breakers:** Wrap external calls; open after repeated failures.
- **Graceful shutdown:** Stop HTTP listener, drain WebSocket hub, close NATS, Postgres, Redis, tracing exporter.
- **Rate limiting:** Token bucket at delivery layer.
- **Backpressure:** Hub disconnects clients that stop reading.
- **Streaming export:** Sends CSV in chunks so memory stays flat.
- **Time window guard:** Refuses searches over 90 days to keep queries bounded.

---

## 9. Observability

- **Metrics:** `/metrics` exposes ingestion counts, search latency histograms, HTTP request stats, active WebSocket connections.
- **Tracing:** Optional OTLP exporter; spans wrap HTTP handlers, DB calls, search calls, NATS processing.
- **Logging:** Structured JSON with correlation IDs propagated via headers.

---

## 10. Development Checklist

When adding a feature:
1. Start in the **Domain** layer. Add/extend entities, validation, or events.
2. Define or update **ports** if a new capability is required.
3. Implement or adjust **Application** services to coordinate use cases.
4. Create/extend **Infrastructure** adapters that satisfy the ports.
5. Update **Delivery** handlers to expose the new behavior.
6. Wire dependencies in `internal/app/container.go`.
7. Add tests at the right level (domain → service → handler).
8. Run `make test`, `make test-integration`, and frontend tests.
9. Update docs if behavior or endpoints change.

---

## 11. Common Extension Points

- **New data source:** Create a port interface, implement it under `internal/infrastructure/<category>/`.
- **New API endpoint:** Add handler in `internal/delivery/http`, hook into existing services.
- **Additional metrics:** Register in `internal/observability/metrics` and export via `/metrics`.
- **Alternate cache or search backend:** Add a new adapter implementing the same port; wire it through config.

---

## 12. Summary

The CAATSM Dashboard backend stays maintainable by:
- Keeping business logic in the Domain layer.
- Letting Application services coordinate work through port interfaces.
- Plugging external systems directly in Infrastructure without extra wrappers.
- Making Delivery responsible only for transport concerns.
- Enforcing guardrails that protect performance, security, and reliability.

Build new features by respecting these boundaries and the system will remain predictable, testable, and easy to scale.