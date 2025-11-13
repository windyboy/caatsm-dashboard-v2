package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
	natsrepo "github.com/windy/caatsm-dashboard/internal/repository/nats"
	"go.uber.org/zap"
)

// Worker coordinates streaming telegram ingestion and indexing.
type Worker struct {
	consumer repository.StreamConsumer
	store    repository.TelegramStore
	index    repository.SearchIndex
	redisCli redis.UniversalClient
	logger   *zap.Logger
}

// NewWorker creates a Worker instance.
func NewWorker(consumer repository.StreamConsumer, store repository.TelegramStore, index repository.SearchIndex, logger *zap.Logger) *Worker {
	return &Worker{
		consumer: consumer,
		store:    store,
		index:    index,
		logger:   logger,
	}
}

// SetRedisClient sets the Redis client for publishing stats update events.
func (w *Worker) SetRedisClient(redisCli redis.UniversalClient) {
	w.redisCli = redisCli
}

// Handle processes a single telegram message.
func (w *Worker) Handle(ctx context.Context, telegram *models.Telegram) error {
	// Validate telegram data
	if telegram.MessageID == "" {
		return fmt.Errorf("telegram missing message_id")
	}
	if telegram.Time.IsZero() {
		w.logger.Warn("telegram has zero time, using current time",
			zap.String("message_id", telegram.MessageID),
		)
		// Don't modify the telegram, but log the issue
	}

	w.logger.Info("worker received telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
		zap.Time("time", telegram.Time),
		zap.String("flight", telegram.FlightNumber),
		zap.String("route", fmt.Sprintf("%s -> %s", telegram.Source, telegram.Destination)),
	)

	// Save to PostgreSQL
	if err := w.store.Save(ctx, telegram); err != nil {
		w.logger.Error("failed to save telegram to database",
			zap.Error(err),
			zap.String("message_id", telegram.MessageID),
			zap.String("type", telegram.Type),
		)
		return fmt.Errorf("save telegram: %w", err)
	}

	// Index to Meilisearch
	if err := w.index.Index(ctx, telegram); err != nil {
		w.logger.Error("failed to index telegram",
			zap.Error(err),
			zap.String("message_id", telegram.MessageID),
			zap.String("type", telegram.Type),
		)
		return fmt.Errorf("index telegram: %w", err)
	}

	w.logger.Info("successfully processed telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	// Publish stats update event to Redis for real-time dashboard updates
	// Include full telegram data so the dashboard can display the message
	if w.redisCli != nil {
		w.logger.Info("publishing event to redis", zap.String("message_id", telegram.MessageID))
		event := map[string]interface{}{
			"type":     "telegram_processed",
			"time":     time.Now().Unix(),
			"telegram": telegram, // Include full telegram data for live stream
		}
		data, err := json.Marshal(event)
		if err == nil {
			if err := w.redisCli.Publish(ctx, "stats:update", data).Err(); err != nil {
				w.logger.Warn("failed to publish stats update event", zap.Error(err))
			} else {
				w.logger.Info("published stats update event to redis",
					zap.String("message_id", telegram.MessageID),
					zap.String("type", telegram.Type),
					zap.Int("payload_size", len(data)))
			}
		} else {
			w.logger.Warn("failed to marshal event", zap.Error(err))
		}
	} else {
		w.logger.Warn("redis client is nil, cannot publish event", zap.String("message_id", telegram.MessageID))
	}

	return nil
}

// Run starts the worker loop.
func (w *Worker) Run(ctx context.Context) error {
	// Set up handler for NATS consumer
	if natsConsumer, ok := w.consumer.(*natsrepo.Consumer); ok {
		natsConsumer.SetHandler(w.Handle)
	}

	// Start consuming messages
	if err := w.consumer.Start(ctx); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}

	w.logger.Info("sync worker started")

	// Wait for context cancellation
	<-ctx.Done()

	w.logger.Info("sync worker stopped")
	return w.consumer.Close()
}
