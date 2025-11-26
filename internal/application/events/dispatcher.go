package events

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// Handler defines an interface for event handlers.
type Handler interface {
	Handle(ctx context.Context, event domain.Event) error
}

// Dispatcher coordinates domain event handling.
type Dispatcher struct {
	handlers map[string][]Handler
	logger   *zap.Logger
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher(logger *zap.Logger) *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

// Subscribe registers a handler for a specific event type.
func (d *Dispatcher) Subscribe(eventType string, handler Handler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
	d.logger.Debug("handler subscribed",
		zap.String("event_type", eventType),
	)
}

// Dispatch dispatches an event to all registered handlers.
// Errors from handlers are logged but don't stop other handlers from executing.
func (d *Dispatcher) Dispatch(ctx context.Context, event domain.Event) error {
	eventType := event.EventType()
	handlers := d.handlers[eventType]

	if len(handlers) == 0 {
		d.logger.Debug("no handlers registered for event",
			zap.String("event_type", eventType),
		)
		return nil
	}

	var lastErr error
	for i, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			d.logger.Warn("handler failed",
				zap.String("event_type", eventType),
				zap.Int("handler_index", i),
				zap.Error(err),
			)
			lastErr = err
			// Continue with other handlers
		}
	}

	return lastErr // Return last error, but don't fail entire dispatch
}

// DispatchAsync dispatches events asynchronously.
// Use this for fire-and-forget event handling.
func (d *Dispatcher) DispatchAsync(ctx context.Context, event domain.Event) {
	go func() {
		if err := d.Dispatch(ctx, event); err != nil {
			d.logger.Error("async dispatch failed",
				zap.String("event_type", event.EventType()),
				zap.Error(err),
			)
		}
	}()
}

