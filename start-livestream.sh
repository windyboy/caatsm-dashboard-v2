#!/bin/bash
# Quick Start Script for WebSocket Live Stream
# This script starts the message publisher so you can see live updates

echo "🚀 Starting WebSocket Live Stream Publisher"
echo "=========================================="
echo ""

# Check if services are running
echo "📦 Checking services..."
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3002/api/health 2>/dev/null)
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5173 2>/dev/null)

if [ "$BACKEND_STATUS" != "200" ]; then
    echo "❌ Backend not running on port 3002"
    echo "   Please start backend: make dev"
    echo ""
    exit 1
fi

if [ "$FRONTEND_STATUS" != "200" ]; then
    echo "⚠️  Frontend not running on port 5173"
    echo "   You can still test, but open frontend: make frontend-dev"
    echo ""
fi

echo "✅ Backend is running"
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "✅ Frontend is running"
fi
echo ""

# Ask user which mode
echo "Select publishing mode:"
echo "  1) Fast test (50 messages, 1/second) - Recommended"
echo "  2) Slow continuous (1 message/5 seconds)"
echo "  3) Sync worker (production-like, reads from NATS)"
echo "  4) Manual test (publish single message)"
echo ""
read -p "Enter choice (1-4) [1]: " CHOICE
CHOICE=${CHOICE:-1}

echo ""
echo "=========================================="
echo ""

case $CHOICE in
    1)
        echo "📤 Starting fast publisher (50 messages @ 1/second)..."
        echo ""
        echo "🌐 Open browser to: http://localhost:5173"
        echo "📊 Or use diagnostic: http://localhost:5173/test-websocket.html"
        echo ""
        echo "Press Ctrl+C to stop"
        echo ""
        task publish-stream:fast
        ;;
    2)
        echo "📤 Starting slow publisher (continuous @ 5 seconds)..."
        echo ""
        echo "🌐 Open browser to: http://localhost:5173"
        echo "📊 Or use diagnostic: http://localhost:5173/test-websocket.html"
        echo ""
        echo "Press Ctrl+C to stop"
        echo ""
        task publish-stream:slow
        ;;
    3)
        echo "📤 Starting sync worker (reads from NATS)..."
        echo ""
        echo "Note: You need to publish to NATS first"
        echo "Run in another terminal: task publish-stream:fast"
        echo ""
        echo "🌐 Open browser to: http://localhost:5173"
        echo ""
        echo "Press Ctrl+C to stop"
        echo ""
        go run ./cmd/sync -config config/config.local.toml
        ;;
    4)
        echo "📨 Publishing single test message..."
        MSG_ID="MANUAL-$(date +%s)"
        docker exec caatsm-dashboard-v2-redis-1 valkey-cli PUBLISH stats:update \
            "{\"type\":\"telegram_processed\",\"data\":{\"telegram\":{\"message_id\":\"$MSG_ID\",\"type\":\"TEST\",\"flight_number\":\"MANUAL123\",\"source\":\"VTBS\",\"destination\":\"VVTS\",\"priority\":1,\"content\":\"Manual test message\",\"time\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"raw_data\":\"{}\"}}}" 2>&1
        
        echo ""
        echo "✅ Test message published with ID: $MSG_ID"
        echo ""
        echo "🌐 Check your browser at: http://localhost:5173"
        echo "   The message should appear immediately in the live stream"
        echo ""
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "=========================================="

