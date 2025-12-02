# Monitoring & Observability

## Metrics Endpoints

- **Application metrics**: `GET /metrics` (Prometheus format)
- **Health check**: `GET /api/health`

## Key Metrics

### Ingestion Pipeline

- `telegrams_ingested_total` - Total messages processed
- `ingest_errors_total` - Failed ingestion attempts
- `ingest_latency_seconds` - P50, P95, P99 latency histograms

### Real-time WebSocket

- `ws_active_connections` - Current active connections
- `ws_messages_broadcast_total` - Messages sent to clients
- `ws_messages_dropped_total` - Messages dropped due to slow clients
- `ws_slow_clients_disconnected_total` - Clients auto-disconnected

### Indexer

- `indexer_queue_depth` - Pending messages in Redis Stream
- `indexer_processed_total` - Messages indexed to Meilisearch
- `indexer_errors_total` - Indexing failures

### Storage

- `db_hypertable_size_bytes` - PostgreSQL telegrams table size
- `db_compressed_chunks_count` - Number of compressed TimescaleDB chunks

## Service Level Objectives (SLOs)

### Ingestion SLO

- **Target**: 99% of messages stored in PostgreSQL within 5 seconds of NATS receipt
- **Measurement**: `histogram_quantile(0.99, ingest_latency_seconds) < 5`

### API Latency SLO

- **Target**: 95% of `/api/search` requests complete under 200ms
- **Measurement**: `histogram_quantile(0.95, http_request_duration_seconds{endpoint="/api/search"}) < 0.2`

### Indexer Backlog SLO

- **Target**: Indexer backlog stays under 10,000 messages for 95% of time
- **Measurement**: `indexer_queue_depth < 10000`

## Recommended Alerts

### Critical Alerts

**IngestionDown**

```promql
rate(telegrams_ingested_total[5m]) == 0
```

Trigger: No messages ingested for 5 minutes

Action: Check NATS connectivity, inspect ingestion service logs

**DatabaseDown**

```promql
up{job="caatsm-api"} == 0
```

Trigger: Health check failing

Action: Check PostgreSQL connectivity, verify connection pool

**IndexerBacklogHigh**

```promql
indexer_queue_depth > 50000
```

Trigger: Redis Stream backlog exceeds 50k messages

Action: Scale indexer workers, check Meilisearch health

### Warning Alerts

**HighIngestionLatency**

```promql
histogram_quantile(0.95, ingest_latency_seconds) > 10
```

Trigger: P95 ingestion latency > 10s for 5 minutes

Action: Check database performance, review slow queries

**WebSocketDropRate**

```promql
rate(ws_messages_dropped_total[5m]) / rate(ws_messages_broadcast_total[5m]) > 0.1
```

Trigger: >10% of WebSocket messages dropped

Action: Investigate slow clients, consider buffer size tuning

**SearchLatencyHigh**

```promql
histogram_quantile(0.95, http_request_duration_seconds{endpoint="/api/search"}) > 1
```

Trigger: P95 search latency > 1s

Action: Check Meilisearch index health, verify cache hit rates

## Dashboards

### Recommended Grafana Panels

1. **Ingestion Overview**
   - Messages/sec ingested (rate)
   - P95 latency
   - Error rate

2. **Real-time WebSocket**
   - Active connections (gauge)
   - Messages broadcast/sec
   - Drop rate %

3. **Indexer Health**
   - Queue depth (gauge)
   - Indexing rate
   - Error count

4. **Storage**
   - Database size growth
   - Compressed vs uncompressed chunks
   - Oldest message timestamp (for retention verification)

## Log Correlation

All logs include `request_id` field for tracing requests across services.

Example query (for structured JSON logs):

```bash
jq 'select(.request_id == "abc123")' app.log
```

## Example Prometheus Recording Rules

```yaml
groups:
  - name: caatsm_slos
    interval: 30s
    rules:
      - record: caatsm:ingestion_success_rate:5m
        expr: |
          rate(telegrams_ingested_total[5m])
          / (rate(telegrams_ingested_total[5m]) + rate(ingest_errors_total[5m]))

      - record: caatsm:api_latency_p95:5m
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket{endpoint="/api/search"}[5m])
          )

      - record: caatsm:ws_drop_rate:5m
        expr: |
          rate(ws_messages_dropped_total[5m])
          / rate(ws_messages_broadcast_total[5m])
```

## Example Alert Rules

```yaml
groups:
  - name: caatsm_alerts
    rules:
      - alert: CaatsmIngestionDown
        expr: rate(telegrams_ingested_total[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "CAATSM ingestion has stopped"
          description: "No messages have been ingested in the last 5 minutes"

      - alert: CaatsmHighLatency
        expr: caatsm:api_latency_p95:5m > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CAATSM API latency is high"
          description: "P95 latency for /api/search is {{ $value }}s (threshold: 1s)"

      - alert: CaatsmIndexerBacklog
        expr: indexer_queue_depth > 50000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "CAATSM indexer backlog is high"
          description: "Redis Stream has {{ $value }} pending messages"
```

## Monitoring Best Practices

1. **Set up alerts for all critical SLOs** - Don't wait for users to report issues
2. **Monitor error budgets** - Track how much downtime/errors you have left in your SLO window
3. **Dashboard for on-call** - Create a single dashboard with all critical metrics
4. **Runbooks for alerts** - Document what each alert means and how to fix it
5. **Test alerts** - Periodically trigger alerts to ensure notification channels work
6. **Review and tune** - Adjust thresholds based on actual production behavior

## Health Check Implementation

The `/api/health` endpoint checks:

- PostgreSQL connectivity
- Redis connectivity
- Meilisearch connectivity
- NATS JetStream status

Example response:

```json
{
  "status": "healthy",
  "timestamp": "2025-12-02T10:00:00Z",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "meilisearch": "ok",
    "nats": "ok"
  }
}
```

Use this endpoint for:
- Kubernetes liveness probes
- Load balancer health checks
- Monitoring system uptime checks

