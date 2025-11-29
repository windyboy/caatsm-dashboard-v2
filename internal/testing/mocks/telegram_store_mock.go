package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// RepositoryMock is a mock implementation of ports.Repository
type RepositoryMock struct {
	mock.Mock
}

// Ensure RepositoryMock implements ports.Repository
var _ ports.Repository = (*RepositoryMock)(nil)

// Save mocks the Save method
func (m *RepositoryMock) Save(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

// BulkSave mocks the BulkSave method
func (m *RepositoryMock) BulkSave(ctx context.Context, telegrams []*domain.Telegram) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

// Search mocks the Search method
func (m *RepositoryMock) Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResult), args.Error(1)
}

// TrafficSummary mocks the TrafficSummary method
func (m *RepositoryMock) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	args := m.Called(ctx, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TrafficSummary), args.Error(1)
}

// RouteStats mocks the RouteStats method
func (m *RepositoryMock) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RouteStat), args.Error(1)
}

// StreamSearch mocks the StreamSearch method
func (m *RepositoryMock) StreamSearch(ctx context.Context, filters domain.SearchFilters) (<-chan *domain.Telegram, <-chan error) {
	args := m.Called(ctx, filters)
	telegramCh := make(chan *domain.Telegram, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(telegramCh)
		defer close(errCh)

		// If mock returns a result, send it to channel
		if result := args.Get(0); result != nil {
			if searchResult, ok := result.(*domain.SearchResult); ok {
				for i := range searchResult.Telegrams {
					telegramCh <- &searchResult.Telegrams[i]
				}
				return
			}
		}

		// If error, send error
		if err := args.Error(1); err != nil {
			errCh <- err
			return
		}
	}()

	return telegramCh, errCh
}

// TelegramStoreMock is kept for backward compatibility, now wraps RepositoryMock
// Deprecated: Use RepositoryMock instead
type TelegramStoreMock struct {
	*RepositoryMock
}
