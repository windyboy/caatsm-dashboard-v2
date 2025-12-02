package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
)

type mockEventSubscriber struct {
	mock.Mock
}

func (m *mockEventSubscriber) Subscribe(ctx context.Context, channel string, handler func(context.Context, app.Event) error) error {
	args := m.Called(ctx, channel, handler)
	return args.Error(0)
}

type mockWebSocketHub struct {
	mock.Mock
}

func (m *mockWebSocketHub) Register(client interface{}) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *mockWebSocketHub) Unregister(client interface{}) {
	m.Called(client)
}

func (m *mockWebSocketHub) Broadcast(message app.WSMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *mockWebSocketHub) GetActiveConnections() int {
	args := m.Called()
	return args.Int(0)
}

func TestRealtimeService_handleMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		event         app.Event
		setupMocks    func(*mockWebSocketHub)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful broadcast with TelegramPersisted",
			event: &domain.TelegramPersisted{
				Telegram: &app.Telegram{
					MessageID:    "TEST-001",
					Type:         "AFTN",
					Time:         time.Now(),
					FlightNumber: "CA100",
					Priority:     1,
				},
				At: time.Now(),
			},
			setupMocks: func(hub *mockWebSocketHub) {
				hub.On("Broadcast", mock.MatchedBy(func(msg app.WSMessage) bool {
					return msg.Type == "message" && msg.Data != nil
				})).Return(nil)
				hub.On("GetActiveConnections").Return(5)
			},
			expectedError: false,
		},
		{
			name: "broadcast error",
			event: &domain.TelegramPersisted{
				Telegram: &app.Telegram{
					MessageID:    "TEST-002",
					Type:         "AFTN",
					Time:         time.Now(),
					FlightNumber: "CA100",
					Priority:     1,
				},
				At: time.Now(),
			},
			setupMocks: func(hub *mockWebSocketHub) {
				hub.On("Broadcast", mock.Anything).Return(errors.New("broadcast error"))
			},
			expectedError: true,
			errorContains: "broadcast",
		},
		{
			name: "event without telegram (generic event)",
			event: &domain.SearchPerformed{
				Query:       "test",
				ResultCount: 10,
				Duration:    time.Second,
				Timestamp:   time.Now(),
			},
			setupMocks: func(hub *mockWebSocketHub) {
				// Should not broadcast for non-telegram events
			},
			expectedError: false, // Should skip, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriber := &mockEventSubscriber{}
			hub := &mockWebSocketHub{}

			tt.setupMocks(hub)

			svc := NewRealtimeService(subscriber, hub, logger)

			err := svc.handleMessage(ctx, tt.event)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// For non-telegram events, we expect no error (just skip)
				if tt.event.EventType() != "telegram.persisted" {
					require.NoError(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			hub.AssertExpectations(t)
		})
	}
}

func TestRealtimeService_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subscriber := &mockEventSubscriber{}
	hub := &mockWebSocketHub{}

	// Event handler function type - app.Event is a type alias for domain.Event
	// Use mock.Anything to accept the function since type aliases cause issues with mock.AnythingOfType
	subscriber.On("Subscribe", mock.Anything, "msg:broadcast", mock.Anything).Return(nil)

	svc := NewRealtimeService(subscriber, hub, logger)

	// Start service in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	// Wait a bit to ensure Subscribe is called
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop service
	cancel()

	// Wait for service to stop
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("service did not stop in time")
	}

	subscriber.AssertExpectations(t)
}

