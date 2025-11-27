package sync

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/repository"
	natsrepo "github.com/windy/caatsm-dashboard/internal/repository/nats"
	"go.uber.org/zap"
)

// Worker coordinates streaming telegram ingestion and indexing.
type Worker struct {
	consumer repository.StreamConsumer
	store    repository.TelegramStore
	search   repository.SearchIndex
	eventBus event.EventBus
	logger   *zap.Logger
}

// NewWorker creates a Worker instance.
func NewWorker(
	consumer repository.StreamConsumer,
	store repository.TelegramStore,
	search repository.SearchIndex,
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
func (w *Worker) Handle(ctx context.Context, telegram *persistence.Telegram) error {
	w.logger.Info("worker received telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
		zap.Time("time", telegram.Time),
		zap.String("flight", telegram.FlightNumber),
		zap.String("route", fmt.Sprintf("%s -> %s", telegram.Source, telegram.Destination)),
	)

	domainTelegram := domain.ToDomain(telegram)

	if err := domainTelegram.Validate(); err != nil {
		w.logger.Warn("invalid telegram data",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("%w: %v", ErrBadData, err)
	}

	// Normalize the telegram
	domainTelegram.Normalize()

	// Save to PostgreSQL
	if err := w.store.Save(ctx, domainTelegram); err != nil {
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
	if err := w.search.Index(ctx, domainTelegram); err != nil {
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

	// Publish event for real-time updates (non-blocking)
	if w.eventBus != nil {
		eventData := map[string]any{
			"type":     "telegram_processed",
			"telegram": domainTelegram,
		}
		if err := w.eventBus.Publish(ctx, "telegram_processed", eventData); err != nil {
			w.logger.Warn("failed to publish event",
				zap.String("message_id", telegram.MessageID),
				zap.Error(err),
			)
		} else {
			w.logger.Debug("event published",
				zap.String("message_id", telegram.MessageID),
			)
		}
	}

	w.logger.Info("successfully processed telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	return nil
}

// Run starts the worker loop.
func (w *Worker) Run(ctx context.Context) error {
	if natsConsumer, ok := w.consumer.(*natsrepo.Consumer); ok {
		natsConsumer.SetHandler(w.Handle)
	}

	if err := w.consumer.Start(ctx); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}

	w.logger.Info("sync worker started")

	<-ctx.Done()

	w.logger.Info("sync worker stopped")
	return w.consumer.Close()
}
