# Agent Instructions for CAATSM Dashboard

## Architecture

The project follows **Clean Architecture** with four layers (fully optimized, zero wrapper patterns):
- **Delivery Layer** (`internal/delivery/`) - HTTP/WebSocket handlers with streaming export
- **Application Layer** (`internal/app/`) - Services, ports, container pattern
- **Domain Layer** (`internal/domain/`) - Business logic, entities, events, validation (with time range limits)
- **Infrastructure Layer** (`internal/infrastructure/`) - **Direct port implementations** (PostgreSQL, Meilisearch, Valkey, NATS, WebSocket hub)

**Additional Packages**:
- **Observability** (`internal/observability/`) - Cross-cutting concerns: logging, metrics, tracing
- **Server** (`internal/server/`) - HTTP server setup and configuration
- **Sync** (`internal/sync/`) - NATS synchronization worker
- **Testing** (`internal/testing/`) - Shared test utilities and helpers

**Key Architecture Principles**:
- Dependencies flow inward: outer layers depend on inner layers, never the reverse
- **Infrastructure directly implements `app/ports` interfaces** - no wrapper layers
- **No `internal/repository/` package** - removed wrapper pattern (Dec 2025)
- Container only exposes port interfaces, ensuring proper dependency inversion

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
go test ./internal/domain -v -run TestValidateTelegram
go test ./internal/app/services -v -run TestDashboardService
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
- Follow clean architecture: domain → application → infrastructure → delivery
- Domain layer must have no external dependencies (✅ enforced)
- Application layer depends only on domain and port interfaces (✅ enforced)
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
│   ├── app/               # Services, ports, container (application layer)
│   │   ├── services/      # Application services
│   │   └── ports/         # Port interfaces (Repository, Cache, SearchIndex, etc.)
│   ├── domain/            # Business logic, entities, validation
│   ├── infrastructure/    # Direct port implementations
│   │   ├── persistence/   # PostgreSQL (implements ports.Repository)
│   │   ├── search/        # Meilisearch (implements ports.SearchIndex)
│   │   ├── cache/         # Valkey/Redis (implements ports.Cache)
│   │   ├── streaming/     # NATS consumer (implements ports.StreamConsumer)
│   │   ├── event/         # Event bus
│   │   └── ws/            # WebSocket hub
│   ├── observability/     # Logging, metrics, tracing
│   ├── server/            # Server setup and configuration
│   ├── sync/              # Synchronization worker
│   └── testing/           # Test utilities and helpers
├── frontend/              # SvelteKit frontend
│   ├── src/
│   │   ├── routes/       # Pages
│   │   └── lib/          # Components, stores, services
│   └── tests/            # E2E and unit tests
├── migrations/            # Database migrations (goose)
├── config/               # TOML configuration
└── Taskfile.yaml         # Task automation
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
1. Write domain logic in `internal/domain/`
2. Define port interface in `internal/app/ports/` (if needed)
3. Add service in `internal/app/services/`
4. Implement infrastructure in `internal/infrastructure/` (directly implements ports)
5. Add HTTP/WS handlers in `internal/delivery/`
6. Wire dependencies in `internal/app/app.go` Container
7. Write tests at each layer
8. Run `make lint` and `make test`

**Note**: Infrastructure implementations should directly implement port interfaces from `app/ports/`. No wrapper layers or intermediate abstractions.

**Database Changes:**
1. Create migration: `goose -dir migrations create <name> sql`
2. Edit migration file (add Up and Down sections)
3. Run: `make migrate`

**Frontend Changes:**
1. Edit components in `frontend/src/lib/components/`
2. Edit routes in `frontend/src/routes/`
3. Test with: `make frontend-test-unit`
4. E2E test with: `make frontend-test`

## Environment Variables

Key environment variables (set in `.env.local` or config file):
- `CAATSM_DATABASE_DSN` - PostgreSQL connection string
- `CAATSM_CONFIG` - Path to config file (default: `config/config.local.toml`)
- `CLI_ARGS` - Arguments for CLI commands (used with task)

## Dependencies

**Backend:**
- Go 1.25+
- PostgreSQL 14+
- Redis/Valkey 7+
- NATS 2.10+
- Meilisearch 1.5+

**Frontend:**
- Deno 1.40+ (preferred) or Node.js 20+
- SvelteKit 2.x
- UnoCSS for styling

**Tools:**
- `air` - Hot reload for Go
- `goose` - Database migrations
- `golangci-lint` - Go linting
- `task` - Task automation
- `make` - Quick commands
