# Domain Rules

## Principles
- Domain is single source of truth
- Lives in `internal/app/`
- Pure Go, deterministic, testable
- No I/O, DB, HTTP, search, cache, network

## Contains
- Message rules
- Filtering & time window semantics
- Aggregation rules
- Domain errors

## File Locations
`telegram.go`, `filters.go`, `query.go`, `events.go`
