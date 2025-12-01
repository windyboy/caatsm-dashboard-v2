# Agent Instructions for CAATSM Dashboard

## Essential Commands
- `make dev` - Start development (backend + frontend)
- `make test` - Run all tests
- `make lint` - Run linters
- `make dev-up` - Start dependencies (Postgres, Meilisearch, Valkey, NATS)
- `make migrate` - Run database migrations

See `docs/DEVELOPMENT.md` for complete command reference and workflows.

## Tooling
Use deterministic tools for code style: `gofmt`, `golangci-lint`, `Prettier`, `tsc`

## References
- Project overview: `.cursor/rules/general.mdc`
- Backend patterns: `.cursor/rules/backend.mdc`
- Frontend patterns: `.cursor/rules/frontend.mdc`
- Architecture: `docs/ARCHITECTURE.md`
- Development guide: `docs/DEVELOPMENT.md`
