# Domain Rules

## Principles
- Domain is single source of truth
- Lives in `internal/domain/`
- Pure Go, deterministic, testable
- No I/O, DB, HTTP, search, cache, network

## Contains
- Message rules, filtering & time window semantics
- Aggregation rules, domain errors

## Files
`telegram.go`, `filters.go`, `query.go`, `events.go`
