package sync

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"go.uber.org/zap"
)

// Worker coordinates streaming telegram ingestion and indexing.
type Worker struct {
	consumer ports.StreamConsumer
	store    ports.Repository
	search   ports.SearchIndex
	eventBus event.EventBus
	logger   *zap.Logger
}

// NewWorker creates a Worker instance.
func NewWorker(
	consumer ports.StreamConsumer,
	store ports.Repository,
	search ports.SearchIndex,
	eventBus event.EventBus,
	logger *zap.Logger,
) *Worker {
	return &Worker{
		consumer: consumer,
		store:    store,
		search:   search,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle processes a telegram message.
func (w *Worker) Handle(ctx context.Context, telegram *domain.Telegram) error {
	w.logger.Info("worker received telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
		zap.Time("time", telegram.Time),
		zap.String("flight", telegram.FlightNumber),
		zap.String("route", fmt.Sprintf("%s -> %s", telegram.Source, telegram.Destination)),
	)

	if err := telegram.Validate(); err != nil {
		w.logger.Warn("invalid telegram data",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("%w: %v", ErrBadData, err)
	}

	// Normalize the telegram
	telegram.Normalize()

	// Save to PostgreSQL
	if err := w.store.Save(ctx, telegram); err != nil {
		w.logger.Error("failed to save telegram",
			zap.Error(err),
			zap.String("message_id", telegram.MessageID),
		)
		return fmt.Errorf("%w: %v", ErrAppFailure, err)
	}

	w.logger.Info("telegram saved to database",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	// Index to Meilisearch (non-blocking - log errors but continue)
	if w.search != nil {
		if err := w.search.Index(ctx, telegram); err != nil {
			w.logger.Warn("failed to index telegram",
				zap.String("message_id", telegram.MessageID),
				zap.Error(err),
			)
			// Continue - indexing failure should not block the flow
		} else {
			w.logger.Info("telegram indexed",
				zap.String("message_id", telegram.MessageID),
			)
		}
	}

	// Publish event for real-time updates (non-blocking)
	if w.eventBus != nil {
		eventData := map[string]any{
			"type":     "telegram_processed",
			"telegram": telegram, // domain.Telegram now has JSON tags
		}
		if err := w.eventBus.Publish(ctx, "telegram_processed", eventData); err != nil {
			w.logger.Warn("failed to publish event",
				zap.String("message_id", telegram.MessageID),
				zap.Error(err))
		} else {
			w.logger.Debug("event published",
				zap.String("message_id", telegram.MessageID))
		}
	}

	w.logger.Info("successfully processed telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	return nil
}

// NATSConsumer is an interface for NATS-specific consumer operations.
type NATSConsumer interface {
	ports.StreamConsumer
	SetHandler(handler func(context.Context, *domain.Telegram) error)
	SetLogger(logger *zap.Logger)
}

// Run starts the worker loop.
func (w *Worker) Run(ctx context.Context) error {
	// Set handler and logger if consumer supports it
	if natsConsumer, ok := w.consumer.(NATSConsumer); ok {
		natsConsumer.SetHandler(w.Handle)
		natsConsumer.SetLogger(w.logger)
	}

	if err := w.consumer.Start(ctx); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}

	w.logger.Info("sync worker started")

	<-ctx.Done()

	w.logger.Info("sync worker stopped")
	return w.consumer.Close()
}
