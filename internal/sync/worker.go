package sync

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/application"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
	natsrepo "github.com/windy/caatsm-dashboard/internal/repository/nats"
	"go.uber.org/zap"
)

// Worker coordinates streaming telegram ingestion and indexing.
type Worker struct {
	consumer        repository.StreamConsumer
	telegramService application.TelegramService
	logger          *zap.Logger
}

// NewWorker creates a Worker instance.
func NewWorker(consumer repository.StreamConsumer, telegramService application.TelegramService, logger *zap.Logger) *Worker {
	return &Worker{
		consumer:        consumer,
		telegramService: telegramService,
		logger:          logger,
	}
}

// Handle processes a telegram message.
func (w *Worker) Handle(ctx context.Context, telegram *models.Telegram) error {
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

	if err := w.telegramService.SaveTelegram(ctx, domainTelegram); err != nil {
		w.logger.Error("failed to save telegram via application service",
			zap.Error(err),
			zap.String("message_id", telegram.MessageID),
			zap.String("type", telegram.Type),
		)
		return fmt.Errorf("%w: %v", ErrAppFailure, err)
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
