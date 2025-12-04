package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap"
)

// RealtimeService handles real-time message broadcasting to WebSocket clients.
type RealtimeService struct {
	subscriber          app.EventSubscriber
	hub                 app.WebSocketHubPort
	statsSvc            *StatsService
	logger              *zap.Logger
	lastStatsUpdate     time.Time
	statsUpdateMu       sync.Mutex
	statsUpdateDebounce time.Duration
}

// NewRealtimeService creates a new real-time service.
func NewRealtimeService(
	subscriber app.EventSubscriber,
	hub app.WebSocketHubPort,
	statsSvc *StatsService,
	logger *zap.Logger,
) *RealtimeService {
	return &RealtimeService{
		subscriber:          subscriber,
		hub:                 hub,
		statsSvc:            statsSvc,
		logger:              logger,
		statsUpdateDebounce: 500 * time.Millisecond, // Debounce updates to avoid too frequent refreshes
	}
}

// Start starts the real-time service as a background task.
func (s *RealtimeService) Start(ctx context.Context) error {
	// Subscribe to Redis Pub/Sub channel msg:broadcast
	handler := func(ctx context.Context, event app.Event) error {
		return s.handleMessage(ctx, event)
	}

	if err := s.subscriber.Subscribe(ctx, "msg:broadcast", handler); err != nil {
		return fmt.Errorf("subscribe to msg:broadcast: %w", err)
	}

	// Start periodic stats update goroutine
	go s.startPeriodicStatsUpdate(ctx)

	s.logger.Info("realtime service started",
		zap.String("channel", "msg:broadcast"),
	)

	return nil
}

// startPeriodicStatsUpdate starts a goroutine that periodically updates and broadcasts statistics.
func (s *RealtimeService) startPeriodicStatsUpdate(ctx context.Context) {
	// Update stats immediately on startup
	if s.statsSvc != nil {
		s.updateStatsAsync(ctx)
	}

	// Then update every 1 second for more real-time updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("periodic stats update stopped", zap.Error(ctx.Err()))
			return
		case <-ticker.C:
			// Only update if there are active WebSocket connections
			if s.hub.GetActiveConnections() > 0 && s.statsSvc != nil {
				s.requestStatsUpdate(ctx)
			}
		}
	}
}

// handleMessage handles a received message event.
func (s *RealtimeService) handleMessage(ctx context.Context, event app.Event) error {
	// Extract telegram from event
	var telegram *app.Telegram

	// Try type assertion for known event types
	if persistedEvent, ok := event.(*domain.TelegramPersisted); ok {
		telegram = persistedEvent.Telegram
	} else if persistedEvent, ok := event.(domain.TelegramPersisted); ok {
		telegram = persistedEvent.Telegram
	} else {
		// Try to extract from generic event (from Redis Pub/Sub)
		var err error
		telegram, err = s.extractTelegramFromGenericEvent(event)
		if err != nil {
			s.logger.Debug("could not extract telegram from event, skipping",
				zap.String("event_type", event.EventType()),
				zap.Error(err),
			)
			return nil // Not an error, just skip this event
		}
	}

	if telegram == nil {
		return fmt.Errorf("could not extract telegram from event")
	}

	// Broadcast to WebSocket Hub
	wsMsg := app.WSMessage{
		Type: "message",
		Data: telegram,
	}

	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast message",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("message broadcasted to websocket clients",
		zap.String("message_id", telegram.MessageID),
		zap.Int("active_connections", s.hub.GetActiveConnections()),
	)

	// Asynchronously update statistics with debouncing
	if s.statsSvc != nil {
		s.requestStatsUpdate(context.Background())
	}

	return nil
}

// requestStatsUpdate requests a stats update with debouncing to avoid too frequent updates
func (s *RealtimeService) requestStatsUpdate(ctx context.Context) {
	s.statsUpdateMu.Lock()
	now := time.Now()
	lastUpdate := s.lastStatsUpdate
	s.statsUpdateMu.Unlock()

	// If last update was too recent, skip this update request
	if now.Sub(lastUpdate) < s.statsUpdateDebounce {
		return
	}

	// Update last update time
	s.statsUpdateMu.Lock()
	s.lastStatsUpdate = now
	s.statsUpdateMu.Unlock()

	go s.updateStatsAsync(ctx)
}

// updateStatsAsync asynchronously updates and broadcasts statistics
func (s *RealtimeService) updateStatsAsync(ctx context.Context) {
	// Use last 24 hours as time window to match frontend expectation
	window := TimeWindow{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	stats, err := s.statsSvc.GetStats(ctx, window)
	if err != nil {
		s.logger.Warn("failed to get stats for update",
			zap.Error(err),
		)
		return
	}

	// Get messages per second
	messagesPerSec, err := s.statsSvc.GetMessagesPerSec(ctx)
	if err != nil {
		s.logger.Debug("failed to get messages per sec, using 0",
			zap.Error(err),
		)
		messagesPerSec = 0.0
	}

	// Broadcast stats update
	statsData := &ws.StatsData{
		Total:         stats.TotalMessages,
		ByType:        stats.ByType,
		ActiveRoutes:  stats.ActiveRoutes,
		MessagesPerSec: messagesPerSec,
		TimeWindow:    stats.TimeWindow,
	}
	wsMsg := app.WSMessage{
		Type: "stats",
		Data: statsData,
	}
	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast stats update",
			zap.Error(err),
		)
	}
}

// extractTelegramFromGenericEvent extracts telegram from a generic event (from Redis Pub/Sub).
func (s *RealtimeService) extractTelegramFromGenericEvent(event app.Event) (*app.Telegram, error) {
	// The event from Redis Pub/Sub will be a GenericEvent with Data field
	// EventPublisherAdapter publishes events as: {"type": "event.type", "data": <event object>}
	// So for TelegramPersisted, the data will be: {"Telegram": {...}, "At": "..."}
	
	// Try to extract Data field using JSON marshaling
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventJSON, &eventMap); err != nil {
		return nil, fmt.Errorf("unmarshal event JSON: %w", err)
	}

	// Look for telegram in the data field (case-insensitive)
	var data interface{}
	var found bool
	for key, val := range eventMap {
		if key == "Data" || key == "data" {
			data = val
			found = true
			break
		}
	}
	
	if !found {
		return nil, fmt.Errorf("no data field found in event")
	}

	// The data should be the event object (e.g., TelegramPersisted)
	// Extract Telegram field from it
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Look for Telegram field (case-insensitive)
		var telegramData interface{}
		for key, val := range dataMap {
			if key == "Telegram" || key == "telegram" {
				telegramData = val
				break
			}
		}
		
		if telegramData != nil {
			return parseEventData(telegramData)
		}
	}

	// Fallback: try to parse data directly as telegram
	return parseEventData(data)
}

// parseEventData attempts to parse event data from JSON.
func parseEventData(data interface{}) (*app.Telegram, error) {
	// If data is already a telegram, return it
	if tg, ok := data.(*app.Telegram); ok {
		return tg, nil
	}

	// Try to unmarshal from JSON bytes
	if jsonBytes, ok := data.([]byte); ok {
		var telegram app.Telegram
		if err := json.Unmarshal(jsonBytes, &telegram); err != nil {
			return nil, fmt.Errorf("unmarshal telegram: %w", err)
		}
		return &telegram, nil
	}

	// Try to unmarshal from map
	if dataMap, ok := data.(map[string]interface{}); ok {
		jsonBytes, err := json.Marshal(dataMap)
		if err != nil {
			return nil, fmt.Errorf("marshal data map: %w", err)
		}
		var telegram app.Telegram
		if err := json.Unmarshal(jsonBytes, &telegram); err != nil {
			return nil, fmt.Errorf("unmarshal telegram: %w", err)
		}
		return &telegram, nil
	}

	return nil, fmt.Errorf("unable to extract telegram from data type: %T", data)
}

