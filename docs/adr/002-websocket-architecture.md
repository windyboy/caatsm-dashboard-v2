# ADR-002: Real-time Communication Protocol

**Status**: Accepted  
**Date**: 2024-12-XX  
**Deciders**: Architecture Team  
**Tags**: architecture, websocket, real-time, communication

---

## Context

The system requires real-time delivery of telegram messages to web clients. We need to select a communication protocol and architecture.

**Requirements**:
- Sub-100ms latency for real-time updates
- Bidirectional communication (client commands)
- Support 100+ concurrent connections
- Browser compatibility (Chrome, Firefox, Safari, Edge)
- Minimal infrastructure dependencies

**Constraints**:
- Single-instance deployment (no horizontal scaling)
- Message volume: ~10-100 messages/second
- Message size: < 10KB per message
- Long-lived connections (minutes to hours)

**Decision Drivers**:
- Latency requirements
- Browser compatibility
- Operational simplicity
- Development velocity

---

## Decision

**We will use WebSocket protocol with an in-memory hub architecture.**

### Architecture

```
Ingestion Service
    ↓ (in-memory channel)
Real-time Service
    ↓
WebSocket Hub (in-memory)
    ↓
WebSocket Clients
```

**Key Points**:
- WebSocket for bidirectional communication
- In-memory hub within HTTP server process
- Best-effort delivery (no persistence)
- Client-side resync on reconnect

---

## Consequences

### Positive

- **Low Latency**: ~1-5ms from ingestion to client (in-memory hub)
- **Bidirectional**: Full-duplex communication
- **No External Dependencies**: In-memory hub requires no additional infrastructure
- **Browser Native**: Built-in browser API support

### Negative

- **Best-Effort Only**: No message persistence or replay
- **Single Instance**: Doesn't scale horizontally
- **Connection State Loss**: All connections lost on server restart
- **Memory Overhead**: Each connection consumes memory

### Mitigations

- **Best-Effort**: Clients resync via HTTP API on reconnect
- **Single Instance**: Acceptable for current requirements
- **Connection Loss**: Automatic reconnection with exponential backoff
- **Memory Usage**: Connection limits and timeouts prevent unbounded growth

---

## Alternatives Considered

### Alternative 1: Server-Sent Events (SSE)

**Description**: Unidirectional HTTP streaming.

**Pros**: Simpler, automatic reconnection, works through proxies

**Cons**: Unidirectional only, browser connection limits, less efficient

**Rejected Because**: Unidirectional limitation and higher latency.

---

### Alternative 2: HTTP Long Polling

**Description**: Client polls server, server holds request until data available.

**Pros**: Works with all HTTP infrastructure, simple

**Cons**: Higher latency (1-5 seconds), more overhead, not truly real-time

**Rejected Because**: Latency too high for real-time requirements.

---

### Alternative 3: Redis Pub/Sub with WebSocket

**Description**: Redis Pub/Sub distributes messages to multiple server instances.

**Pros**: Enables horizontal scaling, decoupled architecture

**Cons**: Adds latency (~5-10ms), requires Redis, more complex

**Rejected Because**: Not needed for single-instance deployment; can be added later if scaling required.

---

### Alternative 4: gRPC-Web Streaming

**Description**: gRPC bidirectional streaming over HTTP/2.

**Pros**: Efficient binary protocol, type-safe

**Cons**: Requires gRPC-Web proxy, more complex, less web-friendly

**Rejected Because**: Higher complexity than WebSocket; overkill for requirements.

---

## Decision Criteria

| Criterion | Weight | WebSocket | SSE | Long Polling | Redis Pub/Sub |
|-----------|--------|-----------|-----|--------------|---------------|
| Latency | 30% | 10/10 | 7/10 | 2/10 | 8/10 |
| Bidirectional | 20% | 10/10 | 0/10 | 5/10 | 10/10 |
| Browser Support | 15% | 10/10 | 9/10 | 10/10 | 10/10 |
| Operational Simplicity | 20% | 10/10 | 9/10 | 9/10 | 7/10 |
| Scalability | 10% | 7/10 | 7/10 | 5/10 | 10/10 |
| Development Velocity | 5% | 10/10 | 9/10 | 9/10 | 7/10 |
| **Weighted Score** | - | **9.7/10** | **7.0/10** | **5.1/10** | **8.4/10** |

**Selected**: WebSocket with In-Memory Hub (highest weighted score)

---

## Notes

- WebSocket hub manages connections, broadcasting, and backpressure
- Keep-alive: Ping frames every 45 seconds
- ReadDeadline: 60 seconds (refreshed on message)
- Bounded buffers per client prevent blocking
- Clients automatically reconnect and resync via HTTP API
- Can migrate to Redis Pub/Sub if horizontal scaling is required

---

## References

- [Architecture Design Document](../architecture.md)
- [ADR-001: Message Ingestion Architecture](./001-standalone-worker.md)
