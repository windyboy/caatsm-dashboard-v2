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

// noOpEventPublisher is a no-op implementation of app.EventPublisher.
// It safely handles cases where EventBus is nil by silently discarding events.
type noOpEventPublisher struct{}

// Publish implements app.EventPublisher as a no-op that always returns nil.
func (n *noOpEventPublisher) Publish(ctx context.Context, event app.Event) error {
	return nil
}

// NewEventPublisherAdapter creates a new adapter that wraps EventBus to implement app.EventPublisher.
// If bus is nil, returns a no-op EventPublisher that safely discards all events.
// This prevents panics when Publish is called with a nil bus.
func NewEventPublisherAdapter(bus EventBus) app.EventPublisher {
	if bus == nil {
		return &noOpEventPublisher{}
	}
	return &EventPublisherAdapter{
		bus: bus,
	}
}

// Publish implements app.EventPublisher by converting app.Event to EventBus format.
func (a *EventPublisherAdapter) Publish(ctx context.Context, event app.Event) error {
	return a.bus.Publish(ctx, event.EventType(), event)
}
