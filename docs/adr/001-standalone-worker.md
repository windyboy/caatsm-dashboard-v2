# ADR-001: Message Ingestion Architecture

**Status**: Accepted  
**Date**: 2024-12-XX  
**Deciders**: Architecture Team  
**Tags**: architecture, ingestion, process-architecture

---

## Context

The system requires continuous ingestion of telegram messages from NATS JetStream. We need to decide whether to run ingestion as a separate process or within the HTTP server.

**Constraints**:
- NATS JetStream provides at-least-once delivery
- PostgreSQL with `ON CONFLICT DO NOTHING` ensures idempotency
- Real-time updates require sub-100ms latency
- Single-instance deployment (no horizontal scaling requirement)
- Small team prefers operational simplicity

**Decision Drivers**:
- Operational complexity vs. fault isolation
- Latency requirements for real-time updates
- Resource efficiency
- Development and maintenance effort

---

## Decision

**We will implement message ingestion as a background goroutine within the HTTP server process (`cmd/server`).**

### Architecture

```
NATS JetStream
    ↓
HTTP Server (single process)
    ├─ Ingestion Service (goroutine)
    ├─ HTTP Handlers
    ├─ WebSocket Hub
    └─ Indexer Service (goroutine)
```

**Key Points**:
- Ingestion runs as background goroutine in HTTP server
- In-memory channels for real-time updates (low latency)
- Redis Streams for reliable indexing queue
- Shared process resources (connection pools, memory)

---

## Consequences

### Positive

- **Operational Simplicity**: Single process, single binary, unified monitoring
- **Low Latency**: In-memory channels provide ~1ms latency (vs. ~5ms with Redis Pub/Sub)
- **Resource Efficiency**: Shared connection pools reduce memory by ~30-40%
- **Development Simplicity**: Easier local development and debugging

### Negative

- **Coupled Lifecycle**: HTTP server restarts stop ingestion temporarily
- **No Independent Scaling**: Cannot scale ingestion separately from query service
- **Resource Competition**: Ingestion and query share same process resources

### Mitigations

- **Deployment Impact**: NATS redelivery ensures no data loss; brief interruption acceptable
- **Resource Competition**: Use bounded concurrency, monitor resource usage
- **Future Scaling**: Architecture abstracted behind interfaces; can migrate to standalone worker if needed

---

## Alternatives Considered

### Alternative 1: Standalone Worker Process

**Description**: Separate process (`cmd/sync`) for NATS ingestion.

**Pros**:
- Independent failure isolation
- Independent scaling
- Zero-downtime deployments possible

**Cons**:
- Two processes to manage
- Higher latency (~5ms with Redis Pub/Sub)
- More operational overhead

**Rejected Because**: Operational overhead and latency penalty not justified for current requirements.

---

### Alternative 2: Microservices Architecture

**Description**: Separate services for ingestion, indexing, query, WebSocket.

**Pros**:
- Maximum fault isolation
- Independent scaling per service

**Cons**:
- High operational complexity (4+ services)
- Network latency (~10-20ms)
- Distributed system challenges

**Rejected Because**: Over-engineering for current scale and team size.

---

### Alternative 3: Event-Driven with Message Queue

**Description**: Message queue (RabbitMQ/Kafka) between ingestion and downstream services.

**Pros**:
- Full decoupling
- Message persistence and replay

**Cons**:
- Additional infrastructure dependency
- Higher latency (~10-50ms)
- More complex setup

**Rejected Because**: Unnecessary complexity; Redis Streams already provides queue functionality.

---

## Decision Criteria

| Criterion | Weight | Integrated | Standalone | Microservices |
|-----------|--------|------------|------------|---------------|
| Operational Simplicity | 30% | 10/10 | 6/10 | 2/10 |
| Latency | 25% | 10/10 | 7/10 | 4/10 |
| Fault Isolation | 20% | 7/10 | 10/10 | 10/10 |
| Resource Efficiency | 15% | 10/10 | 7/10 | 4/10 |
| Development Velocity | 10% | 10/10 | 7/10 | 3/10 |
| **Weighted Score** | - | **9.1/10** | **7.2/10** | **4.3/10** |

**Selected**: Integrated Architecture (highest weighted score)

---

## Notes

- Ingestion service runs as background goroutine started during HTTP server initialization
- Graceful shutdown ensures in-flight messages are processed
- NATS redelivery handles messages not ACKed during shutdown
- Can migrate to standalone worker if independent scaling becomes required

---

## References

- [Architecture Design Document](../architecture.md)
