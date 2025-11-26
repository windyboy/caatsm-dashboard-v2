# Testing Guide

This document describes how to test the CAATSM Dashboard functionality, including the Deno + Svelte frontend.

## Prerequisites

1. Ensure all dependency services are running:
   ```bash
   task dev:up
   # or
   docker compose -f docker-compose.dev.yml up -d
   ```

2. Start the Go backend:
   ```bash
   make dev
   # or
   task dev
   # or
   ./bin/caatsm -config config/config.local.toml
   ```

3. Start the frontend development server:
   ```bash
   # Using Deno (recommended)
   task frontend:dev
   # or
   cd frontend && deno task dev
   
   # Using Node.js
   cd frontend
   npm install
   npm run dev
   ```

## WebSocket Testing

### Manual Testing Steps

1. **Open Dashboard Page**
   - Visit `http://localhost:5173`
   - You should see "Live Stream" component and statistics cards

2. **Check WebSocket Connection**
   - Open browser developer tools (F12)
   - Switch to Network tab
   - Filter for "WS" (WebSocket)
   - You should see a connection to `ws://localhost:3002/ws`

3. **Verify Real-time Message Reception**
   - If messages are flowing through NATS
   - You should see new messages appear in Live Stream
   - Statistics numbers should update in real-time

4. **Test Connection Reconnection**
   - In developer console, you should see "WebSocket connected" logs
   - If connection drops, you should see reconnection attempt logs

### Automated Testing

Run Playwright E2E tests:

```bash
# Using Taskfile/Makefile
task frontend:test
# or
make frontend-test

# Or manually
cd frontend
npm run test
# or with Deno
deno task test
```

Run unit tests:

```bash
task frontend:test:unit
# or
make frontend-test-unit
```

## REST API Testing

### Manual Testing Steps

1. **Test Search Functionality**
   - Visit `http://localhost:5173/search`
   - Enter search keywords
   - Click "Search" button
   - You should see search results

2. **Test Autocomplete**
   - Type at least 2 characters in the search box
   - You should see autocomplete suggestion dropdown

3. **Test Statistics Endpoints**
   - Visit Dashboard page
   - Statistics cards should display data
   - Check API requests in browser developer tools Network tab

4. **Test Export Functionality**
   - Execute a search on the search page
   - Export options should be available (if implemented)

### API Endpoint Testing

Use curl or Postman to test APIs:

```bash
# Test search
curl "http://localhost:3002/api/search?query=test"

# Test total statistics
curl "http://localhost:3002/api/stats/total"

# Test priority statistics
curl "http://localhost:3002/api/stats/priority"

# Test type statistics
curl "http://localhost:3002/api/stats/type"

# Test autocomplete
curl "http://localhost:3002/api/autocomplete?term=test&size=5"
```

All APIs should return JSON responses.

## WebSocket Message Format Testing

### Message Types

WebSocket messages should follow this format:

```json
{
  "type": "message|stats-total|stats-priority|stats-type",
  "data": { ... }
}
```

### Test Message Reception

Run in browser console:

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:3002/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

## Integration Testing

### Complete Flow Testing

1. **Start All Services**
   ```bash
   # Terminal 1: Start dependencies
   task dev:up
   
   # Terminal 2: Start Go backend
   task dev
   
   # Terminal 3: Start frontend
   task frontend:dev
   ```

2. **Test Real-time Data Flow**
   - If you have a message publisher, publish test messages to NATS
   - Observe if Dashboard updates in real-time
   - Check if statistics numbers update correctly

3. **Test Search and Real-time Updates Together**
   - Execute a search on the search page
   - Simultaneously observe Dashboard real-time updates
   - Both should work independently

## Performance Testing

### WebSocket Connection Count

Test multiple clients connecting simultaneously:

```bash
# Use multiple browser tabs or windows
# Each tab should independently receive messages
```

### Message Processing Performance

- Observe performance when large volumes of messages flow in
- Check if frontend limits message list length (should be max 50 messages)
- Check memory usage

## Error Handling Testing

### Network Disconnection

1. Disconnect network connection
2. WebSocket should attempt to reconnect
3. After network restoration, should automatically reconnect

### Backend Service Stop

1. Stop Go backend service
2. WebSocket should detect disconnection
3. Should show reconnection attempts
4. After restarting backend, should automatically reconnect

## Browser Compatibility

Test on the following browsers:
- Chrome/Edge (Chromium)
- Firefox
- Safari

## Known Issues

- If you encounter type errors, it may be a TypeScript configuration issue, doesn't affect runtime functionality
- WebSocket reconnection may take a few seconds

## Troubleshooting

### WebSocket Connection Failure

1. Check if Go backend is running on `localhost:3002`
2. Check firewall settings
3. Check browser console error messages

### API Request Failure

1. Check Vite proxy configuration
2. Check CORS settings
3. Check backend logs

### Frontend Build Failure

1. Run `npm install` to reinstall dependencies (or use Deno)
2. Run `npm run check` to check type errors (or `deno task check`)
3. Clear `.svelte-kit` directory and retry

## WebSocket Handler Tests

### Test File Location
`internal/handlers/websocket_test.go`

### Test Cases Covered
- Valid telegram_processed event handling (data field format)
- Valid telegram_processed event handling (top-level telegram format)
- Invalid JSON event handling
- Event missing type field
- Non-telegram_processed event (stats_update)
- Stats service error handling
- WebSocketMessage JSON serialization

### Running Tests
```bash
go test ./internal/handlers/... -v
```

## EventBroadcaster Tests

### Test File Location
`internal/handlers/broadcaster_test.go`

### Test Cases Covered
- Subscribe/Unsubscribe functionality
- Broadcast to multiple clients
- Channel full handling (capacity 10)
- Close functionality (all channels closed)
- PublishStatsUpdate with nil Redis client
- JSON validation

### Running Tests
```bash
go test ./internal/handlers/... -v
```
