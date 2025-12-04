package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

// Subscriber implements app.EventSubscriber for Redis Pub/Sub.
type Subscriber struct {
	client redis.UniversalClient
	logger *zap.Logger
}

// NewRedisSubscriber creates a new Redis Pub/Sub subscriber.
func NewRedisSubscriber(client redis.UniversalClient, logger *zap.Logger) app.EventSubscriber {
	return &Subscriber{
		client: client,
		logger: logger,
	}
}

// Subscribe subscribes to a Redis Pub/Sub channel and calls the handler for each message.
func (s *Subscriber) Subscribe(ctx context.Context, channel string, handler func(context.Context, app.Event) error) error {
	if s.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	pubsub := s.client.Subscribe(ctx, channel)
	defer func() {
		if err := pubsub.Close(); err != nil {
			s.logger.Warn("failed to close pubsub", zap.Error(err))
		}
	}()

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("subscribe to channel %s: %w", channel, err)
	}

	s.logger.Info("subscribed to redis channel", zap.String("channel", channel))

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("unsubscribing from channel", zap.String("channel", channel))
			return nil
		case msg, ok := <-ch:
			if !ok {
				s.logger.Info("pubsub channel closed", zap.String("channel", channel))
				return nil
			}

			if msg == nil {
				continue
			}

			if msg.Channel != channel {
				continue
			}

			// Parse event from message payload
			event, err := s.parseEvent(msg.Payload)
			if err != nil {
				s.logger.Warn("failed to parse event",
					zap.String("channel", channel),
					zap.String("payload", msg.Payload),
					zap.Error(err),
				)
				continue
			}

			// Call handler
			if err := handler(ctx, event); err != nil {
				s.logger.Warn("handler error",
					zap.String("channel", channel),
					zap.String("event_type", event.EventType()),
					zap.Error(err),
				)
				// Continue processing other messages even if handler fails
			}
		}
	}
}

// parseEvent parses a JSON payload into an app.Event.
func (s *Subscriber) parseEvent(payload string) (app.Event, error) {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &eventData); err != nil {
		return nil, fmt.Errorf("unmarshal event data: %w", err)
	}

	// Check event type
	eventType, ok := eventData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event type")
	}

	// For now, we'll create a generic event wrapper
	// In a more sophisticated implementation, we could deserialize to specific event types
	occurredAt := time.Now()
	if occurredAtStr, ok := eventData["occurred_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, occurredAtStr); err == nil {
			occurredAt = t
		}
	}
	return &GenericEvent{
		Type:       eventType,
		Data:       eventData["data"],
		occurredAt: occurredAt,
	}, nil
}

// GenericEvent is a generic event implementation for Redis Pub/Sub messages.
type GenericEvent struct {
	Type       string
	Data       interface{}
	occurredAt time.Time
}

func (e *GenericEvent) EventType() string {
	return e.Type
}

func (e *GenericEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// GetData returns the event data for easier access.
func (e *GenericEvent) GetData() interface{} {
	return e.Data
}


