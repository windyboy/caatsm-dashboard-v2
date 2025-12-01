# CAATSM Dashboard - Development Guide

This guide covers development setup, testing, debugging, and deployment testing for the CAATSM Dashboard.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Development Workflow](#development-workflow)
- [Testing Strategy](#testing-strategy)
- [Code Quality](#code-quality)
- [Debugging](#debugging)
- [Performance Testing](#performance-testing)
- [Deployment Testing](#deployment-testing)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Go**: 1.25.4+
- **Deno**: 2.0+ (recommended) or Node.js 20+
- **Docker**: 20.10+ with Docker Compose
- **Git**: 2.30+

### External Dependencies

- **PostgreSQL**: 15+ (TimescaleDB recommended)
- **Meilisearch**: 1.5+
- **NATS**: 2.12.2+
- **Valkey/Redis**: 7+

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/windy/caatsm-dashboard-v2.git
cd caatsm-dashboard-v2

# Install all dependencies
make install
# or
task install
```

### 2. Start Development Environment

```bash
# Start all dependencies (Postgres, Meilisearch, NATS, Valkey)
make dev-up
# or
task dev:up

# Run database migrations
task migrate
```

### 3. Start Development Servers

```bash
# Terminal 1: Backend API (with hot reload)
make dev

# Terminal 2: Frontend (with hot reload)
make frontend-dev
```

### 4. Access Application

- **Frontend**: http://localhost:5173
- **API**: http://localhost:3002
- **API Docs**: http://localhost:3002/api/health

### 5. Generate Test Data (Optional)

```bash
# Generate sample telegrams
task generate-test-data

# Publish test messages to NATS
task publish-stream:fast  # 50 messages quickly
task publish-stream:slow  # Messages every 5 seconds
```

## Development Workflow

### Backend Development (Go)

#### Project Structure

```
cmd/
├── server/          # Main API server
├── sync/           # Message ingestion worker
├── generate-test-data/
└── publish-stream/

internal/
├── app/            # Application layer (ports, container)
├── domain/         # Business logic & entities
├── service/        # Application services
├── infrastructure/ # External service adapters
├── delivery/       # HTTP/WebSocket handlers
├── repository/     # Database layer
├── server/         # Server setup
└── testing/        # Test utilities

pkg/
└── errors/         # Shared error types
```

#### Adding New Features

1. **Define Domain Entities** (`internal/domain/`)
   ```go
   type NewEntity struct {
       ID   string `json:"id"`
       Name string `json:"name"`
   }
   ```

2. **Add Port Interface** (`internal/app/ports.go`)
   ```go
   type NewService interface {
       Create(ctx context.Context, entity *NewEntity) error
   }
   ```

3. **Implement Service** (`internal/service/`)
   ```go
   type NewService struct {
       repo Repository
   }

   func (s *NewService) Create(ctx context.Context, entity *NewEntity) error {
       return s.repo.Save(ctx, entity)
   }
   ```

4. **Add HTTP Handler** (`internal/delivery/http/`)
   ```go
   func (h *Handler) CreateNewEntity(c echo.Context) error {
       // Parse request, call service, return response
   }
   ```

5. **Wire Dependencies** (`internal/server/server.go`)
   ```go
   newSvc := service.NewNewService(store)
   container.NewService = newSvc
   ```

#### Code Conventions

- **Formatting**: `gofmt` (enforced by CI)
- **Imports**: stdlib → third-party → internal
- **Naming**: PascalCase for exported, camelCase for unexported
- **Errors**: Return early, wrap with context
- **Functions**: <50 lines, single responsibility
- **Security**: Validate inputs, use whitelists for SQL

### Frontend Development (SvelteKit)

#### Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/    # Reusable components
│   │   ├── services/      # API clients
│   │   ├── stores/        # State management
│   │   ├── types/         # TypeScript definitions
│   │   └── utils/         # Utility functions
│   ├── routes/            # Page routes
│   └── app.html           # HTML template
├── tests/                 # Test files
├── package.json
├── svelte.config.js
├── tsconfig.json
└── vite.config.ts
```

#### Component Development

```svelte
<!-- lib/components/NewComponent.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let title: string;
  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('custom', { data: 'value' });
  }
</script>

<h1>{title}</h1>
<button on:click={handleClick}>Click me</button>

<style>
  h1 { color: blue; }
</style>
```

#### State Management

```ts
// lib/stores/dashboard.ts
import { writable } from 'svelte/store';

export const dashboardData = writable(null);
export const loading = writable(false);
export const error = writable(null);
```

#### API Integration

```ts
// lib/services/api.ts
export class ApiService {
  async getDashboard() {
    const response = await fetch('/api/dashboard');
    return response.json();
  }
}
```

#### Code Conventions

- **Formatting**: Prettier (2 spaces, semicolons, double quotes)
- **TypeScript**: Strict mode enabled
- **Components**: PascalCase naming, <200 lines
- **Stores**: camelCase with `$` prefix for reactive
- **Error Handling**: Try/catch with proper typing

## Testing Strategy

### Backend Testing

#### Unit Tests

```bash
# Run all unit tests
make test-unit
# or
go test -short ./...

# Run specific package tests
go test ./internal/domain -v
go test ./internal/service -v

# Run with coverage
go test -cover ./internal/domain
```

#### Integration Tests

```bash
# Run integration tests (requires Docker)
make test-integration
# or
go test -tags=integration ./...

# Start test dependencies
task dev:up
```

#### Test Structure

```go
// internal/domain/telegram_test.go
func TestTelegram_Validate(t *testing.T) {
    tests := []struct {
        name     string
        telegram Telegram
        wantErr  bool
    }{
        {
            name: "valid telegram",
            telegram: Telegram{
                MessageID: "ABC123",
                Type:      "AFTN",
                Time:      time.Now(),
                Priority:  1,
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.telegram.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Frontend Testing

#### Unit Tests (Vitest)

```bash
# Run unit tests
make frontend-test-unit
# or
cd frontend && deno task test:unit
# or
npm run test:unit
```

#### E2E Tests (Playwright)

```bash
# Run E2E tests
make frontend-test
# or
cd frontend && deno task test
# or
npm test
```

#### Test Structure

```ts
// tests/unit/components/Dashboard.test.ts
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import Dashboard from '../../../src/lib/components/Dashboard.svelte';

describe('Dashboard', () => {
  it('renders dashboard title', () => {
    const { getByText } = render(Dashboard, {
      props: { title: 'Test Dashboard' }
    });

    expect(getByText('Test Dashboard')).toBeInTheDocument();
  });
});
```

### Test Data Management

#### Generate Test Data

```bash
# Generate sample telegrams in database
task generate-test-data

# Publish messages to NATS for testing
task publish-stream:fast  # 50 messages at 1/sec
task publish-stream:slow  # Messages every 5 seconds
```

#### Test Fixtures

```go
// internal/testing/fixtures.go
func CreateTestTelegram() *domain.Telegram {
    return &domain.Telegram{
        MessageID:    "TEST123",
        Type:        "AFTN",
        Time:        time.Now(),
        FlightNumber: "AA101",
        Source:      "KJFK",
        Destination: "KLAX",
        Priority:    1,
        Content:     "Test message content",
    }
}
```

## Code Quality

### Linting

```bash
# Run Go linter
make lint
# or
golangci-lint run ./...

# Run frontend linter
cd frontend && deno lint
# or
cd frontend && npm run lint
```

### Code Coverage

```bash
# Backend coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend coverage
cd frontend && deno task test:coverage
```

### Pre-commit Hooks

Consider setting up pre-commit hooks:

```bash
# Install pre-commit (if using)
pre-commit install

# Or create .git/hooks/pre-commit
#!/bin/sh
make lint
make test-unit
```

## Debugging

### Backend Debugging

#### Logging

The application uses structured logging with Zap:

```go
logger.Info("processing telegram",
    zap.String("message_id", telegram.MessageID),
    zap.String("type", telegram.Type),
    zap.Time("time", telegram.Time),
)
```

#### Debug Mode

```bash
# Set log level to debug
export CAATSM_LOG_LEVEL=debug

# Start with debug logging
make dev
```

#### Database Debugging

```bash
# Connect to database
psql "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable"

# Check recent messages
SELECT message_id, type, time, flight_number, source, destination
FROM telegrams
ORDER BY time DESC
LIMIT 10;

# Check Meilisearch index
curl "http://localhost:7700/indexes/telegrams/search?q=*"
```

### Frontend Debugging

#### Browser DevTools

- **Network tab**: Check API calls
- **Console**: View client-side errors
- **Application tab**: Inspect local storage, session storage

#### WebSocket Debugging

```javascript
// In browser console
const ws = new WebSocket('ws://localhost:3002/ws');
ws.onmessage = (event) => console.log('Received:', event.data);
ws.onopen = () => console.log('Connected');
```

#### Svelte DevTools

Install Svelte DevTools browser extension for component inspection.

### Common Debug Commands

```bash
# Check service health
curl http://localhost:3002/api/health

# View application logs
docker compose logs app -f

# Check NATS streams
docker compose exec nats nats stream ls

# Monitor Redis keys
docker compose exec redis redis-cli keys "*"

# Check Meilisearch health
curl http://localhost:7700/health
```

## Performance Testing

### Load Testing

#### API Load Testing

```bash
# Using hey (HTTP load testing)
hey -n 1000 -c 10 http://localhost:3002/api/dashboard

# Using wrk
wrk -t12 -c400 -d30s http://localhost:3002/api/dashboard
```

#### WebSocket Load Testing

```bash
# Connect multiple WebSocket clients
npm install -g artillery
artillery quick --count 50 --num 10 ws://localhost:3002/ws
```

### Profiling

#### Go Profiling

```go
import _ "net/http/pprof"

// Access at http://localhost:3002/debug/pprof/
go tool pprof http://localhost:3002/debug/pprof/profile
```

#### Memory Analysis

```bash
# Enable memory profiling
export CAATSM_PPROF_ENABLED=true

# Get heap profile
go tool pprof http://localhost:3002/debug/pprof/heap
```

### Database Performance

```sql
-- Check slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- Analyze table statistics
ANALYZE telegrams;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE tablename = 'telegrams';
```

## Deployment Testing

### Local Production Testing

```bash
# Build production image
make docker-build

# Start production stack
make docker-up

# Test production endpoints
curl https://localhost:3002/api/health
curl https://localhost:3002/api/dashboard
```

### Configuration Testing

```bash
# Test with production config
docker run --rm \
  -v $(pwd)/config/config.toml:/app/config/config.toml \
  caatsm-dashboard:latest \
  ./caatsm -config /app/config/config.toml --help
```

### Integration Testing

```bash
# Test full message flow
# 1. Start stack
make docker-up

# 2. Publish test message
task publish-stream:fast

# 3. Verify in database
docker compose exec postgres psql -U caatsm -d caatsm -c "SELECT COUNT(*) FROM telegrams;"

# 4. Verify in search
curl "http://localhost:7700/indexes/telegrams/search?q=*"

# 5. Test WebSocket
# Open browser to http://localhost:3002 and check real-time updates
```

### Health Checks

```bash
# Test all service health
curl http://localhost:3002/api/health
curl http://localhost:7700/health
curl http://localhost:8222/
redis-cli -h localhost ping
```

## Troubleshooting

### Common Issues

#### Backend Won't Start

```bash
# Check dependencies
docker compose ps

# Check logs
docker compose logs app

# Test database connection
docker compose exec postgres pg_isready -U caatsm -d caatsm
```

#### Frontend Build Fails

```bash
# Clear build cache
cd frontend && rm -rf .svelte-kit build node_modules

# Reinstall dependencies
make frontend-install

# Check Node/Deno version
node --version
deno --version
```

#### Tests Failing

```bash
# Clean test cache
go clean -testcache

# Run with verbose output
go test -v ./internal/domain

# Check test dependencies
task dev:up
```

#### WebSocket Not Working

```bash
# Check Redis connection
docker compose exec redis redis-cli ping

# Check WebSocket logs
docker compose logs app | grep -i websocket

# Test WebSocket manually
# Use browser dev tools or websocat
```

### Getting Help

1. **Check logs**: `docker compose logs -f [service]`
2. **Review docs**: See `docs/troubleshooting.md`
3. **Check issues**: GitHub issues for similar problems
4. **Community**: Ask in project discussions

### Development Checklist

- [ ] Code follows style guidelines
- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated
- [ ] No secrets committed
- [ ] Database migrations tested
- [ ] Frontend builds successfully
- [ ] WebSocket connections work
- [ ] API endpoints respond correctly
- [ ] Error handling tested
- [ ] Performance acceptable

This guide should cover all aspects of development and testing for the CAATSM Dashboard. For more specific issues, refer to the troubleshooting documentation or create an issue in the repository.