# Test Coverage Report

## Current Coverage

### Domain Layer
- **Coverage: 100.0%** ✅ (Target: ≥90%)
- Files:
  - `telegram.go` - Telegram entity and business rules
  - `validator.go` - Validation logic
  - `parser.go` - Parser interface and implementation
  - `converter.go` - Domain/Model conversion
  - `errors.go` - Domain errors

### Application Layer
- **Coverage: 83.1%** ✅ (Target: ≥80%)
- Files:
  - `telegram_service.go` - TelegramService use cases
  - `query_service.go` - QueryService operations
  - `interfaces.go` - Service interfaces

### Worker Layer
- **Coverage: 57.1%** (Target: ≥60%)
- Files:
  - `worker.go` - Worker message handling
  - `errors.go` - Error classification

## Test Files

### Unit Tests
1. `internal/domain/telegram_test.go` - Telegram entity tests
2. `internal/domain/validator_test.go` - Validator tests
3. `internal/domain/parser_test.go` - Parser tests
4. `internal/domain/converter_test.go` - Converter tests
5. `internal/application/telegram_service_test.go` - TelegramService tests
6. `internal/application/query_service_test.go` - QueryService tests
7. `internal/sync/worker_test.go` - Worker tests
8. `internal/infrastructure/event/redis_eventbus_test.go` - EventBus tests

### Integration Tests
1. `internal/application/telegram_service_integration_test.go` - Application integration
2. `internal/sync/worker_integration_test.go` - Worker integration

## Test Utilities

### Mocks
- `internal/testing/mocks/telegram_store_mock.go`
- `internal/testing/mocks/search_index_mock.go`
- `internal/testing/mocks/eventbus_mock.go`
- `internal/testing/mocks/telegram_service_mock.go`

### Test Containers
- `internal/testing/integration/postgres_container.go`
- `internal/testing/integration/redis_container.go`
- `internal/testing/integration/meili_container.go`

### Helpers
- `internal/testing/helpers.go` - Test environment setup
- `internal/testing/fixtures.go` - Test data generators
- `internal/testing/test_suite.go` - Test suite base class

## Running Coverage Reports

```bash
# Domain layer coverage
go test ./internal/domain/... -short -coverprofile=domain_coverage.out
go tool cover -html=domain_coverage.out -o domain_coverage.html

# Application layer coverage
go test ./internal/application/... -short -coverprofile=app_coverage.out
go tool cover -html=app_coverage.out -o app_coverage.html

# Overall coverage
go test ./internal/... -short -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Coverage Goals Status

- ✅ Domain layer: 100.0% (Target: ≥90%)
- ✅ Application layer: 83.1% (Target: ≥80%)
- ⚠️ Worker layer: 57.1% (Target: ≥60%) - Needs improvement
- Overall: TBD (Target: ≥60%)

