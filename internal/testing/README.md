# Testing Framework

This directory contains the testing infrastructure for CAATSM Dashboard.

## Structure

```
internal/testing/
├── mocks/              # Mock implementations for testing
│   ├── telegram_store_mock.go
│   ├── search_index_mock.go
│   ├── eventbus_mock.go
│   └── telegram_service_mock.go
├── integration/        # Test container helpers
│   ├── postgres_container.go
│   ├── redis_container.go
│   └── meili_container.go
├── helpers.go          # Test environment setup
└── fixtures.go         # Test data generators
```

## Usage

### Unit Tests

Unit tests use mocks and don't require external services:

```go
import (
    "github.com/windy/caatsm-dashboard/internal/testing/mocks"
)

func TestMyService(t *testing.T) {
    storeMock := new(mocks.TelegramStoreMock)
    // Setup expectations
    storeMock.On("Save", mock.Anything, mock.Anything).Return(nil)
    // Test...
}
```

### Integration Tests

Integration tests use testcontainers and require Docker:

```go
//go:build integration

import (
    testhelpers "github.com/windy/caatsm-dashboard/internal/testing"
)

func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    env, err := testhelpers.SetupTestEnv(ctx)
    require.NoError(t, err)
    defer env.Cleanup(ctx)
    
    // Use env.Pool, env.Redis, env.Meili
}
```

### Test Fixtures

Use fixtures to generate test data:

```go
import "github.com/windy/caatsm-dashboard/internal/testing"

telegram := testing.NewTelegram("TEST-001")
telegramWithType := testing.NewTelegramWithType("TEST-002", "aftn")
filter := testing.NewSearchFilter()
```

## Running Tests

```bash
# Run all unit tests
make test-unit

# Run all integration tests (requires Docker)
make test-integration

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Coverage Goals

- Domain layer: ≥ 90%
- Application layer: ≥ 80%
- Overall: ≥ 60%

