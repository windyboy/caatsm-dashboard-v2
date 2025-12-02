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

type mockRedisStreamConsumer struct {
	mock.Mock
	handler func(context.Context, *app.Telegram) error
}

func (m *mockRedisStreamConsumer) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRedisStreamConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRedisStreamConsumer) SetHandler(handler func(context.Context, *app.Telegram) error) {
	m.handler = handler
}

type mockSearchIndex struct {
	mock.Mock
}

func (m *mockSearchIndex) EnsureIndex(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockSearchIndex) Index(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

func (m *mockSearchIndex) BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

func (m *mockSearchIndex) Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error) {
	args := m.Called(ctx, query, filter, limit, offset, sort)
	return args.Get(0), args.Error(1)
}

func (m *mockSearchIndex) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (*app.AutocompleteResponse, error) {
	args := m.Called(ctx, query, limit, attributes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*app.AutocompleteResponse), args.Error(1)
}

func TestIndexerService_handleMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		telegram      *app.Telegram
		setupMocks    func(*mockSearchIndex)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful indexing",
			telegram: &app.Telegram{
				MessageID:    "TEST-001",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Priority:     1,
			},
			setupMocks: func(search *mockSearchIndex) {
				// Index expects *domain.Telegram, but handleMessage passes *app.Telegram
				// Since app.Telegram is a type alias for domain.Telegram, mock.Anything works
				search.On("Index", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "indexing error",
			telegram: &app.Telegram{
				MessageID:    "TEST-002",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Priority:     1,
			},
			setupMocks: func(search *mockSearchIndex) {
				// Index expects *domain.Telegram, but handleMessage passes *app.Telegram
				// Since app.Telegram is a type alias for domain.Telegram, mock.Anything works
				search.On("Index", mock.Anything, mock.Anything).Return(errors.New("meilisearch error"))
			},
			expectedError: true,
			errorContains: "index to meilisearch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := &mockRedisStreamConsumer{}
			search := &mockSearchIndex{}

			tt.setupMocks(search)

			svc := NewIndexerService(
				consumer,
				search,
				logger,
			)

			// handleMessage expects *app.Telegram, and since app.Telegram is a type alias
			// for domain.Telegram, we can pass it directly
			err := svc.handleMessage(ctx, tt.telegram)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			search.AssertExpectations(t)
		})
	}
}

func TestIndexerService_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := &mockRedisStreamConsumer{}
	search := &mockSearchIndex{}

	consumer.On("Start", mock.Anything).Return(nil)
	consumer.On("Close").Return(nil)

	svc := NewIndexerService(consumer, search, logger)

	// Start service in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	// Wait a bit to ensure Start is called
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

	consumer.AssertExpectations(t)
}

