package services

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// Mock Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResult), args.Error(1)
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

func (m *mockRepository) Save(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

func (m *mockRepository) BulkSave(ctx context.Context, telegrams []*domain.Telegram) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

func (m *mockRepository) StreamSearch(ctx context.Context, filters domain.SearchFilters) (<-chan *domain.Telegram, <-chan error) {
	args := m.Called(ctx, filters)
	telegramCh := make(chan *domain.Telegram, 100)
	errCh := make(chan error, 1)

	// If mock returns a result, send it to channel
	if result := args.Get(0); result != nil {
		if searchResult, ok := result.(*domain.SearchResult); ok {
			go func() {
				defer close(telegramCh)
				defer close(errCh)
				for i := range searchResult.Telegrams {
					telegramCh <- &searchResult.Telegrams[i]
				}
			}()
			return telegramCh, errCh
		}
	}

	// If error, send error
	if err := args.Error(1); err != nil {
		go func() {
			defer close(telegramCh)
			defer close(errCh)
			errCh <- err
		}()
		return telegramCh, errCh
	}

	// Default: empty stream
	go func() {
		defer close(telegramCh)
		defer close(errCh)
	}()
	return telegramCh, errCh
}

// Mock Cache
type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Mock SearchIndex
type mockSearchIndex struct {
	mock.Mock
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

func (m *mockSearchIndex) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (interface{}, error) {
	args := m.Called(ctx, query, limit, attributes)
	return args.Get(0), args.Error(1)
}

func (m *mockSearchIndex) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}
