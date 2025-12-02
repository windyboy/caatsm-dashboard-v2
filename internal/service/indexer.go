package service

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

// IndexerService handles asynchronous indexing of telegrams to Meilisearch.
type IndexerService struct {
	streamConsumer app.RedisStreamConsumer
	search         app.SearchIndex
	logger         *zap.Logger
}

// NewIndexerService creates a new indexer service.
func NewIndexerService(
	streamConsumer app.RedisStreamConsumer,
	search app.SearchIndex,
	logger *zap.Logger,
) *IndexerService {
	return &IndexerService{
		streamConsumer: streamConsumer,
		search:         search,
		logger:         logger,
	}
}

// handleMessage processes a single message for indexing.
func (s *IndexerService) handleMessage(ctx context.Context, telegram *app.Telegram) error {
	if err := s.search.Index(ctx, telegram); err != nil {
		s.logger.Warn("failed to index telegram",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("index to meilisearch: %w", err)
	}

	s.logger.Debug("telegram indexed successfully",
		zap.String("message_id", telegram.MessageID),
	)

	return nil
}

// Start starts the indexer service as a background task.
func (s *IndexerService) Start(ctx context.Context) error {
	// Set handler
	s.streamConsumer.SetHandler(s.handleMessage)

	// Start the consumer
	if err := s.streamConsumer.Start(ctx); err != nil {
		return fmt.Errorf("start stream consumer: %w", err)
	}

	s.logger.Info("indexer service started")

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("indexer service stopping")
	if err := s.streamConsumer.Close(); err != nil {
		s.logger.Warn("error closing stream consumer", zap.Error(err))
	}

	return nil
}

