package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

// Connect creates a NATS connection and JetStream context.
func Connect(ctx context.Context, cfg config.NATSConfig) (*nats.Conn, nats.JetStreamContext, error) {
	timeout := cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	opts := []nats.Option{
		nats.Name("caatsm-dashboard"),
		nats.Timeout(timeout),
		nats.RetryOnFailedConnect(true),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, nil, fmt.Errorf("jetstream context: %w", err)
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			_ = conn.Drain()
		}()
	}

	return conn, js, nil
}

// Consumer wraps a JetStream consumer for telegram ingestion.
type Consumer struct {
	js       nats.JetStreamContext
	stream   string
	consumer string
	sub      *nats.Subscription
	handler  func(context.Context, *app.Telegram) error
	logger   *zap.Logger
}

// NewNATSConsumer creates a Consumer from JetStream context.
func NewNATSConsumer(js nats.JetStreamContext, stream, consumer string) *Consumer {
	return &Consumer{js: js, stream: stream, consumer: consumer}
}

// Ensure Consumer implements app.StreamConsumer
var _ app.StreamConsumer = (*Consumer)(nil)

// SetLogger sets the logger for the consumer.
func (c *Consumer) SetLogger(logger *zap.Logger) {
	c.logger = logger
}

// SetHandler sets the message handler.
func (c *Consumer) SetHandler(handler func(context.Context, *app.Telegram) error) {
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
		// Consumer already exists - fetch and compare configuration
		// Check for consumer name already in use error using typed error comparison
		if errors.Is(err, jetstream.ErrConsumerNameAlreadyInUse) || errors.Is(err, jetstream.ErrConsumerExists) {
			info, fetchErr := c.js.ConsumerInfo(c.stream, c.consumer)
			if fetchErr != nil {
				return fmt.Errorf("ensure consumer: failed to fetch existing consumer info: %w (original error: %w)", fetchErr, err)
			}

			existing := info.Config
			if c.compareConsumerConfig(&existing, cfg) {
				// Configuration matches - log info and return success
				if c.logger != nil {
					c.logger.Info("consumer already exists with matching configuration",
						zap.String("stream", c.stream),
						zap.String("consumer", c.consumer),
					)
				} else {
					log.Printf("INFO: consumer already exists with matching configuration: stream=%s consumer=%s", c.stream, c.consumer)
				}
				return nil
			}

			// Configuration differs - log differences and return error for caller to decide
			diff := c.diffConsumerConfig(&existing, cfg)
			if c.logger != nil {
				c.logger.Warn("consumer exists with different configuration",
					zap.String("stream", c.stream),
					zap.String("consumer", c.consumer),
					zap.String("differences", diff),
					zap.Bool("has_active_subscriptions", info.NumPending > 0 || info.NumAckPending > 0),
				)
			} else {
				log.Printf("WARN: consumer exists with different configuration: stream=%s consumer=%s differences=%s has_active_subscriptions=%v",
					c.stream, c.consumer, diff, info.NumPending > 0 || info.NumAckPending > 0)
			}

			return fmt.Errorf("ensure consumer: existing consumer has different configuration (differences: %s). manual intervention required to avoid message loss", diff)
		}
		return fmt.Errorf("ensure consumer: %w", err)
	}

	return nil
}

// compareConsumerConfig compares two consumer configurations and returns true if they match.
func (c *Consumer) compareConsumerConfig(existing, desired *nats.ConsumerConfig) bool {
	return existing.Durable == desired.Durable &&
		existing.DeliverPolicy == desired.DeliverPolicy &&
		existing.AckPolicy == desired.AckPolicy &&
		existing.AckWait == desired.AckWait &&
		existing.MaxDeliver == desired.MaxDeliver &&
		existing.MaxWaiting == desired.MaxWaiting
}

// diffConsumerConfig returns a human-readable string describing configuration differences.
func (c *Consumer) diffConsumerConfig(existing, desired *nats.ConsumerConfig) string {
	var diffs []string

	if existing.Durable != desired.Durable {
		diffs = append(diffs, fmt.Sprintf("Durable: %s != %s", existing.Durable, desired.Durable))
	}
	if existing.DeliverPolicy != desired.DeliverPolicy {
		diffs = append(diffs, fmt.Sprintf("DeliverPolicy: %v != %v", existing.DeliverPolicy, desired.DeliverPolicy))
	}
	if existing.AckPolicy != desired.AckPolicy {
		diffs = append(diffs, fmt.Sprintf("AckPolicy: %v != %v", existing.AckPolicy, desired.AckPolicy))
	}
	if existing.AckWait != desired.AckWait {
		diffs = append(diffs, fmt.Sprintf("AckWait: %v != %v", existing.AckWait, desired.AckWait))
	}
	if existing.MaxDeliver != desired.MaxDeliver {
		diffs = append(diffs, fmt.Sprintf("MaxDeliver: %d != %d", existing.MaxDeliver, desired.MaxDeliver))
	}
	if existing.MaxWaiting != desired.MaxWaiting {
		diffs = append(diffs, fmt.Sprintf("MaxWaiting: %d != %d", existing.MaxWaiting, desired.MaxWaiting))
	}

	if len(diffs) == 0 {
		return "none"
	}
	return strings.Join(diffs, "; ")
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
				if c.logger != nil {
					c.logger.Warn("fetch messages failed", zap.Error(err))
				}
				continue
			}

			for _, msg := range msgs {
				var telegram app.Telegram
				if err := json.Unmarshal(msg.Data, &telegram); err != nil {
					if c.logger != nil {
						c.logger.Warn("failed to unmarshal telegram",
							zap.Error(err),
							zap.String("subject", msg.Subject),
							zap.ByteString("data_preview", msg.Data[:min(100, len(msg.Data))]),
						)
					}
					// NAK to allow retry - unmarshal failures may be temporary
					// (e.g., schema mismatch during deployment, code bugs)
					// MaxDeliver limit (10) prevents infinite retries
					if nakErr := msg.Nak(); nakErr != nil {
						c.logNakError("nak after unmarshal failure", msg, nakErr, err, "")
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
