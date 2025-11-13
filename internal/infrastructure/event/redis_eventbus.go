package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisEventBus implements EventBus using Redis pub/sub
type RedisEventBus struct {
	client redis.UniversalClient
	channel string
}

// NewRedisEventBus creates a new RedisEventBus
func NewRedisEventBus(client redis.UniversalClient, channel string) EventBus {
	return &RedisEventBus{
		client:  client,
		channel: channel,
	}
}

// Publish publishes an event to Redis
func (e *RedisEventBus) Publish(ctx context.Context, eventType string, data interface{}) error {
	eventData := map[string]interface{}{
		"type": eventType,
		"data": data,
	}

	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := e.client.Publish(ctx, e.channel, eventBytes).Err(); err != nil {
		return fmt.Errorf("publish to redis: %w", err)
	}

	return nil
}

