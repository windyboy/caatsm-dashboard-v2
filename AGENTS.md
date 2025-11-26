# Agent Instructions for CAATSM Dashboard

## Architecture

The project follows **Clean Architecture** with four layers:
- **Transport Layer** (`internal/transport/`) - HTTP/WebSocket handlers
- **Application Layer** (`internal/application/`) - Use cases, event handlers, ports
- **Domain Layer** (`internal/domain/`) - Business logic, entities, events, validation
- **Infrastructure Layer** (`internal/infrastructure/`) - PostgreSQL, Meilisearch, Valkey, WebSocket hub

Dependencies flow inward: outer layers depend on inner layers, never the reverse.

## Build/Lint/Test Commands

**Go Backend:**
- Build: `make build` or `task build`
- Test all: `make test` or `task test`
- Test unit only: `make test-unit`
- Test integration: `make test-integration`
- Test single: `go test ./internal/domain -v -run TestValidateTelegram`
- Lint: `make lint` or `task lint`
- Dev server: `make dev` or `task dev`
- Migrate: `task migrate` (uses goose or psql, extracts DSN from config)

**Frontend (Deno/Svelte):**
- Setup: `task frontend:setup` or `make frontend-setup` (installs Deno, caches dependencies)
- Install Deno: `task install:deno` or `make install-deno`
- Dev: `task frontend:dev` or `make frontend-dev` (prefers Deno, falls back to npm)
- Test E2E: `task frontend:test` or `make frontend-test` (Playwright)
- Test unit: `task frontend:test:unit` or `make frontend-test-unit` (Vitest)
- Build: `task frontend:build` or `make frontend-build`
- Install deps: `task frontend:install` or `make frontend-install` (pre-cache with Deno)

**Development:**
- Start dependencies: `task dev:up` (Docker/Podman compatible)
- Stop dependencies: `task dev:down`
- View logs: `task dev:logs`
- Setup config: `task dev:config` (creates .env.local from example)

## Code Style Guidelines

**Go:**
- Use `gofmt` for formatting
- Exported types/functions: PascalCase (e.g., `Telegram`, `Validate()`)
- Unexported: camelCase (e.g., `messageID`)
- Error handling: Return errors, use custom error types like `ErrInvalidTelegram`
- Imports: stdlib → third-party → internal
- Structs: Clear field names, use time.Time for timestamps
- Functions: Descriptive names, early returns for errors
- Security: Always validate user input, use whitelists for SQL column names
- Architecture: Keep layers separate, dependencies point inward

**TypeScript/Svelte:**
- Use Prettier for formatting (2 spaces, semicolons, double quotes)
- TypeScript strict mode
- Component naming: PascalCase (e.g., `LiveStream.svelte`)
- Store naming: camelCase with $ prefix for reactive
- Error handling: Try/catch blocks, proper typing

**General:**
- No comments unless explaining complex business logic
- Use dependency injection pattern
- Follow clean architecture: domain → application → infrastructure → transport
- Domain layer must have no external dependencies
- Application layer depends only on domain and port interfaces
- Infrastructure implements application ports