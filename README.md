# CAATSM Dashboard

**CAATSM** (Civil Aviation Aerogram Traffic Stream Monitor) Dashboard provides realtime monitoring, search, and analytics for aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC networks.

## Features

- ⚡️ **High Performance**: Go + Echo API with HTMX/templ front-end
- 📬 **Real-time Ingestion**: NATS JetStream for message streaming
- 🔍 **Full-text Search**: Meilisearch for fast, typo-tolerant search
- 📊 **Analytics**: PostgreSQL + TimescaleDB for time-series data
- 🚀 **Scalable**: Redis caching, connection pooling, and async processing
- 📈 **Observability**: Prometheus + Grafana for metrics and monitoring
- 🔒 **Secure**: BasicAuth/JWT authentication, rate limiting, audit logging

## Architecture

```
┌───────────────────────────────────────┐
│                前端层                  │
│ HTMX + Alpine.js + Tailwind + templ    │
│  ├─ Dashboard 仪表盘 (SSE 实时刷新)      │
│  ├─ 搜索结果视图 (HTMX 局部渲染)          │
│  └─ 历史记录 / 导出页面 (分页)            │
└───────────────────┬───────────────────┘
                    │ REST + SSE (TLS)
┌───────────────────▼───────────────────┐
│              应用服务层                │
│ Go Echo + Clean Architecture           │
│  ├─ Handler 层：HTTP 控制器             │
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
        │ Redis Cache │
        │ Session/SSE │
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
- Node.js 20+ (for Tailwind assets)
- Docker & Docker Compose (for containerized services)
- PostgreSQL 15+ (or TimescaleDB)
- Meilisearch
- NATS Server
- Redis 7+

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
```

### Database Setup

Run database migrations:

```bash
# Using psql
psql -U caatsm -d caatsm -f migrations/001_create_telegrams_table.up.sql
psql -U caatsm -d caatsm -f migrations/002_create_audit_logs_table.up.sql

# Or use a migration tool like goose
goose -dir migrations postgres "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable" up
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

#### 3. Run the Application

```bash
# Generate templ components
make generate

# Build Tailwind CSS
make tailwind-build

# Run with hot reload (uses default config or .env.local)
make dev

# Or run with specific config file
./bin/caatsm -config config/config.local.toml

# Or run directly
make run
```

**Note**: 
- The development compose file (`docker-compose.dev.yml`) only includes dependencies and is compatible with both Docker and Podman.
- The application runs locally for better development experience with hot reload.
- If Meilisearch generates a new master key, update it in `.env.local` or `config/config.local.toml`.

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
- Redis (port 6379)
- Prometheus (port 9090)
- Grafana (port 3000)

## Development

### Project Structure

```
caatsm/
├─ cmd/
│  ├─ server/             # Main service entry point
│  └─ sync/               # Background sync worker
├─ config/
│  ├─ config.go           # Configuration structs
│  ├─ loader.go           # Viper config loader
│  └─ config.toml         # Default configuration
├─ internal/
│  ├─ app/                # Dependency injection container
│  ├─ handlers/           # HTTP controllers
│  ├─ services/           # Business logic layer
│  ├─ repository/         # Data access layer
│  │  ├─ postgres/        # PostgreSQL implementation
│  │  ├─ meili/           # Meilisearch implementation
│  │  ├─ nats/            # NATS consumer implementation
│  │  └─ cache/           # Redis cache implementation
│  ├─ platform/           # External client wrappers
│  ├─ sync/               # Message sync worker
│  ├─ metrics/            # Prometheus metrics
│  ├─ observability/      # Logging, tracing, middleware
│  ├─ auth/               # Authentication middleware
│  └─ models/             # Data models
├─ views/                 # templ templates
├─ assets/                # Tailwind/PostCSS sources
├─ public/                # Generated static assets
├─ migrations/            # Database migrations
├─ scripts/               # Utility scripts
└─ deploy/                # Deployment configs
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
make tailwind      # Build Tailwind CSS (watch mode)
make tailwind-build # Build Tailwind CSS once
make docker-build  # Build Docker image
make docker-up     # Start Docker Compose stack
make docker-down   # Stop Docker Compose stack
```

### Taskfile

Alternatively, use Taskfile:

```bash
task dev            # Start hot reload server
task build          # Build server binary
task test           # Run tests
task lint           # Run linter
task tailwind       # Build Tailwind CSS (watch)
task dev:config     # Setup local development configuration (.env.local)
task dev:up         # Start development dependencies (Docker/Podman)
task dev:down       # Stop development dependencies
task dev:logs       # Show logs from development dependencies
task dev:run        # Run application with local config file
task dev:meili-key  # Extract Meilisearch master key from logs
task docker:up      # Start full Docker Compose stack
task docker:down    # Stop Docker Compose stack
```

## API Endpoints

### Search

- `POST /api/search` - Search telegrams
- `GET /api/autocomplete?term=...&size=5` - Autocomplete suggestions

### Statistics

- `GET /api/stats/total` - Total message count (24h)
- `GET /api/stats/priority` - Priority breakdown

### Export

- `GET /api/export?format=csv&query=...` - Export search results (CSV/Excel/PDF)

### Health

- `GET /api/health` - Health check
- `GET /metrics` - Prometheus metrics

## Deployment

### Production Deployment

1. **Environment Variables**: Set all required environment variables:

```bash
export CAATSM_DATABASE_DSN="postgres://user:pass@host:5432/caatsm?sslmode=require"
export CAATSM_MEILISEARCH_HOST="https://meilisearch.example.com"
export CAATSM_MEILISEARCH_API_KEY="your-api-key"
export CAATSM_NATS_URL="nats://nats.example.com:4222"
export CAATSM_REDIS_ADDR="redis.example.com:6379"
```

2. **Database Migrations**: Run migrations before starting the application:

```bash
goose -dir migrations postgres "$CAATSM_DATABASE_DSN" up
```

3. **Build and Run**:

```bash
make build
./bin/caatsm -config /path/to/config.toml
```

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

### Grafana Dashboards

Import Grafana dashboards from `deploy/grafana/` (coming soon).

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

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
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [HTMX](https://htmx.org/) - Hypermedia-driven web applications
