package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventBroadcaster manages WebSocket connections and broadcasts events to all connected clients.
type EventBroadcaster struct {
	clients  map[chan []byte]bool // map value: true = active, false = closed
	mu       sync.RWMutex
	redisCli redis.UniversalClient
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
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
// The caller is responsible for closing the channel after calling Unsubscribe.
// This avoids race conditions where Broadcast might send on a channel that was
// closed while holding the write lock.
func (b *EventBroadcaster) Unsubscribe(client chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Map value: true = open/active, false = closed
	if isOpen, ok := b.clients[client]; ok && isOpen {
		// Mark as closed and remove from map
		b.clients[client] = false
		delete(b.clients, client)
		b.logger.Debug("client unsubscribed", zap.Int("total_clients", len(b.clients)))
	}
	// Note: Caller must close the channel after Unsubscribe returns
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
// This method blocks and should be called as a goroutine: go b.StartRedisListener()
func (b *EventBroadcaster) StartRedisListener() {
	b.wg.Add(1)
	defer b.wg.Done()

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
	defer func() {
		if err := pubsub.Close(); err != nil {
			b.logger.Warn("error closing pubsub", zap.Error(err))
		} else {
			b.logger.Info("pubsub closed")
		}
	}()

	// Subscription is confirmed when Subscribe() returns successfully
	// pubsub.Channel() will automatically handle subscription confirmation
	b.logger.Info("redis listener subscribed and ready", zap.String("channel", "stats:update"))

	ch := pubsub.Channel()

	// Process pub/sub messages
	for {
		select {
		case <-b.ctx.Done():
			b.logger.Info("redis listener stopped")
			return
		case msg, ok := <-ch:
			if !ok {
				b.logger.Info("redis pubsub channel closed, stopping listener")
				return
			}
			if msg != nil {
				// Process messages from the subscribed channel
				if msg.Channel == "stats:update" && msg.Payload != "" {
					// Verify it's valid JSON before broadcasting
					var eventData map[string]any
					if err := json.Unmarshal([]byte(msg.Payload), &eventData); err == nil {
						eventType, _ := eventData["type"].(string)
						b.logger.Info("received event from redis",
							zap.String("channel", msg.Channel),
							zap.String("event_type", eventType),
							zap.Int("payload_size", len(msg.Payload)))

						b.Broadcast([]byte(msg.Payload))
						b.mu.RLock()
						clientCount := len(b.clients)
						b.mu.RUnlock()
						b.logger.Info("broadcasted event to clients",
							zap.String("event_type", eventType),
							zap.Int("client_count", clientCount))
					} else {
						b.logger.Warn("ignored invalid JSON message from redis",
							zap.String("channel", msg.Channel),
							zap.String("payload", msg.Payload),
							zap.Error(err))
					}
				} else {
					b.logger.Debug("ignored message",
						zap.String("channel", msg.Channel),
						zap.Bool("has_payload", msg.Payload != ""),
						zap.String("expected_channel", "stats:update"))
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

	event := map[string]any{
		"type": "stats_update",
		"time": time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return b.redisCli.Publish(ctx, "stats:update", data).Err()
}

// Close shuts down the broadcaster and waits for goroutines to finish.
func (b *EventBroadcaster) Close() {
	b.logger.Info("closing broadcaster")
	b.cancel()

	// Wait for Redis listener goroutine to finish
	b.logger.Info("waiting for Redis listener to stop")
	b.wg.Wait()
	b.logger.Info("Redis listener stopped")

	// Clear all client registrations
	// Note: We do not close channels here to avoid race conditions with Broadcast.
	// Channels should be closed by their owners (callers of Subscribe).
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients = make(map[chan []byte]bool)
	b.logger.Info("all WebSocket clients disconnected")
}
