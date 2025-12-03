# Agent Instructions for CAATSM Dashboard

## Essential Commands
**Format**: Use `make <target>` or `task <target>` (e.g., `make backend-dev`).

### Build/Lint/Test
- **Backend**: `backend-build`, `backend-lint`, `backend-test` (all), `backend-test-unit` (unit only), `backend-test-integration`
- **Single Go test**: `go test -run TestName ./internal/service`
- **Frontend**: `frontend-build`, `frontend-test-unit`, `frontend-test` (E2E)
- **Single frontend test**: `cd frontend && npm run test:unit -- --run tests/unit/Component.test.ts`

### Development
- **Backend dev**: `backend-dev` (hot reload), `backend-dev-run` (direct)
- **Frontend dev**: `frontend-dev`
- **Environment**: `dev-up` (start deps), `dev-down`, `dev-logs`

## Code Style Guidelines

### Backend (Go)
- **Formatting**: `gofmt` (enforced by CI)
- **Imports**: Standard library first, then third-party, then internal (blank line separators)
- **Naming**: PascalCase for exported, camelCase for unexported; descriptive names
- **Error handling**: Return early, wrap with context: `fmt.Errorf("operation failed: %w", err)`
- **Testing**: Table-driven tests, testify/assert for assertions
- **Architecture**: Follow Clean Architecture (delivery → service → domain ← infrastructure)

### Frontend (SvelteKit + TypeScript)
- **Formatting**: Prettier (2 spaces, semicolons, double quotes)
- **TypeScript**: Strict mode; explicit types; no `any`
- **Components**: PascalCase, <200 lines, Svelte 5 runes (`$state`, `$derived`, `$effect`)
- **State**: Svelte stores with `$` prefix for reactivity
- **Imports**: Absolute paths with `$lib/` alias

### General
- Use deterministic tools: `gofmt`, `golangci-lint`, `prettier`, `tsc`
- No comments unless explaining complex business logic
- Follow existing patterns in codebase

## Cursor Rules
- **Backend**: `.cursor/rules/backend.mdc` - Go + Echo, Clean Architecture, 90-day query limits
- **Frontend**: `.cursor/rules/frontend.mdc` - SvelteKit + TypeScript, real-time first, runes
- **General**: `.cursor/rules/general.mdc` - Lean real-time dashboard, small purposeful changes

## Architecture Overview
Clean Architecture: Delivery (handlers) → Service (business logic) → Domain (entities) ← Infrastructure (adapters).
PostgreSQL canonical, Meilisearch for search, Redis for cache/realtime, NATS for ingestion.
