package natsrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/windy/caatsm-dashboard/internal/models"
)

// Consumer wraps a JetStream consumer for telegram ingestion.
type Consumer struct {
	js       nats.JetStreamContext
	stream   string
	consumer string
	sub      *nats.Subscription
	handler  func(context.Context, *models.Telegram) error
}

// New creates a Consumer from JetStream context.
func New(js nats.JetStreamContext, stream, consumer string) *Consumer {
	return &Consumer{js: js, stream: stream, consumer: consumer}
}

// SetHandler sets the message handler.
func (c *Consumer) SetHandler(handler func(context.Context, *models.Telegram) error) {
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
				var telegram models.Telegram
				if err := json.Unmarshal(msg.Data, &telegram); err != nil {
					_ = msg.Ack()
					continue
				}

				if telegram.MessageID == "" {
					_ = msg.Ack()
					continue
				}

				if err := c.handler(ctx, &telegram); err != nil {
					_ = msg.Nak()
					continue
				}

				_ = msg.Ack()
			}
		}
	}
}

// Close performs cleanup for the consumer.
func (c *Consumer) Close() error {
	if c.sub != nil {
		return c.sub.Unsubscribe()
	}
	return nil
}
