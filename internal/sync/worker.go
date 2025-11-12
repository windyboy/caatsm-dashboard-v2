package sync

import (
	"context"
	"fmt"

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

// Handle processes a single telegram message.
func (w *Worker) Handle(ctx context.Context, telegram *models.Telegram) error {
	// Save to PostgreSQL
	if err := w.store.Save(ctx, telegram); err != nil {
		w.logger.Error("failed to save telegram to database", zap.Error(err), zap.String("message_id", telegram.MessageID))
		return fmt.Errorf("save telegram: %w", err)
	}

	// Index to Meilisearch
	if err := w.index.Index(ctx, telegram); err != nil {
		w.logger.Error("failed to index telegram", zap.Error(err), zap.String("message_id", telegram.MessageID))
		return fmt.Errorf("index telegram: %w", err)
	}

	w.logger.Debug("processed telegram", zap.String("message_id", telegram.MessageID))
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
