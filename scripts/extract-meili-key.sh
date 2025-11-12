#!/bin/sh
# Extract Meilisearch master key from container logs

set -e

if command -v podman > /dev/null 2>&1 && ! command -v docker > /dev/null 2>&1; then
  COMPOSE_CMD="podman compose"
else
  COMPOSE_CMD="docker compose"
fi

KEY=$($COMPOSE_CMD -f docker-compose.dev.yml logs meilisearch 2>/dev/null | \
  grep -A 1 "We generated a new secure master key" | \
  tail -1 | \
  grep -oE '--master-key [^ ]+' | \
  sed 's/--master-key //' | \
  tr -d '><' | \
  head -1 || true)

if [ -n "$KEY" ]; then
  echo "$KEY"
  echo ""
  echo "To use this key, update your .env.local:"
  echo "CAATSM_MEILISEARCH_API_KEY=$KEY"
else
  echo "Could not find generated master key in logs."
  echo "If you set MEILI_MASTER_KEY in docker-compose.dev.yml, use that value."
  echo "Otherwise, check the logs manually:"
  echo "  $COMPOSE_CMD -f docker-compose.dev.yml logs meilisearch"
fi

