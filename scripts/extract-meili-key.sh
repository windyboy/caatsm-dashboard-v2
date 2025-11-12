#!/bin/sh
# Extract Meilisearch master key from container logs
# Usage: docker compose logs meilisearch | ./scripts/extract-meili-key.sh
#    or: podman compose logs meilisearch | ./scripts/extract-meili-key.sh

set -e

# Step 1: Read logs from stdin and find the line with "We generated a new secure master key" and the next line
# Step 2: Get only the second line (which contains the key)
# Step 3: Remove container name prefix (e.g., "meilisearch-1  | ")
# Step 4: Remove leading ">> " if present
# Step 5: Extract the key value after "--master-key"
# Step 6: Remove trailing " <<" and anything after it
# Step 7: Trim any remaining whitespace
KEY=$(grep -A 1 "We generated a new secure master key" | \
  tail -n 1 | \
  sed 's/^[^|]*[|][ ]*//' | \
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
  echo "  docker compose -f docker-compose.dev.yml --env-file .env.local logs meilisearch | ./scripts/extract-meili-key.sh"
  echo "  or: podman compose -f docker-compose.dev.yml --env-file .env.local logs meilisearch | ./scripts/extract-meili-key.sh"
fi

