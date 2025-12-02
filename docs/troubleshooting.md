# Troubleshooting

Quick fixes for common issues.

## Health Checks

```bash
curl -f http://localhost:3002/api/health
curl -f http://localhost:7700/health
redis-cli ping
nats server check
```

## Common Issues

| Issue | Fix |
|-------|-----|
| Backend won't start | Check config validation errors in logs |
| Database errors | Verify `CAATSM_DATABASE_DSN`, run `make backend-migrate` |
| Search returns empty | Check Meilisearch health, verify API key |
| WebSocket disconnects | Check allowed origins, increase buffer size |
| Config validation fails | Set required env vars (TLS, auth, DSN) |

## Reset Environment

```bash
make dev-down
docker compose -f docker-compose.dev.yml down -v
docker compose -f docker-compose.dev.yml up -d
make backend-migrate
```
