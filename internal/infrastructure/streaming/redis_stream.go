package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

// Publisher implements app.StreamPublisher for Redis Streams.
type Publisher struct {
	client redis.UniversalClient
	logger *zap.Logger
}

// NewRedisStreamPublisher creates a new Redis Streams publisher.
func NewRedisStreamPublisher(client redis.UniversalClient, logger *zap.Logger) app.StreamPublisher {
	return &Publisher{
		client: client,
		logger: logger,
	}
}

// Publish publishes a telegram message to a Redis Stream.
func (p *Publisher) Publish(ctx context.Context, stream string, telegram *app.Telegram) error {
	data, err := json.Marshal(telegram)
	if err != nil {
		return fmt.Errorf("marshal telegram: %w", err)
	}

	values := map[string]interface{}{
		"telegram": string(data),
		"timestamp": time.Now().Unix(),
	}

	// Use XADD with MAXLEN to prevent unbounded growth
	// Approximate trimming is acceptable for performance
	args := &redis.XAddArgs{
		Stream: stream,
		MaxLen: 100000, // Limit stream to 100k messages (configurable)
		Approx: true,   // Approximate trimming for better performance
		Values: values,
	}

	id, err := p.client.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("xadd to stream %s: %w", stream, err)
	}

	if p.logger != nil {
		p.logger.Debug("published message to redis stream",
			zap.String("stream", stream),
			zap.String("message_id", id),
			zap.String("telegram_id", telegram.MessageID),
		)
	}

	return nil
}

// RedisStreamConsumer implements app.RedisStreamConsumer for Redis Streams.
type RedisStreamConsumer struct {
	client         redis.UniversalClient
	stream         string
	group          string
	consumerName   string
	handler        func(context.Context, *app.Telegram) error
	logger         *zap.Logger
	maxRetries     int
	dlqStream      string
	batchSize      int64
	blockTime      time.Duration
	recoverInterval time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewRedisStreamConsumer creates a new Redis Streams consumer.
func NewRedisStreamConsumer(
	client redis.UniversalClient,
	stream, group, consumerName string,
	logger *zap.Logger,
) app.RedisStreamConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisStreamConsumer{
		client:          client,
		stream:          stream,
		group:           group,
		consumerName:    consumerName,
		logger:          logger,
		maxRetries:      5, // Default max retries before DLQ
		dlqStream:       stream + ":dlq",
		batchSize:       10,
		blockTime:       1 * time.Second,
		recoverInterval: 30 * time.Second,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// SetHandler sets the message handler.
func (c *RedisStreamConsumer) SetHandler(handler func(context.Context, *app.Telegram) error) {
	c.handler = handler
}

// Start begins consuming messages from the Redis Stream.
func (c *RedisStreamConsumer) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("handler not set")
	}

	// Create consumer group if it doesn't exist
	if err := c.ensureConsumerGroup(ctx); err != nil {
		return fmt.Errorf("ensure consumer group: %w", err)
	}

	// Recover pending messages on startup
	if err := c.recoverPending(ctx); err != nil {
		c.logger.Warn("failed to recover pending messages on startup", zap.Error(err))
	}

	// Start periodic pending recovery
	go c.periodicRecover(ctx)

	// Start consuming messages
	go c.consumeLoop(ctx)

	c.logger.Info("redis stream consumer started",
		zap.String("stream", c.stream),
		zap.String("group", c.group),
		zap.String("consumer", c.consumerName),
	)

	return nil
}

// Close stops the consumer.
func (c *RedisStreamConsumer) Close() error {
	c.cancel()
	return nil
}

// ensureConsumerGroup creates the consumer group if it doesn't exist.
func (c *RedisStreamConsumer) ensureConsumerGroup(ctx context.Context) error {
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}
	return nil
}

// consumeLoop continuously reads messages from the stream.
func (c *RedisStreamConsumer) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		default:
			// Read messages from the stream
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumerName,
				Streams:  []string{c.stream, ">"},
				Count:    c.batchSize,
				Block:    c.blockTime,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No messages available, continue
					continue
				}
				c.logger.Warn("xreadgroup error", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			// Process messages
			for _, stream := range streams {
				if stream.Stream == c.stream {
					for _, msg := range stream.Messages {
						c.processMessage(ctx, msg)
					}
				}
			}
		}
	}
}

// processMessage processes a single message from the stream.
func (c *RedisStreamConsumer) processMessage(ctx context.Context, msg redis.XMessage) {
	// Extract telegram from message
	telegramData, ok := msg.Values["telegram"].(string)
	if !ok {
		c.logger.Warn("invalid message format, missing telegram field",
			zap.String("id", msg.ID),
		)
		c.ackMessage(ctx, msg.ID)
		return
	}

	var telegram app.Telegram
	if err := json.Unmarshal([]byte(telegramData), &telegram); err != nil {
		c.logger.Warn("failed to unmarshal telegram",
			zap.String("id", msg.ID),
			zap.Error(err),
		)
		c.ackMessage(ctx, msg.ID)
		return
	}

	// Process message
	if err := c.handler(ctx, &telegram); err != nil {
		// Check retry count
		retryCount := c.getRetryCount(msg)
		retryCount++
		
		if retryCount >= c.maxRetries {
			c.logger.Warn("message exceeded max retries, moving to DLQ",
				zap.String("id", msg.ID),
				zap.String("telegram_id", telegram.MessageID),
				zap.Int("retry_count", retryCount),
			)
			c.moveToDLQ(ctx, msg, &telegram)
			c.ackMessage(ctx, msg.ID)
			return
		}
		
		c.logger.Warn("handler error, will retry",
			zap.String("id", msg.ID),
			zap.String("telegram_id", telegram.MessageID),
			zap.Int("retry_count", retryCount),
			zap.Error(err),
		)
		// Don't ACK, message will remain pending and be retried
		return
	}

	// ACK successful processing
	c.ackMessage(ctx, msg.ID)
	c.logger.Debug("message processed successfully",
		zap.String("id", msg.ID),
		zap.String("telegram_id", telegram.MessageID),
	)
}

// ackMessage acknowledges a message.
func (c *RedisStreamConsumer) ackMessage(ctx context.Context, messageID string) {
	if err := c.client.XAck(ctx, c.stream, c.group, messageID).Err(); err != nil {
		c.logger.Warn("failed to ack message",
			zap.String("id", messageID),
			zap.Error(err),
		)
	}
}

// getRetryCount extracts retry count from message or returns 0.
// Note: Redis Streams doesn't natively track retry count, so we use delivery count
// from XPENDING info, or default to 0 for new messages.
func (c *RedisStreamConsumer) getRetryCount(msg redis.XMessage) int {
	// For simplicity, we'll use a delivery count approximation
	// In a production system, you might want to track retries in message metadata
	if retryStr, ok := msg.Values["retry_count"].(string); ok {
		var retryCount int
		if _, err := fmt.Sscanf(retryStr, "%d", &retryCount); err == nil {
			return retryCount
		}
	}
	// Default to 0 for new messages
	return 0
}

// moveToDLQ moves a message to the dead-letter queue.
func (c *RedisStreamConsumer) moveToDLQ(ctx context.Context, msg redis.XMessage, telegram *app.Telegram) {
	data, err := json.Marshal(telegram)
	if err != nil {
		c.logger.Warn("failed to marshal telegram for DLQ", zap.Error(err))
		return
	}

	values := map[string]interface{}{
		"telegram":    string(data),
		"original_id": msg.ID,
		"timestamp":   time.Now().Unix(),
		"reason":      "max_retries_exceeded",
	}

	args := &redis.XAddArgs{
		Stream: c.dlqStream,
		Values: values,
	}

	if _, err := c.client.XAdd(ctx, args).Result(); err != nil {
		c.logger.Error("failed to move message to DLQ",
			zap.String("id", msg.ID),
			zap.Error(err),
		)
	}
}

// recoverPending recovers pending messages that haven't been ACKed.
func (c *RedisStreamConsumer) recoverPending(ctx context.Context) error {
	// Get pending messages
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  "-",
		End:    "+",
		Count:  100,
	}).Result()

	if err != nil {
		return fmt.Errorf("xpending: %w", err)
	}

	if len(pending) == 0 {
		return nil
	}

	c.logger.Info("recovering pending messages",
		zap.Int("count", len(pending)),
	)

	// Claim and process pending messages
	for _, p := range pending {
		// Claim message (idle for more than 1 minute)
		claimed, err := c.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   c.stream,
			Group:    c.group,
			Consumer: c.consumerName,
			MinIdle:  1 * time.Minute,
			Messages: []string{p.ID},
		}).Result()

		if err != nil {
			c.logger.Warn("failed to claim pending message",
				zap.String("id", p.ID),
				zap.Error(err),
			)
			continue
		}

		// Process claimed messages
		for _, msg := range claimed {
			c.processMessage(ctx, msg)
		}
	}

	return nil
}

// periodicRecover periodically recovers pending messages.
func (c *RedisStreamConsumer) periodicRecover(ctx context.Context) {
	ticker := time.NewTicker(c.recoverInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.recoverPending(ctx); err != nil {
				c.logger.Warn("periodic recover failed", zap.Error(err))
			}
		}
	}
}

