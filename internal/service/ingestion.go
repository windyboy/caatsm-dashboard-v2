package service

import (
	"context"
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// IngestionService handles message ingestion from NATS JetStream.
type IngestionService struct {
	consumer       app.StreamConsumer
	store          app.Repository
	eventPublisher app.EventPublisher
	streamPublisher app.StreamPublisher
	logger         *zap.Logger
}

// NewIngestionService creates a new ingestion service.
func NewIngestionService(
	consumer app.StreamConsumer,
	store app.Repository,
	eventPublisher app.EventPublisher,
	streamPublisher app.StreamPublisher,
	logger *zap.Logger,
) *IngestionService {
	return &IngestionService{
		consumer:        consumer,
		store:           store,
		eventPublisher:  eventPublisher,
		streamPublisher: streamPublisher,
		logger:          logger,
	}
}

// Handle processes a single telegram message.
func (s *IngestionService) Handle(ctx context.Context, telegram *app.Telegram) error {
	startTime := time.Now()

	// Validate message structure
	if err := telegram.Validate(); err != nil {
		s.logger.Warn("invalid telegram data",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Normalize message data
	telegram.Normalize()

	// Store to PostgreSQL synchronously (idempotent via ON CONFLICT DO NOTHING)
	if err := s.store.Save(ctx, telegram); err != nil {
		s.logger.Error("failed to save telegram",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("save to database: %w", err)
	}

	s.logger.Info("telegram saved to database",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	// Publish to Redis Pub/Sub msg:broadcast (fire-and-forget, best-effort)
	event := domain.TelegramPersisted{
		Telegram: telegram,
		At:       time.Now(),
	}
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		// Log but don't fail - this is best-effort
		s.logger.Warn("failed to publish event to redis pub/sub",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
	}

	// Push to Redis Stream job:index (reliable queue, at-least-once)
	if err := s.streamPublisher.Publish(ctx, "job:index", telegram); err != nil {
		s.logger.Error("failed to publish to redis stream",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		// This is a critical failure - indexing must happen
		// But we've already saved to DB, so we return error but don't fail ingestion
		// The message will be retried by NATS if we return error
		return fmt.Errorf("publish to stream: %w", err)
	}

	latency := time.Since(startTime)
	s.logger.Debug("telegram processed successfully",
		zap.String("message_id", telegram.MessageID),
		zap.Duration("latency", latency),
	)

	return nil
}

// Start starts the ingestion service as a background task.
func (s *IngestionService) Start(ctx context.Context) error {
	// Set handler if consumer supports it
	if natsConsumer, ok := s.consumer.(interface {
		SetHandler(handler func(context.Context, *app.Telegram) error)
		SetLogger(logger *zap.Logger)
	}); ok {
		natsConsumer.SetHandler(s.Handle)
		natsConsumer.SetLogger(s.logger)
	}

	// Start the consumer
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}

	s.logger.Info("ingestion service started")

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("ingestion service stopping")
	if err := s.consumer.Close(); err != nil {
		s.logger.Warn("error closing consumer", zap.Error(err))
	}

	return nil
}

