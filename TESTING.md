# Testing Guide

## Overview

CAATSM Dashboard uses a comprehensive testing framework with unit tests, integration tests, and test utilities.

## Test Structure

### Unit Tests
- **Domain Layer**: `internal/domain/*_test.go`
  - Telegram validation
  - Business rules
  - Parser logic
  - Converter functions

- **Application Layer**: `internal/application/*_test.go`
  - TelegramService use cases
  - QueryService operations
  - Error handling

- **Worker**: `internal/sync/worker_test.go`
  - Message handling
  - Error classification
  - Retry logic

### Integration Tests
- **Application Integration**: `internal/application/*_integration_test.go`
- **Worker Integration**: `internal/sync/worker_integration_test.go`

Integration tests use testcontainers and require Docker.

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests Only
```bash
make test-integration
```

### With Race Detector
```bash
make race
```

### Specific Package
```bash
go test ./internal/domain/... -v
go test ./internal/application/... -v
```

## Test Coverage

Generate coverage report:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Coverage Goals:
- Domain layer: ≥ 90%
- Application layer: ≥ 80%
- Overall: ≥ 60%

## Test Utilities

### Mocks
Located in `internal/testing/mocks/`:
- `TelegramStoreMock`
- `SearchIndexMock`
- `EventBusMock`
- `TelegramServiceMock`

### Test Containers
Located in `internal/testing/integration/`:
- `PostgresContainer`
- `RedisContainer`
- `MeiliContainer`

### Fixtures
Located in `internal/testing/fixtures.go`:
- `NewTelegram()`
- `NewTelegramWithType()`
- `NewTelegramWithPriority()`
- `NewSearchFilter()`
- `NewTimeWindow()`

### Test Environment
Use `testhelpers.SetupTestEnv()` to create a complete test environment:
```go
env, err := testhelpers.SetupTestEnv(ctx)
defer env.Cleanup(ctx)
```

## Writing Tests

### Unit Test Example
```go
func TestMyFunction(t *testing.T) {
    storeMock := new(mocks.TelegramStoreMock)
    storeMock.On("Save", mock.Anything, mock.Anything).Return(nil)
    
    service := NewMyService(storeMock)
    err := service.DoSomething()
    
    require.NoError(t, err)
    storeMock.AssertExpectations(t)
}
```

### Integration Test Example
```go
//go:build integration

func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    env, err := testhelpers.SetupTestEnv(ctx)
    require.NoError(t, err)
    defer env.Cleanup(ctx)
    
    // Test with real database
    store := pgstore.New(env.Pool)
    // ...
}
```

## Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Use mocks** for unit tests to avoid external dependencies
3. **Use testcontainers** for integration tests
4. **Clean up** test data between tests
5. **Test error cases** as well as success cases
6. **Use descriptive test names** that explain what is being tested

## CI Integration

Tests run automatically in CI:
- Unit tests run on every push
- Integration tests run optionally (can be skipped)
- Coverage reports are uploaded to codecov

