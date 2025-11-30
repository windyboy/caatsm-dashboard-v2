package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// EventBus defines the interface for publishing events
type EventBus interface {
	// Publish publishes an event to the event bus
	Publish(ctx context.Context, eventType string, data any) error
}

// RedisEventBus implements EventBus using Valkey/Redis pub/sub
type RedisEventBus struct {
	client  redis.UniversalClient
	channel string
}

// NewRedisEventBus creates a new RedisEventBus (works with both Redis and Valkey)
func NewRedisEventBus(client redis.UniversalClient, channel string) EventBus {
	return &RedisEventBus{
		client:  client,
		channel: channel,
	}
}

// Publish publishes an event to Valkey/Redis
func (e *RedisEventBus) Publish(ctx context.Context, eventType string, data any) error {
	eventData := map[string]any{
		"type": eventType,
		"data": data,
	}

	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := e.client.Publish(ctx, e.channel, eventBytes).Err(); err != nil {
		return fmt.Errorf("publish to valkey/redis: %w", err)
	}

	return nil
}

// EventPublisherAdapter adapts EventBus to EventPublisher interface.
// It converts Event to the format expected by EventBus.
type EventPublisherAdapter struct {
	bus EventBus
}

// NewEventPublisherAdapter creates a new adapter that wraps EventBus to implement EventPublisher.
func NewEventPublisherAdapter(bus EventBus) EventPublisher {
	return &EventPublisherAdapter{
		bus: bus,
	}
}

// Publish implements EventPublisher by converting Event to EventBus format.
func (a *EventPublisherAdapter) Publish(ctx context.Context, event Event) error {
	return a.bus.Publish(ctx, event.EventType(), event)
}
