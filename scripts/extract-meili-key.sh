#!/bin/sh
# Extract Meilisearch master key from container logs

set -e

if command -v podman > /dev/null 2>&1 && ! command -v docker > /dev/null 2>&1; then
  COMPOSE_CMD="podman compose"
else
  COMPOSE_CMD="docker compose"
fi

# Step 1: Get logs and find the line with "We generated a new secure master key" and the next line
# Step 2: Get only the second line (which contains the key)
# Step 3: Remove container name prefix (e.g., "meilisearch-1  | ")
# Step 4: Remove leading ">> " if present
# Step 5: Extract the key value after "--master-key"
# Step 6: Remove trailing " <<" and anything after it
# Step 7: Trim any remaining whitespace
KEY=$($COMPOSE_CMD -f docker-compose.dev.yml logs meilisearch 2>/dev/null | \
  grep -A 1 "We generated a new secure master key" | \
  tail -n 1 | \
  sed 's/^[^|]*| *//' | \
  sed 's/^>> *//' | \
  sed 's/.*--master-key *//' | \
  sed 's/ *<<.*$//' | \
  sed 's/^[[:space:]]*//' | \
  sed 's/[[:space:]]*$//' | \
  head -n 1 || true)

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

