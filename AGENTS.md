# Agent Instructions for CAATSM Dashboard

This document provides essential information for AI agents working on the CAATSM Dashboard project. For detailed documentation, see the `docs/` directory.

## Essential Commands

> **Format**: Commands are available via both `make` and `task`. Use `make <target>` or `task <target>` (e.g., `make backend-dev` or `task backend:dev`).

### Backend Development
- `backend-dev` / `backend:dev` - Start backend with hot reload (air if available, else go run)
- `backend-dev-run` / `backend:dev:run` - Run backend server directly (no hot reload)
- `backend-build` / `backend:build` - Build the API server binary
- `backend-migrate` / `backend:migrate` - Run database migrations

### Backend Testing
- `backend-test` / `backend:test` - Run all Go tests with coverage
- `backend-test-unit` / `backend:test:unit` - Run unit tests only (short)
- `backend-test-integration` / `backend:test:integration` - Run integration tests only
- `backend-test-race` / `backend:test:race` - Run tests with race detector

### Backend Code Quality
- `backend-lint` / `backend:lint` - Run golangci-lint

### Backend Utilities
- `backend-generate-test-data` / `backend:generate-test-data` - Generate sample telegram data (50 records)
- `backend-publish-stream` / `backend:publish-stream` - Publish live messages to NATS
- `backend-publish-stream-fast` / `backend:publish-stream:fast` - Publish messages quickly (1/sec, 50 total)
- `backend-publish-stream-slow` / `backend:publish-stream:slow` - Publish messages slowly (5s interval)

### Frontend Development
- `frontend-dev` / `frontend:dev` - Run frontend dev server
- `frontend-build` / `frontend:build` - Build frontend for production
- `frontend-test` / `frontend:test` - Run frontend E2E tests
- `frontend-test-unit` / `frontend:test:unit` - Run frontend unit tests

### Environment
- `dev-up` / `dev:up` - Start development dependencies (Postgres, Meilisearch, Valkey, NATS)
- `dev-down` / `dev:down` - Stop development dependencies
- `dev-logs` / `dev:logs` - Tail dependency logs
- `dev-config` / `dev:config` - Create .env.local from example if missing

See `docs/development.md` for complete command reference and workflows.

## System Overview

CAATSM Dashboard is a real-time telegram monitoring and querying system with:
- **Real-time Monitoring**: Receive and display parsed telegram messages in real-time
- **Query & Search**: Full-text search, filtering, and time-range queries
- **Analytics**: Aggregated metrics and dashboard data

### Key Principles
- **PostgreSQL is canonical**: All query correctness validated against PostgreSQL
- **Eventually consistent**: Search index and realtime views are eventually consistent
- **Failure isolation**: Search/realtime failures don't block ingestion
- **Idempotency**: Duplicate messages handled via `ON CONFLICT DO NOTHING`

## Architecture

### Backend Architecture (Clean Architecture)

The backend follows Clean Architecture with four layers:

1. **Delivery** (`internal/delivery/`): HTTP/WebSocket handlers, request parsing, response formatting
2. **Domain** (`internal/domain/`): Core business entities, validation, business rules
3. **Application/Service** (`internal/app/`, `internal/service/`): Orchestrates business logic, defines ports
4. **Infrastructure** (`internal/infrastructure/`): Concrete implementations of ports (DB, search, cache, streaming)

**Dependency Flow**: Delivery → Application → Domain ← Infrastructure

### Application Services

1. **Ingestion Service**: Consumes from NATS JetStream, stores to PostgreSQL, publishes to Redis Pub/Sub and Streams
2. **Query Service**: Executes search queries, coordinates between PostgreSQL and Meilisearch, manages caching
3. **Statistics Service**: Calculates aggregated metrics, manages statistics caching
4. **Real-time Service**: Broadcasts messages to WebSocket clients (best-effort delivery)
5. **Indexer Service**: Consumes from Redis Streams, updates Meilisearch index
6. **Export Service**: Streams large datasets for export

### Frontend Architecture (SvelteKit)

- **Framework**: SvelteKit with TypeScript
- **Styling**: UnoCSS + DaisyUI
- **Structure**:
  - `src/lib/components/` - Reusable components
  - `src/lib/services/` - API clients
  - `src/lib/stores/` - State management
  - `src/routes/` - Page routes

### Data Flow

**Message Ingestion**:
```
NATS JetStream → Ingestion Service → PostgreSQL
                                    ├─→ Redis Pub/Sub (realtime)
                                    └─→ Redis Streams → Indexer → Meilisearch
```

**Query Flow**:
```
Client Request → Query Service → Check Redis cache
                                 └─→ If miss: Meilisearch + PostgreSQL → Merge & cache
```

**Real-time Updates**:
```
New Message → Redis Pub/Sub → Realtime Service → WebSocket Hub → Clients
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.25.4+ | Server implementation |
| Web Framework | Echo | HTTP server and routing |
| Message Broker | NATS JetStream 2.12.2 | Message ingestion |
| Database | PostgreSQL 15+ / TimescaleDB | Primary storage (180-day retention) |
| Search | Meilisearch 1.5+ | Full-text search |
| Cache/Broker | Valkey/Redis 7+ | Pub/Sub, Streams, Cache |
| Frontend | SvelteKit + TypeScript | Web dashboard |
| Styling | UnoCSS + DaisyUI | CSS framework |
| Observability | Prometheus + Zap | Metrics and logging |

## Data Storage

- **PostgreSQL/TimescaleDB**: Primary storage (source of truth), time-series optimization with hypertables, 180-day retention
- **Meilisearch**: Full-text search and autocomplete, eventually consistent with PostgreSQL
- **Redis**: Pub/Sub (realtime), Streams (reliable job queue), Cache (query results)

## Configuration

Configuration is loaded from (in order of precedence):
1. In-code defaults
2. TOML files in `config/`
3. Environment variables with `CAATSM_` prefix (highest precedence)

**Key Environment Variables**:
- `CAATSM_DATABASE_DSN` - PostgreSQL connection string
- `CAATSM_MEILISEARCH_HOST` / `API_KEY` - Meilisearch endpoint
- `CAATSM_REDIS_ADDR` / `PASSWORD` - Redis connection
- `CAATSM_NATS_URL` - NATS JetStream endpoint
- `CAATSM_ENVIRONMENT` - `development`, `staging`, or `production`
- `CAATSM_SERVER_TLS_ENABLED` - Enable HTTPS (required in production)
- `CAATSM_AUTH_ENABLE_BASIC` / `AUTH_JWT_SECRET` - Authentication

See `docs/configuration.md` for configuration details.

## Security Considerations

- **TLS**: Required in production for all connections
- **Authentication**: Basic auth or JWT required in production
- **Input Validation**: Domain layer validates all inputs, whitelist-based SQL
- **Rate Limiting**: Default 10 req/sec, configurable
- **Secrets**: Never commit secrets; use environment variables or secret stores
- **WebSocket**: Origin restrictions, per-IP and global connection limits

See `docs/security.md` for complete security guidelines.

## Observability

**Key Metrics**:
- Ingestion: `ingest_total`, `ingest_errors`, `ingest_latency_p95`
- Real-time: `ws_active_connections`, `ws_messages_broadcast_total`
- Indexing: `indexer_queue_depth`, `indexer_processed_total`
- Storage: `db_hypertable_size`, `db_compressed_chunks_count`

**Logging**: Structured JSON logs with request ID propagation
**Metrics**: Prometheus format at `/metrics` endpoint

## Development Guidelines

### Backend (Go)
- **Formatting**: `gofmt` (enforced by CI)
- **Linting**: `golangci-lint`
- **Error Handling**: Return early, wrap with context using `fmt.Errorf("context: %w", err)`
- **Testing**: Table-driven tests, separate unit/integration tests
- **Architecture**: Follow Clean Architecture boundaries strictly

### Frontend (SvelteKit)
- **Formatting**: Prettier (2 spaces, semicolons, double quotes)
- **TypeScript**: Strict mode enabled
- **Components**: PascalCase naming, keep under 200 lines
- **State**: Use Svelte stores with `$` prefix for reactivity

### Code Style
Use deterministic tools: `gofmt`, `golangci-lint`, `Prettier`, `tsc`

## Package Structure

```
cmd/
├── server/          # HTTP Server entry point
└── generate-test-data/

internal/
├── app/            # Port interfaces and dependency container
├── domain/         # Business entities and validation
├── service/        # Application services
├── delivery/       # HTTP/WebSocket handlers
├── repository/     # PostgreSQL repository
└── infrastructure/ # Adapters (NATS, Redis, Meilisearch, WebSocket)

frontend/
├── src/
│   ├── lib/        # Components, services, stores, types, utils
│   └── routes/     # Page routes
└── tests/          # Test files
```

## Limitations

- **No guaranteed realtime delivery**: WebSocket is best-effort
- **No long-term storage**: Data beyond 180 days is dropped
- **No ordering guarantees**: Accepts eventual consistency
- **No strict latency guarantees**: Optimized for read-heavy workload

## References

- **Architecture**: `docs/architecture.md` - Detailed architecture design
- **Development**: `docs/development.md` - Complete development guide
- **Configuration**: `docs/configuration.md` - Configuration reference
- **Security**: `docs/security.md` - Security guidelines
- **Troubleshooting**: `docs/troubleshooting.md` - Common issues and solutions
