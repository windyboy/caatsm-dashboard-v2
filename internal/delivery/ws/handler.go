package ws

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB - maximum message size allowed from peer
)

// Handler wraps the WebSocket hub and provides HTTP handlers.
type Handler struct {
	hub          *ws.Hub
	logger       *zap.Logger
	authConfig   config.AuthConfig
	statsService interface {
		TrafficSummary(ctx context.Context, window interface{}) (interface{}, error)
	}
	queryService interface {
		Recent(ctx context.Context, limit int) ([]*app.Telegram, error)
	}
	redisCli          redis.UniversalClient
	upgrader          websocket.Upgrader
	redisListenerOnce sync.Once
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	hub *ws.Hub,
	logger *zap.Logger,
	authConfig config.AuthConfig,
	statsService interface {
		TrafficSummary(ctx context.Context, window interface{}) (interface{}, error)
	},
	queryService interface {
		Recent(ctx context.Context, limit int) ([]*app.Telegram, error)
	},
	redisCli redis.UniversalClient,
	allowedOrigins []string,
) *Handler {
	// Prepare allowed origins: trim whitespace and filter empty entries
	origins := prepareAllowedOrigins(allowedOrigins)

	return &Handler{
		hub:          hub,
		logger:       logger,
		authConfig:   authConfig,
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
		// If parsing fails, return empty string to deny by default (fail-secure)
		return ""
	}

	// If scheme or host is empty, the URL is invalid - deny by default
	if u.Scheme == "" || u.Host == "" {
		return ""
	}

	// Normalize scheme and host to lowercase
	scheme := strings.ToLower(u.Scheme)

	// Only allow http and https schemes for WebSocket origins (fail-secure)
	if scheme != "http" && scheme != "https" {
		return ""
	}

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
	// Validate JWT token if authentication is enabled
	if h.authConfig.JWTSecret != "" {
		tokenParam := c.QueryParam("token")
		if tokenParam == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token parameter")
		}

		token, err := jwt.Parse(tokenParam, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(h.authConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", zap.Error(err))
		return err
	}

	// Set message size limit for security (use hub config if available, otherwise default)
	msgSizeLimit := int64(maxMessageSize)
	if h.hub.Config.Timeouts != nil && h.hub.Config.Timeouts.MaxMessageSize > 0 {
		msgSizeLimit = h.hub.Config.Timeouts.MaxMessageSize
	}
	conn.SetReadLimit(msgSizeLimit)

	h.logger.Info("WebSocket connection established", zap.String("remote_addr", c.RealIP()))

	// Create client and register with hub
	client := ws.NewClient(conn, h.hub, h.logger)
	if err := h.hub.Register(client); err != nil {
		h.logger.Warn("failed to register client", zap.Error(err))
		return err
	}

	// Create a detached context tied to the hub's lifecycle (not the HTTP request)
	// This ensures the client pumps and Redis listener persist after the HTTP upgrade
	clientCtx := h.hub.Context()

	// Start read/write pump goroutines for the WebSocket client
	// readPump: handles inbound messages and triggers cleanup on error/cancellation
	go client.ReadPump(clientCtx)

	// writePump: sends messages from the send channel, periodic pings, and closes connection when done
	go client.WritePump(clientCtx)

	// Send initial data after WritePump is started to ensure messages can be sent
	// Use goroutine to avoid blocking and give WritePump time to start
	go func() {
		// Small delay to ensure WritePump has started processing
		time.Sleep(10 * time.Millisecond)

		// Use clientCtx (hub context) instead of request context
		// Request context is canceled after WebSocket upgrade, causing "context canceled" errors
		if err := h.sendInitialData(clientCtx, client); err != nil {
			h.logger.Warn("failed to send initial data", zap.Error(err))
		}
	}()

	// Note: Redis listener is now handled by RealtimeService to avoid duplicate message handling.
	// Removed: h.startRedisListener() - RealtimeService handles all Redis Pub/Sub subscriptions.

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
	// Use last 24 hours as default time window to match frontend expectation
	window := app.TimeWindow{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	summaryResult, err := h.statsService.TrafficSummary(ctx, window)
	if err != nil {
		return fmt.Errorf("get traffic summary: %w", err)
	}

	// Type assert to get the summary
	summary, ok := summaryResult.(*app.TrafficSummary)
	if !ok {
		return fmt.Errorf("unexpected summary type")
	}

	// Get messages per second (try to get from adapter if it supports it)
	var messagesPerSec float64
	if rateService, ok := h.statsService.(interface {
		GetMessagesPerSec(ctx context.Context) (float64, error)
	}); ok {
		if rate, err := rateService.GetMessagesPerSec(ctx); err == nil {
			messagesPerSec = rate
		}
	}

	// Send unified stats message
	statsData := &ws.StatsData{
		Total:          summary.TotalMessages,
		ByType:         summary.ByType,
		ActiveRoutes:   summary.ActiveRoutes,
		MessagesPerSec: messagesPerSec,
		TimeWindow:     summary.TimeWindow,
	}
	msg := &ws.Message{
		Type: ws.MessageTypeStats,
		Data: statsData,
	}
	if !client.SendMessageJSON(msg) {
		return fmt.Errorf("failed to send stats")
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

		msg := &ws.Message{
			Type: ws.MessageTypeMessage,
			Data: telegram, // app.Telegram now has JSON tags
		}

		if !client.SendMessageJSON(msg) {
			return fmt.Errorf("failed to send message: buffer full")
		}
	}

	return nil
}
