# CAATSM Dashboard

CAATSM (Civil Aviation Aerogram Traffic Stream Monitor) Dashboard tracks aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC networks. It collects messages in real time, stores them safely, and serves them through a Go API and a SvelteKit frontend.

## Highlights

- Go backend powered by Echo with consistent request validation
- NATS JetStream ingestion processed by the ingestion service (integrated in main server)
- PostgreSQL (TimescaleDB-compatible image) storage via pgx
- Meilisearch full-text search
- Valkey/Redis cache for stats, counters, and realtime fan-out
- WebSocket hub with backpressure protection and 10 req/sec rate limiting
- Real-time message updates via WebSocket
- Prometheus metrics and structured Zap logging
- Config guards enforcing TLS/auth secrets and 90-day time windows

## Clean Architecture Layout

The code follows a pragmatic layered architecture:

1. **Delivery** (`internal/delivery/`): HTTP and WebSocket handlers plus validation.
2. **Domain** (`internal/domain/`): Core domain entities (Telegram, SearchFilters, TimeWindow), business logic, and validation rules.
3. **Application** (`internal/app/`): Ports (interfaces), dependency container, and domain entity re-exports for backward compatibility.
4. **Service** (`internal/service/`): Application services that orchestrate business logic using domain entities.
5. **Infrastructure** (`internal/infrastructure/`): Concrete adapters for Postgres, Meilisearch, Valkey, NATS, events, and the WebSocket hub that satisfy the application ports.
6. **Repository** (`internal/repository/`): PostgreSQL repository implementation that implements the Repository port interface.

Dependencies always flow inward. Infrastructure and repository implementations directly implement the interfaces from `internal/app/ports.go`.

Domain entities and business logic live in `internal/domain/`, application services in `internal/service/`, and infrastructure implementations in `internal/infrastructure/` and `internal/repository/`.

Extra support packages:

- `internal/observability/`
- `internal/server/`
- `internal/service/` (includes ingestion and indexer services)
- `internal/testing/`

## Requirements

- Go 1.25.4+
- Deno 2.0+ (recommended) or Bun (recommended as fallback) or Node.js 18+ (20+ recommended)
- Docker + Docker Compose (optional but recommended)
- PostgreSQL 15+
- Meilisearch 1.5+
- Valkey 9+ (or Redis 7+)
- NATS 2.12.2+

## Quick Start

1. Clone the repo:

   ```
   git clone https://github.com/windyboy/caatsm-dashboard-v2.git
   cd caatsm-dashboard-v2
   ```

2. Install dependencies:

   ```
   make install
   # or
   task install
   ```

   This downloads Go modules, installs Deno when missing, and caches frontend deps.

3. Create local config:

   ```
   task dev:config
   ```



   This copies `env.local.example` to `.env.local` if it doesn't exist and keeps `config/config.local.toml` as the per-machine override file.



4. Start dependencies (Postgres, Meilisearch, Valkey, NATS):

   ```
   make dev-up
   ```

5. Run migrations:

   ```
   make backend-migrate
   # or
   task backend:migrate
   # or
   goose -dir migrations postgres "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable" up
   ```

6. Start backend with hot reload:

   ```
   make backend-dev
   # or
   task backend:dev
   ```

7. Start frontend:

   ```
   make frontend-dev
   # or
   task frontend:dev
   ```

8. Open http://localhost:5173 in your browser.

## Project Overview

```
caatsm-dashboard/
├── cmd/                # Main entry points
├── config/             # Config structs and loader
├── internal/           # Clean Architecture layers and helpers
├── frontend/           # SvelteKit app
├── docs/               # Architecture, config, security, troubleshooting
├── migrations/         # Goose migrations
├── scripts/            # Utility scripts
├── deploy/             # Deployment manifests
└── Makefile / Taskfile # Task runners
```

See `docs/architecture.md` for deep detail.

## Running the Backend

- `make backend-dev` / `task backend:dev` starts the API with Air hot reload.
- `make backend-dev-run` / `task backend:dev:run` runs the server directly (no hot reload).
- Binary output lives in `./bin/`.
- Run without hot reload:

  ```
  ./bin/caatsm -config config/config.local.toml
  ```

## Running the Frontend

Recommended (Deno):

```
task frontend:setup   # one time
task frontend:dev
```

Alternative (Bun):

```
cd frontend
bun install
bun run dev
```

Or with Node.js:

```
cd frontend
npm install
npm run dev
```

The dev server proxies API calls to `http://localhost:3002`.

HMR is disabled when running the dev server with Deno (to avoid Vite 7 WebSocket issues). Use Bun if you need hot reload.

## Configuration

- Main file: `config/config.toml`
- Local override: `config/config.local.toml`
- Environment variables use `CAATSM_` prefix.
- Loader order:
  1. `.env.local`
  2. `.env`
  3. Config file
  4. In-code defaults
- Production validation checks for:
  - TLS enabled
  - Non-default secrets or API keys
  - Auth enabled
  - Proper database DSN

See `docs/configuration.md` for configuration details.

## Data Flow Summary

1. Messages enter NATS JetStream.
2. Ingestion service (in main server) consumes, validates, and stores them in Postgres.
3. Messages are published to Redis Pub/Sub for real-time updates and Redis Streams for indexing.
4. Indexer service (in main server) consumes from Redis Streams and updates Meilisearch.
5. Application services read from Postgres, cache hot data in Valkey, and push stats.
6. Delivery layer exposes REST and WebSocket endpoints.
7. Frontend receives updates in real time via WebSocket.

## Data Guarantees

### Message Identity & Idempotency

- Each telegram has a unique, immutable `message_id` (primary key)
- Duplicate messages are silently ignored via `ON CONFLICT DO NOTHING`
- Safe to replay messages from NATS without duplication

### Delivery Guarantees by Layer

- **NATS → PostgreSQL**: At-least-once delivery, exactly-once storage
- **PostgreSQL → Meilisearch**: At-least-once indexing, eventually consistent (reindexable)
- **Redis Pub/Sub → WebSocket**: Best-effort, no persistence or replay

### Data Retention

- PostgreSQL stores messages for **180 days** (enforced by TimescaleDB retention policy)
- Chunks older than 7 days are compressed automatically
- Meilisearch index can be rebuilt from PostgreSQL at any time
- No long-term cold storage or compliance archiving (out of scope)

### Consistency Model

- PostgreSQL is the **source of truth** (canonical)
- Meilisearch is **eventually consistent** with PostgreSQL
- Real-time WebSocket views are **best-effort**; clients must resync via `/api/search` after reconnect

## Database Notes

- TimescaleDB extensions are optional.
- Migrations use Goose annotations (`-- +goose Up` / `Down`).
- Index recommendations are listed in docs.

## Testing

### Backend

```
make backend-test             # run all tests
make backend-test-unit        # unit tests
make backend-test-integration # integration (requires Docker)
make backend-test-race        # go test with race detector
```

Run specific packages:

```
go test ./internal/app -v
go test ./internal/infrastructure/persistence -v
```

### Frontend

```
make frontend-test-unit  # Vitest
make frontend-test       # Playwright
```

Or direct:

```
deno task test
deno task test:unit
```

More guidance in `testing.md`.

## Useful Commands

Using Makefile:

- `make backend-build` - Build the API server binary
- `make backend-dev` - Start backend with hot reload
- `make backend-dev-run` - Run backend server directly
- `make backend-test` - Run all Go tests
- `make backend-test-unit` - Run unit tests only
- `make backend-test-integration` - Run integration tests
- `make backend-lint` - Run golangci-lint
- `make backend-migrate` - Run database migrations
- `make backend-generate-test-data` - Generate sample telegram data
- `make backend-publish-stream` - Publish live messages to NATS
- `make docker-build` - Build Docker image
- `make docker-up` - Start docker-compose stack
- `make docker-down` - Stop docker-compose stack
- `make clean` - Remove build artifacts
- `make help` - Show all available commands

Using Taskfile:

- `task backend:build` - Build the API server binary
- `task backend:dev` - Start backend with hot reload
- `task backend:dev:run` - Run backend server directly
- `task backend:test` - Run all Go tests
- `task backend:test:unit` - Run unit tests only
- `task backend:test:integration` - Run integration tests
- `task backend:lint` - Run golangci-lint
- `task backend:migrate` - Run database migrations
- `task backend:generate-test-data` - Generate sample telegram data
- `task backend:publish-stream` - Publish live messages to NATS (slow/fast variants available)
- `task dev:up` - Start development dependencies
- `task dev:down` - Stop development dependencies



## REST API Outline

- `GET /api/search` - Search telegrams with filters (query, type, source, destination, priority, time range)
- `GET /api/stats` - Get traffic statistics (total messages, by priority, by type)
- `GET /api/health` - Health check endpoint
- `POST /api/admin/reindex` - Admin endpoint to trigger reindexing
- `GET /metrics` - Prometheus metrics (if enabled)

Full schemas live in `api/openapi.yaml`.

## WebSocket Endpoint

- `GET /ws`
- Message types:
  - `message` - Individual telegram messages
  - `stats` - Unified stats update containing total, byType, activeRoutes, messagesPerSec, and timeWindow
- Ping/pong heartbeat
- Slow clients auto-disconnect when buffers fill

## Deployment

### Docker Compose

```
docker compose up --build
```

Services exposed:
- API: 3002
- Postgres: 5432
- Meilisearch: 7700
- NATS: 4222
- Valkey: 6379
- Prometheus: 9090
- Grafana: 3000

### Docker Image

```
docker build -t caatsm-dashboard:latest .
docker run -d \
  -p 3002:3002 \
  -v $(pwd)/config/config.toml:/app/config/config.toml \
  caatsm-dashboard:latest
```

### Production Checklist

- Supply real secrets through environment variables or a secret manager.
- Enable TLS.
- Run migrations before startup.
- Build frontend (`task frontend:build`) and backend (`make backend-build`).
- Serve static files from `frontend/build`.
- Configure rate limiting and CORS.

Kubernetes manifests are under `deploy/`.

## Monitoring and Observability

- Metrics at `/metrics` (Prometheus scrape target):
  - `telegrams_ingested_total`
  - `search_latency_seconds`
  - Optional WebSocket metrics (`websocket_messages_*`) when hub metrics are enabled.

- Tracing configuration keys exist in `config`. Exporter wiring is still pending, but production validation requires `tracing.enabled=true`; keep it false in other environments until instrumentation ships.

- Structured logging uses correlation IDs via the Echo middleware stack.


## Security Notes



- Enforce HTTPS in production (TLS 1.3) and satisfy the config guard checks for TLS, auth, and non-default secrets.

- Enable Basic Auth (`auth.enable_basic`) or provide a JWT secret before exposing the API.
- Rotate database, Redis, and Meilisearch credentials; never ship default passwords or keys.
- Store secrets in environment variables or a managed secret store instead of the repo.
- Keep request throttling enabled (default 10 req/sec) and monitor before adjusting.
- Respect the 90-day time window guard and validated sort fields to block unbounded queries.
- CSV export streams in chunks to cap memory usage; prefer temporary storage for downloaded files.


For full guidance see `docs/security.md`.

## Troubleshooting

See `docs/troubleshooting.md` for checklists on:

- Database connectivity
- WebSocket issues
- Search index problems
- Config validation failures
- Performance bottlenecks

## Contributing

1. Fork the repo.
2. Create a feature branch.
3. Make changes following Clean Architecture rules.
4. Run `make backend-lint` and `make backend-test`.
5. Open a pull request.

## License

MIT © 2025 Windy
