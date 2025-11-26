package handlers

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	infraevents "github.com/windy/caatsm-dashboard/internal/infrastructure/events"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"go.uber.org/zap"
)

// IndexingHandler handles indexing-related domain events.
type IndexingHandler struct {
	index    repository.SearchIndex
	eventBus *infraevents.DomainEventBus
	logger   *zap.Logger
}

// NewIndexingHandler creates a new indexing handler.
func NewIndexingHandler(
	index repository.SearchIndex,
	eventBus *infraevents.DomainEventBus,
	logger *zap.Logger,
) *IndexingHandler {
	return &IndexingHandler{
		index:    index,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle implements the appevents.Handler interface.
func (h *IndexingHandler) Handle(ctx context.Context, evt domain.Event) error {
	switch e := evt.(type) {
	case domain.TelegramPersisted:
		return h.HandleTelegramPersisted(ctx, e)
	default:
		// Ignore other event types
		return nil
	}
}

// HandleTelegramPersisted handles TelegramPersisted events by indexing the telegram.
func (h *IndexingHandler) HandleTelegramPersisted(ctx context.Context, evt domain.TelegramPersisted) error {
	if evt.Telegram == nil {
		return fmt.Errorf("telegram is nil in TelegramPersisted event")
	}

	// Convert domain telegram to persistence model
	modelTelegram := domain.FromDomain(evt.Telegram)
	if modelTelegram == nil {
		return fmt.Errorf("failed to convert domain telegram to model")
	}

	// Index to search engine
	if err := h.index.Index(ctx, modelTelegram); err != nil {
		h.logger.Warn("failed to index telegram",
			zap.String("message_id", evt.Telegram.MessageID),
			zap.Error(err),
		)
		// Return error but don't fail the entire pipeline
		return fmt.Errorf("index telegram: %w", err)
	}

	h.logger.Info("telegram indexed",
		zap.String("message_id", evt.Telegram.MessageID),
		zap.String("type", evt.Telegram.Type),
	)

	// Publish TelegramIndexed event
	if h.eventBus != nil {
		indexedEvent := domain.TelegramIndexed{
			Telegram: evt.Telegram,
			At:       evt.OccurredAt(),
		}
		if err := h.eventBus.PublishDomainEvent(ctx, indexedEvent); err != nil {
			h.logger.Warn("failed to publish TelegramIndexed event",
				zap.String("message_id", evt.Telegram.MessageID),
				zap.Error(err),
			)
			// Don't fail the indexing operation
		}
	}

	return nil
}

