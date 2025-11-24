package event

import (
	"context"
)

// EventBus defines the interface for publishing events
type EventBus interface {
	// Publish publishes an event to the event bus
	Publish(ctx context.Context, eventType string, data any) error
}
