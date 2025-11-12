#!/bin/sh
# Extract Meilisearch master key from docker-compose.dev.yml, .env.local, or container logs
# Usage: docker compose logs meilisearch 2>&1 | ./scripts/extract-meili-key.sh
#    or: podman compose logs meilisearch 2>&1 | ./scripts/extract-meili-key.sh
# Note: Use 2>&1 (not 2>/dev/null) to include stderr, as the key message goes to stderr

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
KEY=""

# Step 1: If logs are piped in, ONLY extract from logs (don't fall back to .env.local)
# The log format is:
#   We generated a new secure master key for you (you can safely use this token):
#   >> --master-key KEY <<
# Step 1a: Read logs from stdin and find the line with "--master-key"
# Step 1b: Remove container name prefix (e.g., "meilisearch-1  | " or "caatsm-dashboard_meilisearch_1  | ")
# Step 1c: Remove leading ">> " if present
# Step 1d: Extract the key value after "--master-key"
# Step 1e: Remove trailing " <<" and anything after it
# Step 1f: Trim any remaining whitespace
if [ ! -t 0 ]; then
  # Read stdin into a variable so we can use it multiple times
  LOGS=$(cat)
  KEY=$(echo "$LOGS" | grep -E -- "--master-key" | \
    sed 's/^[^|]*[|][ ]*//' | \
    sed 's/^>>[ ]*//' | \
    sed 's/.*--master-key[ ]*//' | \
    sed 's/[ ]*<<.*$//' | \
    sed 's/^[[:space:]]*//' | \
    sed 's/[[:space:]]*$//' | \
    head -n 1 || true)
  if [ -n "$KEY" ]; then
    echo "$KEY"
    echo ""
    echo "Found generated master key in logs"
    # Update .env.local if it exists
    if [ -f "$PROJECT_ROOT/.env.local" ]; then
      ENV_FILE="$PROJECT_ROOT/.env.local"
      TMP_FILE=$(mktemp)
      MEILI_KEY_UPDATED=false
      API_KEY_UPDATED=false
      
      # Quote the key value with double quotes to handle special characters
      QUOTED_KEY="\"$KEY\""
      
      # Check if keys exist in original file
      MEILI_KEY_EXISTS=false
      API_KEY_EXISTS=false
      if grep -q "^MEILI_MASTER_KEY=" "$ENV_FILE"; then
        MEILI_KEY_EXISTS=true
      fi
      if grep -q "^CAATSM_MEILISEARCH_API_KEY=" "$ENV_FILE"; then
        API_KEY_EXISTS=true
      fi
      
      # Process the file line by line
      while IFS= read -r line || [ -n "$line" ]; do
        # Update MEILI_MASTER_KEY if it exists
        if echo "$line" | grep -q "^MEILI_MASTER_KEY="; then
          echo "MEILI_MASTER_KEY=$QUOTED_KEY" >> "$TMP_FILE"
          MEILI_KEY_UPDATED=true
        # Update CAATSM_MEILISEARCH_API_KEY if it exists
        elif echo "$line" | grep -q "^CAATSM_MEILISEARCH_API_KEY="; then
          echo "CAATSM_MEILISEARCH_API_KEY=$QUOTED_KEY" >> "$TMP_FILE"
          API_KEY_UPDATED=true
        # Add MEILI_MASTER_KEY after MEILI_ENV if it doesn't exist yet
        elif echo "$line" | grep -q "^MEILI_ENV=" && [ "$MEILI_KEY_EXISTS" = false ] && [ "$MEILI_KEY_UPDATED" = false ]; then
          echo "$line" >> "$TMP_FILE"
          echo "MEILI_MASTER_KEY=$QUOTED_KEY" >> "$TMP_FILE"
          MEILI_KEY_UPDATED=true
        # Add CAATSM_MEILISEARCH_API_KEY after CAATSM_MEILISEARCH_HOST if it doesn't exist yet
        elif echo "$line" | grep -q "^CAATSM_MEILISEARCH_HOST=" && [ "$API_KEY_EXISTS" = false ] && [ "$API_KEY_UPDATED" = false ]; then
          echo "$line" >> "$TMP_FILE"
          echo "CAATSM_MEILISEARCH_API_KEY=$QUOTED_KEY" >> "$TMP_FILE"
          API_KEY_UPDATED=true
        else
          echo "$line" >> "$TMP_FILE"
        fi
      done < "$ENV_FILE"
      
      # Move temp file to .env.local
      mv "$TMP_FILE" "$ENV_FILE"
      
      if [ "$MEILI_KEY_UPDATED" = true ] || [ "$API_KEY_UPDATED" = true ]; then
        if [ "$MEILI_KEY_UPDATED" = true ]; then
          if [ "$MEILI_KEY_EXISTS" = true ]; then
            echo "Updated MEILI_MASTER_KEY in .env.local"
          else
            echo "Added MEILI_MASTER_KEY to .env.local"
          fi
        fi
        if [ "$API_KEY_UPDATED" = true ]; then
          if [ "$API_KEY_EXISTS" = true ]; then
            echo "Updated CAATSM_MEILISEARCH_API_KEY in .env.local"
          else
            echo "Added CAATSM_MEILISEARCH_API_KEY to .env.local"
          fi
        fi
        echo ""
        echo "✓ .env.local has been updated with the new key"
      fi
    else
      echo "To use this key, update your .env.local:"
      echo "MEILI_MASTER_KEY=$KEY"
      echo "CAATSM_MEILISEARCH_API_KEY=$KEY"
    fi
    exit 0
  else
    echo "Could not find generated master key in logs."
    echo ""
    echo "If you set MEILI_MASTER_KEY in docker-compose.dev.yml or .env.local, use that value."
    echo "Otherwise, check the logs manually for a generated key."
    exit 1
  fi
fi

# Step 2: Check .env.local for MEILI_MASTER_KEY
if [ -f "$PROJECT_ROOT/.env.local" ]; then
  KEY=$(grep -E "^MEILI_MASTER_KEY=" "$PROJECT_ROOT/.env.local" | cut -d '=' -f2- | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' || true)
  if [ -n "$KEY" ]; then
    echo "$KEY"
    echo ""
    echo "Found MEILI_MASTER_KEY in .env.local"
    echo "To use this key, update your .env.local:"
    echo "CAATSM_MEILISEARCH_API_KEY=$KEY"
    exit 0
  fi
fi

# Step 3: Check docker-compose.dev.yml for MEILI_MASTER_KEY
if [ -f "$PROJECT_ROOT/docker-compose.dev.yml" ]; then
  COMPOSE_LINE=$(grep -E "MEILI_MASTER_KEY:" "$PROJECT_ROOT/docker-compose.dev.yml" | head -n 1 || true)
  if [ -n "$COMPOSE_LINE" ]; then
    # Extract the value after MEILI_MASTER_KEY:
    COMPOSE_VALUE=$(echo "$COMPOSE_LINE" | sed 's/.*MEILI_MASTER_KEY:[[:space:]]*//' | sed 's/[[:space:]]*$//')
    # If it's ${MEILI_MASTER_KEY:-default}, extract the default value
    if echo "$COMPOSE_VALUE" | grep -qE '\$\{MEILI_MASTER_KEY:-'; then
      KEY=$(echo "$COMPOSE_VALUE" | sed 's/.*\$\{MEILI_MASTER_KEY:-\([^}]*\)\}.*/\1/')
    # If it's just ${MEILI_MASTER_KEY}, it needs to come from env
    elif echo "$COMPOSE_VALUE" | grep -qE '\$\{MEILI_MASTER_KEY\}'; then
      KEY=""
    # Otherwise, it's a literal value
    else
      KEY="$COMPOSE_VALUE"
    fi
    if [ -n "$KEY" ]; then
      echo "$KEY"
      echo ""
      echo "Found MEILI_MASTER_KEY in docker-compose.dev.yml"
      echo "To use this key, update your .env.local:"
      echo "CAATSM_MEILISEARCH_API_KEY=$KEY"
      exit 0
    fi
  fi
fi

# If we get here, no key was found
echo "Could not find master key."
echo ""
echo "Checked (in order of priority):"
if [ -n "$LOGS_FROM_STDIN" ]; then
  echo "  1. Container logs (generated key) - no generated key found in logs"
  echo "  2. .env.local (MEILI_MASTER_KEY)"
  echo "  3. docker-compose.dev.yml (MEILI_MASTER_KEY)"
else
  echo "  1. .env.local (MEILI_MASTER_KEY)"
  echo "  2. docker-compose.dev.yml (MEILI_MASTER_KEY)"
fi
echo ""
echo "To set the key manually:"
echo "  1. Check logs for generated key (use 2>&1 to include stderr):"
echo "     docker compose -f docker-compose.dev.yml --env-file .env.local logs meilisearch 2>&1 | ./scripts/extract-meili-key.sh"
echo "     or: podman compose -f docker-compose.dev.yml --env-file .env.local logs meilisearch 2>&1 | ./scripts/extract-meili-key.sh"
echo "  2. Check docker-compose.dev.yml for MEILI_MASTER_KEY value"
echo "  3. Check .env.local for MEILI_MASTER_KEY value"
exit 1

