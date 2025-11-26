# CAATSM Dashboard Architecture

## Overview

The CAATSM Dashboard follows **Clean Architecture** principles with clear separation of concerns across layers. The architecture is designed for maintainability, testability, and scalability.

## Architecture Layers

The system is organized into four main layers, with dependencies flowing inward:

```
┌─────────────────────────────────────────┐
│         Transport Layer                 │
│   HTTP/WebSocket Handlers               │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│      Application Layer                  │
│   Use Cases & Orchestration             │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Domain Layer                    │
│   Business Logic & Entities             │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│      Infrastructure Layer               │
│   External Concerns (DB, Cache, etc.)   │
└─────────────────────────────────────────┘
```

**Dependency Rule**: Dependencies point **inward**. Outer layers depend on inner layers, but inner layers never depend on outer layers.

## Layer Details

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Pure business logic with no external dependencies.

**Contains**:
- Domain entities (`telegram.go`)
- Business rules and validation (`validator.go`)
- Domain events (`events.go`)
- Domain filters (`filters.go`)
- Value objects and domain-specific types

**Principles**:
- No framework dependencies
- No infrastructure dependencies
- Self-contained business logic
- Immutable where possible

**Key Files**:
- `telegram.go` - Core domain entity with validation and normalization
- `events.go` - Domain events (TelegramReceived, TelegramValidated, etc.)
- `filters.go` - Domain-level search filters
- `errors.go` - Domain-specific error types

### 2. Application Layer (`internal/application/`)

**Purpose**: Use cases and orchestration. Implements business workflows.

**Contains**:
- Use case services (search, stats, export, health)
- Port interfaces (repository interfaces)
- Event handlers
- Event dispatcher

**Depends On**:
- Domain layer (uses domain entities and events)
- Port interfaces (depends on abstractions, not implementations)

**Key Packages**:
- `application/search/` - Search use cases
- `application/stats/` - Statistics use cases
- `application/export/` - Export use cases
- `application/health/` - Health checking and degradation policies
- `application/handlers/` - Event handlers (persistence, indexing)
- `application/events/` - Event dispatching
- `application/ports/` - Port interfaces (repository abstractions)

### 3. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: External concerns and implementations. Adapts external systems to application ports.

**Contains**:
- Repository implementations
- External service clients
- WebSocket hub
- Event bus implementations
- Resilience patterns (circuit breakers)

**Depends On**:
- Application ports (implements interfaces defined in application layer)
- Domain types (for data conversion)

**Key Packages**:
- `infrastructure/persistence/postgres/` - PostgreSQL repository
- `infrastructure/search/meilisearch/` - Meilisearch indexing
- `infrastructure/cache/valkey/` - Valkey/Redis caching
- `infrastructure/events/` - Event bus implementations
- `infrastructure/ws/` - WebSocket hub
- `infrastructure/resilience/` - Circuit breakers

### 4. Transport Layer (`internal/transport/`)

**Purpose**: HTTP/WebSocket handling, request/response formatting.

**Contains**:
- HTTP handlers
- WebSocket handlers
- Request parsing
- Response formatting

**Depends On**:
- Application services (calls use cases)
- No direct domain or infrastructure access

**Key Packages**:
- `transport/http/` - HTTP handlers
- `transport/ws/` - WebSocket handler wrapper

## Data Flow

### Request Flow

```
HTTP Request
    ↓
Transport Layer (HTTP Handler)
    ↓ Parse request, validate input
Application Layer (Use Case Service)
    ↓ Orchestrate business logic
Domain Layer (Domain Entity)
    ↓ Business rules, validation
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
- Validation at startup
- Environment variable support (via Viper)
- TOML file support

All configuration is validated before the application starts.

## Dependency Injection

The `internal/app/app.go` container wires all dependencies:
- Creates external clients (PostgreSQL, Meilisearch, Redis)
- Initializes repositories
- Creates application services
- Wires event handlers
- Configures health checks

## Resilience Patterns

### Health Checks

The health service (`application/health/`) checks:
- PostgreSQL connectivity and latency
- Meilisearch availability
- Redis/Valkey connectivity

### Degradation Policies

When components are degraded:
- **Meilisearch down**: Fallback to PostgreSQL search
- **PostgreSQL slow**: Use cached data
- **Redis down**: Continue without cache

Policies are configurable via `PolicyManager`.

### Circuit Breakers

Circuit breakers (`infrastructure/resilience/`) protect external service calls:
- Automatic failure detection
- State transitions (Closed → Open → Half-Open → Closed)
- Automatic recovery

## Observability

### Structured Logging

All logs include:
- Correlation IDs (flow through all layers)
- Request IDs
- Layer information
- Context enrichment via `LoggerFromContext()`

### Correlation IDs

Correlation IDs are:
- Generated automatically for each request
- Passed through HTTP headers (`X-Correlation-ID`)
- Stored in context
- Included in all log entries

## WebSocket Architecture

The WebSocket hub (`infrastructure/ws/`) provides:
- Connection management
- Backpressure control
- Slow client detection
- Connection limits (per IP, global)
- Metrics (connections, messages, drops)

## Testing Strategy

### Unit Tests
- Domain layer: Test business rules in isolation
- Application layer: Mock repositories, test use cases
- Infrastructure layer: Integration tests with testcontainers

### Integration Tests
- End-to-end flows with real dependencies
- Testcontainers for external services
- Mocks for expensive operations

## Migration Strategy

The architecture supports gradual migration:
- Old and new services coexist (`SearchService` vs `SearchServiceV2`)
- Type aliases maintain compatibility
- Handlers can be migrated incrementally

## Key Design Decisions

1. **Clean Architecture**: Clear layer separation enables testability and maintainability
2. **Domain Events**: Decouple processing steps, allow independent failure
3. **Port Interfaces**: Application layer depends on abstractions, not implementations
4. **Type Aliases**: Maintain backward compatibility during migration
5. **Resilience First**: Health checks, degradation policies, circuit breakers
6. **Observability**: Correlation IDs, structured logging, metrics

## Future Enhancements

- Distributed tracing (OpenTelemetry)
- More sophisticated degradation strategies
- Event sourcing for audit trails
- CQRS pattern for read/write separation
- API versioning

