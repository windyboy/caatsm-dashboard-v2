# CAATSM Dashboard Testing Guide

This guide explains how to verify the CAATSM Dashboard from end to end using clear, step-by-step language. It covers the backend (Go), the frontend (SvelteKit), real-time features, and smoke checks for critical endpoints.

## 1. Before You Start

1. Install the toolchain:
   - Go 1.25 or newer
   - Deno 2 (preferred) or Node.js 20
   - Docker and Docker Compose (for integration tests)
2. Copy the local environment file: `cp env.local.example .env.local`
3. Bring up local services (PostgreSQL, Meilisearch, Valkey, NATS) before you run tests.

To start the service stack, open a terminal in the project root and run:

```/dev/null/bootstrap.sh#L1-3
make dev-up
task migrate
task dev:config
```

## 2. Quick System Check

Run this smoke script to confirm the essentials are alive:

```/dev/null/smoke.sh#L1-6
curl -f http://localhost:3002/api/health
curl -f http://localhost:7700/health
redis-cli ping
nats server check
psql postgres://caatsm:caatsm@localhost:5432/caatsm -c "SELECT 1"
```

Proceed only after every command finishes without errors.

## 3. Backend Test Matrix

### 3.1 Unit Tests (fast feedback)

Scope: domain logic, services, helper packages.

```/dev/null/backend-unit.sh#L1-2
make test-unit
go test ./internal/domain ./internal/app/services -v
```

Key things to watch:
- Failing tests usually indicate validation rules (90-day window, sort whitelist) or event sequencing issues.
- If a new test relies on time, fix the clock dependency by using the existing time provider utilities in `internal/testing`.

### 3.2 Integration Tests (full stack with containers)

Scope: Postgres, Meilisearch, Valkey, NATS, streaming export.

```/dev/null/backend-integration.sh#L1-4
make dev-up
make test-integration
task test:integration
docker compose -f docker-compose.dev.yml logs --tail=50
```

Focus areas:
- CSV export streams data in chunks; expect long-running tests rather than memory spikes.
- Sync worker scenarios validate that NATS messages persist to Postgres and index into Meilisearch.
- Full data flow integration tests (TestFullDataFlowIntegration) verify complete API-to-database pipelines.
- WebSocket handler tests cover message broadcasting, slow client disconnects, and connection errors.
- Error scenario tests validate input validation, malformed requests, and edge cases.
- Health checks must return `status=degraded` if any dependency is stopped mid-test.

### 3.3 Race Detector (concurrency safety)

Run this weekly or before releases:

```/dev/null/backend-race.sh#L1-1
make test-race
```

Any race warning must be resolved before merging.

## 4. Frontend Test Matrix

### 4.1 Unit Tests with Vitest

Scope: Svelte stores, WebSocket client, helper utilities.

```/dev/null/frontend-unit.sh#L1-3
make frontend-test-unit
deno task test:unit
npm run test:unit
```

Look for:
- Message store keeps only the most recent 50 entries.
- Stats store resets correctly on disconnect.
- WebSocket client backoff behaves as expected.

### 4.2 End-to-End Tests with Playwright

Scope: full UI workflow, real-time dashboard, search flows.

```/dev/null/frontend-e2e.sh#L1-3
make frontend-test
deno task test
npm run test
```

Recommended assertions:
- Live stream displays new telegrams after running `task publish-stream:fast`.
- Search page enforces the 90-day range limit.
- Export button downloads CSV without blocking the UI.
- WebSocket connection handles disconnect and reconnect scenarios gracefully.
- Error states are properly displayed for connection failures.

## 5. Manual Real-Time Verification

1. Start backend (`make dev`) and frontend (`make frontend-dev`).
2. Publish sample messages:

```/dev/null/publish.sh#L1-1
task publish-stream:fast
```

3. Confirm in the browser (http://localhost:5173):
   - Live stream widgets update within a second.
   - Stats cards increment in sync with the incoming messages.
   - WebSocket status indicator remains “Connected”.

If updates stop, check:
- Browser console for WebSocket warnings.
- Backend logs for “slow client” disconnects.
- Metrics endpoint for `websocket_connections`.

## 6. API Spot Checks

Run these curl calls to validate core endpoints:

```/dev/null/api-checks.sh#L1-6
curl -f "http://localhost:3002/api/search?query=test"
curl -f "http://localhost:3002/api/stats/total"
curl -f "http://localhost:3002/api/autocomplete?term=TE&size=5"
curl -f "http://localhost:3002/api/export?format=csv&limit=100" -o /tmp/export.csv
curl -f "http://localhost:3002/api/health"
```

Expected outcomes:
- Search returns JSON, never more than 1000 results per page.
- Export streams CSV without loading everything in memory.
- Health endpoint lists each dependency with `status` and (when available) `latency_ms`.

## 7. Performance Sanity Checks

1. Generate sample data:

```/dev/null/data.sh#L1-1
task generate-test-data
```

2. Export 10k records:

```/dev/null/perf.sh#L1-1
curl "http://localhost:3002/api/export?format=csv&limit=10000" -o /tmp/large_export.csv
```

3. Observe memory usage with your system monitor. The process should stay stable because export is chunked.

## 8. Troubleshooting Checklist

| Symptom | Likely Cause | Resolution |
| --- | --- | --- |
| Search returns empty | Meilisearch not indexed | Re-run `task sync`, check API key |
| WebSocket disconnects | Slow client or rate limit hit | Increase buffer in config or inspect client logic |
| Health check degraded | Dependency down | Restart service via `make dev-up` and rerun tests |
| `config validation failed` | Missing TLS/auth in prod mode | Set required config keys before restart |

Refer to `docs/troubleshooting.md` for deeper playbooks.

## 9. Release Gate

A build is ready for release when:
1. `make test`, `make test-integration`, and `make test-race` are green.
2. `make frontend-test-unit` and `make frontend-test` succeed.
3. Smoke checks and manual real-time verification pass.
4. Export, search, and health API spot checks work in the staging environment.
5. Observability dashboards show metrics for ingestion, search latency, and WebSocket connections.

Document any deviations and capture logs, screenshots, or trace IDs in the release notes.

## 10. Continuous Improvement

- Add new regression tests whenever defects are fixed.
- Keep Playwright scenarios aligned with real user flows.
- Rotate integration tests into CI to catch drift in dependencies.
- Use `make security-scan` before major releases to ensure gosec and Trivy reports are clean.

Following this guide keeps the CAATSM Dashboard reliable, observable, and production-ready.