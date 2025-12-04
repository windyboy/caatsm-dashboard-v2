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
	subscriber      app.EventSubscriber
	hub             app.WebSocketHubPort
	statsSvc        *StatsService
	statsCounterSvc *StatsCounterService
	logger          *zap.Logger
	// Smart full stats trigger state
	accumulatedCount   int
	lastFullStatsQuery time.Time
	smartTriggerMu     sync.Mutex
}

// NewRealtimeService creates a new real-time service.
func NewRealtimeService(
	subscriber app.EventSubscriber,
	hub app.WebSocketHubPort,
	statsSvc *StatsService,
	statsCounterSvc *StatsCounterService,
	logger *zap.Logger,
) *RealtimeService {
	return &RealtimeService{
		subscriber:         subscriber,
		hub:                hub,
		statsSvc:           statsSvc,
		statsCounterSvc:    statsCounterSvc,
		logger:             logger,
		lastFullStatsQuery: time.Now(),
	}
}

// Start starts the real-time service as a background task.
// Each Subscribe call is blocking, so we need to start them in separate goroutines.
func (s *RealtimeService) Start(ctx context.Context) error {
	// Subscribe to Redis Pub/Sub channel msg:broadcast for telegram messages
	messageHandler := func(ctx context.Context, event app.Event) error {
		return s.handleMessage(ctx, event)
	}

	go func() {
		if err := s.subscriber.Subscribe(ctx, "msg:broadcast", messageHandler); err != nil {
			s.logger.Error("failed to subscribe to msg:broadcast", zap.Error(err))
		}
	}()

	// Subscribe to Redis Pub/Sub channel stats:incremented for incremental stats updates
	statsHandler := func(ctx context.Context, event app.Event) error {
		return s.handleStatsIncrement(ctx, event)
	}

	go func() {
		if err := s.subscriber.Subscribe(ctx, "stats:incremented", statsHandler); err != nil {
			s.logger.Error("failed to subscribe to stats:incremented", zap.Error(err))
		}
	}()

	// Subscribe to Redis Pub/Sub channel health:updated for health status updates
	healthHandler := func(ctx context.Context, event app.Event) error {
		return s.handleHealthUpdate(ctx, event)
	}

	go func() {
		if err := s.subscriber.Subscribe(ctx, "health:updated", healthHandler); err != nil {
			s.logger.Error("failed to subscribe to health:updated", zap.Error(err))
		}
	}()

	s.logger.Info("realtime service started")
	return nil
}

// handleMessage handles a received message event.
func (s *RealtimeService) handleMessage(_ context.Context, event app.Event) error {
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
			return nil // Not an error, just skip this event
		}
	}

	if telegram == nil {
		s.logger.Error("telegram is nil after extraction",
			zap.String("event_type", event.EventType()),
		)
		return fmt.Errorf("could not extract telegram from event")
	}

	// Broadcast to WebSocket Hub
	wsMsg := app.WSMessage{
		Type: ws.MessageTypeMessage,
		Data: telegram,
	}

	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast message",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return err
	}

	// Note: Statistics are now updated via stats:incremented event subscription
	// We no longer trigger stats updates from message events

	return nil
}

// updateStatsAsync asynchronously updates and broadcasts statistics from database (fallback only)
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
		messagesPerSec = 0.0
	}

	// Broadcast stats update
	statsData := &ws.StatsData{
		Total:          stats.TotalMessages,
		ByType:         stats.ByType,
		ActiveRoutes:   stats.ActiveRoutes,
		MessagesPerSec: messagesPerSec,
		TimeWindow:     stats.TimeWindow,
	}
	wsMsg := app.WSMessage{
		Type: ws.MessageTypeStats,
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

// handleStatsIncrement handles incremental statistics update events.
func (s *RealtimeService) handleStatsIncrement(ctx context.Context, event app.Event) error {
	// Extract StatsIncremented event
	var statsEvent *domain.StatsIncremented

	// Try type assertion for known event types
	if incrEvent, ok := event.(*domain.StatsIncremented); ok {
		statsEvent = incrEvent
	} else if incrEvent, ok := event.(domain.StatsIncremented); ok {
		statsEvent = &incrEvent
	} else {
		// Try to extract from generic event (from Redis Pub/Sub)
		var err error
		statsEvent, err = s.extractStatsIncrementFromGenericEvent(event)
		if err != nil {
			s.logger.Warn("could not extract stats increment from event, skipping",
				zap.String("event_type", event.EventType()),
				zap.Error(err),
			)
			return nil // Not an error, just skip this event
		}
	}

	if statsEvent == nil {
		s.logger.Error("stats event is nil after extraction")
		return fmt.Errorf("could not extract stats increment from event")
	}

	// Only broadcast if there are active WebSocket connections
	if s.hub.GetActiveConnections() == 0 {
		return nil
	}

	// Broadcast incremental stats update
	deltaData := &ws.StatsDeltaData{
		Total:     statsEvent.Total,
		ByType:    statsEvent.ByType,
		Route:     statsEvent.Route,
		Timestamp: statsEvent.Timestamp.Format(time.RFC3339),
	}

	wsMsg := app.WSMessage{
		Type: ws.MessageTypeStatsDelta,
		Data: deltaData,
	}

	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast stats delta",
			zap.Error(err),
		)
		return err
	}

	// Update smart trigger state and check if we need full stats sync
	s.smartTriggerMu.Lock()
	s.accumulatedCount++
	shouldSync := s.shouldSyncFullStats()
	if shouldSync {
		s.accumulatedCount = 0
		s.lastFullStatsQuery = time.Now()
	}
	s.smartTriggerMu.Unlock()

	// Trigger full stats sync if needed
	if shouldSync {
		go s.syncFullStatsAsync(ctx)
	}

	return nil
}

// extractStatsIncrementFromGenericEvent extracts StatsIncremented from a generic event.
func (s *RealtimeService) extractStatsIncrementFromGenericEvent(event app.Event) (*domain.StatsIncremented, error) {
	// Try to access GenericEvent's Data field directly
	// GenericEvent from redis_subscriber has Data as interface{}
	if genericEvent, ok := event.(interface{ GetData() interface{} }); ok {
		data := genericEvent.GetData()
		if data != nil {
			// Try to unmarshal data as StatsIncremented
			dataBytes, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("marshal data: %w", err)
			}

			var statsEvent domain.StatsIncremented
			if err := json.Unmarshal(dataBytes, &statsEvent); err != nil {
				return nil, fmt.Errorf("unmarshal stats increment: %w", err)
			}

			return &statsEvent, nil
		}
	}

	// Fallback: try to extract from JSON marshaling
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventJSON, &eventMap); err != nil {
		return nil, fmt.Errorf("unmarshal event JSON: %w", err)
	}

	// Look for data field (case-insensitive)
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

	// Try to unmarshal data as StatsIncremented
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var statsEvent domain.StatsIncremented
	if err := json.Unmarshal(dataBytes, &statsEvent); err != nil {
		return nil, fmt.Errorf("unmarshal stats increment: %w", err)
	}

	return &statsEvent, nil
}

// shouldSyncFullStats determines if we should query full stats from database.
// Returns true if:
// - Accumulated 100 or more incremental updates, OR
// - 5 seconds have passed since last full query
func (s *RealtimeService) shouldSyncFullStats() bool {
	if s.accumulatedCount >= 100 {
		return true
	}

	if time.Since(s.lastFullStatsQuery) > 5*time.Second {
		return true
	}

	return false
}

// syncFullStatsAsync queries full stats from database and broadcasts to sync with counters.
func (s *RealtimeService) syncFullStatsAsync(ctx context.Context) {
	if s.statsCounterSvc == nil {
		// Fall back to StatsService if counter service not available
		if s.statsSvc != nil {
			s.updateStatsAsync(ctx)
		}
		return
	}

	// Get current snapshot from Redis counters
	snapshot, err := s.statsCounterSvc.GetCurrentSnapshot(ctx)
	if err != nil {
		s.logger.Warn("failed to get stats snapshot from counter service",
			zap.Error(err),
		)
		// Fall back to database query
		if s.statsSvc != nil {
			s.updateStatsAsync(ctx)
		}
		return
	}

	// Get messages per second from StatsService
	var messagesPerSec float64
	if s.statsSvc != nil {
		mps, err := s.statsSvc.GetMessagesPerSec(ctx)
		if err != nil {
			messagesPerSec = 0.0
		} else {
			messagesPerSec = mps
		}
	}

	// Broadcast full stats update
	statsData := &ws.StatsData{
		Total:          snapshot.TotalMessages,
		ByType:         snapshot.ByType,
		ActiveRoutes:   snapshot.ActiveRoutes,
		MessagesPerSec: messagesPerSec,
		TimeWindow:     snapshot.TimeWindow,
	}
	wsMsg := app.WSMessage{
		Type: ws.MessageTypeStatsFull,
		Data: statsData,
	}
	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast full stats sync",
			zap.Error(err),
		)
	}
}

// handleHealthUpdate handles health status update events.
func (s *RealtimeService) handleHealthUpdate(_ context.Context, event app.Event) error {
	// Extract HealthUpdated event
	var healthEvent *domain.HealthUpdated

	// Try type assertion for known event types
	if updatedEvent, ok := event.(*domain.HealthUpdated); ok {
		healthEvent = updatedEvent
	} else if updatedEvent, ok := event.(domain.HealthUpdated); ok {
		healthEvent = &updatedEvent
	} else {
		// Try to extract from generic event (from Redis Pub/Sub)
		var err error
		healthEvent, err = s.extractHealthUpdateFromGenericEvent(event)
		if err != nil {
			return nil // Not an error, just skip this event
		}
	}

	if healthEvent == nil {
		return fmt.Errorf("could not extract health update from event")
	}

	// Only broadcast if there are active WebSocket connections
	if s.hub.GetActiveConnections() == 0 {
		return nil
	}

	// Extract health data from event
	healthData, ok := healthEvent.Health.(app.HealthCheckResult)
	if !ok {
		// Try to convert from map if it came from Redis
		healthMap, ok := healthEvent.Health.(map[string]interface{})
		if !ok {
			s.logger.Warn("health event contains invalid health data type",
				zap.String("type", fmt.Sprintf("%T", healthEvent.Health)),
			)
			return nil
		}
		// Convert map to HealthCheckResult
		healthBytes, err := json.Marshal(healthMap)
		if err != nil {
			s.logger.Warn("failed to marshal health map",
				zap.Error(err),
			)
			return nil
		}
		if err := json.Unmarshal(healthBytes, &healthData); err != nil {
			s.logger.Warn("failed to unmarshal health data",
				zap.Error(err),
			)
			return nil
		}
	}

	// Broadcast health update
	healthWSData := &ws.HealthData{
		Status:      healthData.Status,
		Version:     healthData.Version,
		Uptime:      healthData.Uptime,
		Timestamp:   healthData.Timestamp,
		PostgreSQL:  healthData.PostgreSQL,
		Meilisearch: healthData.Meilisearch,
		Redis:       healthData.Redis,
		NATS:        healthData.NATS,
	}

	wsMsg := app.WSMessage{
		Type: ws.MessageTypeHealth,
		Data: healthWSData,
	}

	if err := s.hub.Broadcast(wsMsg); err != nil {
		s.logger.Warn("failed to broadcast health update",
			zap.Error(err),
		)
		return err
	}

	return nil
}

// extractHealthUpdateFromGenericEvent extracts HealthUpdated from a generic event.
func (s *RealtimeService) extractHealthUpdateFromGenericEvent(event app.Event) (*domain.HealthUpdated, error) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventJSON, &eventMap); err != nil {
		return nil, fmt.Errorf("unmarshal event JSON: %w", err)
	}

	// Look for data field
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

	// Try to unmarshal data as HealthUpdated
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var healthEvent domain.HealthUpdated
	if err := json.Unmarshal(dataBytes, &healthEvent); err != nil {
		return nil, fmt.Errorf("unmarshal health update: %w", err)
	}

	return &healthEvent, nil
}
