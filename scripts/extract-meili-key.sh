#!/bin/sh
# Extract Meilisearch master key from logs, .env.local, or docker-compose.dev.yml
# Usage: docker compose logs meilisearch 2>&1 | ./scripts/extract-meili-key.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/.env.local"

# Extract from logs if piped via stdin
if [ ! -t 0 ]; then
  KEY=$(grep -oE '--master-key[[:space:]]+[^[:space:]]+' | sed 's/--master-key[[:space:]]*//' | head -1)
  
  if [ -n "$KEY" ]; then
    echo "$KEY"
    
    # Update .env.local if it exists
    if [ -f "$ENV_FILE" ]; then
      # Update or add both keys
      grep -q "^MEILI_MASTER_KEY=" "$ENV_FILE" && \
        sed -i.bak "s|^MEILI_MASTER_KEY=.*|MEILI_MASTER_KEY=\"$KEY\"|" "$ENV_FILE" || \
        echo "MEILI_MASTER_KEY=\"$KEY\"" >> "$ENV_FILE"
      
      grep -q "^CAATSM_MEILISEARCH_API_KEY=" "$ENV_FILE" && \
        sed -i.bak "s|^CAATSM_MEILISEARCH_API_KEY=.*|CAATSM_MEILISEARCH_API_KEY=\"$KEY\"|" "$ENV_FILE" || \
        echo "CAATSM_MEILISEARCH_API_KEY=\"$KEY\"" >> "$ENV_FILE"
      
      rm -f "$ENV_FILE.bak" 2>/dev/null
      echo "✓ Updated .env.local"
    fi
    exit 0
  fi
  echo "Key not found in logs" >&2
  exit 1
fi

# Check .env.local
if [ -f "$ENV_FILE" ]; then
  KEY=$(grep "^MEILI_MASTER_KEY=" "$ENV_FILE" | cut -d '=' -f2- | tr -d '"' | tr -d "'" | xargs)
  [ -n "$KEY" ] && echo "$KEY" && exit 0
fi

# Check docker-compose.dev.yml
if [ -f "$PROJECT_ROOT/docker-compose.dev.yml" ]; then
  KEY=$(grep "MEILI_MASTER_KEY:" "$PROJECT_ROOT/docker-compose.dev.yml" | \
    sed 's/.*MEILI_MASTER_KEY:[[:space:]]*//' | \
    sed 's/\${MEILI_MASTER_KEY:-\([^}]*\)}.*/\1/' | \
    sed 's/\${MEILI_MASTER_KEY}.*//' | xargs)
  [ -n "$KEY" ] && echo "$KEY" && exit 0
fi

echo "Key not found" >&2
exit 1
