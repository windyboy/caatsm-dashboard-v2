package event

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/app"
)

// EventPublisherAdapter adapts EventBus to app.EventPublisher interface.
// It converts app.Event to the format expected by EventBus.
type EventPublisherAdapter struct {
	bus EventBus
}

// NewEventPublisherAdapter creates a new adapter that wraps EventBus to implement app.EventPublisher.
func NewEventPublisherAdapter(bus EventBus) app.EventPublisher {
	return &EventPublisherAdapter{
		bus: bus,
	}
}

// Publish implements app.EventPublisher by converting app.Event to EventBus format.
func (a *EventPublisherAdapter) Publish(ctx context.Context, event app.Event) error {
	return a.bus.Publish(ctx, event.EventType(), event)
}
