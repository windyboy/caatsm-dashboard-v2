package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// Handler wraps the WebSocket hub and provides HTTP handlers.
type Handler struct {
	hub          *ws.Hub
	logger       *zap.Logger
	statsService interface {
		TrafficSummary(ctx context.Context, window interface{}) (interface{}, error)
	}
	queryService interface {
		Recent(ctx context.Context, limit int) ([]*domain.Telegram, error)
	}
	redisCli          redis.UniversalClient
	upgrader          websocket.Upgrader
	redisListenerOnce sync.Once
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
	allowedOrigins []string,
) *Handler {
	// Prepare allowed origins: trim whitespace and filter empty entries
	origins := prepareAllowedOrigins(allowedOrigins)

	return &Handler{
		hub:          hub,
		logger:       logger,
		statsService: statsService,
		queryService: queryService,
		redisCli:     redisCli,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     makeCheckOriginFunc(origins),
		},
	}
}

// prepareAllowedOrigins trims whitespace and filters empty entries from the origins list.
func prepareAllowedOrigins(origins []string) []string {
	if len(origins) == 0 {
		// Safe default for development
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}

	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	// If all entries were empty, use safe defaults
	if len(result) == 0 {
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}

	return result
}

// normalizeOrigin normalizes an origin URL for case-insensitive comparison per RFC 6454.
// It lowercases the scheme and host, and removes default ports (80 for http, 443 for https).
func normalizeOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	// Parse the origin URL
	u, err := url.Parse(origin)
	if err != nil {
		// If parsing fails, return the original (will likely fail matching anyway)
		return origin
	}

	// If scheme or host is empty, the URL is invalid - return original
	if u.Scheme == "" || u.Host == "" {
		return origin
	}

	// Normalize scheme and host to lowercase
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)

	// Remove default ports
	if scheme == "http" && strings.HasSuffix(host, ":80") {
		host = strings.TrimSuffix(host, ":80")
	} else if scheme == "https" && strings.HasSuffix(host, ":443") {
		host = strings.TrimSuffix(host, ":443")
	}

	// Reconstruct the origin
	return scheme + "://" + host
}

// makeCheckOriginFunc creates a CheckOrigin function that validates against allowed origins.
// Per RFC 6454, origin comparisons are case-insensitive for scheme and host.
func makeCheckOriginFunc(allowedOrigins []string) func(*http.Request) bool {
	// Pre-normalize all allowed origins for efficiency
	normalizedAllowed := make([]string, len(allowedOrigins))
	for i, origin := range allowedOrigins {
		normalizedAllowed[i] = normalizeOrigin(origin)
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow same-origin requests (empty Origin header)
		if origin == "" {
			return true
		}

		// Normalize the request origin
		normalizedOrigin := normalizeOrigin(origin)

		// Check against allowed origins
		for _, allowed := range normalizedAllowed {
			if normalizedOrigin == allowed {
				return true
			}
		}
		return false
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *Handler) HandleWebSocket(c echo.Context) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", zap.Error(err))
		return err
	}

	h.logger.Info("WebSocket connection established", zap.String("remote_addr", c.RealIP()))

	// Create client and register with hub
	client := ws.NewClient(conn, h.hub, h.logger)
	h.hub.Register(client)

	// Send initial data using request context (safe for this quick operation)
	ctx := c.Request().Context()
	if err := h.sendInitialData(ctx, client); err != nil {
		h.logger.Warn("failed to send initial data", zap.Error(err))
	}

	// Create a detached context tied to the hub's lifecycle (not the HTTP request)
	// This ensures the client pumps and Redis listener persist after the HTTP upgrade
	clientCtx := h.hub.Context()

	// Start read/write pump goroutines for the WebSocket client
	// readPump: handles inbound messages and triggers cleanup on error/cancellation
	go client.ReadPump(clientCtx)

	// writePump: sends messages from the send channel, periodic pings, and closes connection when done
	go client.WritePump(clientCtx)

	// Start Redis listener once with hub context (not request context) so it doesn't exit prematurely
	// Use sync.Once to ensure only one listener is started across all client connections
	if h.redisCli != nil {
		h.redisListenerOnce.Do(func() {
			h.logger.Info("starting Redis listener with hub lifecycle context")
			go h.startRedisListener(clientCtx)
		})
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
			return fmt.Errorf("failed to send message: buffer full")
		}
	}

	return nil
}

func (h *Handler) startRedisListener(ctx context.Context) {
	if h.redisCli == nil {
		return
	}

	pubsub := h.redisCli.Subscribe(ctx, "stats:update")
	defer func() {
		_ = pubsub.Close()
	}()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
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
}
