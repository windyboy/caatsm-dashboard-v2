package event

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// EventPublisherAdapter adapts EventBus to ports.EventPublisher interface.
// It converts domain.Event to the format expected by EventBus.
type EventPublisherAdapter struct {
	bus EventBus
}

// NewEventPublisherAdapter creates a new adapter that wraps EventBus to implement ports.EventPublisher.
func NewEventPublisherAdapter(bus EventBus) ports.EventPublisher {
	return &EventPublisherAdapter{
		bus: bus,
	}
}

// Publish implements ports.EventPublisher by converting domain.Event to EventBus format.
func (a *EventPublisherAdapter) Publish(ctx context.Context, event domain.Event) error {
	return a.bus.Publish(ctx, event.EventType(), event)
}

