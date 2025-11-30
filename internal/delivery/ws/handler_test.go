package ws

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap/zaptest"
)

func TestPrepareAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice returns defaults",
			input:    []string{},
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "nil slice returns defaults",
			input:    nil,
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "trims whitespace",
			input:    []string{" http://example.com ", "  https://test.com  "},
			expected: []string{"http://example.com", "https://test.com"},
		},
		{
			name:     "filters empty entries",
			input:    []string{"http://example.com", "", "  ", "https://test.com"},
			expected: []string{"http://example.com", "https://test.com"},
		},
		{
			name:     "all empty entries returns defaults",
			input:    []string{"", "  ", "   "},
			expected: []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:     "valid origins unchanged",
			input:    []string{"http://localhost:3000", "https://example.com"},
			expected: []string{"http://localhost:3000", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareAllowedOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d origins, got %d", len(tt.expected), len(result))
				return
			}
			for i, origin := range result {
				if origin != tt.expected[i] {
					t.Errorf("origin[%d] = %s, want %s", i, origin, tt.expected[i])
				}
			}
		})
	}
}

func TestMakeCheckOriginFunc(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expected       bool
	}{
		{
			name:           "empty origin header returns true",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "",
			expected:       true,
		},
		{
			name:           "allowed origin returns true",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "http://localhost:3000",
			expected:       true,
		},
		{
			name:           "disallowed origin returns false",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://evil.com",
			expected:       false,
		},
		{
			name:           "multiple allowed origins",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com", "http://localhost:5173"},
			requestOrigin:  "https://example.com",
			expected:       true,
		},
		{
			name:           "case insensitive scheme and host per RFC 6454",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "HTTP://LOCALHOST:3000",
			expected:       true,
		},
		{
			name:           "mixed case domain matches",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://Example.Com",
			expected:       true,
		},
		{
			name:           "uppercase scheme matches",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "HTTPS://example.com",
			expected:       true,
		},
		{
			name:           "default port 80 removed for http",
			allowedOrigins: []string{"http://localhost"},
			requestOrigin:  "http://localhost:80",
			expected:       true,
		},
		{
			name:           "default port 443 removed for https",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com:443",
			expected:       true,
		},
		{
			name:           "non-default port must match exactly",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:8080",
			expected:       false,
		},
		{
			name:           "non-default port preserved",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkOrigin := makeCheckOriginFunc(tt.allowedOrigins)
			req := &http.Request{
				Header: http.Header{},
			}
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			result := checkOrigin(req)
			if result != tt.expected {
				t.Errorf("checkOrigin(%s) = %v, want %v", tt.requestOrigin, result, tt.expected)
			}
		})
	}
}

func TestNormalizeOrigin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase scheme and host",
			input:    "HTTP://LOCALHOST:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "mixed case domain",
			input:    "https://Example.Com",
			expected: "https://example.com",
		},
		{
			name:     "already normalized",
			input:    "http://localhost:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "remove default http port 80",
			input:    "http://localhost:80",
			expected: "http://localhost",
		},
		{
			name:     "remove default https port 443",
			input:    "https://example.com:443",
			expected: "https://example.com",
		},
		{
			name:     "preserve non-default http port",
			input:    "http://localhost:3000",
			expected: "http://localhost:3000",
		},
		{
			name:     "preserve non-default https port",
			input:    "https://example.com:8443",
			expected: "https://example.com:8443",
		},
		{
			name:     "complex case with uppercase scheme and domain",
			input:    "HTTPS://API.EXAMPLE.COM:443",
			expected: "https://api.example.com",
		},
		{
			name:     "invalid URL returns original",
			input:    "not a valid url",
			expected: "not a valid url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeOrigin(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeOrigin(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWebSocketHandler_MessageBroadcast tests that messages are broadcasted to connected clients
func TestWebSocketHandler_MessageBroadcast(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup hub and handler
	hub := ws.NewHub(ws.DefaultConfig(), logger)
	defer hub.Close()

	handler := NewHandler(
		hub,
		logger,
		&stubStatsService{summary: &app.TrafficSummary{}}, // stub stats service
		&stubQueryService{telegrams: []app.Telegram{}},    // stub query service
		nil, // realtime service not needed
		nil, // event publisher not needed
	)

	// Setup Echo server
	e := echo.New()
	e.GET("/ws", handler.HandleWebSocket)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(e)
	server.Listener = listener
	server.Start()
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Read initial stats message from client (sent automatically on connection)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var initialMsg app.WSMessage
	require.NoError(t, json.Unmarshal(data, &initialMsg))
	assert.Equal(t, "stats", initialMsg.Type)

	// Now broadcast a message via hub
	testMsg := app.WSMessage{
		Type: "test-message",
		Data: map[string]string{"key": "value"},
	}
	msgBytes, _ := json.Marshal(testMsg)
	hub.Broadcast(msgBytes)

	// Read messages until we find the broadcast message
	// There might be recent messages sent first, so we need to skip those
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var broadcastMsg app.WSMessage
	found := false
	for i := 0; i < 10; i++ { // Read up to 10 messages to find the broadcast
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if err := json.Unmarshal(data, &broadcastMsg); err != nil {
			continue
		}
		if broadcastMsg.Type == "test-message" {
			found = true
			break
		}
		// Reset deadline for next read
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	}
	require.True(t, found, "broadcast message not received")
	assert.Equal(t, "test-message", broadcastMsg.Type)
	assert.Equal(t, map[string]interface{}{"key": "value"}, broadcastMsg.Data)
}

// TestWebSocketHandler_SlowClientDisconnect tests that slow clients are disconnected
func TestWebSocketHandler_SlowClientDisconnect(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup hub with small buffer to trigger backpressure quickly
	config := ws.DefaultConfig()
	config.ClientBufferSize = 5 // Small buffer to trigger backpressure
	hub := ws.NewHub(config, logger)
	defer hub.Close()

	handler := NewHandler(
		hub,
		logger,
		&stubStatsService{summary: &app.TrafficSummary{}},
		&stubQueryService{telegrams: []app.Telegram{}},
		nil,
		nil,
	)

	e := echo.New()
	e.GET("/ws", handler.HandleWebSocket)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(e)
	server.Listener = listener
	server.Start()
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Read initial stats and recent messages to clear them
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, _ = conn.ReadMessage() // Initial stats
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, _ = conn.ReadMessage()          // Recent messages (if sent, timeout if not)
	_ = conn.SetReadDeadline(time.Time{}) // Reset deadline

	// Don't read messages to simulate slow client
	// Broadcast enough messages rapidly to fill the buffer (buffer size is 5)
	// Send messages faster than WritePump can consume them
	for i := 0; i < 20; i++ {
		testMsg := app.WSMessage{
			Type: "test",
			Data: map[string]int{"count": i},
		}
		msgBytes, _ := json.Marshal(testMsg)
		hub.Broadcast(msgBytes)
	}

	// Wait for disconnect due to slow client (buffer should fill and trigger disconnection)
	// Give enough time for the hub to detect, disconnect, and close the connection
	// The WritePump needs time to process the channel close and close the connection
	time.Sleep(300 * time.Millisecond)

	// Try to read - should fail because connection should be closed
	// The connection should be closed by WritePump when the send channel is closed
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	// Should get close error or connection closed error
	// Note: The connection might be closed gracefully, so we check for any error
	assert.Error(t, err, "expected read to fail on closed connection")

	// Also verify by trying to write
	writeErr := conn.WriteMessage(websocket.TextMessage, []byte("test"))
	// Write might succeed if connection isn't fully closed yet, but read should fail
	if writeErr == nil {
		// If write succeeded, try reading again with a longer timeout
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		_, _, readErr := conn.ReadMessage()
		assert.Error(t, readErr, "connection should be closed, read should fail")
	}
}

// TestWebSocketHandler_ConnectionError tests handling of invalid WebSocket upgrade requests
func TestWebSocketHandler_ConnectionError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	hub := ws.NewHub(ws.DefaultConfig(), logger)
	defer hub.Close()

	handler := NewHandler(
		hub,
		logger,
		nil,
		nil,
		nil,
		nil,
	)

	e := echo.New()
	e.GET("/ws", handler.HandleWebSocket)

	// Test with invalid request (missing upgrade headers)
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleWebSocket(c)
	// Should return error for non-WebSocket request
	assert.Error(t, err)
}

type stubStatsService struct {
	summary *app.TrafficSummary
}

func (s *stubStatsService) TrafficSummary(ctx context.Context, window interface{}) (interface{}, error) {
	return s.summary, nil
}

type stubQueryService struct {
	telegrams []app.Telegram
}

func (s *stubQueryService) Recent(ctx context.Context, limit int) ([]*app.Telegram, error) {
	result := make([]*app.Telegram, 0, len(s.telegrams))
	for i := range s.telegrams {
		result = append(result, &s.telegrams[i])
	}
	return result, nil
}
