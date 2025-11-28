package natsrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

// Consumer wraps a JetStream consumer for telegram ingestion.
type Consumer struct {
	js       nats.JetStreamContext
	stream   string
	consumer string
	sub      *nats.Subscription
	handler  func(context.Context, *persistence.Telegram) error
	logger   *zap.Logger
}

// New creates a Consumer from JetStream context.
func New(js nats.JetStreamContext, stream, consumer string) *Consumer {
	return &Consumer{js: js, stream: stream, consumer: consumer}
}

// SetLogger sets the logger for the consumer.
func (c *Consumer) SetLogger(logger *zap.Logger) {
	c.logger = logger
}

// SetHandler sets the message handler.
func (c *Consumer) SetHandler(handler func(context.Context, *persistence.Telegram) error) {
	c.handler = handler
}

// EnsureStream creates or updates the JetStream stream configuration.
func (c *Consumer) EnsureStream(ctx context.Context) error {
	cfg := &nats.StreamConfig{
		Name:      c.stream,
		Subjects:  []string{"telegrams.>"},
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour * 30, // 30 days
		Storage:   nats.FileStorage,
		Replicas:  1,
	}

	_, err := c.js.AddStream(cfg)
	if err != nil {
		// Stream might already exist, try to update it
		_, err = c.js.UpdateStream(cfg)
		if err != nil {
			return fmt.Errorf("ensure stream: %w", err)
		}
	}

	return nil
}

// EnsureConsumer creates or updates the JetStream consumer configuration.
func (c *Consumer) EnsureConsumer(ctx context.Context) error {
	cfg := &nats.ConsumerConfig{
		Durable:       c.consumer,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    10,
		MaxWaiting:    128, // Must match PullMaxWaiting in PullSubscribe
	}

	_, err := c.js.AddConsumer(c.stream, cfg)
	if err != nil {
		// Consumer might already exist, try to delete and recreate
		if strings.Contains(err.Error(), "consumer name already in use") {
			// Delete existing consumer
			if deleteErr := c.js.DeleteConsumer(c.stream, c.consumer); deleteErr != nil {
				// If delete fails, consumer might be in use or doesn't exist
				// Return original error
				return fmt.Errorf("ensure consumer: %w (delete failed: %v)", err, deleteErr)
			}
			// Try to create again after deletion
			_, err = c.js.AddConsumer(c.stream, cfg)
			if err != nil {
				return fmt.Errorf("recreate consumer: %w", err)
			}
			return nil
		}
		return fmt.Errorf("ensure consumer: %w", err)
	}

	return nil
}

// Start begins processing messages.
func (c *Consumer) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("handler not set")
	}

	if err := c.EnsureStream(ctx); err != nil {
		return fmt.Errorf("ensure stream: %w", err)
	}

	if err := c.EnsureConsumer(ctx); err != nil {
		return fmt.Errorf("ensure consumer: %w", err)
	}

	sub, err := c.js.PullSubscribe("telegrams.>", c.consumer, nats.PullMaxWaiting(128))
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	c.sub = sub

	go c.processMessages(ctx)

	return nil
}

func (c *Consumer) processMessages(ctx context.Context) {
	batchSize := 10
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msgs, err := c.sub.Fetch(batchSize, nats.MaxWait(1*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				// Log error but continue processing
				continue
			}

			for _, msg := range msgs {
				var telegram persistence.Telegram
				if err := json.Unmarshal(msg.Data, &telegram); err != nil {
					if ackErr := msg.Ack(); ackErr != nil {
						c.logAckError("ack after unmarshal failure", msg, ackErr, nil)
					}
					continue
				}

				if telegram.MessageID == "" {
					if ackErr := msg.Ack(); ackErr != nil {
						c.logAckError("ack after empty message_id", msg, ackErr, nil)
					}
					continue
				}

				if err := c.handler(ctx, &telegram); err != nil {
					if nakErr := msg.Nak(); nakErr != nil {
						c.logNakError("nak after handler failure", msg, nakErr, err, telegram.MessageID)
					}
					continue
				}

				if ackErr := msg.Ack(); ackErr != nil {
					c.logAckError("ack after successful processing", msg, ackErr, nil)
				}
			}
		}
	}
}

// logAckError logs an error from an Ack operation with contextual information.
func (c *Consumer) logAckError(operation string, msg *nats.Msg, ackErr error, handlerErr error) {
	fields := c.getMessageFields(msg)
	fields = append(fields, zap.String("operation", operation), zap.Error(ackErr))
	if handlerErr != nil {
		fields = append(fields, zap.NamedError("handler_error", handlerErr))
	}

	if c.logger != nil {
		c.logger.Error("failed to ack message", fields...)
	} else {
		log.Printf("ERROR: failed to ack message: operation=%s subject=%s seq=%d error=%v",
			operation, msg.Subject, c.getMessageSeq(msg), ackErr)
		if handlerErr != nil {
			log.Printf("  handler_error=%v", handlerErr)
		}
	}
}

// logNakError logs an error from a Nak operation with contextual information.
func (c *Consumer) logNakError(operation string, msg *nats.Msg, nakErr error, handlerErr error, messageID string) {
	fields := c.getMessageFields(msg)
	fields = append(fields,
		zap.String("operation", operation),
		zap.Error(nakErr),
		zap.NamedError("handler_error", handlerErr),
		zap.String("telegram_message_id", messageID),
	)

	if c.logger != nil {
		c.logger.Error("failed to nak message", fields...)
	} else {
		log.Printf("ERROR: failed to nak message: operation=%s subject=%s seq=%d message_id=%s error=%v handler_error=%v",
			operation, msg.Subject, c.getMessageSeq(msg), messageID, nakErr, handlerErr)
	}
}

// getMessageFields extracts contextual fields from a NATS message for logging.
func (c *Consumer) getMessageFields(msg *nats.Msg) []zap.Field {
	fields := []zap.Field{
		zap.String("subject", msg.Subject),
		zap.String("stream", c.stream),
		zap.String("consumer", c.consumer),
	}

	if seq := c.getMessageSeq(msg); seq > 0 {
		fields = append(fields, zap.Uint64("sequence", seq))
	}

	metadata, err := msg.Metadata()
	if err == nil && metadata != nil {
		if metadata.NumDelivered > 0 {
			fields = append(fields, zap.Uint64("num_delivered", metadata.NumDelivered))
		}
		if !metadata.Timestamp.IsZero() {
			fields = append(fields, zap.Time("timestamp", metadata.Timestamp))
		}
	}

	return fields
}

// getMessageSeq extracts the sequence number from a NATS message.
func (c *Consumer) getMessageSeq(msg *nats.Msg) uint64 {
	metadata, err := msg.Metadata()
	if err == nil && metadata != nil {
		return metadata.Sequence.Stream
	}
	return 0
}

// Close performs cleanup for the consumer.
func (c *Consumer) Close() error {
	if c.sub != nil {
		return c.sub.Unsubscribe()
	}
	return nil
}
