# Agent Instructions for CAATSM Dashboard

## Architecture

The project follows **Clean Architecture** with four main layers:
- **Domain Layer** (`internal/domain/`) - Domain entities, business logic, and validation (Telegram, SearchFilters, TimeWindow, etc.)
- **Application Layer** (`internal/app/`) - Ports, container pattern, and service interfaces
- **Service Layer** (`internal/service/`) - Application services and business logic orchestration
- **Infrastructure Layer** (`internal/infrastructure/`) - Direct port implementations (PostgreSQL, Meilisearch, Valkey, NATS, WebSocket hub)
- **Delivery Layer** (`internal/delivery/`) - HTTP/WebSocket handlers with streaming export

**Note**: This project maintains a clear separation between domain entities (in `internal/domain/`), application services (in `internal/service/`), and infrastructure implementations. Domain entities are re-exported through `internal/app/ports.go` for backward compatibility.

**Additional Packages**:
- **Observability** (`internal/observability/`) - Cross-cutting concerns: logging, metrics, tracing
- **Server** (`internal/server/`) - HTTP server setup and configuration
- **Sync** (`internal/sync/`) - NATS synchronization worker
- **Testing** (`internal/testing/`) - Shared test utilities and helpers
- **Errors** (`pkg/errors/`) - Shared error handling utilities

**Key Architecture Principles**:
- Dependencies flow inward: outer layers depend on inner layers, never the reverse
- **Infrastructure directly implements `app/ports` interfaces** - no wrapper layers
- Container only exposes port interfaces, ensuring proper dependency inversion
- Domain logic (validation, business rules) is co-located with entities in `internal/domain/`
- Application services in `internal/service/` orchestrate business logic using domain entities
- Infrastructure implementations are wired in the server layer to avoid circular imports

**Production Features**:
- Time range validation: Max 90 days to prevent unbounded queries
- Streaming CSV export: Handles 50k+ records without OOM via chunking
- Production config guards: Prevents insecure defaults in production environment
- Rate limiting: 10 req/sec (configurable)
- WebSocket backpressure: Auto-disconnects slow clients

## Build/Lint/Test Commands

### Quick Reference (Makefile)

For everyday development, use simple `make` commands:

**Core Development:**
- `make build` - Build the application
- `make test` - Run all tests with coverage
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests only
- `make test-race` - Run tests with race detector
- `make lint` - Run golangci-lint
- `make dev` - Run backend with hot reload (air)
- `make clean` - Clean build artifacts

**Frontend:**
- `make frontend-dev` - Run frontend dev server (Deno/npm)
- `make frontend-build` - Build frontend for production
- `make frontend-test` - Run E2E tests (Playwright)
- `make frontend-test-unit` - Run unit tests (Vitest)

**Development Environment:**
- `make dev-up` - Start development dependencies (Docker)
- `make dev-down` - Stop development dependencies
- `make migrate` - Run database migrations

**Setup:**
- `make install` - Install all dependencies (one-time setup)
- `make help` - Show all available commands

### Complete Reference (Taskfile)

For advanced features and full control, use `task` commands:

**Go Backend:**
```bash
task build                # Build server binary
task test                 # Run all tests with race detection and coverage
task test:unit            # Run unit tests only (short mode)
task test:integration     # Run integration tests only
task test:race            # Run tests with race detector
task test:coverage        # Run tests with coverage report
task lint                 # Run golangci-lint
task dev                  # Run backend with hot reload
task dev:run              # Run with local config file
task tidy                 # Run go mod tidy
```

**Installation:**
```bash
task install              # Install all dependencies (Go + frontend)
task install:deno         # Install Deno if not available
```

**Frontend (Deno/Svelte):**
```bash
task frontend:setup       # Full setup: install Deno + cache dependencies
task install:deno         # Install Deno if not available
task frontend:install     # Pre-cache dependencies with Deno (or npm fallback)
task frontend:dev         # Run dev server (Deno, fallback to npm)
task frontend:build       # Build for production (Deno, fallback to npm)
task frontend:test        # Run E2E tests (Playwright via Deno, fallback to npm)
task frontend:test:unit   # Run unit tests (Vitest via Deno, fallback to npm)
```

**Development Environment:**
```bash
task dev:up               # Start dependencies (Postgres, Redis, NATS, Meilisearch)
task dev:down             # Stop dependencies
task dev:logs             # Show logs from dependencies
task dev:config           # Create .env.local from example
task dev:meili-key        # Extract Meilisearch master key
```

**Database & Testing:**
```bash
task migrate              # Run database migrations (goose/psql)
task generate-test-data   # Generate test telegram data (50 records)
task publish-stream       # Publish live messages to NATS for testing
task publish-stream:fast  # Publish quickly (1/sec, 50 total)
task publish-stream:slow  # Publish slowly (5 sec interval)
```

**Docker:**
```bash
task docker:build         # Build Docker image
task docker:up            # Start docker-compose stack
task docker:down          # Stop docker-compose stack
```

**Cleanup:**
```bash
task clean                # Clean build artifacts
task clean:all            # Clean all including node_modules
```

### Running Single Tests

To run a specific test:
```bash
go test ./internal/app -v -run TestValidateTelegram
go test ./internal/app -v -run TestDashboardService
```

## Code Style Guidelines

**Go:**
- Use `gofmt` for formatting
- Exported types/functions: PascalCase (e.g., `Telegram`, `Validate()`)
- Unexported: camelCase (e.g., `messageID`)
- Error handling: Return errors, use custom error types like `ErrInvalidTelegram`
- Imports: stdlib → third-party → internal
- Structs: Clear field names, use `time.Time` for timestamps
- Functions: Descriptive names, early returns for errors
- Security: Always validate user input, use whitelists for SQL column names
- Architecture: Keep layers separate, dependencies point inward

**TypeScript/Svelte:**
- Use Prettier for formatting (2 spaces, semicolons, double quotes)
- TypeScript strict mode
- Component naming: PascalCase (e.g., `LiveStream.svelte`)
- Store naming: camelCase with `$` prefix for reactive
- Error handling: Try/catch blocks, proper typing

**General:**
- No comments unless explaining complex business logic
- Use dependency injection pattern
- Follow clean architecture: delivery → service → application → infrastructure
- Domain entities and business logic live in `internal/domain/` (not a separate domain package)
- Application layer depends only on port interfaces (✅ enforced)
- Infrastructure **directly implements** application ports (✅ enforced, no wrapper layers)
- All legacy handlers/services/models/repository wrappers have been removed (✅ complete)

## Project Structure

```
caatsm-dashboard/
├── cmd/                    # Entry points
│   ├── server/            # Main API server
│   ├── sync/              # NATS sync worker
│   ├── generate-test-data/
│   ├── publish-stream/
│   └── extract-dsn/
├── internal/
│   ├── delivery/          # HTTP/WebSocket handlers (delivery layer)
│   │   ├── http/          # HTTP handlers, routes, middleware
│   │   └── ws/            # WebSocket handlers and broadcasters
│   ├── app/               # Application layer (ports, container, domain re-exports)
│   │   ├── app.go         # Container and dependency wiring
│   │   ├── ports.go       # Port interfaces (Repository, Cache, SearchIndex, etc.)
│   │   ├── cache.go       # Cache-related types
│   │   ├── event.go       # Event-related types
│   │   └── ports_backup/  # Legacy backup (not used)
│   ├── domain/            # Domain layer (entities, business logic, validation)
│   │   ├── telegram.go    # Telegram entity and validation
│   │   ├── query.go       # Query types and filters
│   │   ├── filters.go     # Search filter logic
│   │   ├── events.go      # Domain events
│   │   ├── errors.go      # Error types
│   │   └── validation_test.go
│   ├── service/           # Service layer (application services)
│   │   ├── dashboard.go   # DashboardService
│   │   ├── search.go      # SearchService
│   │   ├── stats.go       # StatsService
│   │   ├── export.go      # ExportService
│   │   ├── realtime.go    # RealtimeService
│   │   └── types.go       # Service types
│   ├── infrastructure/    # Direct port implementations
│   │   ├── cache/         # Valkey/Redis (implements ports.Cache)
│   │   ├── search/        # Meilisearch (implements ports.SearchIndex)
│   │   ├── streaming/     # NATS consumer (implements ports.StreamConsumer)
│   │   ├── event/         # Event bus
│   │   └── ws/            # WebSocket hub
│   ├── repository/        # PostgreSQL repository (implements ports.Repository)
│   │   ├── models.go      # Database models
│   │   └── store.go       # Repository implementation
│   ├── observability/     # Logging, metrics, tracing
│   ├── server/            # Server setup and configuration
│   ├── sync/              # Synchronization worker
│   └── testing/           # Test utilities and helpers
├── pkg/                   # Shared packages
│   └── errors/           # Error handling utilities
├── frontend/              # SvelteKit frontend
│   ├── src/
│   │   ├── routes/       # Pages and routes
│   │   └── lib/          # Shared code
│   │       ├── components/  # Svelte components
│   │       ├── services/   # Frontend services (API clients, WebSocket)
│   │       ├── stores/     # Svelte stores (state management)
│   │       ├── types/      # TypeScript type definitions
│   │       └── utils/      # Utility functions
│   └── tests/            # E2E and unit tests
├── migrations/            # Database migrations (goose)
├── config/               # TOML configuration files
├── api/                  # API specifications
│   └── openapi.yaml     # OpenAPI 3.0 specification
├── docs/                 # Additional documentation
│   ├── ARCHITECTURE.md
│   ├── configuration.md
│   ├── security.md
│   └── troubleshooting.md
├── scripts/              # Utility scripts
├── deploy/               # Deployment configurations
├── Taskfile.yaml         # Task automation
└── Makefile              # Quick command shortcuts
```

## Quick Start Guide

1. **Initial Setup:**
   ```bash
   make install          # Install all dependencies
   task dev:config       # Create .env.local
   ```

2. **Start Development Environment:**
   ```bash
   make dev-up          # Start Postgres, Redis, NATS, Meilisearch
   make migrate         # Run database migrations
   ```

3. **Start Development Servers:**
   ```bash
   # Terminal 1 - Backend
   make dev
   
   # Terminal 2 - Frontend
   make frontend-dev
   ```

4. **Generate Test Data (Optional):**
   ```bash
   task generate-test-data
   task publish-stream:fast
   ```

5. **Run Tests:**
   ```bash
   make test            # All tests
   make frontend-test   # E2E tests
   ```

## Common Workflows

**Adding a New Feature:**
1. Define domain entities/types in `internal/domain/` (e.g., `telegram.go`, `query.go`)
2. Define port interface in `internal/app/ports.go` (if needed)
3. Add service methods in `internal/service/` (e.g., `dashboard.go`, `search.go`)
4. Implement infrastructure in `internal/infrastructure/` (directly implements ports)
5. Add HTTP/WS handlers in `internal/delivery/`
6. Wire dependencies in `internal/server/` (server layer handles wiring to avoid circular imports)
7. Write tests at each layer
8. Run `make lint` and `make test`

**Note**:
- Domain entities (Telegram, SearchFilters, TimeWindow) and business logic live in `internal/domain/`, not a separate domain package
- Infrastructure implementations should directly implement port interfaces from `app/ports.go`. No wrapper layers or intermediate abstractions.

**Database Changes:**
1. Create migration: `goose -dir migrations create <name> sql`
2. Edit migration file (add Up and Down sections)
3. Run: `make migrate`

**Frontend Changes:**
1. Edit components in `frontend/src/lib/components/`
2. Edit routes in `frontend/src/routes/`
3. Edit services in `frontend/src/lib/services/`
4. Edit stores in `frontend/src/lib/stores/`
5. Test with: `make frontend-test-unit`
6. E2E test with: `make frontend-test`

## Environment Variables

Key environment variables (set in `.env.local` or config file):
- `CAATSM_DATABASE_DSN` - PostgreSQL connection string
- `CAATSM_CONFIG` - Path to config file (default: `config/config.local.toml`)
- `CLI_ARGS` - Arguments for CLI commands (used with task)

## Dependencies

**Backend:**
- Go 1.25.4+
- PostgreSQL 14+
- Valkey 9+ (or Redis 7+)
- NATS 2.12.2+
- Meilisearch 1.5+

**Frontend:**
- Deno 2.0+ (preferred) or Node.js 20+
- SvelteKit 2.x with Svelte 5
- UnoCSS for styling

**Tools:**
- `air` - Hot reload for Go
- `goose` - Database migrations
- `golangci-lint` - Go linting
- `task` - Task automation
- `make` - Quick commands
