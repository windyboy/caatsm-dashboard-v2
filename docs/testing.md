# Testing Guide

## Test Commands

```bash
# All tests
make backend-test

# Unit tests only (fast)
make backend-test-unit

# Integration tests (requires Docker)
make backend-test-integration

# Race detector
make backend-test-race
```

## Test Coverage Matrix

| Layer | Unit Tests | Integration Tests | Current Files | Coverage |
|-------|-----------|-------------------|---------------|----------|
| **Domain** | ✅ Full | N/A | `telegram_test.go`, `filters_test.go`, `validation_test.go` | ~85% |
| **Service** | ✅ Partial | ❌ TODO | `ingestion_test.go`, `indexer_test.go`, `realtime_service_test.go`, `admin_test.go` | ~60% |
| **Repository** | ❌ TODO | ❌ TODO | None | ~0% |
| **Infrastructure** | ✅ Partial | ✅ Partial | `cache/valkey_test.go`, `event/*_test.go`, `streaming/redis_stream_test.go` | ~50% |
| **Delivery** | ✅ Partial | ✅ Basic | `handlers_test.go`, `middleware_test.go`, `http_ws_integration_test.go` | ~65% |

## Test Strategy

### Domain Layer (Pure Unit Tests)

- No external dependencies
- Table-driven tests
- Focus: business logic validation

Example: `internal/domain/telegram_test.go` tests `Validate()`, `Normalize()`, `IsHighPriority()`

**Current coverage**: Strong validation and filter tests

### Service Layer (Unit + Mock Integration)

- Mock ports (Repository, SearchIndex, Cache)
- Test orchestration logic
- Verify error handling and retries

**Current coverage**: Basic service tests exist

**TODO**: Add comprehensive service tests with mocked dependencies

### Repository Layer (Integration Tests)

- Requires real PostgreSQL (Testcontainers)
- Test SQL correctness
- Verify idempotency (`ON CONFLICT` behavior)

**Current coverage**: None

**TODO**: Create `internal/repository/store_test.go` with Testcontainers

### Infrastructure (Integration Tests)

- Test real Redis, Meilisearch adapters
- Verify reconnection logic
- Test error handling

**Current coverage**: Basic Redis Streams and event bus tests exist

### Delivery Layer (HTTP Integration Tests)

- Test full HTTP request/response cycle
- Verify middleware stack (auth, rate limiting, CORS)
- WebSocket connection handling

**Current coverage**: Basic handler tests exist, expand coverage

## Running Specific Tests

```bash
# Single package
go test ./internal/domain -v

# Single test function
go test ./internal/domain -run TestTelegram_Validate -v

# With coverage report
go test ./internal/domain -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## CI Integration

Add to CI pipeline:

```yaml
- name: Run tests with race detector
  run: make backend-test-race

- name: Check coverage
  run: |
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | \
    awk '{if ($1 < 70) exit 1}'  # Fail if coverage < 70%
```

## Test Priorities

### High Priority (Next Sprint)

1. Repository integration tests with real PostgreSQL
2. Service layer tests with mocked ports
3. WebSocket hub concurrency tests

### Medium Priority

1. Meilisearch adapter integration tests
2. NATS consumer error handling tests

### Low Priority

1. Performance benchmarks
2. Chaos engineering tests (network failures)
3. Load tests for WebSocket broadcasting

## Test Data Management

### Test Fixtures

Test data lives in `internal/testing/fixtures/`:
- `telegrams.json` - Sample telegram messages
- `filters.json` - Sample search filters

### Testcontainers Setup

For integration tests requiring real databases:

```go
func setupTestDB(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()
    
    // Start PostgreSQL container
    req := testcontainers.ContainerRequest{
        Image:        "timescale/timescaledb:latest-pg15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready"),
    }
    
    container, err := testcontainers.GenericContainer(ctx, req)
    require.NoError(t, err)
    
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
    
    // Get connection string and connect
    // ... setup code ...
}
```

## Writing Good Tests

### Table-Driven Tests

```go
func TestTelegram_Validate(t *testing.T) {
    tests := []struct {
        name     string
        telegram Telegram
        wantErr  bool
        errField string
    }{
        {
            name: "valid telegram",
            telegram: Telegram{
                MessageID: "TEST-001",
                Type:      "AFTN",
                Time:      time.Now(),
                Priority:  2,
            },
            wantErr: false,
        },
        // ... more cases ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.telegram.Validate()
            if tt.wantErr {
                require.Error(t, err)
                if tt.errField != "" {
                    assert.Contains(t, err.Error(), tt.errField)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Parallel Tests

```go
func TestExpensiveOperation(t *testing.T) {
    tests := []struct {
        name string
        // ... test cases ...
    }{
        // ... cases ...
    }
    
    for _, tt := range tests {
        tt := tt // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run in parallel
            // ... test body ...
        })
    }
}
```

### Mocking Interfaces

```go
type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) Save(ctx context.Context, telegram *domain.Telegram) error {
    args := m.Called(ctx, telegram)
    return args.Error(0)
}

func TestIngestionService_Handle(t *testing.T) {
    repo := new(mockRepository)
    repo.On("Save", mock.Anything, mock.Anything).Return(nil)
    
    svc := NewIngestionService(nil, repo, nil, nil, logger)
    
    err := svc.Handle(context.Background(), &domain.Telegram{...})
    
    assert.NoError(t, err)
    repo.AssertExpectations(t)
}
```

## Test Naming Conventions

- Test functions: `TestFunctionName_Scenario`
- Subtests: Descriptive sentence starting with lowercase
- Files: `*_test.go` alongside source files

Examples:
- `TestTelegram_Validate_EmptyMessageID`
- `TestStore_Save_DuplicateMessage`
- `TestHub_Broadcast_SlowClient`

## Coverage Goals

- **Critical paths**: 90%+ (ingestion, persistence)
- **Business logic**: 80%+ (domain, services)
- **Infrastructure**: 70%+ (adapters, WebSocket hub)
- **Overall target**: 75%+

## Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/repository

# Profile CPU usage
go test -bench=BenchmarkSearch -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Profile memory
go test -bench=BenchmarkSearch -memprofile=mem.prof
go tool pprof mem.prof
```

Example benchmark:

```go
func BenchmarkStore_Search(b *testing.B) {
    store := setupBenchDB(b)
    filter := domain.SearchFilters{
        Pagination: domain.Pagination{Limit: 50},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := store.Search(context.Background(), filter)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Integration Test Tags

Mark integration tests with build tags:

```go
//go:build integration
// +build integration

package repository_test

func TestStore_Integration(t *testing.T) {
    // ... integration test requiring real DB ...
}
```

Run with: `go test -tags=integration ./...`

## Frontend Testing

See `frontend/README.md` for frontend-specific testing:
- Unit tests: Vitest
- Component tests: Svelte Testing Library
- E2E tests: Playwright

```bash
cd frontend
deno task test:unit   # Unit tests
deno task test:e2e    # E2E tests (uses Bun for Playwright runner)

# Or using Bun:
bun run test          # E2E tests (default)
bun run test:unit     # Unit tests

# Or using npm:
npm run test:e2e      # E2E tests
npm run test:unit     # Unit tests
```
