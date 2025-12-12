# Agent Instructions for CAATSM Dashboard

## Essential Commands
Use `make <target>` or `task <target>` (e.g., `make backend-dev`).

### Build/Lint/Test
- **Backend**: `backend-build`, `backend-lint`, `backend-test`, `backend-test-unit`, `backend-test-integration`
- **Single Go test**: `go test -run TestName ./internal/service`
- **Frontend**: `frontend-build`, `frontend-test-unit` (Vitest), `frontend-test` (Playwright E2E)

### Development
- **Backend**: `backend-dev` (hot reload), `backend-dev-run`
- **Frontend**: `frontend-dev`
- **Environment**: `dev-up`, `dev-down`, `dev-logs`

## Code Style Guidelines

### Backend (Go)
- **Formatting**: `gofmt` + `golangci-lint`
- **Imports**: stdlib → third-party → internal (blank lines)
- **Naming**: PascalCase exported, camelCase unexported
- **Error handling**: Early returns, `fmt.Errorf("msg: %w", err)`
- **Testing**: Table-driven, testify/assert
- **Constraints**: 90-day windows, idempotent consumers, bounded queries (≤100)

### Frontend (SvelteKit + TypeScript)
- **Formatting**: Prettier (2 spaces, semicolons, double quotes)
- **TypeScript**: Strict mode, explicit types, no `any`
- **Components**: PascalCase, <200 lines, Svelte 5 runes (`$state`, `$derived`, `$effect`)
- **State**: Svelte stores with `$` prefix
- **Imports**: `$lib/` absolute paths

### General
- Use deterministic tools: `gofmt`, `golangci-lint`, `prettier`, `tsc`
- No comments unless complex business logic
- Follow existing patterns
- Real-time first: Never block WebSocket handlers

## Cursor Rules & Architecture
- **Backend**: `.cursor/rules/backend.mdc` (Go + Echo, Clean Arch, 90-day limits)
- **Frontend**: `frontend/.cursor/rules/frontend.mdc` (SvelteKit + TS, real-time first)
- **General**: `.cursor/rules/general.mdc` (lean real-time dashboard)
- **Architecture**: Clean Architecture: Delivery → Service → Domain ← Infrastructure
- **Stack**: PostgreSQL/Timescale, Meilisearch, Valkey, NATS