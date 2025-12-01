# Agent Instructions for CAATSM Dashboard

## Build/Lint/Test Commands

**Core Commands:**
- `make build` - Build application
- `make test` - Run all tests with coverage
- `make lint` - Run golangci-lint
- `make dev` - Run backend with hot reload
- `make sync` or `task sync` - Run sync worker to process NATS messages
- `make docker-up` - Start full production stack (includes sync worker)

**Running Single Tests:**
```bash
go test ./internal/domain -v -run TestValidateTelegram
go test ./internal/service -v -run TestSearchService
```

**Frontend:**
- `make frontend-dev` - Run dev server
- `make frontend-test` - Run E2E tests

## Code Style Guidelines

**Go:**
- Formatting: `gofmt`
- Naming: PascalCase for exported, camelCase for unexported
- Errors: Return errors early, use custom types like `ErrInvalidTelegram`
- Imports: stdlib → third-party → internal
- Functions: <50 lines, descriptive names, early error returns
- Security: Validate inputs, use whitelists for SQL

**TypeScript/Svelte:**
- Formatting: Prettier (2 spaces, semicolons, double quotes)
- Strict mode enabled
- Components: PascalCase (e.g., `LiveStream.svelte`)
- Stores: camelCase with `$` prefix for reactive
- Error handling: Try/catch with proper typing

**General:**
- No comments unless complex business logic
- Clean Architecture: delivery → service → application → infrastructure
- Domain entities in `internal/domain/`
- Infrastructure directly implements ports

## Cursor Rules

**Backend (Go):**
- Senior Go engineer persona; small, production-safe changes
- Stack: Go + Echo-style HTTP, Postgres/Timescale, NATS, Meilisearch, Valkey
- Architecture: Delivery → application → infrastructure via ports
- Code: Clear functions, explicit errors, contexts with timeouts
- Organization: Functions <50 lines, meaningful names, single responsibility
- Errors: Return early, wrap with context, custom domain types
- Data: Bounded queries, idempotent consumers, proper observability

**Frontend (SvelteKit):**
- Senior SvelteKit engineer; simple, predictable UI
- Stack: SvelteKit + TypeScript, Deno/Node, UnoCSS
- Components: <200 lines, descriptive names, composition over inheritance
- State: Simple stores, reactive statements, TypeScript interfaces
- Data: Mirror backend constraints, handle all states explicitly
- Testing: Fast unit tests, focused E2E coverage

**General:**
- Configuration: Explicit, documented, fail fast on missing
- Testing: Fast/deterministic unit tests, integration only at boundaries
- Git: Descriptive commits, feature branches, PR merges
- Security: Input validation, no secrets logging, regular updates</content>
<parameter name="filePath">AGENTS.md