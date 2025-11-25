# CAATSM Dashboard

**CAATSM** (Civil Aviation Aerogram Traffic Stream Monitor) Dashboard provides realtime monitoring, search, and analytics for aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC networks.

## Features

- ⚡️ **High Performance**: Go + Echo API with Deno + Svelte front-end
- 📬 **Real-time Ingestion**: NATS JetStream for message streaming
- 🔍 **Full-text Search**: Meilisearch for fast, typo-tolerant search
- 📊 **Analytics**: PostgreSQL + TimescaleDB for time-series data
- 🚀 **Scalable**: Valkey caching, connection pooling, and async processing
- 📈 **Observability**: Prometheus + Grafana for metrics and monitoring
- 🔒 **Security**: Optional Basic Auth middleware with planned JWT, rate limiting, and audit logging enhancements
- 🌐 **WebSocket**: Real-time updates via WebSocket (replacing SSE)

## Architecture

```
┌───────────────────────────────────────┐
│                前端层                  │
│ Deno + Svelte + SvelteKit              │
│  ├─ Dashboard 仪表盘 (WebSocket 实时)   │
│  ├─ 搜索结果视图 (REST API)             │
│  └─ 历史记录 / 导出页面 (分页)            │
└───────────┬───────────────────────────┘
            │ REST + WebSocket
┌───────────▼───────────────────────────┐
│              应用服务层                │
│ Go Echo + Clean Architecture           │
│  ├─ Handler 层：HTTP 控制器             │
│  ├─ WebSocket 处理器                   │
│  ├─ Service 层：业务逻辑/集成            │
│  ├─ Repository 层：数据抽象             │
│  └─ Metrics/Tracing：Prom + OTEL        │
└──────────────┬──────────────┬──────────────┘
               │              │
     ┌─────────▼─────────┐    │
     │  消息流 & 同步服务   │    │
     │ NATS JetStream +    │    │
     │ SyncWorker/Indexer  │    │
     └─────────┬──────────┘    │
               │                │
     ┌─────────▼────────┐ ┌────▼───────────┐
     │ PostgreSQL/       │ │ Meilisearch     │
     │ TimescaleDB        │ │ (全文检索 & 排序) │
     │ (结构化存储)        │ └─────────────────┘
     └─────────┬─────────┘
               │
        ┌──────▼──────┐
        │ Valkey Cache│
        │ Pub/Sub     │
        └──────┬──────┘
               │
    ┌──────────▼──────────┐
    │ Observability Stack │
    │ Prometheus/Grafana  │
    │ Loki/Tempo (可选)    │
    └──────────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.25+
- **Deno 2.0+** (recommended for frontend) or **Node.js 20+** (alternative)
- Docker & Docker Compose (for containerized services)
- PostgreSQL 15+ (or TimescaleDB)
- Meilisearch
- NATS Server
- Valkey 9+

### Development Utilities & Dependencies

The following development tools and dependencies are required for the dev commands:

#### Required Tools

- **air** - Hot reload tool for Go development
  ```bash
  go install github.com/air-verse/air@latest
  ```

- **templ** - Templating tool (automatically installed via `go run`)
  - Used via: `go run github.com/a-h/templ/cmd/templ@v0.3.960 generate`
  - Required for: `make generate`, `task generate`, and `task dev`
  - Note: Still used for legacy pages, new frontend uses Svelte

- **golangci-lint** - Go linter (optional but recommended)
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **task** - Task runner (optional, for Taskfile commands)
  ```bash
  # macOS
  brew install go-task/tap/go-task
  
  # Linux
  sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
  ```

- **goose** - Database migration tool (optional)
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

#### Frontend Dependencies

The frontend supports both **Deno 2.0+** (recommended) and **Node.js 20+**. 

**Quick Setup with Deno (Recommended)**:
```bash
# Install Deno and setup frontend (one command)
task frontend:setup
# or
make frontend-setup

# Then start development
task frontend:dev
# or
make frontend-dev
```

**Manual Setup with Deno**:
```bash
# Install Deno
task install:deno
# or
make install-deno

# Cache frontend dependencies
task frontend:install
# or
make frontend-install

# Start development
cd frontend
deno task dev  # No installation needed!
```

**With Node.js**:
```bash
cd frontend
npm install
npm run dev
```

> **Note**: Deno is recommended as it matches the original plan and provides zero-configuration development. Dependencies are automatically downloaded by Deno. Use `task frontend:setup` or `make frontend-setup` for a complete one-command setup.

### Installation

```bash
git clone https://github.com/windy/caatsm-dashboard.git
cd caatsm-dashboard
go mod tidy
```

### Configuration

All configuration values live in `config/config.toml` and can be overridden with environment variables prefixed with `CAATSM_`. 

For local development:
- Use `config/config.local.toml` for TOML-based configuration (uses `localhost` addresses)
- Use `.env.local` for environment variables (copy from `env.local.example`)
- The loader automatically loads `.env.local` (if exists) before `.env`

See `env.example` and `env.local.example` for starter sets.

```toml
[server]
host = "0.0.0.0"
port = 3002

[database]
dsn = "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable"

[meilisearch]
host = "http://localhost:7700"
api_key = "masterKey"
index = "telegrams"

[nats]
url = "nats://localhost:4222"
stream = "telegrams"
consumer = "dashboard-sync"

[redis]
addr = "localhost:6379"  # Valkey (Redis-compatible)
```

### Database Setup

Run database migrations:

```bash
# Using goose (recommended)
goose -dir migrations postgres "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable" up

# Or using task
task migrate

# Note: Migration files use goose annotations (-- +goose Up/Down)
# If using psql directly, you'll need to extract the "Up" portion manually
```

### Run Locally

#### 1. Setup Local Configuration (Optional)

For local development, you can use either:

**Option A: TOML Configuration File**
```bash
# Use the local development config file
./bin/caatsm -config config/config.local.toml
```

**Option B: Environment Variables**
```bash
# Copy the example file and customize
cp env.local.example .env.local

# Edit .env.local with your settings
# If Meilisearch generates a new master key, update CAATSM_MEILISEARCH_API_KEY
```

The configuration loader will automatically:
- Load `.env.local` (if exists) for local overrides
- Load `.env` for shared defaults
- Use `config/config.local.toml` if specified with `-config` flag

#### 2. Start Development Dependencies

```bash
# Using Docker:
docker compose -f docker-compose.dev.yml up -d

# Or using Podman:
podman compose -f docker-compose.dev.yml up -d

# Or using Task:
task dev:up
```

#### 3. Run the Backend

```bash
# Generate templ components (for legacy pages)
make generate

# Run with hot reload (uses default config or .env.local)
make dev

# Or run with specific config file
./bin/caatsm -config config/config.local.toml

# Or run directly
make run
```

#### 4. Run the Frontend

**Quick Start (Recommended)**:
```bash
# One-command setup and start (uses Deno, falls back to npm)
task frontend:setup && task frontend:dev
# or
make frontend-setup && make frontend-dev
```

**With Deno (Recommended)**:
```bash
# Setup Deno and cache dependencies
task frontend:setup
# or
make frontend-setup

# Start development server
task frontend:dev
# or
make frontend-dev
# or manually:
cd frontend
deno task dev

# The frontend will be available at http://localhost:5173
# It will proxy API requests to the Go backend at http://localhost:3002
```

**With Node.js**:
```bash
cd frontend

# Install dependencies (first time only)
npm install

# Start development server
npm run dev

# The frontend will be available at http://localhost:5173
# It will proxy API requests to the Go backend at http://localhost:3002
```

> **Recommendation**: Use `task frontend:setup` or `make frontend-setup` for a complete one-command setup. It installs Deno if needed, caches dependencies, and falls back to npm if Deno is unavailable.

**Note**: 
- The development compose file (`docker-compose.dev.yml`) only includes dependencies and is compatible with both Docker and Podman.
- The application runs locally for better development experience with hot reload.
- If Meilisearch generates a new master key, update it in `.env.local` or `config/config.local.toml`.

#### 5. Run Background Workers (Optional)

For message ingestion and indexing:

```bash
# Run sync worker to consume messages from NATS
task sync
# or
go run ./cmd/sync -config config/config.local.toml
```

### Docker Compose

Start the entire stack:

```bash
docker compose up --build
```

This will start:
- Application server (port 3002)
- PostgreSQL (port 5432)
- Meilisearch (port 7700)
- NATS (port 4222)
- Valkey (port 6379)
- Prometheus (port 9090)
- Grafana (port 3000)

## Development

### Project Structure

```
caatsm/
├── cmd/
│   ├── server/             # Main service entry point
│   ├── sync/               # Background sync worker
│   ├── generate-test-data/ # Test data generator
│   └── publish-stream/     # NATS message publisher for testing
├── config/
│   ├── config.go           # Configuration structs
│   ├── loader.go           # Viper config loader
│   └── config.toml         # Default configuration
├── internal/
│   ├── app/                # Dependency injection container
│   ├── domain/             # Domain layer (business entities & rules)
│   ├── application/        # Application layer (use cases)
│   ├── infrastructure/      # Infrastructure layer (external adapters)
│   ├── handlers/           # HTTP controllers
│   │   ├── websocket.go    # WebSocket handler
│   │   └── handlers.go     # REST API handlers
│   ├── services/           # Business logic layer
│   ├── repository/         # Data access layer
│   │   ├── postgres/        # PostgreSQL implementation
│   │   ├── meili/           # Meilisearch implementation
│   │   ├── nats/            # NATS consumer implementation
│   │   └── cache/           # Valkey cache implementation
│   ├── platform/           # External client wrappers
│   ├── sync/               # Message sync worker
│   ├── metrics/            # Prometheus metrics
│   ├── observability/      # Logging, tracing, middleware
│   ├── auth/               # Authentication middleware
│   ├── models/             # Data models
│   └── testing/            # Test utilities & mocks
├── frontend/               # Deno + Svelte frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/  # Svelte components
│   │   │   ├── stores/      # Svelte stores
│   │   │   ├── services/    # API and WebSocket clients
│   │   │   └── utils/       # Utilities
│   │   └── routes/          # SvelteKit routes
│   ├── deno.json            # Deno configuration
│   └── svelte.config.js     # SvelteKit configuration
├── views/                   # Legacy templ templates (for old pages)
├── assets/                  # Legacy UnoCSS sources
├── public/                   # Generated static assets
├── migrations/               # Database migrations
├── scripts/                  # Utility scripts
└── deploy/                   # Deployment configs
```

### Available Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests
make lint          # Run linter
make run           # Run the application
make dev           # Run with hot reload (air)
make generate      # Generate templ components
make docker-build  # Build Docker image
make docker-up     # Start Docker Compose stack
make docker-down   # Stop Docker Compose stack
make install-deno  # Install Deno if not available
make frontend-setup # Full frontend setup (Deno + dependencies)
make frontend-dev  # Run frontend dev server
make frontend-build # Build frontend for production
make frontend-test # Run frontend tests
```

### Taskfile

Alternatively, use Taskfile:

```bash
task dev            # Start hot reload server
task build          # Build server binary
task test           # Run tests
task lint           # Run linter
task install:deno   # Install Deno if not available
task frontend:setup # Full frontend setup (Deno + dependencies)
task frontend:dev   # Run frontend dev server
task frontend:build # Build frontend for production
task frontend:test  # Run frontend tests
task frontend:install # Pre-cache frontend dependencies with Deno
task dev:up         # Start development dependencies (Docker/Podman)
task dev:down       # Stop development dependencies
task dev:logs       # Show logs from development dependencies
task dev:run        # Run application with local config file
task sync           # Run sync worker to consume messages from NATS
task migrate        # Run database migrations
task generate-test-data  # Generate test telegram data for development
task publish-stream      # Publish live messages to NATS for testing
task docker:up      # Start full Docker Compose stack
task docker:down    # Stop Docker Compose stack
```

### Frontend Development

**Quick Setup**:
```bash
# One-command setup (installs Deno, caches dependencies)
task frontend:setup
# or
make frontend-setup
```

**With Deno (Recommended)**:
```bash
# Using Taskfile/Makefile (recommended)
task frontend:dev   # Start development server
task frontend:build # Build for production
task frontend:test  # Run tests
# or
make frontend-dev
make frontend-build
make frontend-test

# Or manually in frontend directory
cd frontend
deno task dev       # Start development server
deno task build     # Build for production
deno task preview   # Preview production build
deno task check     # Type check
deno fmt            # Format code
deno lint           # Lint code
```

**With Node.js**:
```bash
cd frontend

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npm run check

# Format code
npm run format

# Run E2E tests
npm run test
```

> **Recommendation**: Use **Deno** for development. Run `task frontend:setup` or `make frontend-setup` for a complete one-command setup. Deno requires no installation and provides a better developer experience with built-in TypeScript, formatter, and linter.

## API Endpoints

### REST API Endpoints (JSON)

- `GET /api/search` - Search telegrams (returns JSON)
- `POST /api/search` - Search telegrams (returns JSON)
- `GET /api/autocomplete?term=...&size=5` - Autocomplete suggestions (returns JSON)
- `GET /api/stats/total` - Total message count (24h, returns JSON)
- `GET /api/stats/priority` - Priority breakdown (returns JSON)
- `GET /api/stats/type` - Message type breakdown (returns JSON)
- `GET /api/export?format=csv&query=...` - Export search results (CSV/Excel/PDF)
- `GET /api/health` - Health check
- `GET /metrics` - Prometheus metrics

### WebSocket Endpoint

- `GET /ws` - WebSocket connection for real-time updates
  - Message format: `{ "type": "message|stats-total|stats-priority|stats-type", "data": {...} }`
  - Supports ping/pong heartbeat
  - Automatic reconnection on client side

## Deployment

### Production Deployment

1. **Environment Variables**: Set all required environment variables:

```bash
export CAATSM_DATABASE_DSN="postgres://user:pass@host:5432/caatsm?sslmode=require"
export CAATSM_MEILISEARCH_HOST="https://meilisearch.example.com"
export CAATSM_MEILISEARCH_API_KEY="your-api-key"
export CAATSM_NATS_URL="nats://nats.example.com:4222"
export CAATSM_REDIS_ADDR="valkey.example.com:6379"  # Valkey (Redis-compatible)
```

2. **Database Migrations**: Run migrations before starting the application:

```bash
goose -dir migrations postgres "$CAATSM_DATABASE_DSN" up
# or
task migrate
```

3. **Build Frontend**:

```bash
# Using Taskfile/Makefile (recommended, uses Deno if available)
task frontend:build
# or
make frontend-build

# Or manually with Deno
cd frontend
deno task build

# Or manually with Node.js
cd frontend
npm install
npm run build
```

4. **Build and Run Backend**:

```bash
make build
# or
task build

./bin/caatsm -config /path/to/config.toml
```

The Go backend will serve the built Svelte frontend from `frontend/build`.

### Docker Deployment

```bash
docker build -t caatsm-dashboard:latest .
docker run -d \
  -p 3002:3002 \
  -v $(pwd)/config/config.toml:/app/config/config.toml \
  caatsm-dashboard:latest
```

### Kubernetes Deployment

See `deploy/` directory for Kubernetes manifests (coming soon).

## Monitoring

### Prometheus Metrics

The application exposes Prometheus metrics at `/metrics`:

- `telegrams_ingested_total` - Total number of telegrams processed
- `search_latency_seconds` - Search request latency
- `http_requests_total` - HTTP request count
- `http_request_duration_seconds` - HTTP request duration
- `websocket_connections` - Active WebSocket connections (to be added)

### Grafana Dashboards

Import Grafana dashboards from `deploy/grafana/` (coming soon).

## Testing

```bash
# Run all tests
make test
# or
task test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

For detailed testing documentation, see [TESTING.md](TESTING.md) and [internal/testing/README.md](internal/testing/README.md).

## Recent Improvements

### Type Safety
- Improved type safety in WebSocket handler with proper type assertions
- Enhanced error handling to prevent panics from invalid type conversions
- Better validation of event data structures

### Test Coverage
- Added comprehensive tests for WebSocket handler (`internal/handlers/websocket_test.go`)
- Added EventBroadcaster tests (`internal/handlers/broadcaster_test.go`)
- Created service mocks for StatsService and QueryService
- Test coverage for handlers layer targeting ≥80%

### Linting Configuration
- Updated linting rules for better code quality
- Consistent use of `any` type instead of `interface{}` in mocks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linter
5. Submit a pull request

## License

MIT © 2025 Windy

## Acknowledgments

- [Echo](https://echo.labstack.com/) - High-performance HTTP framework
- [Meilisearch](https://www.meilisearch.com/) - Fast, typo-tolerant search
- [NATS](https://nats.io/) - Cloud-native messaging
- [Svelte](https://svelte.dev/) - Modern frontend framework
- [SvelteKit](https://kit.svelte.dev/) - Full-stack Svelte framework
- [Vite](https://vitejs.dev/) - Next generation frontend tooling
- [UnoCSS](https://unocss.com/) - Instant atomic CSS engine
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation for Go
