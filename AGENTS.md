# Agent Instructions for CAATSM Dashboard

## Build/Lint/Test Commands

**Go Backend:**
- Build: `make build` or `task build`
- Test all: `make test` or `task test`
- Test unit only: `make test-unit`
- Test integration: `make test-integration`
- Test single: `go test ./internal/domain -v -run TestValidateTelegram`
- Lint: `make lint` or `task lint`
- Dev server: `make dev` or `task dev`

**Frontend (Deno/Svelte):**
- Setup: `task frontend:setup` or `make frontend-setup` (installs Deno, caches dependencies)
- Install Deno: `task install:deno` or `make install-deno`
- Dev: `task frontend:dev` or `make frontend-dev` (prefers Deno, falls back to npm)
- Test: `task frontend:test` or `make frontend-test`
- Build: `task frontend:build` or `make frontend-build`
- Install deps: `task frontend:install` or `make frontend-install` (pre-cache with Deno)

## Code Style Guidelines

**Go:**
- Use `gofmt` for formatting
- Exported types/functions: PascalCase (e.g., `Telegram`, `Validate()`)
- Unexported: camelCase (e.g., `messageID`)
- Error handling: Return errors, use custom error types like `ErrInvalidTelegram`
- Imports: stdlib → third-party → internal
- Structs: Clear field names, use time.Time for timestamps
- Functions: Descriptive names, early returns for errors

**TypeScript/Svelte:**
- Use Prettier for formatting (2 spaces, semicolons, double quotes)
- TypeScript strict mode
- Component naming: PascalCase (e.g., `LiveStream.svelte`)
- Store naming: camelCase with $ prefix for reactive
- Error handling: Try/catch blocks, proper typing

**General:**
- No comments unless explaining complex business logic
- Use dependency injection pattern
- Follow clean architecture: domain → application → infrastructure