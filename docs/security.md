# CAATSM Dashboard Security Guide

This guide summarizes the security expectations for the CAATSM Dashboard. It covers every part of the stack: Go services, SvelteKit frontend, background workers, and shared infrastructure. Use it as the baseline policy for development, staging, and production deployments.

---

## 1. Security Goals

- Protect aviation data from unauthorized access or tampering.
- Keep services resilient against denial-of-service and credential stuffing attacks.
- Ensure all secrets and encryption keys remain confidential.
- Provide visibility into incidents and support quick recovery.
- Meet regulatory expectations for TLS, audit trails, and data retention.

---

## 2. Environments

| Environment | Purpose | Security posture |
| --- | --- | --- |
| Development | Local laptops and test machines | Relaxed TLS, mock secrets allowed |
| Staging | Production-like validation | TLS recommended, real integrations, rotated secrets |
| Production | Live traffic | TLS required, hardened configs, monitored 24/7 |

Always set `CAATSM_ENVIRONMENT` to the correct value. Production-specific checks fail fast when that value is `production`.

---

## 3. Configuration and Secrets

### Required practice

- Never commit secrets to the repository.
- Prefer environment variables or a managed secret store.
- Rotate credentials on a fixed schedule (quarterly minimum).
- Use least-privilege accounts for databases and cloud resources.
- Log secret _usage_ but never log the actual secret values.


### Supported secret backends

- **Environment variables:** simplest approach; inject at runtime.

- **External secret managers (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager, etc.):** Fetch secrets with your deployment automation before startup—the application expects them as environment variables or mounted config and does not ship built-in adapters.


Ensure the application can unwrap secrets before calling `cfg.Validate()`. Validation should run after every refresh to catch misconfigurations early.

---

## 4. TLS and Network Security

- Enforce TLS 1.2+ (1.3 preferred) for all inbound and outbound traffic.
- In production, set `CAATSM_SERVER_TLS_ENABLED=true` and provide certificate/key paths.
- For upstream services:
  - Use `rediss://` or TLS-enabled Redis/Valkey.
  - Use HTTPS for Meilisearch and NATS TLS endpoints.
  - Verify certificates where possible.
- Limit inbound firewall rules to trusted networks and load balancers.
- Apply network policies (Kubernetes, security groups, etc.) to isolate app, data stores, and observability tooling.
- Disable legacy protocols (SSL, TLS 1.0/1.1).

---

## 5. Authentication and Authorization

### API Authentication

- `CAATSM_AUTH_ENABLE_BASIC`: use strong username/password pairs for operator endpoints.
- `CAATSM_AUTH_JWT_SECRET`: rotate secrets periodically; never use defaults in production.
- Consider integrating with SSO by implementing a new port if corporate policy requires it.

### Rate Limiting

- Delivery layer rate limits route traffic to 10 requests per second by default.
- For production, tune rate limits based on real load, but never disable them.
- Monitor HTTP 429 counts to spot abuse.

### WebSocket Controls

- Set `websocket` configuration to restrict allowed origins.
- Enforce per-IP and global connection limits (defaults: 5 per IP, 10,000 total).
- Idle or slow consumers are disconnected to prevent resource exhaustion.

---

## 6. Input Validation and Data Protection

- Domain layer rejects search queries that span more than 90 days.
- All sort fields are checked against a whitelist; do not concatenate SQL strings manually.
- Sanitize user input that flows into logs or metrics.
- Use parameterized queries via pgx to prevent SQL injection.

### Data at Rest

- Enable encryption for PostgreSQL storage (cloud-managed or OS-level).
- Use encrypted disks or storage classes for containers and virtual machines.
- Store exported CSV files in encrypted buckets if persisted beyond the streaming response.

### Data in Transit

- Use TLS for:
  - PostgreSQL (`sslmode=require`)
  - Meilisearch (HTTPS)
  - Redis/Valkey (`tls=true`)
  - NATS (`tls://` or secure websocket endpoints)

---

## 7. Observability and Logging

- Enable structured logging with correlation IDs.
- Mask or hash sensitive identifiers before logging.
- Keep application logs for at least 30 days; adjust retention to meet policy.
- Ship logs to a centralized store with access controls (e.g., Cloud Logging, ELK, Loki).

### Metrics

- `/metrics` exposes Prometheus-format counters, gauges, and histograms.
- Restrict access to `/metrics` via firewall or network policy.
- Monitor:
  - `telegrams_ingested_total`
  - `search_latency_seconds`
  - `http_requests_total`
  - `http_request_duration_seconds`
  - `websocket_connections`


### Tracing



- Tracing configuration flags exist, but exporter and instrumentation are not yet wired; keep `CAATSM_TRACING_ENABLED=false` until support is added.
- Plan for a secure endpoint (Jaeger, Tempo, etc.) before turning telemetry on.

- Capture spans for HTTP handlers, database operations, search calls, and streaming operations once instrumentation lands.


---

## 8. Dependency and Patch Management

- Pin Go modules in `go.mod` and confirm checksum integrity with `go.sum`.
- Keep `deno.lock` or `package-lock.json` under version control for frontend reproducibility.
- Monthly dependency update routine:
  1. Create an update branch.
  2. Run `go get -u ./...` and `go mod tidy`.
  3. Refresh frontend dependencies (`deno cache --reload` or `npm update`).
  4. Run all automated tests (`make test`, `make frontend-test-unit`, `make frontend-test`).
  5. Run security scanners.
- Security scanning:
  - `make security-scan` (runs gosec and Trivy).
  - Address high or critical issues immediately; document exceptions.
- Subscribe to security advisories for key dependencies (Go, Echo, pgx, Meilisearch, Redis, NATS).

---

## 9. Incident Response

1. **Detect:** Use alerts on error rates, latency spikes, 4xx/5xx anomalies, and security events.
2. **Triage:** Identify scope (service, data, user impact) and severity.
3. **Contain:** Disable affected endpoints, rotate secrets, block malicious traffic.
4. **Eradicate:** Patch vulnerabilities, remove malicious code, regenerate certificates.
5. **Recover:** Redeploy validated builds, restore data from secure backups if necessary.
6. **Review:** Document root cause, timeline, lessons learned, and required follow-up actions.

Keep a runbook with on-call contacts, escalation paths, logging locations, and monitoring dashboards.

---

## 10. Deployment Checklist

Perform this checklist before any production rollout:

- [ ] TLS certificates installed or managed at the load balancer.
- [ ] `CAATSM_ENVIRONMENT=production`.
- [ ] All secrets sourced from secure backend or environment variables; no defaults remain.
- [ ] `config/config.prod.toml` reviewed for hostnames, timeouts, and rate limits.
- [ ] Database, NATS, Redis, and Meilisearch credentials use least privilege.
- [ ] `/api/health` integrated with liveness/readiness probes.
- [ ] `/metrics` reachable only from monitoring network.
- [ ] `CAATSM_TRACING_ENABLED` set appropriately (keep false until instrumentation is delivered) and telemetry endpoint ready for future rollout.
- [ ] System and application logs forwarded to central storage.
- [ ] Disaster recovery plan verified (database backups, restore tested within last quarter).

---

## 11. Developer Checklist

When implementing features:

- [ ] Respect Clean Architecture boundaries; do not bypass the domain validators.
- [ ] Handle user input defensively (length, pattern, and range checks).
- [ ] Ensure errors never leak sensitive data to clients.
- [ ] Add tests for new validation logic and security-sensitive code.
- [ ] Update docs when introducing new configuration flags or secrets.
- [ ] Run the full test suite and linting before opening a pull request.
- [ ] Call out any security implications in the PR description.

---

## 12. Continuous Improvement

- Conduct threat modeling reviews at least twice a year.
- Schedule regular penetration tests or red team exercises when possible.
- Track security metrics: mean time to detect (MTTD), mean time to resolve (MTTR), open vulnerabilities.
- Provide ongoing security training for engineers and operators.
- Review this document after every incident or major architectural change.

Following these guidelines keeps the CAATSM Dashboard secure, resilient, and trustworthy for the aviation stakeholders who rely on it.