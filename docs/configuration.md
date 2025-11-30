# Configuration Guide

This document explains how application settings are loaded, how to override them, and which checks keep production deployments safe. Everything below uses plain English and focuses on the steps you will take most often.

---

## 1. What the configuration does

The Go process reads settings from three places:

1. **Defaults in code** – safe starting values that make local development easy.
2. **TOML files in `config/`** – structured overrides that go into version control.
3. **Environment variables with the `CAATSM_` prefix** – the final authority for production.

The loader merges them in that order, so environment variables always win.

---

## 2. Key files

| File | Purpose | Typical use |
| --- | --- | --- |
| `config/config.toml` | Shared defaults | Checked in; rarely edited |
| `config/config.local.toml` | Local overrides | Points services to `localhost` |
| `env.example` | Template | Shows the minimum env vars |
| `env.local.example` | Local template | Copy to `.env.local` when you set up a dev machine |

When you run `task dev:config`, the helper copies `env.local.example` to `.env.local`.

---

## 3. Environment variables

All variables begin with `CAATSM_`. Nested TOML keys map to upper-case names with underscores:

```
[server]
host = "0.0.0.0"
port = 3002
allow_bind_all = true
```

becomes

```
CAATSM_SERVER_HOST=0.0.0.0
CAATSM_SERVER_PORT=3002
CAATSM_SERVER_ALLOW_BIND_ALL=true
```

### Loading order

Configuration sources are merged in the following order (lowest to highest precedence). Later entries override earlier ones, so explicit environment variables always win (consistent with the precedence described in [section 1](#1-what-the-configuration-does)):

1. In-code defaults (lowest precedence)
2. TOML files
3. `.env` (optional shared defaults)
4. `.env.local` (if present)
5. Explicit environment variables supplied by your shell or orchestrator (highest precedence)

### Common variables

| Variable | Description |
| --- | --- |
| `CAATSM_DATABASE_DSN` | PostgreSQL DSN with credentials and TLS flags |
| `CAATSM_MEILISEARCH_HOST` / `API_KEY` | Meilisearch endpoint and key |
| `CAATSM_REDIS_ADDR` / `PASSWORD` | Valkey/Redis connection info |
| `CAATSM_NATS_URL` | NATS JetStream endpoint |
| `CAATSM_SERVER_TLS_ENABLED` | Enable HTTPS (must be `true` in production) |
| `CAATSM_SERVER_ALLOW_BIND_ALL` | Allow binding to `0.0.0.0` in production (required for containerized/multi-interface environments) |
| `CAATSM_AUTH_ENABLE_BASIC` and credentials | Basic auth for admin endpoints |
| `CAATSM_AUTH_JWT_SECRET` | Secret key for JWT validation |
| `CAATSM_TRACING_ENABLED` | Tracing feature flag (instrumentation not yet wired) |
| `CAATSM_ENVIRONMENT` | `development`, `staging`, or `production` |

---

## 4. How the loader validates settings

The loader in `config/loader.go` performs these steps:

1. Load environment files.
2. Parse the selected TOML file (or fall back to `config/config.toml`).
3. Unmarshal into strongly typed structs.
4. Run `Validate()` before returning.

### Validation rules

- `server.port` must be greater than `0`.
- `database.dsn` cannot be empty.
- TLS must be enabled when `environment == "production"`.
- Default usernames, passwords, and API keys are rejected in production.
- At least one auth mechanism (basic auth or JWT) must be active in production.
- Server host cannot be `localhost` or `127.0.0.1` in production.
- Server host cannot be `0.0.0.0` in production unless `server.allow_bind_all` is set to `true` (required for containerized/multi-interface environments).
- Redis, Meilisearch, and NATS URLs must use TLS in production.
- `tracing.enabled` must be true in production; the flag is enforced even though exporter wiring is pending.

If any rule fails, the loader returns an error and the binary exits during startup. That prevents accidental insecure deployments.

---

## 5. Production safeguards

Checklist for live environments:

| Item | Why it matters |
| --- | --- |
| TLS enabled (`CAATSM_SERVER_TLS_ENABLED=true`) | Protects API traffic |
| Strong database DSN | Avoids default credentials and enforces SSL |
| Non-default secrets for Meilisearch, Redis, JWT | Stops attackers from guessing keys |
| Auth enabled | Guards admin endpoints |
| Rate limiting configured | Prevents abuse (defaults to 10 req/sec) |
| Health checks wired | Allows orchestrators to detect degraded services |
| Allowed origins defined | Locks down CORS and WebSocket origins |

---

## 6. Secret management

Avoid placing secrets in code or committed files.

### Option A: Environment variables

Inject secure values at runtime using your deployment tool (Kubernetes, Docker Compose, systemd, etc.). This is the simplest approach for staging and small installations.

### Option B: External secret store


Use your preferred secret manager (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager, etc.) to materialize secrets before the process starts. The repository does not ship built-in adapters, so retrieval should happen in your deployment tooling or via a custom integration you maintain.

Recommended workflow:

1. Fetch secrets through automation (init container, sidecar, CI/CD task, etc.).
2. Export them as environment variables (preferred) or write them to a mounted config file before starting the binary.
3. Run the binary so `Validate()` executes with the real values.

Never log the secrets; only surface which keys failed validation.


---

## 7. Example setups

### Local development

1. Copy `.env.local.example` to `.env.local`.
2. Update `CAATSM_DATABASE_DSN` if you changed the port or password.
3. Leave TLS disabled; the loader allows that when `CAATSM_ENVIRONMENT=development`.
4. Run `make dev-up` to start Postgres, Redis, Meilisearch, and NATS.

### Staging

1. Duplicate `config/config.toml` to `config/config.staging.toml`.
2. Adjust hostnames to point at staging infrastructure.
3. Inject secrets via environment variables or your secret store.
4. Set `CAATSM_ENVIRONMENT=staging` to enable most production checks without forcing TLS.
5. Run `task dev:config` to regenerate `.env.local` for staging defaults as needed.

### Production

1. Create `config/config.prod.toml` with secure defaults: TLS cert paths, timeouts, rate limits.
2. Set `CAATSM_ENVIRONMENT=production`.
3. Provide real secrets through environment variables or your secret-management automation.
4. Ensure any orchestrator health probes call `/api/health`.
5. Set `tracing.enabled=true` to satisfy validation (exporter is currently inert) and turn on metrics exporters as required.

---

## 8. Common validation errors and fixes

| Error message | Root cause | Fix |
| --- | --- | --- |
| `server.port must be greater than 0` | Port value missing or zero | Set `CAATSM_SERVER_PORT` or update TOML |
| `database.dsn is required` | DSN omitted | Supply `CAATSM_DATABASE_DSN` |
| `production: TLS must be enabled` | Running without TLS in production | Set `CAATSM_SERVER_TLS_ENABLED=true` and provide cert/key paths |
| `production: server host cannot be localhost or 127.0.0.1` | Using localhost addresses in production | Use a specific IP address or hostname |
| `production: server host cannot be 0.0.0.0 unless server.allow_bind_all is set to true` | Using `0.0.0.0` without explicit permission | Set `CAATSM_SERVER_ALLOW_BIND_ALL=true` for containerized/multi-interface environments |
| `production: authentication required` | Auth disabled in production | Enable basic auth or JWT config |
| `meilisearch.api_key cannot be default` | Still using `masterKey` | Rotate the key and update `CAATSM_MEILISEARCH_API_KEY` |
| `redis: TLS required in production` | Redis connection not secured | Switch to `rediss://` style address or enable TLS flags |

---

## 9. Deployment checklist

Before shipping a new environment, confirm the following:

- [ ] `config/config.<env>.toml` exists and matches host names.
- [ ] All required `CAATSM_` environment variables are set by the orchestrator.
- [ ] TLS certificates or load balancer termination are ready.
- [ ] `.env` files are excluded from version control.
- [ ] Secret store access policies allow the app to read production secrets.
- [ ] Health checks point to `/api/health`.
- [ ] Metrics endpoint (`/metrics`) is reachable only by your monitoring stack.
- [ ] `tracing.enabled` is set to true (required by validation) and the OTLP endpoint is reachable if/when instrumentation ships.

### NATS 2.12.2 Upgrade Verification (Staging → Production)

When upgrading to NATS 2.12.2 (`nats:2.12.2-alpine`), perform these checks in staging before promoting to production:

**Pre-upgrade (Staging):**
- [ ] Snapshot/backup JetStream metadata before upgrading (use `nats stream backup` or volume snapshots).
- [ ] Test the NATS 2.11 image in a staging cluster with representative JetStream workloads matching production patterns.
- [ ] Validate there are no server/cluster/gateway names with spaces (NATS 2.11 enforces stricter naming).
- [ ] Monitor logs for warnings during staging deployment, especially:
  - `JSStreamTooRequests` errors (indicates buffer limits exceeded)
  - Exit code changes during shutdown
  - Stream state rebuild/rescan messages

**JetStream Configuration Adjustments:**
- [ ] Check per-stream ingest limits: if workloads hit the new 128MB/10k message buffers, either:
  - Adjust stream `MaxMsgs` or `MaxBytes` limits to prevent `JSStreamTooRequests` errors, or
  - Enable `js-meta-only` mode if metadata-only operations are sufficient.
- [ ] Verify stream configurations in `internal/infrastructure/streaming/nats.go` are compatible with 2.12.2 limits.

**Orchestration & Shutdown:**
- [ ] Ensure orchestration scripts (Docker Compose, Kubernetes, systemd) tolerate changed exit codes from NATS 2.12.2.
- [ ] Test graceful shutdown procedures: verify connections drain properly and JetStream state persists.

**Post-upgrade Validation (Staging):**
- [ ] Confirm all JetStream streams and consumers are operational.
- [ ] Verify message processing throughput matches pre-upgrade baselines.
- [ ] Check for any stream state rebuild/rescan operations (may occur after downgrade scenarios).
- [ ] Monitor for 24-48 hours to catch any edge cases or performance regressions.

**Production Promotion:**
- [ ] Only promote to production after successful staging validation (all checks above passing).
- [ ] Schedule upgrade during low-traffic window with rollback plan ready.
- [ ] Have JetStream metadata backup available for immediate rollback if needed.

---

## 10. Keeping configuration healthy

- Run `make lint` and `make test` after changing config structs to catch compilation errors.
- Add unit tests in `config/config_test.go` when you extend validation rules.
- Document new environment variables in `env.example` so teammates know what to set.
- Rotate API keys and secrets regularly; the adapters support hot reload if you implement it.
- Review validation errors in logs; the loader prints actionable hints.

Following these guidelines keeps configuration straightforward and secure across development, staging, and production environments.