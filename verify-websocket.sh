#!/bin/bash
# WebSocket Live Stream Verification Script

echo "🔍 WebSocket Live Stream Verification"
echo "======================================"
echo ""

# Check Docker services
echo "📦 Docker Services:"
docker ps --filter "name=caatsm-dashboard-v2" --format "  ✓ {{.Names}}: {{.Status}}" 2>/dev/null || echo "  ❌ Docker not running or no containers found"
echo ""

# Check database
echo "💾 Database Check:"
DB_COUNT=$(docker exec caatsm-dashboard-v2-postgres-1 psql -U caatsm -d caatsm -t -c "SELECT COUNT(*) FROM telegrams;" 2>/dev/null | tr -d ' ')
if [ -n "$DB_COUNT" ]; then
    echo "  ✓ Telegrams in database: $DB_COUNT"
else
    echo "  ❌ Could not query database"
fi
echo ""

# Check Redis
echo "🔴 Redis Check:"
REDIS_PING=$(docker exec caatsm-dashboard-v2-redis-1 valkey-cli ping 2>/dev/null)
if [ "$REDIS_PING" = "PONG" ]; then
    echo "  ✓ Redis responding: $REDIS_PING"
else
    echo "  ❌ Redis not responding"
fi
echo ""

# Check backend API
echo "🖥️  Backend API Check:"
API_HEALTH=$(curl -s http://localhost:3002/api/health 2>/dev/null)
if [ -n "$API_HEALTH" ]; then
    echo "  ✓ Backend API responding: $API_HEALTH"
else
    echo "  ❌ Backend API not responding on port 3002"
fi
echo ""

# Check frontend
echo "🌐 Frontend Check:"
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5173 2>/dev/null)
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "  ✓ Frontend responding on port 5173"
else
    echo "  ❌ Frontend not responding on port 5173 (status: $FRONTEND_STATUS)"
fi
echo ""

# Check if publisher is running
echo "📤 Publisher Check:"
PUBLISHER_PID=$(ps aux | grep "publish-stream" | grep -v grep | awk '{print $2}' | head -1)
if [ -n "$PUBLISHER_PID" ]; then
    echo "  ✓ Publisher running (PID: $PUBLISHER_PID)"
else
    echo "  ⚠️  No publisher running (start with: task publish-stream:fast)"
fi
echo ""

# Test WebSocket endpoint
echo "🔌 WebSocket Endpoint Test:"
WS_TEST=$(timeout 5 curl -i -N \
    -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
    http://localhost:3002/ws 2>&1 | head -1)

if echo "$WS_TEST" | grep -q "101"; then
    echo "  ✓ WebSocket endpoint accepting connections (HTTP 101)"
else
    echo "  ❌ WebSocket endpoint not accepting connections"
    echo "     Response: $WS_TEST"
fi
echo ""

# Publish test message
echo "📨 Publishing Test Message to Redis:"
PUBLISH_RESULT=$(docker exec caatsm-dashboard-v2-redis-1 valkey-cli PUBLISH stats:update \
    '{"type":"telegram_processed","data":{"telegram":{"message_id":"VERIFY-'$(date +%s)'","type":"TEST","flight_number":"VERIFY001","source":"TEST","destination":"TEST","priority":1,"content":"Verification test message","time":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","raw_data":"{}"}}}' 2>/dev/null)

if [ "$PUBLISH_RESULT" = "1" ]; then
    echo "  ✓ Test message published successfully (1 subscriber received)"
else
    echo "  ⚠️  Test message published but subscribers: $PUBLISH_RESULT"
fi
echo ""

# Summary
echo "======================================"
echo "📋 Summary:"
echo ""

ALL_GOOD=true

if [ -z "$DB_COUNT" ]; then ALL_GOOD=false; fi
if [ "$REDIS_PING" != "PONG" ]; then ALL_GOOD=false; fi
if [ -z "$API_HEALTH" ]; then ALL_GOOD=false; fi
if [ "$FRONTEND_STATUS" != "200" ]; then ALL_GOOD=false; fi

if $ALL_GOOD; then
    echo "  ✅ All systems operational!"
    echo ""
    echo "  🎯 Next Steps:"
    echo "     1. Open http://localhost:5173 in your browser"
    echo "     2. Check Live Stream component for messages"
    echo "     3. For diagnostics: http://localhost:5173/test-websocket.html"
    echo ""
    if [ -z "$PUBLISHER_PID" ]; then
        echo "  💡 Tip: Start publisher for continuous messages:"
        echo "     task publish-stream:fast"
    fi
else
    echo "  ⚠️  Some services need attention"
    echo ""
    echo "  📖 Check WEBSOCKET_STATUS.md for detailed troubleshooting"
fi

echo ""
echo "======================================"

