# CAATSM Dashboard

CAATSM (Civil Aviation Aerogram Traffic Stream Monitor) Dashboard tracks aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC networks. It collects messages in real time, stores them safely, and serves them through a Go API and a SvelteKit frontend.

## Highlights

- Fast Go backend with Echo
- Realtime ingestion through NATS JetStream and background workers
- PostgreSQL + TimescaleDB for analytics
- Meilisearch for full text search and autocomplete
- Valkey/Redis cache for hot data
- WebSocket updates with backpressure control
- Streaming CSV export that handles large result sets
- Prometheus metrics, OpenTelemetry traces, structured logging
- Rate limiting (10 requests per second by default)
- Production config safety checks and 90-day time range guard

## Clean Architecture Layout

The code follows a four-layer Clean Architecture pattern:

1. **Delivery** (`internal/delivery/`): HTTP and WebSocket handlers plus validation.
2. **Application** (`internal/app/`): Services, ports, and the dependency container.
3. **Domain** (`internal/domain/`): Entities, value objects, business rules, and events. No external imports.
4. **Infrastructure** (`internal/infrastructure/`): Concrete adapters for Postgres, Meilisearch, Valkey, NATS, events, and WebSocket hub.

Dependencies always flow inward. Infrastructure implements the interfaces from `internal/app/ports/` directly. No repository wrapper package remains.

Extra support packages:

- `internal/observability/`
- `internal/server/`
- `internal/sync/`
- `internal/testing/`

## Requirements

- Go 1.25+
- Deno 2.0+ (recommended) or Node.js 20+
- Docker + Docker Compose (optional but recommended)
- PostgreSQL 15+
- Meilisearch 1.5+
- Valkey/Redis 7+
- NATS 2.10+

## Quick Start

1. Clone the repo:

   ```
   git clone https://github.com/windy/caatsm-dashboard.git
   cd caatsm-dashboard
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

   This copies `.env.local` from the example file.

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
go test ./internal/domain -v
go test ./internal/app/services -v
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
- `task lint`
- `task migrate`
- `task sync`
- `task docker:up`
- `task docker:down`
- `task generate-test-data`
- `task publish-stream` (slow/fast variants available)

## REST API Outline

- `GET /api/search`
- `POST /api/search`
- `GET /api/autocomplete`
- `GET /api/stats/total`
- `GET /api/stats/priority`
- `GET /api/stats/type`
- `GET /api/export` (CSV streaming)
- `GET /api/health`
- `GET /metrics`

Full schemas live in `api/openapi.yaml`.

## WebSocket Endpoint

- `GET /ws`
- Message types:
  - `message`
  - `stats-total`
  - `stats-priority`
  - `stats-type`
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
  - `http_requests_total`
  - `websocket_connections`

- OpenTelemetry tracing:
  - Configure via `[tracing]` block in config.
  - Export to OTLP endpoint (Jaeger, Zipkin, etc).

- Structured logging uses correlation IDs.

## Security Notes

- Enforce HTTPS in prod (TLS 1.3).
- Do not ship with default API keys or passwords.
- Use environment variables or secret managers for sensitive data.
- Rate limit endpoints (default 10 req/sec; adjustable).
- Validate inputs and restrict sortable fields to whitelisted values.
- Streaming export uses chunking to prevent memory spikes.

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