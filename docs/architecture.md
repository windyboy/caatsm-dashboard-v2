# CAATSM Dashboard - Architecture Design

This document describes the high-level architecture for a real-time telegram monitoring and querying system.

---

## 1. System Overview

### Core Capabilities

- **Real-time Monitoring**: Receive and display parsed telegram messages in real-time
- **Query & Search**: Full-text search, filtering, and time-range queries
- **Analytics**: Aggregated metrics and dashboard data

### Key Principles

- **Direct Storage**: Messages stored with minimal validation, no domain-level enrichment
- **Single Process**: All services (HTTP, ingestion, indexer, realtime) run as goroutines within one process
- **Clean Architecture**: Layered architecture with unidirectional dependencies

---

## 2. Architecture Overview

```
NATS JetStream (Message Source)
    ↓
HTTP Server (Go + Echo)
    ├─ Query API: GET /api/search, /api/stats
    ├─ WebSocket: WS /ws
    └─ Background Tasks:
       • NATS message ingestion
       • Async index updates
    ↓
Application Services
    ↓
Data Storage
    ├─ PostgreSQL/TimescaleDB (Primary)
    ├─ Meilisearch (Full-text Search)
    └─ Redis (Cache, Pub/Sub, Streams)
```

### Architectural Principles

- **PostgreSQL is canonical**: All query correctness validated against PostgreSQL
- **Eventually consistent**: Search index and realtime views are eventually consistent
- **Failure isolation**: Search/realtime failures don't block ingestion
- **Idempotency**: Duplicate messages handled via `ON CONFLICT DO NOTHING`

---

## 3. Core Components

### 3.1 Application Services

1. **Ingestion Service**
   - Consumes from NATS JetStream
   - Validates and stores to PostgreSQL
   - Publishes to Redis Pub/Sub for realtime
   - Pushes to Redis Streams for indexing

2. **Query Service**
   - Executes search queries
   - Coordinates between PostgreSQL and Meilisearch
   - Manages query caching

3. **Statistics Service**
   - Calculates aggregated metrics
   - Manages statistics caching

4. **Real-time Service**
   - Broadcasts messages to WebSocket clients
   - Best-effort delivery (clients resync on reconnect)

5. **Indexer Service**
   - Consumes from Redis Streams
   - Updates Meilisearch index
   - Handles pending message recovery


### 3.2 Data Storage

**PostgreSQL/TimescaleDB**
- Primary storage (source of truth)
- Time-series optimization with hypertables
- Compression for chunks older than 7 days
- 180-day retention policy

**Meilisearch**
- Full-text search
- Eventually consistent with PostgreSQL

**Redis**
- **Pub/Sub**: Realtime event broadcasting
- **Streams**: Reliable job queue for indexing
- **Cache**: Query results and statistics

---

## 4. Data Flow

### 4.1 Message Ingestion

```
NATS JetStream
    ↓
Ingestion Service
    ↓
PostgreSQL (ON CONFLICT DO NOTHING)
    ├─→ ACK to NATS
    ├─→ Redis Pub/Sub (realtime)
    └─→ Redis Streams (indexing)
        ↓
    Indexer Service → Meilisearch
```

### 4.2 Query Flow

```
Client Request
    ↓
Query Service
    ├─→ Check Redis cache
    └─→ If miss:
        ├─→ Meilisearch (full-text)
        └─→ PostgreSQL (canonical)
            ↓
        Merge & cache → Return
```

### 4.3 Real-time Updates

```
New Message → Redis Pub/Sub → Realtime Service → WebSocket Hub → Clients
```

---

## 5. Package Structure

```
cmd/server/          # HTTP Server entry point
internal/
├── app/             # Port interfaces and dependency container
├── domain/           # Business entities and validation
├── service/          # Application services
├── delivery/         # HTTP/WebSocket handlers
├── repository/       # PostgreSQL repository
└── infrastructure/   # Adapters (NATS, Redis, Meilisearch, WebSocket)
```

**Dependency Flow**: Delivery → Services → Domain ← Infrastructure (implements ports)

---

## 6. Deployment Topology

### Current: Single Process (Embedded Services)

All services run as goroutines within the main HTTP server process (`cmd/server`):

- **NATS ingestion service** - Background goroutine consuming from JetStream
- **Redis Streams indexer service** - Background goroutine updating Meilisearch
- **Real-time broadcast service** - Background goroutine managing WebSocket broadcasts
- **HTTP API handlers** - Main Echo server handling REST endpoints
- **WebSocket hub** - Integrated connection manager with backpressure control

See `internal/server/server.go` lines 289-312 for service startup code.

**Benefits**:
- Simple deployment (single binary)
- Shared connection pools and caches
- Easier local development
- Lower operational complexity

**Trade-offs**:
- All workloads share same process resources
- Scaling requires scaling entire stack
- Failure in one service can affect others

### Future: Multi-Process Option

Services are architecturally independent and can be split into separate binaries for horizontal scaling:

```
┌─────────────────┐
│  cmd/server     │  HTTP API only
│  (API endpoints)│
└─────────────────┘

┌─────────────────┐
│ cmd/ingestion   │  Dedicated NATS consumer
│ (NATS → DB)     │  (can run multiple instances)
└─────────────────┘

┌─────────────────┐
│  cmd/indexer    │  Dedicated Meilisearch indexer
│ (Redis → Meili) │  (can run multiple workers)
└─────────────────┘
```

**Migration steps** (when needed):
1. Create new `cmd/ingestion/main.go` and `cmd/indexer/main.go` entrypoints
2. Extract service initialization logic from `internal/server/server.go`
3. Update deployment manifests (Docker Compose, Kubernetes)
4. Configure separate scaling policies per service type

---

## 7. Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go | Server implementation |
| Web Framework | Echo | HTTP server and routing |
| Message Broker | NATS JetStream | Message ingestion |
| Database | PostgreSQL + TimescaleDB | Primary storage |
| Search | Meilisearch | Full-text search |
| Cache/Broker | Redis | Pub/Sub, Streams, Cache |
| Frontend | SvelteKit + @melt-ui/svelte | Web dashboard with accessible UI components |
| Observability | Prometheus + Zap | Metrics and logging |

**Frontend UI Components**: The frontend uses @melt-ui/svelte for accessible, headless UI components. All components follow the builder pattern, providing built-in ARIA support, keyboard navigation, and focus management while maintaining full style control through UnoCSS and custom CSS.

---

## 8. Design Principles

- **Stateless**: Server instances share no memory state
- **Idempotency**: Duplicate messages handled via unique constraints
- **Source of Truth**: PostgreSQL is authoritative
- **Failure Isolation**: Search/realtime failures don't block ingestion
- **Cache-First**: Query results cached to reduce load

---

## 9. Limitations

- **No guaranteed realtime delivery**: WebSocket is best-effort
- **No long-term storage**: Data beyond 180 days is dropped
- **No ordering guarantees**: Accepts eventual consistency
- **No strict latency guarantees**: Optimized for read-heavy workload

---

## 10. Observability

**Key Metrics**:
- Ingestion: `ingest_total`, `ingest_errors`, `ingest_latency_p95`
- Real-time: `ws_active_connections`, `ws_messages_broadcast_total`
- Indexing: `indexer_queue_depth`, `indexer_processed_total`
- Storage: `db_hypertable_size`, `db_compressed_chunks_count`

**Logging**: Structured JSON logs with request ID propagation
