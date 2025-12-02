#!/bin/bash
set -e

echo "Checking architecture layer boundaries..."

# Infrastructure must not import delivery
if go list -f '{{.ImportPath}} {{.Imports}}' ./internal/infrastructure/... 2>/dev/null | grep -q delivery; then
    echo "❌ ERROR: infrastructure must not import delivery"
    go list -f '{{.ImportPath}} {{.Imports}}' ./internal/infrastructure/... | grep delivery
    exit 1
fi

# Infrastructure must not import service
if go list -f '{{.ImportPath}} {{.Imports}}' ./internal/infrastructure/... 2>/dev/null | grep -q service; then
    echo "❌ ERROR: infrastructure must not import service"
    go list -f '{{.ImportPath}} {{.Imports}}' ./internal/infrastructure/... | grep service
    exit 1
fi

# Domain must not import anything from internal (except itself)
if go list -f '{{.ImportPath}} {{.Imports}}' ./internal/domain/... 2>/dev/null | grep -E 'internal/(delivery|service|infrastructure|repository|app)'; then
    echo "❌ ERROR: domain must not import other internal packages"
    go list -f '{{.ImportPath}} {{.Imports}}' ./internal/domain/... | grep -E 'internal/(delivery|service|infrastructure|repository|app)'
    exit 1
fi

# Repository must not import delivery or service
if go list -f '{{.ImportPath}} {{.Imports}}' ./internal/repository/... 2>/dev/null | grep -E 'internal/(delivery|service)'; then
    echo "❌ ERROR: repository must not import delivery or service"
    go list -f '{{.ImportPath}} {{.Imports}}' ./internal/repository/... | grep -E 'internal/(delivery|service)'
    exit 1
fi

# Service must not import delivery
if go list -f '{{.ImportPath}} {{.Imports}}' ./internal/service/... 2>/dev/null | grep -q delivery; then
    echo "❌ ERROR: service must not import delivery"
    go list -f '{{.ImportPath}} {{.Imports}}' ./internal/service/... | grep delivery
    exit 1
fi

echo "✅ Layer boundaries respected"
exit 0

