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
```

becomes

```
CAATSM_SERVER_HOST=0.0.0.0
CAATSM_SERVER_PORT=3002
```

### Loading order

1. `.env.local` (if present)
2. `.env` (optional shared defaults)
3. Explicit environment variables supplied by your shell or orchestrator
4. TOML files
5. In-code defaults

### Common variables

| Variable | Description |
| --- | --- |
| `CAATSM_DATABASE_DSN` | PostgreSQL DSN with credentials and TLS flags |
| `CAATSM_MEILISEARCH_HOST` / `API_KEY` | Meilisearch endpoint and key |
| `CAATSM_REDIS_ADDR` / `PASSWORD` | Valkey/Redis connection info |
| `CAATSM_NATS_URL` | NATS JetStream endpoint |
| `CAATSM_SERVER_TLS_ENABLED` | Enable HTTPS (must be `true` in production) |
| `CAATSM_AUTH_ENABLE_BASIC` and credentials | Basic auth for admin endpoints |
| `CAATSM_AUTH_JWT_SECRET` | Secret key for JWT validation |
| `CAATSM_TRACING_ENABLED` | Toggles OpenTelemetry exporter |
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
- Redis, Meilisearch, and NATS URLs must use TLS in production.

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

The infrastructure layer ships optional adapters for AWS Secrets Manager and HashiCorp Vault. They expose helper methods to fetch:

- PostgreSQL DSN
- JWT secret
- Meilisearch API key
- Other strings you register

To enable one of these providers:

1. Configure the adapter with region/address, path, and credentials.
2. Load secrets at startup.
3. Set the resulting values on the config struct before calling `Validate()`.

Never log the secrets; the loader only prints keys that fail validation.

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
3. Provide real secrets through environment variables or the secret manager adapter.
4. Ensure any orchestrator health probes call `/api/health`.
5. Turn on tracing and metrics exporters as required.

---

## 8. Common validation errors and fixes

| Error message | Root cause | Fix |
| --- | --- | --- |
| `server.port must be greater than 0` | Port value missing or zero | Set `CAATSM_SERVER_PORT` or update TOML |
| `database.dsn is required` | DSN omitted | Supply `CAATSM_DATABASE_DSN` |
| `production: TLS must be enabled` | Running without TLS in production | Set `CAATSM_SERVER_TLS_ENABLED=true` and provide cert/key paths |
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
- [ ] Tracing exporter (if enabled) can reach the OTLP endpoint.

---

## 10. Keeping configuration healthy

- Run `make lint` and `make test` after changing config structs to catch compilation errors.
- Add unit tests in `config/config_test.go` when you extend validation rules.
- Document new environment variables in `env.example` so teammates know what to set.
- Rotate API keys and secrets regularly; the adapters support hot reload if you implement it.
- Review validation errors in logs; the loader prints actionable hints.

Following these guidelines keeps configuration straightforward and secure across development, staging, and production environments.