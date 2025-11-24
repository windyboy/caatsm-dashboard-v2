package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
)

// EventBusMock is a mock implementation of event.EventBus
type EventBusMock struct {
	mock.Mock
}

// Ensure EventBusMock implements event.EventBus
var _ event.EventBus = (*EventBusMock)(nil)

// Publish mocks the Publish method
func (m *EventBusMock) Publish(ctx context.Context, eventType string, data any) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}
