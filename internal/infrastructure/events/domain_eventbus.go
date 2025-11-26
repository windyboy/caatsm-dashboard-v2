package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"go.uber.org/zap"
)

// DomainEventBus implements domain event publishing using the infrastructure EventBus.
type DomainEventBus struct {
	eventBus event.EventBus
	logger   *zap.Logger
}

// NewDomainEventBus creates a new domain event bus.
func NewDomainEventBus(eventBus event.EventBus, logger *zap.Logger) *DomainEventBus {
	return &DomainEventBus{
		eventBus: eventBus,
		logger:   logger,
	}
}

// PublishDomainEvent publishes a domain event to the event bus.
func (d *DomainEventBus) PublishDomainEvent(ctx context.Context, evt domain.Event) error {
	// Convert domain event to event bus format
	eventData := map[string]any{
		"event_type": evt.EventType(),
		"occurred_at": evt.OccurredAt(),
	}

	// Include event-specific data
	switch e := evt.(type) {
	case domain.TelegramReceived:
		eventData["telegram"] = e.Telegram
		eventData["source"] = e.Source
	case domain.TelegramValidated:
		eventData["telegram"] = e.Telegram
	case domain.TelegramPersisted:
		eventData["telegram"] = e.Telegram
	case domain.TelegramIndexed:
		eventData["telegram"] = e.Telegram
	case domain.TelegramProcessed:
		eventData["telegram"] = e.Telegram
	default:
		// Generic event data
		eventBytes, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}
		var data map[string]any
		if err := json.Unmarshal(eventBytes, &data); err == nil {
			for k, v := range data {
				eventData[k] = v
			}
		}
	}

	// Publish via infrastructure event bus
	if err := d.eventBus.Publish(ctx, evt.EventType(), eventData); err != nil {
		return fmt.Errorf("publish domain event: %w", err)
	}

	d.logger.Debug("domain event published",
		zap.String("event_type", evt.EventType()),
		zap.Time("occurred_at", evt.OccurredAt()),
	)

	return nil
}

