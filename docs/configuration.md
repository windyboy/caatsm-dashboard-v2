# Configuration

Configuration is loaded from (in order):
1. In-code defaults
2. TOML files in `config/`
3. Environment variables with `CAATSM_` prefix (highest precedence)

## Key Environment Variables

| Variable | Description |
|----------|-------------|
| `CAATSM_DATABASE_DSN` | PostgreSQL connection string |
| `CAATSM_MEILISEARCH_HOST` / `API_KEY` | Meilisearch endpoint and key |
| `CAATSM_REDIS_ADDR` / `PASSWORD` | Redis/Valkey connection |
| `CAATSM_NATS_URL` | NATS JetStream endpoint |
| `CAATSM_SERVER_TLS_ENABLED` | Enable HTTPS (required in production) |
| `CAATSM_AUTH_ENABLE_BASIC` | Basic auth for admin endpoints |
| `CAATSM_AUTH_JWT_SECRET` | JWT secret key |
| `CAATSM_ENVIRONMENT` | `development`, `staging`, or `production` |

## Production Requirements

- TLS must be enabled
- Authentication required (basic auth or JWT)
- Non-default secrets
- All external connections use TLS

## Local Development

1. Copy `env.local.example` to `.env.local`
2. Update `CAATSM_DATABASE_DSN` if needed
3. Run `make dev-up` to start services
