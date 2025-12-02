package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// RealtimeService handles real-time message broadcasting to WebSocket clients.
type RealtimeService struct {
	subscriber app.EventSubscriber
	hub        app.WebSocketHubPort
	logger     *zap.Logger
}

// NewRealtimeService creates a new real-time service.
func NewRealtimeService(
	subscriber app.EventSubscriber,
	hub app.WebSocketHubPort,
	logger *zap.Logger,
) *RealtimeService {
	return &RealtimeService{
		subscriber: subscriber,
		hub:        hub,
		logger:     logger,
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

	s.logger.Info("realtime service started",
		zap.String("channel", "msg:broadcast"),
	)

	return nil
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

	return nil
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

