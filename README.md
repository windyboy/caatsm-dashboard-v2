# CAATSM Dashboard

CAATSM (Civil Aviation Aerogram Traffic Stream Monitor) Dashboard tracks aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC networks. It collects messages in real time, stores them safely, and serves them through a Go API and a SvelteKit frontend.

## Highlights

- Go backend powered by Echo with consistent request validation
- NATS JetStream ingestion processed by the sync worker
- PostgreSQL (TimescaleDB-compatible image) storage via pgx
- Meilisearch full-text search with typed autocomplete suggestions
- Valkey/Redis cache for stats, counters, and realtime fan-out
- WebSocket hub with backpressure protection and 10 req/sec rate limiting
- Streaming CSV export backed by repository-level streaming
- Prometheus metrics and structured Zap logging
- Config guards enforcing TLS/auth secrets and 90-day time windows

## Clean Architecture Layout

The code follows a pragmatic layered architecture:

1. **Delivery** (`internal/delivery/`): HTTP and WebSocket handlers plus validation.
2. **Application** (`internal/app/`): Services, domain entities (Telegram, SearchFilters, TimeWindow), ports, and the dependency container. This layer owns business rules such as validation, pagination guards, and stats aggregation. Unlike traditional Clean Architecture, domain entities are integrated directly into the application layer rather than a separate `internal/domain/` package.
3. **Infrastructure** (`internal/infrastructure/`): Concrete adapters for Postgres, Meilisearch, Valkey, NATS, events, and the WebSocket hub that satisfy the application ports.

Dependencies always flow inward. Infrastructure implements the interfaces from `internal/app/ports.go` directly, and there are no repository wrapper packages.

Domain-specific validation and DTOs live alongside services in `internal/app/`, keeping business logic close to the ports that expose it.

Extra support packages:

- `internal/observability/`
- `internal/server/`
- `internal/sync/`
- `internal/testing/`

## Requirements

- Go 1.25.4+
- Deno 2.0+ (recommended) or Node.js 20+
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
   task migrate
   # or
   goose -dir migrations postgres "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable" up
   ```

6. Start backend with hot reload:

   ```
   make dev
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

See `docs/ARCHITECTURE.md` for deep detail.

## Running the Backend

- `make dev` starts the API with Air hot reload.
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

Alternative (Node.js):

```
cd frontend
npm install
npm run dev
```

The dev server proxies API calls to `http://localhost:3002`.

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

See `docs/configuration.md` for templates and secret manager examples.

## Data Flow Summary

1. Messages enter NATS JetStream.
2. Sync worker consumes, validates, and stores them in Postgres and Meilisearch.
3. Application services read from Postgres, cache hot data in Valkey, and push stats.
4. Delivery layer exposes REST and WebSocket endpoints.
5. Frontend receives updates in real time.

## Database Notes

- TimescaleDB extensions are optional.
- Migrations use Goose annotations (`-- +goose Up` / `Down`).
- Index recommendations are listed in docs.

## Testing

### Backend

```
make test             # run all tests
make test-unit        # unit tests
make test-integration # integration (requires Docker)
make test-race        # go test with race detector
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

More guidance in `TESTING.md`.

## Useful Commands

Using Makefile:

- `make build`
- `make lint`
- `make docker-build`
- `make docker-up`
- `make docker-down`
- `make clean`
- `make help`


Using Taskfile:



- `task build`

- `task dev`

- `task dev:run`
- `task lint`

- `task migrate`

- `task dev:up`
- `task dev:down`
- `task generate-test-data`

- `task publish-stream` (slow/fast variants available)



## REST API Outline



- `GET /api/dashboard`
- `GET /api/search`

- `POST /api/search`

- `GET /api/stats`
- `GET /api/export` (CSV streaming)
- `GET /api/autocomplete`

- `GET /api/health`
- `GET /metrics`


Full schemas live in `api/openapi.yaml`.

## WebSocket Endpoint

- `GET /ws`
- Message types:
  - `message` - Individual telegram messages
  - `stats` - Unified stats update containing total, byPriority, and byType
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
- Build frontend (`task frontend:build`) and backend (`make build`).
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
4. Run `make lint` and `make test`.
5. Open a pull request.

## License

MIT © 2025 Windy