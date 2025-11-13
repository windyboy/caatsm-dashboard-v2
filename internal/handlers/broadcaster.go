package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventBroadcaster manages SSE connections and broadcasts events to all connected clients.
type EventBroadcaster struct {
	clients  map[chan []byte]bool
	mu       sync.RWMutex
	redisCli redis.UniversalClient
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewEventBroadcaster creates a new event broadcaster.
func NewEventBroadcaster(redisCli redis.UniversalClient, logger *zap.Logger) *EventBroadcaster {
	ctx, cancel := context.WithCancel(context.Background())
	if redisCli == nil {
		logger.Warn("creating EventBroadcaster with nil Redis client")
	} else {
		logger.Info("creating EventBroadcaster with Redis client")
	}
	return &EventBroadcaster{
		clients:  make(map[chan []byte]bool),
		redisCli: redisCli,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Subscribe adds a new client to the broadcaster.
func (b *EventBroadcaster) Subscribe() chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	client := make(chan []byte, 10)
	b.clients[client] = true
	b.logger.Debug("client subscribed", zap.Int("total_clients", len(b.clients)))
	return client
}

// Unsubscribe removes a client from the broadcaster.
func (b *EventBroadcaster) Unsubscribe(client chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.clients[client]; ok {
		close(client)
		delete(b.clients, client)
		b.logger.Debug("client unsubscribed", zap.Int("total_clients", len(b.clients)))
	}
}

// Broadcast sends an event to all connected clients.
func (b *EventBroadcaster) Broadcast(event []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		select {
		case client <- event:
		default:
			// Client channel is full, skip
			b.logger.Warn("client channel full, skipping broadcast")
		}
	}
}

// StartRedisListener starts listening to Redis pub/sub for stats update events.
func (b *EventBroadcaster) StartRedisListener() {
	b.logger.Info("StartRedisListener called")
	if b.redisCli == nil {
		b.logger.Warn("redis client not available, skipping redis listener")
		return
	}

	// Test Redis connection
	b.logger.Info("testing Redis connection...")
	if err := b.redisCli.Ping(b.ctx).Err(); err != nil {
		b.logger.Error("redis ping failed", zap.Error(err))
		return
	}
	b.logger.Info("redis connection verified")

	b.logger.Info("subscribing to Redis channel", zap.String("channel", "stats:update"))
	pubsub := b.redisCli.Subscribe(b.ctx, "stats:update")
	defer pubsub.Close()

	b.logger.Info("started redis listener for stats updates", zap.String("channel", "stats:update"))

	// Wait for subscription confirmation
	// The first message from Channel() is the subscription confirmation
	ch := pubsub.Channel()
	confirmationTimeout := time.NewTimer(5 * time.Second)
	select {
	case msg := <-ch:
		if msg != nil {
			// This is the subscription confirmation message
			b.logger.Info("subscription confirmed",
				zap.String("channel", msg.Channel),
				zap.String("kind", msg.Payload))
		}
		confirmationTimeout.Stop()
	case <-confirmationTimeout.C:
		b.logger.Warn("subscription confirmation timeout, continuing anyway")
	case <-b.ctx.Done():
		return
	}

	b.logger.Info("redis listener is now waiting for messages...")

	// Now process actual messages
	for {
		select {
		case <-b.ctx.Done():
			b.logger.Info("redis listener stopped")
			return
		case msg := <-ch:
			if msg != nil {
				// Filter out subscription confirmation messages
				// Actual messages have the channel name and non-empty payload
				if msg.Channel == "stats:update" && msg.Payload != "" {
					// Try to parse as JSON to verify it's a real message (not subscription count)
					var testData map[string]interface{}
					if err := json.Unmarshal([]byte(msg.Payload), &testData); err == nil {
						b.logger.Info("received stats update event from redis",
							zap.String("channel", msg.Channel),
							zap.String("payload", msg.Payload))
						b.Broadcast([]byte(msg.Payload))
						b.mu.RLock()
						clientCount := len(b.clients)
						b.mu.RUnlock()
						b.logger.Info("broadcasted event to clients", zap.Int("client_count", clientCount))
					} else {
						b.logger.Debug("ignored non-JSON message (likely subscription confirmation)",
							zap.String("payload", msg.Payload),
							zap.Error(err))
					}
				} else {
					b.logger.Debug("ignored message",
						zap.String("channel", msg.Channel),
						zap.String("payload", msg.Payload))
				}
			}
		}
	}
}

// PublishStatsUpdate publishes a stats update event to Redis.
func (b *EventBroadcaster) PublishStatsUpdate(ctx context.Context) error {
	if b.redisCli == nil {
		return fmt.Errorf("redis client not available")
	}

	event := map[string]interface{}{
		"type": "stats_update",
		"time": time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return b.redisCli.Publish(ctx, "stats:update", data).Err()
}

// Close shuts down the broadcaster.
func (b *EventBroadcaster) Close() {
	b.cancel()
	b.mu.Lock()
	defer b.mu.Unlock()

	for client := range b.clients {
		close(client)
	}
	b.clients = make(map[chan []byte]bool)
}
