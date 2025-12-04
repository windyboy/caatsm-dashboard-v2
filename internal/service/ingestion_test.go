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

// Mock implementations for testing

type mockStreamConsumer struct {
	mock.Mock
	handler func(context.Context, *app.Telegram) error
	logger  interface{}
}

func (m *mockStreamConsumer) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockStreamConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStreamConsumer) SetHandler(handler func(context.Context, *app.Telegram) error) {
	m.handler = handler
}

func (m *mockStreamConsumer) SetLogger(logger interface{}) {
	m.logger = logger
}

// Ensure mockStreamConsumer implements app.StreamConsumer
var _ app.StreamConsumer = (*mockStreamConsumer)(nil)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Save(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

func (m *mockRepository) BulkSave(ctx context.Context, telegrams []any) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

func (m *mockRepository) Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResult), args.Error(1)
}

func (m *mockRepository) StreamSearch(ctx context.Context, filter domain.SearchFilters) (<-chan *domain.Telegram, <-chan error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, nil
	}
	return args.Get(0).(<-chan *domain.Telegram), args.Get(1).(<-chan error)
}

func (m *mockRepository) FindByID(ctx context.Context, messageID string) (*domain.Telegram, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Telegram), args.Error(1)
}

func (m *mockRepository) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	args := m.Called(ctx, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TrafficSummary), args.Error(1)
}

func (m *mockRepository) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RouteStat), args.Error(1)
}

func (m *mockRepository) HistoricalStats(ctx context.Context, window domain.TimeWindow, interval string) (*domain.HistoricalStats, error) {
	args := m.Called(ctx, window, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HistoricalStats), args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) Publish(ctx context.Context, event domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type mockStreamPublisher struct {
	mock.Mock
}

func (m *mockStreamPublisher) Publish(ctx context.Context, stream string, telegram *app.Telegram) error {
	args := m.Called(ctx, stream, telegram)
	return args.Error(0)
}

func TestIngestionService_Handle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		telegram      *app.Telegram
		setupMocks    func(*mockRepository, *mockEventPublisher, *mockEventPublisher, *mockStreamPublisher, interface{})
		expectedError bool
		errorContains string
	}{
		{
			name: "successful ingestion",
			telegram: &app.Telegram{
				MessageID:    "TEST-001",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Source:       "HND",
				Destination:  "SFO",
				Priority:     1,
				Content:      "Test message",
			},
			setupMocks: func(repo *mockRepository, pub *mockEventPublisher, statsPub *mockEventPublisher, streamPub *mockStreamPublisher, statsCounter interface{}) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Telegram")).Return(nil)
				// Stats counter is nil in tests, so these won't be called
				pub.On("Publish", mock.Anything, mock.AnythingOfType("domain.TelegramPersisted")).Return(nil)
				streamPub.On("Publish", mock.Anything, "job:index", mock.MatchedBy(func(tg interface{}) bool {
					if tgApp, ok := tg.(*app.Telegram); ok {
						return tgApp != nil && tgApp.MessageID == "TEST-001"
					}
					if tgDomain, ok := tg.(*domain.Telegram); ok {
						return tgDomain != nil && tgDomain.MessageID == "TEST-001"
					}
					return false
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "invalid telegram",
			telegram: &app.Telegram{
				MessageID: "", // Invalid: empty message ID
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  1,
			},
			setupMocks: func(repo *mockRepository, pub *mockEventPublisher, statsPub *mockEventPublisher, streamPub *mockStreamPublisher, statsCounter interface{}) {
				// No mocks should be called for invalid telegram
			},
			expectedError: true,
			errorContains: "validation failed",
		},
		{
			name: "database save error",
			telegram: &app.Telegram{
				MessageID:    "TEST-002",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Priority:     1,
			},
			setupMocks: func(repo *mockRepository, pub *mockEventPublisher, statsPub *mockEventPublisher, streamPub *mockStreamPublisher, statsCounter interface{}) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Telegram")).Return(errors.New("db error"))
			},
			expectedError: true,
			errorContains: "save to database",
		},
		{
			name: "event publish error (non-fatal)",
			telegram: &app.Telegram{
				MessageID:    "TEST-003",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Priority:     1,
			},
			setupMocks: func(repo *mockRepository, pub *mockEventPublisher, statsPub *mockEventPublisher, streamPub *mockStreamPublisher, statsCounter interface{}) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Telegram")).Return(nil)
				// Stats counter is nil in tests, so these won't be called
				pub.On("Publish", mock.Anything, mock.AnythingOfType("domain.TelegramPersisted")).Return(errors.New("pub error"))
				streamPub.On("Publish", mock.Anything, "job:index", mock.MatchedBy(func(tg interface{}) bool {
					_, ok1 := tg.(*app.Telegram)
					_, ok2 := tg.(*domain.Telegram)
					return ok1 || ok2
				})).Return(nil)
			},
			expectedError: false, // Event publish error is non-fatal
		},
		{
			name: "stream publish error",
			telegram: &app.Telegram{
				MessageID:    "TEST-004",
				Type:         "AFTN",
				Time:         time.Now(),
				FlightNumber: "CA100",
				Priority:     1,
			},
			setupMocks: func(repo *mockRepository, pub *mockEventPublisher, statsPub *mockEventPublisher, streamPub *mockStreamPublisher, statsCounter interface{}) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Telegram")).Return(nil)
				// Stats counter is nil in tests, so these won't be called
				pub.On("Publish", mock.Anything, mock.AnythingOfType("domain.TelegramPersisted")).Return(nil)
				streamPub.On("Publish", mock.Anything, "job:index", mock.MatchedBy(func(tg interface{}) bool {
					if tgApp, ok := tg.(*app.Telegram); ok {
						return tgApp != nil && tgApp.MessageID == "TEST-004"
					}
					if tgDomain, ok := tg.(*domain.Telegram); ok {
						return tgDomain != nil && tgDomain.MessageID == "TEST-004"
					}
					return false
				})).Return(errors.New("stream error"))
			},
			expectedError: true,
			errorContains: "publish to stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := &mockStreamConsumer{}
			repo := &mockRepository{}
			eventPub := &mockEventPublisher{}
			statsEventPub := &mockEventPublisher{}
			streamPub := &mockStreamPublisher{}
			var statsCounter *StatsCounterService

			tt.setupMocks(repo, eventPub, statsEventPub, streamPub, nil)

			svc := NewIngestionService(
				consumer,
				repo,
				eventPub,
				statsEventPub,
				streamPub,
				statsCounter,
				logger,
			)

			err := svc.Handle(ctx, tt.telegram)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
			eventPub.AssertExpectations(t)
			streamPub.AssertExpectations(t)
		})
	}
}

func TestIngestionService_Handle_Normalize(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	telegram := &app.Telegram{
		MessageID:    "test-001",
		Type:         "aftn",
		Time:         time.Now(),
		FlightNumber: "ca100",
		Source:       "hnd",
		Destination:  "sfo",
		Priority:     1,
		Content:      "  test content  ",
	}

	repo := &mockRepository{}
	eventPub := &mockEventPublisher{}
	statsEventPub := &mockEventPublisher{}
	streamPub := &mockStreamPublisher{}

	repo.On("Save", mock.Anything, mock.MatchedBy(func(tg *domain.Telegram) bool {
		// Verify normalization: uppercase flight number, source, destination, type
		return tg.FlightNumber == "CA100" &&
			tg.Source == "HND" &&
			tg.Destination == "SFO" &&
			tg.Type == "AFTN" &&
			tg.Content == "test content"
	})).Return(nil)
	// Stats counter is nil in tests, so these won't be called
	eventPub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	streamPub.On("Publish", mock.Anything, "job:index", mock.MatchedBy(func(tg interface{}) bool {
		if tgApp, ok := tg.(*app.Telegram); ok {
			return tgApp != nil && tgApp.MessageID == "test-001"
		}
		if tgDomain, ok := tg.(*domain.Telegram); ok {
			return tgDomain != nil && tgDomain.MessageID == "test-001"
		}
		return false
	})).Return(nil)

	svc := NewIngestionService(
		&mockStreamConsumer{},
		repo,
		eventPub,
		statsEventPub,
		streamPub,
		nil, // Stats counter is nil in tests
		logger,
	)

	err := svc.Handle(ctx, telegram)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}
