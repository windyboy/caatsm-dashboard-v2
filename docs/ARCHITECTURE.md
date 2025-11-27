# CAATSM Dashboard Architecture

## Overview

The CAATSM Dashboard follows **Clean Architecture** principles with clear separation of concerns across layers. The architecture is designed for maintainability, testability, and scalability.

**Current Status**: ✅ **Simplified Architecture Migration Complete** (Nov 27, 2025) - Legacy `internal/application/` removed, new `internal/app/` with Container pattern.

## Architecture Layers

The system follows a **Simplified Clean Architecture** with clear layer separation:

```
┌─────────────────────────────────────────┐
│         Delivery Layer                  │
│   HTTP/WebSocket Handlers               │
│   (internal/delivery/)                  │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│      Application Layer (App)            │
│   Services + Container (DI)             │
│   (internal/app/)                       │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Domain Layer                    │
│   Business Logic & Entities             │
│   (internal/domain/)                    │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│      Infrastructure Layer               │
│   Repository Implementations            │
│   (internal/infrastructure/)            │
└─────────────────────────────────────────┘
```

**Dependency Rule**: Dependencies point **inward**. Outer layers depend on inner layers, never the reverse.

**Key Changes** (Nov 2025 Migration):
- ✅ `internal/application/` → `internal/app/` (simplified)
- ✅ `internal/transport/` → `internal/delivery/` (renamed for clarity)
- ✅ Introduced `app.Container` for centralized dependency injection
- ✅ Sync worker migrated to use Container pattern
- ✅ Removed unnecessary abstraction layers

## Layer Details

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Pure business logic with no external dependencies.

**Contains**:
- Domain entities (`telegram.go`)
- Business rules and validation (`validator.go`, `filters.go`)
- Domain events (`events.go`)
- Value objects and domain-specific types

**Principles**:
- No framework dependencies
- No infrastructure dependencies
- Self-contained business logic
- Immutable where possible

**Key Files**:
- `telegram.go` - Core domain entity with validation and normalization
- `events.go` - Domain events (TelegramReceived, TelegramValidated, etc.)
- `filters.go` - Domain-level search filters with **90-day time range validation**
- `errors.go` - Domain-specific error types

**Production Safeguards**:
- ✅ Time range validation (max 90 days) to prevent unbounded queries
- ✅ Input sanitization and validation
- ✅ Whitelist-based sort field validation

### 2. Application Layer (`internal/app/`) - **Simplified (Nov 2025)**

**Purpose**: Application services with centralized dependency injection via Container pattern.

**Contains**:
- Application services (dashboard, search, stats, export)
- Port interfaces (repository abstractions)
- Container for dependency injection (`app.go`)

**Depends On**:
- Domain layer (uses domain entities and validation)
- Port interfaces (depends on abstractions)

**Key Components**:
- `app/services/dashboard.go` - Unified dashboard service with parallel data fetching
- `app/services/search.go` - Search service with caching
- `app/services/stats.go` - Statistics aggregation
- `app/services/export.go` - **Streaming export** (handles 50k+ records without OOM)
- `app/services/realtime.go` - Real-time manager for WebSocket updates
- `app/ports/repository.go` - Port interfaces
- `app/app.go` - **Container** with dependency injection

**Container Pattern**:
```go
type Container struct {
    Config       *config.AppConfig
    Logger       *zap.Logger
    Repo         ports.Repository
    Cache        ports.Cache
    Search       ports.SearchIndex
    EventBus     event.EventBus
    TelegramStore repository.TelegramStore  // For direct access
    SearchIndex   repository.SearchIndex     // For direct access
    DashboardService *services.DashboardService
}
```

**Production Features**:
- ✅ Streaming CSV export with chunking (1000 records per chunk)
- ✅ Comprehensive health checks via Container
- ✅ Export limit enforcement (max 10,000 records)
- ✅ Parallel data fetching in dashboard service

### 3. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: External concerns and implementations. Adapts external systems to application ports.

**Contains**:
- Repository implementations
- External service clients
- WebSocket hub with backpressure
- Event bus implementations
- Resilience patterns (circuit breakers)

**Depends On**:
- Application ports (implements interfaces defined in application layer)
- Domain types (for data conversion)

**Key Packages**:
- `infrastructure/persistence/postgres/` - PostgreSQL repository with TimescaleDB
- `infrastructure/search/meilisearch/` - Meilisearch indexing
- `infrastructure/cache/valkey/` - Valkey/Redis caching
- `infrastructure/events/` - Event bus implementations (Redis-based)
- `infrastructure/ws/` - WebSocket hub with **backpressure control** and **slow client detection**
- `infrastructure/resilience/` - Circuit breakers

### 4. Delivery Layer (`internal/delivery/`) - **Renamed (Nov 2025)**

**Purpose**: HTTP/WebSocket handling, request/response formatting.

**Contains**:
- HTTP handlers
- WebSocket handlers with event broadcasting
- Request parsing and validation
- Response formatting (JSON)

**Depends On**:
- Application services (calls `app.Container.DashboardService`)
- No direct domain or infrastructure access

**Key Packages**:
- `delivery/http/` - HTTP handlers with **rate limiting** (10 req/sec)
- `delivery/ws/` - WebSocket handler with EventBroadcaster pattern

**Production Features**:
- ✅ Rate limiting middleware (10 req/sec, configurable)
- ✅ Input validation at delivery boundary
- ✅ WebSocket backpressure handling
- ✅ CORS configuration with allowed origins

## Data Flow

### Request Flow (Updated Nov 2025)

```
HTTP Request
    ↓
Delivery Layer (HTTP Handler)
    ↓ Parse request, validate input, rate limit
Application Layer (app.Container.DashboardService)
    ↓ Orchestrate business logic, parallel data fetching
Domain Layer (Domain Entity)
    ↓ Business rules, validation (max 90 days, etc.)
Infrastructure Layer (Repository)
    ↓ Persist/Query
External System (Database, Cache, etc.)
```

### Event Flow

```
Domain Event (e.g., TelegramValidated)
    ↓
Application Layer (Event Dispatcher)
    ↓
Event Handlers (Persistence, Indexing)
    ↓
Infrastructure Layer (Repository implementations)
    ↓
External Systems (Database, Search Engine)
```

## Domain Events

The system uses domain events to decouple processing steps:

1. **TelegramReceived** - Published when a telegram is received
2. **TelegramValidated** - Published after domain validation
3. **TelegramPersisted** - Published after database save
4. **TelegramIndexed** - Published after search indexing
5. **TelegramProcessed** - Published when all steps complete

**Event Handlers**:
- `PersistenceHandler` - Listens to `TelegramValidated`, saves to DB, publishes `TelegramPersisted`
- `IndexingHandler` - Listens to `TelegramPersisted`, indexes to Meilisearch, publishes `TelegramIndexed`

This allows handlers to fail independently without blocking the main flow.

## Configuration

Configuration is centralized in `config/config.go` with:
- Type-safe configuration structs
- **Production validation** at startup (prevents insecure defaults)
- Environment variable support (via Viper)
- TOML file support
- Tracing configuration (OpenTelemetry)

**Production Safeguards**:
- ✅ Validates authentication is enabled in production
- ✅ Prevents default API keys in production
- ✅ Prevents default database credentials in production
- ✅ Validates JWT secrets are changed from defaults

All configuration is validated before the application starts.

## Dependency Injection

The `internal/app/app.go` container wires all dependencies:
- Creates external clients (PostgreSQL, Meilisearch, Redis, NATS)
- Initializes repositories
- Creates application services
- Wires event handlers
- Configures health checks
- Initializes tracing provider

## Resilience Patterns

### Health Checks

The health service (`application/health/`) checks:
- ✅ PostgreSQL connectivity and latency
- ✅ Meilisearch availability
- ✅ Redis/Valkey connectivity
- ✅ NATS JetStream connection
- ✅ WebSocket hub status

Returns degraded status if any component fails, allowing graceful degradation.

### Graceful Shutdown

Comprehensive shutdown sequence (`internal/server/server.go`):
1. Stop accepting new HTTP connections
2. Close WebSocket hub (drain connections)
3. Close NATS connection
4. Close PostgreSQL pool
5. Close Redis client
6. Shutdown tracing provider

Timeout: 30 seconds with proper logging at each step.

### Circuit Breakers

Circuit breakers (`infrastructure/resilience/`) protect external service calls:
- Automatic failure detection
- State transitions (Closed → Open → Half-Open → Closed)
- Automatic recovery

## Observability

### OpenTelemetry Tracing

Full distributed tracing support (`internal/observability/tracing/`):
- OTLP HTTP exporter for Jaeger/Zipkin
- Echo middleware for automatic HTTP span creation
- Custom span attributes for telegrams, search, export, DB operations
- Configurable sampling ratio
- Context propagation across services

**Configuration**:
```toml
[tracing]
enabled = false
service_name = "caatsm-dashboard"
otlp_endpoint = "localhost:4318"
sampling_ratio = 1.0
```

### Structured Logging

All logs include:
- Correlation IDs (flow through all layers)
- Request IDs
- Layer information
- Context enrichment via `LoggerFromContext()`
- **PII redaction** (configurable)

### PII Redaction

PII redaction policy (`internal/observability/redaction.go`):
- ✅ Email redaction
- ✅ IP address redaction (IPv4 and IPv6)
- ✅ Phone number redaction
- ✅ Content truncation (configurable max length)
- ✅ Message ID masking
- ✅ Flight number partial redaction

**Configuration**:
```toml
[logger]
redact_pii = true  # Enable in production
```

### Correlation IDs

Correlation IDs are:
- Generated automatically for each request
- Passed through HTTP headers (`X-Correlation-ID`)
- Stored in context
- Included in all log entries and traces

## WebSocket Architecture

The WebSocket hub (`infrastructure/ws/`) provides:
- Connection management with per-IP limits
- **Backpressure control** (auto-disconnect slow clients)
- **Slow client detection** (message buffering)
- Connection limits (per IP: 5, global: 10,000)
- Metrics (connections, messages, drops)
- Graceful shutdown

## API Documentation

### OpenAPI 3.1 Specification

Complete API specification available at `api/openapi.yaml`:
- All endpoints documented
- Request/response schemas
- Query parameter validation rules
- Error response formats
- Authentication schemes

View with Swagger UI or OpenAPI editors.

## Testing Strategy

### Unit Tests
- Domain layer: Test business rules in isolation
- Application layer: Mock repositories, test use cases
- Infrastructure layer: Integration tests with testcontainers

### Integration Tests
- End-to-end flows with real dependencies
- Testcontainers for external services (PostgreSQL, Redis, Meilisearch, NATS)
- Time range validation tests
- Streaming export tests
- Health check tests

**Location**: `internal/testing/e2e_simple_test.go`

### Test Utilities
- Container helpers in `internal/testing/integration/`
- PostgreSQL, Redis, Meilisearch, NATS container wrappers
- Reusable test fixtures

## Key Design Decisions

1. **Clean Architecture**: Clear layer separation enables testability and maintainability
2. **Domain Events**: Decouple processing steps, allow independent failure
3. **Port Interfaces**: Application layer depends on abstractions, not implementations
4. **Streaming Export**: Chunked CSV export prevents OOM for large datasets
5. **Production Validation**: Config guards prevent insecure defaults in production
6. **Resilience First**: Health checks, graceful shutdown, circuit breakers
7. **Observability**: Distributed tracing, structured logging, PII redaction, metrics

## Production Features

### Security
- ✅ Rate limiting (10 req/sec, configurable)
- ✅ Input validation with whitelist-based filtering
- ✅ Production config validation
- ✅ PII redaction in logs
- ✅ SQL injection prevention

### Performance
- ✅ Streaming CSV export (handles 50k+ records)
- ✅ Chunked export (1000 records per chunk)
- ✅ Time range limits (max 90 days)
- ✅ WebSocket backpressure control
- ✅ Connection pooling with limits

### Reliability
- ✅ Comprehensive health checks
- ✅ Graceful shutdown (30s timeout)
- ✅ Circuit breakers for external services
- ✅ Automatic WebSocket reconnection
- ✅ Slow client detection and disconnection

### Observability
- ✅ OpenTelemetry distributed tracing
- ✅ Structured logging with correlation IDs
- ✅ Prometheus metrics
- ✅ Health check endpoints
- ✅ OpenAPI 3.1 specification

## Technology Stack

**Backend**:
- Go 1.21+
- Echo v4 (HTTP framework)
- PostgreSQL 14+ with TimescaleDB (time-series data)
- Meilisearch 1.5+ (full-text search)
- Valkey/Redis 7+ (cache and event bus)
- NATS 2.10+ with JetStream (message streaming)

**Frontend**:
- SvelteKit 2.x (UI framework)
- Deno 1.40+ or Node.js 20+ (runtime)
- UnoCSS (styling)
- TypeScript (type safety)

**Observability**:
- OpenTelemetry (distributed tracing)
- Zap (structured logging)
- Prometheus (metrics)

## Migration History

✅ **Phase 1 Complete**: Critical production hardening
- Time range validation (max 90 days)
- Streaming CSV export
- Production config guards

✅ **Phase 2 Complete**: Legacy code removal
- Deleted `internal/handlers/` (7 files)
- Deleted `internal/services/` (6 files)
- Deleted `internal/models/`
- Deleted `views/` (templ-based UI)
- Migrated WebSocket to transport layer

✅ **Phase 3 Complete**: Observability & testing
- OpenTelemetry tracing infrastructure
- E2E integration tests
- OpenAPI 3.1 specification
- PII redaction policy

✅ **Phase 4 Complete**: Final production readiness
- Enhanced health checks (NATS, WebSocket)
- Comprehensive graceful shutdown
- Documentation updates

**Result**: 100% clean architecture, zero legacy code, production-ready.
