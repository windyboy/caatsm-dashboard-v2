package handlers

import (
	"context"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	infraevents "github.com/windy/caatsm-dashboard/internal/infrastructure/events"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"go.uber.org/zap"
)

// PersistenceHandler handles persistence-related domain events.
type PersistenceHandler struct {
	store      repository.TelegramStore
	eventBus   *infraevents.DomainEventBus
	logger     *zap.Logger
}

// NewPersistenceHandler creates a new persistence handler.
func NewPersistenceHandler(
	store repository.TelegramStore,
	eventBus *infraevents.DomainEventBus,
	logger *zap.Logger,
) *PersistenceHandler {
	return &PersistenceHandler{
		store:    store,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle implements the appevents.Handler interface.
func (h *PersistenceHandler) Handle(ctx context.Context, evt domain.Event) error {
	switch e := evt.(type) {
	case domain.TelegramValidated:
		return h.HandleTelegramValidated(ctx, e)
	default:
		// Ignore other event types
		return nil
	}
}

// HandleTelegramValidated handles TelegramValidated events by persisting the telegram.
func (h *PersistenceHandler) HandleTelegramValidated(ctx context.Context, evt domain.TelegramValidated) error {
	if evt.Telegram == nil {
		return fmt.Errorf("telegram is nil in TelegramValidated event")
	}

	// Persist to database
	if err := h.store.Save(ctx, evt.Telegram); err != nil {
		h.logger.Error("failed to persist telegram",
			zap.String("message_id", evt.Telegram.MessageID),
			zap.Error(err),
		)
		return fmt.Errorf("persist telegram: %w", err)
	}

	h.logger.Info("telegram persisted",
		zap.String("message_id", evt.Telegram.MessageID),
		zap.String("type", evt.Telegram.Type),
	)

	// Publish TelegramPersisted event
	if h.eventBus != nil {
		persistedEvent := domain.TelegramPersisted{
			Telegram: evt.Telegram,
			At:       evt.OccurredAt(),
		}
		if err := h.eventBus.PublishDomainEvent(ctx, persistedEvent); err != nil {
			h.logger.Warn("failed to publish TelegramPersisted event",
				zap.String("message_id", evt.Telegram.MessageID),
				zap.Error(err),
			)
			// Don't fail the persistence operation
		}
	}

	return nil
}

