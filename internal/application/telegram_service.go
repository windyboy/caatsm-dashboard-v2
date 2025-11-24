package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"go.uber.org/zap"
)

// telegramService implements TelegramService interface
type telegramService struct {
	store    repository.TelegramStore
	index    repository.SearchIndex
	eventBus event.EventBus
	logger   *zap.Logger
}

// NewTelegramService creates a new TelegramService
func NewTelegramService(
	store repository.TelegramStore,
	index repository.SearchIndex,
	eventBus event.EventBus,
	logger *zap.Logger,
) TelegramService {
	return &telegramService{
		store:    store,
		index:    index,
		eventBus: eventBus,
		logger:   logger,
	}
}

// SaveTelegram implements the SaveTelegram use case
// It validates, saves to database, indexes, and publishes events
// Errors in indexing or publishing do not interrupt the main flow
func (s *telegramService) SaveTelegram(ctx context.Context, telegram *domain.Telegram) error {
	// 1. Validate domain rules
	if err := telegram.Validate(); err != nil {
		return fmt.Errorf("validate telegram: %w", err)
	}

	// 2. Normalize the telegram
	telegram.Normalize()

	// 3. Save to PostgreSQL
	modelTelegram := domain.FromDomain(telegram)
	if err := s.store.Save(ctx, modelTelegram); err != nil {
		return fmt.Errorf("save telegram: %w", err)
	}

	s.logger.Info("telegram saved to database",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
	)

	// 4. Index to Meilisearch (non-blocking - errors are logged but don't fail)
	if err := s.index.Index(ctx, modelTelegram); err != nil {
		s.logger.Warn("failed to index telegram",
			zap.String("message_id", telegram.MessageID),
			zap.Error(err),
		)
		// Continue execution - indexing failure should not block the main flow
	} else {
		s.logger.Info("telegram indexed",
			zap.String("message_id", telegram.MessageID),
		)
	}

	// 5. Publish event (non-blocking - errors are logged but don't fail)
	if s.eventBus != nil {
		eventData := map[string]any{
			"type":     "telegram_processed",
			"telegram": modelTelegram,
		}
		if err := s.eventBus.Publish(ctx, "telegram_processed", eventData); err != nil {
			s.logger.Warn("failed to publish event",
				zap.String("message_id", telegram.MessageID),
				zap.Error(err),
			)
			// Continue execution - event publishing failure should not block the main flow
		} else {
			s.logger.Debug("event published",
				zap.String("message_id", telegram.MessageID),
			)
		}
	}

	return nil
}

// publishEvent is a helper method to publish events (kept for backward compatibility)
func (s *telegramService) publishEvent(ctx context.Context, telegram *domain.Telegram) error {
	if s.eventBus == nil {
		return nil
	}

	eventData := map[string]any{
		"type":     "telegram_processed",
		"telegram": domain.FromDomain(telegram),
	}

	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return s.eventBus.Publish(ctx, "telegram_processed", eventBytes)
}
