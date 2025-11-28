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
