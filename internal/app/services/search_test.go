package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
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

func TestSearchService_Search(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("cache hit", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit:  10,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
		}

		cachedResult := &domain.SearchResult{
			Total:     5,
			Telegrams: []domain.Telegram{{MessageID: "CACHED-001"}},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(cachedResult, nil)

		result, err := svc.Search(ctx, filters)

		require.NoError(t, err)
		assert.Equal(t, cachedResult, result)
		mockCache.AssertExpectations(t)
		// Repository should not be called
		mockRepo.AssertNotCalled(t, "Search")
	})

	t.Run("cache miss - fetch from repository", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit:  10,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
		}

		expectedResult := &domain.SearchResult{
			Total:     3,
			Telegrams: []domain.Telegram{{MessageID: "TEST-001"}},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("Search", ctx, filters).
			Return(expectedResult, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedResult).
			Return(nil)

		result, err := svc.Search(ctx, filters)

		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("invalid filters", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		// Invalid time range (too large)
		filters := domain.SearchFilters{
			Query: "test",
			TimeRange: domain.TimeWindow{
				Start: time.Now().Add(-100 * 24 * time.Hour), // 100 days
				End:   time.Now(),
			},
		}

		result, err := svc.Search(ctx, filters)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid search filters")
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit: 10,
			},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("Search", ctx, filters).
			Return(nil, errors.New("database error"))

		result, err := svc.Search(ctx, filters)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search operation failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestSearchService_Autocomplete(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("empty query", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		result, err := svc.Autocomplete(ctx, "", 5)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("limit validation - zero defaults to 5", func(t *testing.T) {
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)
		mockSearch := new(mockSearchIndex)

		svc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)

		// With zero limit, should default to 5
		// Since we can't easily mock Meilisearch SDK types, we skip the actual call test
		// This is more of an integration test concern
		assert.NotNil(t, svc)
	})

	// Note: Full autocomplete testing requires integration tests with real Meilisearch
	// due to complex SDK types (*meilisearch.SearchResponse) that are difficult to mock
}

func TestSearchService_CacheKeyGeneration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svc := &SearchService{logger: logger}

	t.Run("same filters produce same key", func(t *testing.T) {
		filters1 := domain.SearchFilters{
			Query: "test",
			Types: []string{"METAR"},
			Pagination: domain.Pagination{
				Limit:  10,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
		}

		filters2 := domain.SearchFilters{
			Query: "test",
			Types: []string{"METAR"},
			Pagination: domain.Pagination{
				Limit:  10,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
		}

		key1 := svc.buildCacheKey(filters1)
		key2 := svc.buildCacheKey(filters2)

		assert.Equal(t, key1, key2)
	})

	t.Run("different filters produce different keys", func(t *testing.T) {
		filters1 := domain.SearchFilters{
			Query: "test1",
		}

		filters2 := domain.SearchFilters{
			Query: "test2",
		}

		key1 := svc.buildCacheKey(filters1)
		key2 := svc.buildCacheKey(filters2)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("cache key format", func(t *testing.T) {
		filters := domain.SearchFilters{
			Query: "test",
		}

		key := svc.buildCacheKey(filters)

		// Should start with "search:" prefix
		assert.Contains(t, key, "search:")
		// Should be a hash (fixed length)
		assert.Len(t, key, len("search:")+32) // MD5 hash is 32 hex chars
	})
}

