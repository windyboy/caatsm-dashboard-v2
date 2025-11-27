# Testing Guide

This document describes how to test the CAATSM Dashboard functionality, including the Deno + Svelte frontend and Go backend.

## Prerequisites

1. Ensure all dependency services are running:
   ```bash
   make dev-up
   # or
   task dev:up
   # or
   docker compose -f docker-compose.dev.yml up -d
   ```

2. Start the Go backend:
   ```bash
   make dev
   # or
   task dev
   # or
   ./bin/caatsm -config config/config.local.toml
   ```

3. Start the frontend development server:
   ```bash
   # Using Deno (recommended)
   make frontend-dev
   # or
   task frontend:dev
   
   # Using Node.js
   cd frontend
   npm install
   npm run dev
   ```

## Backend Testing

### Unit Tests

Run unit tests with coverage:

```bash
make test
# or
task test
```

Run specific package tests:

```bash
go test ./internal/domain -v
go test ./internal/app/services -v
go test ./internal/observability -v
```

Run tests with race detection:

```bash
make test-race
# or
task test:race
```

### Integration Tests

Integration tests use Testcontainers for real external dependencies:

```bash
# Run all integration tests
task test:integration

# Run specific integration tests
go test -tags=integration ./internal/testing -v
```

**Test Coverage**:
- ✅ PostgreSQL container setup and queries
- ✅ Time range validation (90-day limit)
- ✅ Large dataset handling (100+ records)
- ✅ Pagination and sorting

### Key Test Files

**Domain Layer**:
- `internal/domain/validator_test.go` - Domain validation, time range limits
- `internal/domain/filters_test.go` - Search filter validation
- `internal/domain/telegram_test.go` - Telegram entity validation

**Application Layer** (Migrated to `internal/app/` Nov 2025):
- `internal/app/services/*` - Application services (dashboard, search, stats, export)
- `internal/sync/worker_test.go` - Sync worker unit tests (✅ migrated)
- `internal/sync/worker_integration_test.go` - Sync worker integration tests (✅ migrated)

**Infrastructure Layer**:
- `internal/infrastructure/event/*_test.go` - Event bus tests
- `internal/infrastructure/ws/*_test.go` - WebSocket tests

**Configuration**:
- `config/config_test.go` - Production config validation

**Observability**:
- `internal/observability/redaction_test.go` - PII redaction (email, IP, phone)

### Test Statistics (Nov 2025)

```bash
Total Packages: 42
Packages with Tests: 7
Test Pass Rate: 100%
Key Test Areas:
  ✅ Domain validation
  ✅ Sync worker (newly migrated)
  ✅ Event broadcasting
  ✅ WebSocket handling
  ✅ Observability/redaction
```

**Integration**:
- `internal/testing/e2e_simple_test.go` - End-to-end pipeline tests

### Production Feature Tests

**Time Range Validation**:
```bash
# Test that > 90 day ranges are rejected
go test ./internal/domain -run TestSearchFilters_Validate_TimeRangeTooLarge -v
```

**Streaming Export**:
```bash
# Test CSV export with large datasets
go test ./internal/app/services -run TestExportService_Export -v
```

**Production Config Guards**:
```bash
# Test production validation
go test ./config -run TestAppConfig_Validate_ProductionDefaults -v
```

**PII Redaction**:
```bash
# Test email/IP/phone redaction
go test ./internal/observability -run TestRedact -v
```

## Frontend Testing

### Unit Tests

Run Vitest unit tests:

```bash
make frontend-test-unit
# or
task frontend:test:unit
```

### E2E Tests

Run Playwright E2E tests:

```bash
make frontend-test
# or
task frontend:test
```

## WebSocket Testing

### Manual Testing Steps

1. **Open Dashboard Page**
   - Visit `http://localhost:5173`
   - You should see "Live Stream" component and statistics cards

2. **Check WebSocket Connection**
   - Open browser developer tools (F12)
   - Switch to Network tab
   - Filter for "WS" (WebSocket)
   - You should see a connection to `ws://localhost:3002/ws`

3. **Verify Real-time Message Reception**
   - If messages are flowing through NATS
   - You should see new messages appear in Live Stream
   - Statistics numbers should update in real-time

4. **Test Connection Reconnection**
   - In developer console, you should see "WebSocket connected" logs
   - If connection drops, you should see reconnection attempt logs

### WebSocket Load Testing

Test WebSocket with multiple connections:

```bash
# Open multiple browser tabs (test client limits)
# Default limits: 5 per IP, 10,000 global
```

**Backpressure Testing**:
- Slow clients (not consuming messages fast enough) are auto-disconnected
- Check WebSocket metrics for dropped connections

## REST API Testing

### Manual Testing Steps

1. **Test Search Functionality**
   - Visit `http://localhost:5173/search`
   - Enter search keywords
   - Click "Search" button
   - You should see search results
   - **Test time range**: Try searching with date range > 90 days (should be rejected)

2. **Test Autocomplete**
   - Type at least 2 characters in the search box
   - You should see autocomplete suggestion dropdown

3. **Test Statistics Endpoints**
   - Visit Dashboard page
   - Statistics cards should display data
   - Check API requests in browser developer tools Network tab

4. **Test Export Functionality**
   - Execute a search on the search page
   - Use export button to download CSV
   - **Test streaming**: Export large result sets (should stream without OOM)

### API Endpoint Testing

Use curl or Postman to test APIs:

```bash
# Test search
curl "http://localhost:3002/api/search?query=test"

# Test search with time range (should reject > 90 days)
curl "http://localhost:3002/api/search?start_time=2024-01-01T00:00:00Z&end_time=2024-11-01T00:00:00Z"

# Test total statistics
curl "http://localhost:3002/api/stats/total"

# Test priority statistics
curl "http://localhost:3002/api/stats/priority"

# Test type statistics
curl "http://localhost:3002/api/stats/type"

# Test autocomplete
curl "http://localhost:3002/api/autocomplete?term=test&size=5"

# Test health endpoint
curl "http://localhost:3002/api/health"

# Test export (streaming CSV)
curl "http://localhost:3002/api/export?format=csv&query=test" -o telegrams.csv
```

All APIs should return JSON responses (except export which returns CSV).

### Rate Limiting Testing

Test rate limiting (default: 10 req/sec):

```bash
# Bash loop to test rate limiting
for i in {1..20}; do
  curl -w "%{http_code}\n" "http://localhost:3002/api/health" &
done
wait

# Should see some 429 (Too Many Requests) responses
```

## Health Check Testing

### Test Health Endpoint

```bash
# Full health check (all dependencies)
curl http://localhost:3002/api/health | jq

# Expected response:
# {
#   "status": "ok",
#   "timestamp": "2024-11-26T10:00:00Z",
#   "postgresql": { "status": "ok", "latency_ms": 5 },
#   "meilisearch": { "status": "ok", "latency_ms": 10 },
#   "redis": { "status": "ok", "latency_ms": 2 },
#   "nats": { "status": "ok" },
#   "websocket": { "status": "ok", "message": "active clients: 3" }
# }
```

### Test Degraded Mode

Stop a dependency and check degraded status:

```bash
# Stop Redis
docker compose -f docker-compose.dev.yml stop redis

# Check health (should return 503 with "degraded" status)
curl -w "%{http_code}" http://localhost:3002/api/health
```

## Integration Testing

### Complete Flow Testing

1. **Start All Services**
   ```bash
   # Terminal 1: Start dependencies
   make dev-up
   
   # Terminal 2: Start Go backend
   make dev
   
   # Terminal 3: Start frontend
   make frontend-dev
   ```

2. **Test Real-time Data Flow**
   - Publish test messages to NATS:
     ```bash
     task publish-stream:fast
     ```
   - Observe if Dashboard updates in real-time
   - Check if statistics numbers update correctly

3. **Test Search and Real-time Updates Together**
   - Execute a search on the search page
   - Simultaneously observe Dashboard real-time updates
   - Both should work independently

### Full Pipeline Test

Test the complete message pipeline:

```bash
# 1. Generate test data
task generate-test-data

# 2. Publish to NATS stream
task publish-stream:fast

# 3. Verify in UI
# - Open http://localhost:5173
# - Should see messages in live stream
# - Statistics should update

# 4. Test search
# - Search for generated messages
# - Export results to CSV
```

## Performance Testing

### Export Performance

Test large export:

```bash
# Generate large dataset
task generate-test-data  # Creates 50 records

# Export with streaming (should not cause OOM)
curl "http://localhost:3002/api/export?format=csv&limit=10000" -o large_export.csv

# Monitor memory usage during export
```

### WebSocket Performance

Test multiple concurrent connections:

```bash
# Open multiple browser tabs
# Monitor WebSocket hub metrics
# Check for slow client disconnections
```

### Search Performance

Test search with large result sets:

```bash
# Search without limits (should use default pagination)
curl "http://localhost:3002/api/search?query=*"

# Search with max limit (1000)
curl "http://localhost:3002/api/search?query=*&limit=1000"
```

## Error Handling Testing

### Network Disconnection

1. Disconnect network connection
2. WebSocket should attempt to reconnect
3. After network restoration, should automatically reconnect

### Backend Service Stop

1. Stop Go backend service
2. WebSocket should detect disconnection
3. Should show reconnection attempts
4. After restarting backend, should automatically reconnect

### Invalid Input Testing

```bash
# Test invalid time range (> 90 days)
curl "http://localhost:3002/api/search?start_time=2024-01-01T00:00:00Z&end_time=2024-12-01T00:00:00Z"
# Expected: 400 Bad Request with "range cannot exceed 90 days" error

# Test invalid sort field (SQL injection prevention)
curl "http://localhost:3002/api/search?sort_by=invalid; DROP TABLE telegrams;"
# Expected: 400 Bad Request with validation error

# Test invalid priority
curl "http://localhost:3002/api/search?priority=-1"
# Expected: 400 Bad Request with validation error
```

## OpenTelemetry Tracing Testing

### Enable Tracing

Update `config/config.local.toml`:

```toml
[tracing]
enabled = true
service_name = "caatsm-dashboard"
otlp_endpoint = "localhost:4318"
sampling_ratio = 1.0
```

### Start Jaeger (for viewing traces)

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### View Traces

1. Make some API requests
2. Open Jaeger UI: http://localhost:16686
3. Select service "caatsm-dashboard"
4. View distributed traces

## PII Redaction Testing

### Test Redaction in Logs

Enable PII redaction in config:

```toml
[logger]
redact_pii = true
```

Test that sensitive data is redacted:

```bash
# Send request with sensitive data
curl "http://localhost:3002/api/search?query=test@example.com+192.168.1.1+555-1234"

# Check logs - should show:
# - [EMAIL_REDACTED]
# - [IP_REDACTED]
# - [PHONE_REDACTED]
```

## Browser Compatibility

Test on the following browsers:
- Chrome/Edge (Chromium)
- Firefox
- Safari

## Known Issues

- WebSocket reconnection may take a few seconds
- Large exports (10k+ records) may take time but won't cause OOM due to streaming

## Troubleshooting

### WebSocket Connection Failure

1. Check if Go backend is running on `localhost:3002`
2. Check firewall settings
3. Check browser console error messages
4. Verify WebSocket hub is initialized (check health endpoint)

### API Request Failure

1. Check Vite proxy configuration
2. Check CORS settings (should be configured)
3. Check backend logs
4. Verify rate limiting isn't blocking requests

### Frontend Build Failure

1. Run `npm install` to reinstall dependencies (or use Deno)
2. Run `npm run check` to check type errors (or `deno task check`)
3. Clear `.svelte-kit` directory and retry

### Test Failures

1. Ensure all Docker containers are running:
   ```bash
   docker compose -f docker-compose.dev.yml ps
   ```

2. Check Docker container logs:
   ```bash
   docker compose -f docker-compose.dev.yml logs
   ```

3. Reset test environment:
   ```bash
   make dev-down
   make dev-up
   make migrate
   ```

## Continuous Integration

### GitHub Actions / CI Pipeline

The project includes comprehensive tests suitable for CI:

```bash
# Run all tests in CI mode
make test
make frontend-test-unit

# Integration tests (requires Docker)
task test:integration

# Build verification
make build
make frontend-build
```

## OpenAPI Specification Testing

View and test API with OpenAPI specification:

```bash
# View spec file
cat api/openapi.yaml

# Serve with Swagger UI (using Docker)
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/api/openapi.yaml \
  -v $(pwd)/api:/api \
  swaggerapi/swagger-ui

# Open http://localhost:8080
```

## Test Coverage

Current test coverage includes:

- ✅ Domain validation (time ranges, input sanitization)
- ✅ Streaming export (large datasets)
- ✅ Production config validation
- ✅ PII redaction (email, IP, phone)
- ✅ Health checks (all dependencies)
- ✅ WebSocket backpressure
- ✅ Rate limiting
- ✅ Integration tests with Testcontainers
- ✅ E2E pipeline tests

**Target Coverage**: 80%+ for critical paths
