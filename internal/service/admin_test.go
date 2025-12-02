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

type mockRepositoryForAdmin struct {
	mock.Mock
}

func (m *mockRepositoryForAdmin) Save(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

func (m *mockRepositoryForAdmin) BulkSave(ctx context.Context, telegrams []any) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

func (m *mockRepositoryForAdmin) Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResult), args.Error(1)
}

func (m *mockRepositoryForAdmin) StreamSearch(ctx context.Context, filter domain.SearchFilters) (<-chan *domain.Telegram, <-chan error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, nil
	}
	// Handle both chan and <-chan types
	tgChan := args.Get(0)
	errChan := args.Get(1)
	
	var telegramChan <-chan *domain.Telegram
	var errorChan <-chan error
	
	if ch, ok := tgChan.(<-chan *domain.Telegram); ok {
		telegramChan = ch
	} else if ch, ok := tgChan.(chan *domain.Telegram); ok {
		telegramChan = ch
	} else {
		return nil, nil
	}
	
	if ch, ok := errChan.(<-chan error); ok {
		errorChan = ch
	} else if ch, ok := errChan.(chan error); ok {
		errorChan = ch
	} else {
		return nil, nil
	}
	
	return telegramChan, errorChan
}

func (m *mockRepositoryForAdmin) FindByID(ctx context.Context, messageID string) (*domain.Telegram, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Telegram), args.Error(1)
}

func (m *mockRepositoryForAdmin) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	args := m.Called(ctx, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TrafficSummary), args.Error(1)
}

func (m *mockRepositoryForAdmin) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RouteStat), args.Error(1)
}

func (m *mockRepositoryForAdmin) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

type mockStreamPublisherForAdmin struct {
	mock.Mock
}

func (m *mockStreamPublisherForAdmin) Publish(ctx context.Context, stream string, telegram *app.Telegram) error {
	args := m.Called(ctx, stream, telegram)
	return args.Error(0)
}

func TestAdminService_Reindex(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		from          time.Time
		to            time.Time
		setupMocks    func(*mockRepositoryForAdmin, *mockStreamPublisherForAdmin)
		expectedError bool
		errorContains string
	}{
		// Note: Testing successful reindex with messages is complex due to channel handling.
		// This is better tested via integration tests. Here we focus on error cases and edge cases.
		{
			name: "reindex with no messages",
			from: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			setupMocks: func(repo *mockRepositoryForAdmin, streamPub *mockStreamPublisherForAdmin) {
				telegramChan := make(chan *domain.Telegram)
				errChan := make(chan error, 1)
				close(telegramChan)
				close(errChan)

				repo.On("StreamSearch", mock.Anything, mock.AnythingOfType("domain.SearchFilters")).Return(telegramChan, errChan)
			},
			expectedError: false,
		},
		{
			name: "stream search error",
			from: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			setupMocks: func(repo *mockRepositoryForAdmin, streamPub *mockStreamPublisherForAdmin) {
				telegramChan := make(chan *domain.Telegram)
				errChan := make(chan error, 1)
				errChan <- errors.New("stream search error")
				close(telegramChan)
				close(errChan)

				repo.On("StreamSearch", mock.Anything, mock.AnythingOfType("domain.SearchFilters")).Return(telegramChan, errChan)
			},
			expectedError: true,
			errorContains: "stream search error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepositoryForAdmin{}
			streamPub := &mockStreamPublisherForAdmin{}

			tt.setupMocks(repo, streamPub)

			svc := NewAdminService(repo, streamPub, logger)

			jobID, err := svc.Reindex(ctx, tt.from, tt.to)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, jobID)
			} else {
				require.NoError(t, err)
				require.NotNil(t, jobID)
				assert.NotEmpty(t, jobID.(*ReindexJobID).ID)
				assert.Equal(t, tt.from, jobID.(*ReindexJobID).From)
				assert.Equal(t, tt.to, jobID.(*ReindexJobID).To)
			}

			repo.AssertExpectations(t)
			streamPub.AssertExpectations(t)
		})
	}
}

func TestAdminService_Reindex_Concurrent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	repo := &mockRepositoryForAdmin{}
	streamPub := &mockStreamPublisherForAdmin{}

	// Setup mocks - use blocking channels to ensure first call holds the lock
	// Create channels that won't be closed immediately, so the first call stays in the loop
	telegramChan1 := make(chan *domain.Telegram) // Unbuffered, will block
	errChan1 := make(chan error, 1)
	
	telegramChan2 := make(chan *domain.Telegram)
	errChan2 := make(chan error, 1)
	close(telegramChan2)
	close(errChan2)
	
	// First call gets a blocking channel, second gets an empty closed channel
	repo.On("StreamSearch", mock.Anything, mock.AnythingOfType("domain.SearchFilters")).Return(telegramChan1, errChan1).Once()
	repo.On("StreamSearch", mock.Anything, mock.AnythingOfType("domain.SearchFilters")).Return(telegramChan2, errChan2).Maybe()

	svc := NewAdminService(repo, streamPub, logger)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	// Start first reindex in goroutine - it will block waiting for channel
	done1 := make(chan bool, 1)
	var err1 error
	var jobID1 interface{}
	go func() {
		jobID1, err1 = svc.Reindex(ctx, from, to)
		done1 <- true
	}()

	// Give first goroutine time to acquire lock and enter the blocking select
	time.Sleep(20 * time.Millisecond)

	// Try to start second reindex (should fail due to mutex)
	_, err2 := svc.Reindex(ctx, from, to)
	
	// Close the first channel to unblock the first goroutine
	close(telegramChan1)
	close(errChan1)
	
	// Wait for first to complete
	<-done1
	
	// Verify that exactly one succeeded and one failed
	successCount := 0
	if err1 == nil && jobID1 != nil {
		successCount++
	}
	if err2 == nil {
		successCount++
	}
	
	// The second call should have failed with "already in progress"
	require.Error(t, err2, "second concurrent call should fail")
	assert.Contains(t, err2.Error(), "already in progress", "error should indicate reindex already in progress")
	
	// Exactly one should succeed
	assert.Equal(t, 1, successCount, "exactly one reindex should succeed")
	require.NoError(t, err1, "first reindex should succeed")
}

