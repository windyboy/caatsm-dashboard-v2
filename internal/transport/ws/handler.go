package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Validate origin properly (Phase 3.4)
		return true
	},
}

// Handler wraps the WebSocket hub and provides HTTP handlers.
type Handler struct {
	hub         *ws.Hub
	logger      *zap.Logger
	statsService interface {
		TrafficSummary(ctx context.Context, window interface{}) (interface{}, error)
	}
	queryService interface {
		Recent(ctx context.Context, limit int) ([]*domain.Telegram, error)
	}
	redisCli redis.UniversalClient
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	hub *ws.Hub,
	logger *zap.Logger,
	statsService interface {
		TrafficSummary(ctx context.Context, window interface{}) (interface{}, error)
	},
	queryService interface {
		Recent(ctx context.Context, limit int) ([]*domain.Telegram, error)
	},
	redisCli redis.UniversalClient,
) *Handler {
	return &Handler{
		hub:          hub,
		logger:       logger,
		statsService: statsService,
		queryService: queryService,
		redisCli:     redisCli,
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *Handler) HandleWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", zap.Error(err))
		return err
	}

	h.logger.Info("WebSocket connection established", zap.String("remote_addr", c.RealIP()))

	// Create client and register with hub
	client := ws.NewClient(conn, h.hub, h.logger)
	h.hub.Register(client)

	// Send initial data
	ctx := c.Request().Context()
	if err := h.sendInitialData(ctx, client); err != nil {
		h.logger.Warn("failed to send initial data", zap.Error(err))
	}

	// Start Redis listener if available
	if h.redisCli != nil {
		go h.startRedisListener()
	}

	return nil
}

// sendInitialData sends initial stats and recent messages.
func (h *Handler) sendInitialData(ctx context.Context, client *ws.Client) error {
	// Send initial stats
	if err := h.sendInitialStats(ctx, client); err != nil {
		return fmt.Errorf("send initial stats: %w", err)
	}

	// Send recent messages
	if err := h.sendRecentMessages(ctx, client); err != nil {
		return fmt.Errorf("send recent messages: %w", err)
	}

	return nil
}

// sendInitialStats sends initial statistics.
func (h *Handler) sendInitialStats(ctx context.Context, client *ws.Client) error {
	window := persistence.TimeWindow{}
	summaryResult, err := h.statsService.TrafficSummary(ctx, window)
	if err != nil {
		return fmt.Errorf("get traffic summary: %w", err)
	}

	// Type assert to get the summary
	summary, ok := summaryResult.(*persistence.TrafficSummary)
	if !ok {
		return fmt.Errorf("unexpected summary type")
	}

	// Send total count
	msg := &ws.Message{
		Type: ws.MessageTypeStatsTotal,
		Data: map[string]any{
			"total": summary.TotalMessages,
		},
	}
	if !client.SendMessageJSON(msg) {
		return fmt.Errorf("failed to send stats-total")
	}

	// Send priority breakdown
	msg = &ws.Message{
		Type: ws.MessageTypeStatsPriority,
		Data: map[string]any{
			"byPriority": summary.ByPriority,
		},
	}
	if !client.SendMessageJSON(msg) {
		return fmt.Errorf("failed to send stats-priority")
	}

	// Send type breakdown
	msg = &ws.Message{
		Type: ws.MessageTypeStatsType,
		Data: map[string]any{
			"byType": summary.ByType,
		},
	}
	if !client.SendMessageJSON(msg) {
		return fmt.Errorf("failed to send stats-type")
	}

	return nil
}

// sendRecentMessages sends recent telegrams.
func (h *Handler) sendRecentMessages(ctx context.Context, client *ws.Client) error {
	recentTelegrams, err := h.queryService.Recent(ctx, 50)
	if err != nil {
		return fmt.Errorf("get recent telegrams: %w", err)
	}

	if len(recentTelegrams) == 0 {
		return nil
	}

	// Send messages in reverse chronological order
	for i := len(recentTelegrams) - 1; i >= 0; i-- {
		telegram := recentTelegrams[i]
		telegramModel := domain.FromDomain(telegram)
		if telegramModel == nil {
			h.logger.Warn("failed to convert domain telegram to model",
				zap.String("message_id", telegram.MessageID))
			continue
		}

		msg := &ws.Message{
			Type: ws.MessageTypeMessage,
			Data: telegramModel,
		}
		if !client.SendMessageJSON(msg) {
			h.logger.Warn("failed to send message", zap.String("message_id", telegram.MessageID))
			continue
		}
	}

	return nil
}

// startRedisListener listens to Redis pub/sub and broadcasts events.
func (h *Handler) startRedisListener() {
	if h.redisCli == nil {
		return
	}

	ctx := context.Background()
	pubsub := h.redisCli.Subscribe(ctx, "stats:update")
	defer func() {
		_ = pubsub.Close()
	}()

	ch := pubsub.Channel()

	for msg := range ch {
		if msg != nil && msg.Channel == "stats:update" && msg.Payload != "" {
			// Verify it's valid JSON
			var eventData map[string]any
			if err := json.Unmarshal([]byte(msg.Payload), &eventData); err == nil {
				// Broadcast to all clients via hub
				h.hub.Broadcast([]byte(msg.Payload))
			} else {
				h.logger.Warn("ignored invalid JSON message from redis",
					zap.String("channel", msg.Channel),
					zap.Error(err))
			}
		}
	}
}

