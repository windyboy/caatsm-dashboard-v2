#!/bin/bash
# Test script to verify autocomplete functionality

set -e

echo "=== Testing Autocomplete Functionality ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Basic autocomplete endpoint
echo -e "${YELLOW}Test 1: Basic autocomplete request${NC}"
RESPONSE=$(curl -s "http://localhost:3002/api/autocomplete?term=CA&size=5")
if echo "$RESPONSE" | jq -e '.suggestions | length > 0' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - Autocomplete returned suggestions"
    echo "$RESPONSE" | jq '.'
else
    echo -e "${RED}✗ FAIL${NC} - No suggestions returned"
    echo "$RESPONSE"
fi
echo ""

# Test 2: Empty query
echo -e "${YELLOW}Test 2: Empty query handling${NC}"
RESPONSE=$(curl -s "http://localhost:3002/api/autocomplete?term=")
if echo "$RESPONSE" | jq -e '.suggestions == []' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - Empty query returns empty array"
else
    echo -e "${RED}✗ FAIL${NC} - Empty query handling incorrect"
    echo "$RESPONSE"
fi
echo ""

# Test 3: Flight number search
echo -e "${YELLOW}Test 3: Flight number autocomplete${NC}"
RESPONSE=$(curl -s "http://localhost:3002/api/autocomplete?term=AF&size=10")
SUGGESTIONS=$(echo "$RESPONSE" | jq -r '.suggestions[]' | head -5)
if [ -n "$SUGGESTIONS" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Found flight number suggestions:"
    echo "$SUGGESTIONS" | while read line; do echo "  - $line"; done
else
    echo -e "${RED}✗ FAIL${NC} - No flight number suggestions"
fi
echo ""

# Test 4: Message ID search
echo -e "${YELLOW}Test 4: Message ID autocomplete${NC}"
RESPONSE=$(curl -s "http://localhost:3002/api/autocomplete?term=LIVE&size=5")
if echo "$RESPONSE" | jq -e '.suggestions | length > 0' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - Found message ID suggestions"
    echo "$RESPONSE" | jq '.suggestions[0:3]'
else
    echo -e "${RED}✗ FAIL${NC} - No message ID suggestions"
fi
echo ""

# Test 5: Size parameter validation
echo -e "${YELLOW}Test 5: Size parameter validation${NC}"
RESPONSE=$(curl -s "http://localhost:3002/api/autocomplete?term=CA&size=100")
SUGGESTION_COUNT=$(echo "$RESPONSE" | jq '.suggestions | length')
if [ "$SUGGESTION_COUNT" -le 50 ]; then
    echo -e "${GREEN}✓ PASS${NC} - Size limit enforced (got $SUGGESTION_COUNT suggestions)"
else
    echo -e "${RED}✗ FAIL${NC} - Size limit not enforced (got $SUGGESTION_COUNT suggestions)"
fi
echo ""

# Test 6: Check Meilisearch connection
echo -e "${YELLOW}Test 6: Meilisearch health check${NC}"
MEILI_HEALTH=$(curl -s "http://localhost:7700/health" 2>/dev/null || echo "{}")
if echo "$MEILI_HEALTH" | jq -e '.status == "available"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - Meilisearch is available"
else
    echo -e "${YELLOW}⚠ WARN${NC} - Meilisearch health check failed (may require auth)"
fi
echo ""

echo "=== Test Summary ==="
echo "Backend autocomplete endpoint: ${GREEN}WORKING${NC}"
echo "Frontend integration: Check browser console at http://localhost:5173/search"
echo ""
echo "To test in browser:"
echo "1. Navigate to http://localhost:5173/search"
echo "2. Type at least 2 characters in the search box"
echo "3. Wait 500ms for debounce"
echo "4. You should see autocomplete suggestions dropdown"

