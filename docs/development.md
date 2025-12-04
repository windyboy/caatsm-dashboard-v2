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
- **Deno**: 2.0+ (recommended) or Bun (recommended as fallback) or Node.js 18+ (20+ recommended)
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
# Create .env.local from example (if not exists)
make dev-config
# or
task dev:config

# Start all dependencies (Postgres, Meilisearch, NATS, Valkey)
make dev-up
# or
task dev:up

# Verify services are running
docker compose -f docker-compose.dev.yml ps

# Run database migrations
make backend-migrate
# or
task backend:migrate
```

### 3. Start Development Servers

```bash
# Terminal 1: Backend API (with hot reload)
make backend-dev
# or
task backend:dev

# Terminal 2: Frontend (with hot reload)
make frontend-dev
# or
task frontend:dev
```

### 4. Access Application

- **Frontend**: http://localhost:5173
- **API**: http://localhost:3002
- **API Health**: http://localhost:3002/api/health
- **NATS Monitoring**: http://localhost:8222
- **Meilisearch**: http://localhost:7700
- **PostgreSQL**: localhost:5432
- **Redis/Valkey**: localhost:6379

### 5. Generate Test Data (Optional)

```bash
# Generate sample telegrams in database
make backend-generate-test-data
# or
task backend:generate-test-data

# Publish test messages to NATS
make backend-publish-stream-fast  # 50 messages at 1/sec
# or
task backend:publish-stream:fast

make backend-publish-stream-slow  # Messages every 5 seconds
# or
task backend:publish-stream:slow
```

## Development Environment

### Docker Compose Services

The development environment uses `docker-compose.dev.yml` to run dependencies:

- **PostgreSQL 15+ (TimescaleDB)**: Primary database with time-series extensions
- **Meilisearch 1.5+**: Full-text search engine
- **NATS 2.12.2**: Message broker with JetStream
- **Valkey 9**: Redis-compatible cache and pub/sub

All services:
- Run on the `backend` network
- Use named volumes for data persistence
- Include health checks for dependency management
- Read configuration from `.env.local` if present

### Managing Development Services

```bash
# Start all services
make dev-up

# Stop all services
make dev-down

# View logs
make dev-logs
# or for specific service
docker compose -f docker-compose.dev.yml logs -f postgres

# Check service status
docker compose -f docker-compose.dev.yml ps

# Restart a specific service
docker compose -f docker-compose.dev.yml restart postgres

# Remove all data (fresh start)
make dev-down
docker volume rm caatsm-dashboard-v2_pg_data_dev
docker volume rm caatsm-dashboard-v2_meili_data_dev
docker volume rm caatsm-dashboard-v2_nats_data_dev
docker volume rm caatsm-dashboard-v2_redis_data_dev
```

### Environment Configuration

Create `.env.local` from the example:

```bash
make dev-config
# or
task dev:config
```

Edit `.env.local` to customize:
- Database credentials
- Service ports
- API keys (for development only)

**Note**: Never commit `.env.local` to version control. It's in `.gitignore`.

## Development Workflow

### Backend Development (Go)

#### Project Structure

```
cmd/
├── server/          # Main API server
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

The frontend runs SvelteKit 2.x with Svelte 5 runes, Vite 7, UnoCSS, and @melt-ui/svelte. Development defaults to Deno 2.x; Bun is the preferred fallback when you need full HMR (Deno dev disables Vite 7 HMR to avoid WebSocket issues).

@melt-ui/svelte provides accessible, headless UI components that integrate seamlessly with UnoCSS and our custom styling system. All UI components use the @melt-ui/svelte builder pattern for enhanced accessibility and keyboard navigation.

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

#### Component Development with @melt-ui/svelte

All UI components use @melt-ui/svelte's builder pattern for accessible, headless components:

```svelte
<!-- lib/components/ui/Button.svelte -->
<script lang="ts">
  import { createButton, melt } from '@melt-ui/svelte';
  
  export let variant: 'default' | 'ghost' = 'default';
  export let disabled: boolean = false;
  
  const {
    elements: { root },
    states: { disabled: isDisabled }
  } = createButton({
    disabled: $derived(disabled)
  });
</script>

<button use:melt={$root} class="button" class:button--ghost={variant === 'ghost'}>
  <slot />
</button>
```

Components maintain existing styles from `app.css` while gaining built-in accessibility features (ARIA attributes, keyboard navigation, focus management).

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
make backend-test-unit
# or
task backend:test:unit
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
make backend-test-integration
# or
task backend:test:integration
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
cd frontend && bun run test:unit
# or
cd frontend && npm run test:unit
```

#### E2E Tests (Playwright)

```bash
# Run E2E tests
make frontend-test
# or
cd frontend && deno task test:e2e   # requires Bun for the Playwright runner
# or
cd frontend && bun run test
# or
cd frontend && npm run test:e2e
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
make backend-generate-test-data
# or
task backend:generate-test-data

# Publish messages to NATS for testing
make backend-publish-stream-fast  # 50 messages at 1/sec
# or
task backend:publish-stream:fast
make backend-publish-stream-slow  # Messages every 5 seconds
# or
task backend:publish-stream:slow
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
make backend-lint
# or
task backend:lint
# or
golangci-lint run ./...

# Run frontend linter
cd frontend && deno lint
# or
cd frontend && bun run lint
# or
cd frontend && npm run lint
```

### Code Coverage

```bash
# Backend coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend coverage
cd frontend && deno task test:unit:coverage
# or
cd frontend && bun run test:unit:coverage
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
make backend-dev
# or
task backend:dev
```

#### Database Debugging

```bash
# Connect to database (using Docker)
docker compose -f docker-compose.dev.yml exec postgres psql -U caatsm -d caatsm

# Or using local psql
psql "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable"

# Check recent messages
SELECT message_id, type, time, flight_number, source, destination
FROM telegrams
ORDER BY time DESC
LIMIT 10;

# Check table statistics
SELECT schemaname, tablename, n_live_tup, n_dead_tup, last_vacuum, last_autovacuum
FROM pg_stat_user_tables
WHERE tablename = 'telegrams';

# Check Meilisearch index
curl "http://localhost:7700/indexes/telegrams/search?q=*"

# Check Meilisearch stats
curl "http://localhost:7700/stats"
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
# Check API health
curl http://localhost:3002/api/health

# View dependency logs
make dev-logs
# or specific service
docker compose -f docker-compose.dev.yml logs -f postgres
docker compose -f docker-compose.dev.yml logs -f nats
docker compose -f docker-compose.dev.yml logs -f redis

# Check NATS streams (requires NATS CLI)
docker compose -f docker-compose.dev.yml exec nats nats stream ls
docker compose -f docker-compose.dev.yml exec nats nats stream info TELEGRAMS

# Monitor Redis/Valkey
docker compose -f docker-compose.dev.yml exec redis valkey-cli ping
docker compose -f docker-compose.dev.yml exec redis valkey-cli keys "*"
docker compose -f docker-compose.dev.yml exec redis valkey-cli info stats

# Check Meilisearch health
curl http://localhost:7700/health

# Check NATS monitoring
curl http://localhost:8222/healthz
curl http://localhost:8222/varz
curl http://localhost:8222/jsz

# Test database connection
docker compose -f docker-compose.dev.yml exec postgres pg_isready -U caatsm
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
bun add -g artillery
# or npm install -g artillery
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
# 1. Start development dependencies
make dev-up

# 2. Start backend and frontend
make backend-dev              # Terminal 1
# or
task backend:dev
make frontend-dev     # Terminal 2
# or
task frontend:dev

# 3. Publish test messages
make backend-publish-stream-fast
# or
task backend:publish-stream:fast

# 4. Verify in database
docker compose -f docker-compose.dev.yml exec postgres psql -U caatsm -d caatsm -c "SELECT COUNT(*) FROM telegrams;"

# 5. Verify in search
curl "http://localhost:7700/indexes/telegrams/search?q=*"

# 6. Test WebSocket
# Open browser to http://localhost:5173 and check real-time updates
# Or use browser console: new WebSocket('ws://localhost:3002/ws')
```

### Health Checks

```bash
# Test all service health
curl http://localhost:3002/api/health

# Dependency health checks
curl http://localhost:7700/health                    # Meilisearch
curl http://localhost:8222/healthz                  # NATS
docker compose -f docker-compose.dev.yml exec redis valkey-cli ping  # Redis/Valkey
docker compose -f docker-compose.dev.yml exec postgres pg_isready -U caatsm  # PostgreSQL

# Check Docker service health
docker compose -f docker-compose.dev.yml ps
```

## Troubleshooting

### Common Issues

#### Backend Won't Start

```bash
# Check dependencies are running
docker compose -f docker-compose.dev.yml ps

# Check dependency logs
make dev-logs

# Test database connection
docker compose -f docker-compose.dev.yml exec postgres pg_isready -U caatsm -d caatsm

# Verify all services are healthy
docker compose -f docker-compose.dev.yml ps

# Check if ports are already in use
lsof -i :3002  # API port
lsof -i :5432  # PostgreSQL
lsof -i :7700  # Meilisearch
lsof -i :4222  # NATS
lsof -i :6379  # Redis

# Restart all dependencies
make dev-down && make dev-up
```

#### Frontend Build Fails

```bash
# Clear build cache
cd frontend && rm -rf .svelte-kit build node_modules

# Reinstall dependencies
make frontend-install

# Check runtime version
deno --version
bun --version
node --version
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
# Check Redis/Valkey connection
docker compose -f docker-compose.dev.yml exec redis valkey-cli ping

# Check Redis pub/sub channels
docker compose -f docker-compose.dev.yml exec redis valkey-cli pubsub channels

# Check WebSocket logs (if running in Docker)
docker compose logs app | grep -i websocket

# Test WebSocket manually in browser console
# const ws = new WebSocket('ws://localhost:3002/ws');
# ws.onopen = () => console.log('Connected');
# ws.onmessage = (e) => console.log('Message:', e.data);

# Or use websocat
websocat ws://localhost:3002/ws
```

### Getting Help

1. **Check logs**: `make dev-logs` or `docker compose -f docker-compose.dev.yml logs -f [service]`
2. **Review docs**: 
   - `docs/troubleshooting.md` - Common issues and solutions
   - `docs/architecture.md` - System architecture
   - `docs/configuration.md` - Configuration guide
   - `AGENTS.md` - Quick reference for AI agents
3. **Check service health**: `docker compose -f docker-compose.dev.yml ps`
4. **Verify configuration**: Check `.env.local` and `config/config.local.toml`
5. **Check issues**: GitHub issues for similar problems

### Development Checklist

- [ ] Code follows style guidelines
- [ ] Tests pass (`make backend-test`)
- [ ] Linting passes (`make backend-lint`)
- [ ] Documentation updated
- [ ] No secrets committed
- [ ] Database migrations tested
- [ ] Frontend builds successfully
- [ ] WebSocket connections work
- [ ] API endpoints respond correctly
- [ ] Error handling tested
- [ ] Performance acceptable

This guide should cover all aspects of development and testing for the CAATSM Dashboard. For more specific issues, refer to the troubleshooting documentation or create an issue in the repository.
