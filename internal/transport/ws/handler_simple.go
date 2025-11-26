package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should validate the origin
		return true
	},
}

// WebSocketMessage represents a WebSocket message sent to clients.
type WebSocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// SimpleHandler handles WebSocket connections using the EventBroadcaster pattern.
type SimpleHandler struct {
	container   *app.Container
	broadcaster *EventBroadcaster
}

// NewSimpleHandler creates a new simple WebSocket handler.
func NewSimpleHandler(container *app.Container, broadcaster *EventBroadcaster) *SimpleHandler {
	return &SimpleHandler{
		container:   container,
		broadcaster: broadcaster,
	}
}

// HandleWebSocket handles WebSocket connections for real-time updates.
func (h *SimpleHandler) HandleWebSocket(c echo.Context) error {
	// Upgrade HTTP connection to WebSocket
	ws, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.container.Logger.Error("failed to upgrade to websocket", zap.Error(err))
		return err
	}
	defer func() { _ = ws.Close() }()

	h.container.Logger.Info("WebSocket connection established", zap.String("remote_addr", c.RealIP()))

	// Set read deadline and message size limits
	_ = ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetReadLimit(maxMessageSize)
	ws.SetPongHandler(func(string) error {
		_ = ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Subscribe to events from broadcaster
	clientChan := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(clientChan)

	// Create context for this connection
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	// Channel to signal when connection should close
	done := make(chan struct{})

	// Start ping ticker
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Start goroutine to send initial data and handle events
	go func() {
		defer close(done)

		// Send initial stats
		if err := h.sendInitialStats(ctx, ws); err != nil {
			h.container.Logger.Warn("failed to send initial stats", zap.Error(err))
		}

		// Load and send recent messages
		if err := h.sendRecentMessages(ctx, ws); err != nil {
			h.container.Logger.Warn("failed to send recent messages", zap.Error(err))
		}

		// Listen for events from broadcaster
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-clientChan:
				if !ok {
					h.container.Logger.Info("broadcaster channel closed")
					return
				}

				if err := h.handleBroadcastEvent(ctx, ws, event); err != nil {
					h.container.Logger.Warn("failed to handle broadcast event", zap.Error(err))
					return
				}
			}
		}
	}()

	// Main loop: handle pings and read messages
	for {
		select {
		case <-done:
			h.container.Logger.Info("WebSocket connection closed")
			return nil
		case <-pingTicker.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.container.Logger.Warn("failed to send ping", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			h.container.Logger.Info("WebSocket connection context cancelled")
			return nil
		}
	}
}

// sendInitialStats sends initial statistics to the WebSocket client.
func (h *SimpleHandler) sendInitialStats(ctx context.Context, ws *websocket.Conn) error {
	window := persistence.TimeWindow{}
	summary, err := h.container.StatsServiceV2.TrafficSummary(ctx, window)
	if err != nil {
		return fmt.Errorf("get traffic summary: %w", err)
	}

	// Send total count
	msg := WebSocketMessage{
		Type: "stats-total",
		Data: map[string]any{
			"total": summary.TotalMessages,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		return fmt.Errorf("write stats-total: %w", err)
	}

	// Send priority breakdown
	msg = WebSocketMessage{
		Type: "stats-priority",
		Data: map[string]any{
			"byPriority": summary.ByPriority,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		return fmt.Errorf("write stats-priority: %w", err)
	}

	// Send type breakdown
	msg = WebSocketMessage{
		Type: "stats-type",
		Data: map[string]any{
			"byType": summary.ByType,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		return fmt.Errorf("write stats-type: %w", err)
	}

	return nil
}

// sendRecentMessages loads and sends recent messages to the WebSocket client.
func (h *SimpleHandler) sendRecentMessages(ctx context.Context, ws *websocket.Conn) error {
	recentTelegrams, err := h.container.QueryService.Recent(ctx, 50)
	if err != nil {
		return fmt.Errorf("get recent telegrams: %w", err)
	}

	if len(recentTelegrams) == 0 {
		return nil
	}

	h.container.Logger.Info("sending recent messages", zap.Int("count", len(recentTelegrams)))

	// Send messages in reverse chronological order (oldest first)
	// This ensures newest messages appear at top
	for i := len(recentTelegrams) - 1; i >= 0; i-- {
		telegram := recentTelegrams[i]
		telegramModel := domain.FromDomain(telegram)
		if telegramModel == nil {
			h.container.Logger.Warn("failed to convert domain telegram to model",
				zap.String("message_id", telegram.MessageID))
			continue
		}

		msg := WebSocketMessage{
			Type: "message",
			Data: telegramModel,
		}
		if err := h.writeWebSocketMessage(ws, msg); err != nil {
			h.container.Logger.Warn("failed to write message", zap.Error(err))
			// Continue sending other messages even if one fails
			continue
		}
	}

	return nil
}

// handleBroadcastEvent processes events from the broadcaster and sends them to the WebSocket client.
func (h *SimpleHandler) handleBroadcastEvent(ctx context.Context, ws *websocket.Conn, event []byte) error {
	// Parse event JSON to determine event type
	var eventData map[string]any
	if err := json.Unmarshal(event, &eventData); err != nil {
		h.container.Logger.Warn("failed to parse event", zap.Error(err), zap.ByteString("raw", event))
		return nil // Don't fail on parse error, just skip
	}

	eventType, ok := eventData["type"].(string)
	if !ok || eventType == "" {
		h.container.Logger.Debug("event missing or invalid type field", zap.Any("event", eventData))
		return nil
	}

	// If it's a telegram_processed event, send the message
	if eventType == "telegram_processed" {
		var telegramData map[string]any
		var found bool

		// First try to get telegram from the "data" field (RedisEventBus wrapped format)
		if dataField, dataOk := eventData["data"].(map[string]any); dataOk {
			if telegram, telegramOk := dataField["telegram"].(map[string]any); telegramOk {
				telegramData = telegram
				found = true
			}
		}

		// If not found in data field, try top level (for backward compatibility)
		if !found {
			if telegram, telegramOk := eventData["telegram"].(map[string]any); telegramOk {
				telegramData = telegram
				found = true
			}
		}

		if found {
			var telegram persistence.Telegram
			telegramBytes, err := json.Marshal(telegramData)
			if err != nil {
				h.container.Logger.Warn("failed to marshal telegram data", zap.Error(err))
				return nil
			}
			if err := json.Unmarshal(telegramBytes, &telegram); err != nil {
				h.container.Logger.Warn("failed to unmarshal telegram", zap.Error(err))
				return nil
			}
			msg := WebSocketMessage{
				Type: "message",
				Data: telegram,
			}
			if err := h.writeWebSocketMessage(ws, msg); err != nil {
				return fmt.Errorf("write message: %w", err)
			}
			h.container.Logger.Debug("sent message event",
				zap.String("message_id", telegram.MessageID),
				zap.String("type", telegram.Type))
		}
	}

	// Always update stats when we receive any event
	window := persistence.TimeWindow{}
	summary, err := h.container.StatsServiceV2.TrafficSummary(ctx, window)
	if err != nil {
		h.container.Logger.Warn("failed to get traffic summary", zap.Error(err))
		return nil // Don't fail on stats error
	}

	// Send total count update
	msg := WebSocketMessage{
		Type: "stats-total",
		Data: map[string]any{
			"total": summary.TotalMessages,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		h.container.Logger.Warn("failed to write stats-total update", zap.Error(err))
		// Continue to try sending other stats
	}

	// Send priority breakdown update
	msg = WebSocketMessage{
		Type: "stats-priority",
		Data: map[string]any{
			"byPriority": summary.ByPriority,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		h.container.Logger.Warn("failed to write stats-priority update", zap.Error(err))
		// Continue processing
	}

	// Send type breakdown update
	msg = WebSocketMessage{
		Type: "stats-type",
		Data: map[string]any{
			"byType": summary.ByType,
		},
	}
	if err := h.writeWebSocketMessage(ws, msg); err != nil {
		h.container.Logger.Warn("failed to write stats-type update", zap.Error(err))
		// Continue processing
	}

	return nil
}

// writeWebSocketMessage writes a WebSocket message to the client.
func (h *SimpleHandler) writeWebSocketMessage(ws *websocket.Conn, msg WebSocketMessage) error {
	_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.WriteJSON(msg)
}

// Register registers the WebSocket route.
func (h *SimpleHandler) Register(e *echo.Echo) {
	e.GET("/ws", h.HandleWebSocket)
}

