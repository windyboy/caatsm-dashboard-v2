package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap/zaptest"
)

func TestHandler_handleBroadcastEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	statsMock := new(mocks.StatsServiceMock)
	queryMock := new(mocks.QueryServiceMock)

	container := &app.Container{
		Logger:       logger,
		StatsService: statsMock,
		QueryService: queryMock,
	}

	broadcaster := NewEventBroadcaster(nil, logger)
	handler := &Handler{
		container:   container,
		broadcaster: broadcaster,
	}

	tests := []struct {
		name          string
		event         []byte
		setupMocks    func()
		expectError   bool
		expectMessage bool
	}{
		{
			name: "valid telegram_processed event with data field",
			event: []byte(`{
				"type": "telegram_processed",
				"data": {
					"telegram": {
						"message_id": "TEST-001",
						"type": "aftn",
						"time": "2024-01-01T00:00:00Z",
						"flight_number": "AA123",
						"source": "KJFK",
						"destination": "KLAX",
						"priority": 1,
						"content": "test",
						"raw_data": "test"
					}
				}
			}`),
			setupMocks: func() {
				statsMock.On("TrafficSummary", mock.Anything, mock.Anything).
					Return(&models.TrafficSummary{
						TotalMessages: 100,
						ByPriority:    map[int]int64{1: 50},
						ByType:        map[string]int64{"aftn": 50},
					}, nil)
			},
			expectError:   false,
			expectMessage: true,
		},
		{
			name: "valid telegram_processed event with top-level telegram",
			event: []byte(`{
				"type": "telegram_processed",
				"telegram": {
					"message_id": "TEST-002",
					"type": "aftn",
					"time": "2024-01-01T00:00:00Z",
					"flight_number": "AA456",
					"source": "KJFK",
					"destination": "KLAX",
					"priority": 2,
					"content": "test",
					"raw_data": "test"
				}
			}`),
			setupMocks: func() {
				statsMock.On("TrafficSummary", mock.Anything, mock.Anything).
					Return(&models.TrafficSummary{
						TotalMessages: 100,
						ByPriority:    map[int]int64{2: 50},
						ByType:        map[string]int64{"aftn": 50},
					}, nil)
			},
			expectError:   false,
			expectMessage: true,
		},
		{
			name:  "invalid JSON event",
			event: []byte(`invalid json`),
			setupMocks: func() {
			},
			expectError:   false,
			expectMessage: false,
		},
		{
			name:  "event missing type field",
			event: []byte(`{"data": {"key": "value"}}`),
			setupMocks: func() {
			},
			expectError:   false,
			expectMessage: false,
		},
		{
			name:  "non-telegram_processed event",
			event: []byte(`{"type": "stats_update", "time": 1234567890}`),
			setupMocks: func() {
				statsMock.On("TrafficSummary", mock.Anything, mock.Anything).
					Return(&models.TrafficSummary{
						TotalMessages: 100,
						ByPriority:    map[int]int64{},
						ByType:        map[string]int64{},
					}, nil)
			},
			expectError:   false,
			expectMessage: false,
		},
		{
			name:  "stats service error",
			event: []byte(`{"type": "stats_update"}`),
			setupMocks: func() {
				statsMock.On("TrafficSummary", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectError:   false,
			expectMessage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statsMock.ExpectedCalls = nil
			statsMock.Calls = nil
			tt.setupMocks()

			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}

			var serverConn *websocket.Conn
			upgradeComplete := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				serverConn, err = upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Errorf("failed to upgrade connection: %v", err)
					return
				}
				close(upgradeComplete)
			}))
			defer server.Close()

			wsURL := "ws" + server.URL[4:]
			clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err, "failed to dial websocket server")
			defer clientConn.Close()

			// Wait for server to complete upgrade
			select {
			case <-upgradeComplete:
				// Upgrade completed successfully
			case <-time.After(1 * time.Second):
				t.Fatal("websocket upgrade timed out")
			}
			require.NotNil(t, serverConn, "server connection should be established")
			defer serverConn.Close()

			ctx := context.Background()
			err = handler.handleBroadcastEvent(ctx, serverConn, tt.event)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			statsMock.AssertExpectations(t)

			// Verify message was sent/received according to expectMessage
			messageReceived := false
			clientConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			for {
				var msg WebSocketMessage
				err := clientConn.ReadJSON(&msg)
				if err != nil {
					break
				}
				if msg.Type == "message" {
					messageReceived = true
				}
			}

			if tt.expectMessage {
				assert.True(t, messageReceived, "expected a message type WebSocket message to be sent")
			} else {
				assert.False(t, messageReceived, "expected no message type WebSocket message to be sent")
			}
		})
	}
}

func TestWebSocketMessage_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		msg      WebSocketMessage
		wantType string
	}{
		{
			name: "message type",
			msg: WebSocketMessage{
				Type: "message",
				Data: map[string]any{"key": "value"},
			},
			wantType: "message",
		},
		{
			name: "stats-total type",
			msg: WebSocketMessage{
				Type: "stats-total",
				Data: map[string]any{"total": 100},
			},
			wantType: "stats-total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			require.NoError(t, err)

			var decoded WebSocketMessage
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, decoded.Type)
			assert.NotNil(t, decoded.Data)
		})
	}
}
