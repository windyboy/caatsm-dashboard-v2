#!/bin/bash
set -euo pipefail

echo "Setting up frontend dependencies..."

if command -v deno >/dev/null 2>&1; then
  echo "Deno detected: $(deno --version | head -n 1)"
else
  echo "Deno not found. Install from https://deno.land/#installation if you prefer running via deno."
fi

if command -v npm >/dev/null 2>&1; then
  npm install
elif command -v bun >/dev/null 2>&1; then
  bun install
else
  echo "Install npm or bun to fetch node modules." >&2
  exit 1
fi

echo "Done. Start the dev server with 'npm run dev' or 'deno task dev'."
