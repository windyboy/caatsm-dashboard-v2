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

func TestStatsService_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("cache hit", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		timeRange := domain.TimeWindow{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		}

		cachedStats := &domain.TrafficSummary{
			TotalMessages: 1000,
			ByType:        map[string]int64{"METAR": 500, "TAF": 500},
			ByPriority:    map[int]int64{1: 800, 2: 200},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(cachedStats, nil)

		result, err := svc.GetStats(ctx, timeRange)

		require.NoError(t, err)
		assert.Equal(t, cachedStats, result)
		mockCache.AssertExpectations(t)
		// Repository should not be called
		mockRepo.AssertNotCalled(t, "TrafficSummary")
	})

	t.Run("cache miss - fetch from repository", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		timeRange := domain.TimeWindow{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		}

		expectedStats := &domain.TrafficSummary{
			TotalMessages: 500,
			ByType:        map[string]int64{"METAR": 300, "TAF": 200},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("TrafficSummary", ctx, timeRange).
			Return(expectedStats, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).
			Return(nil)

		result, err := svc.GetStats(ctx, timeRange)

		require.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("cache type mismatch - fetch fresh data", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		timeRange := domain.TimeWindow{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		}

		// Cache returns wrong type
		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return("invalid type", nil)

		expectedStats := &domain.TrafficSummary{
			TotalMessages: 100,
		}

		mockRepo.On("TrafficSummary", ctx, timeRange).
			Return(expectedStats, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).
			Return(nil)

		result, err := svc.GetStats(ctx, timeRange)

		require.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		timeRange := domain.TimeWindow{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("TrafficSummary", ctx, timeRange).
			Return(nil, errors.New("database error"))

		result, err := svc.GetStats(ctx, timeRange)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get stats failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestStatsService_GetRouteStats(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("successful fetch", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		limit := 10
		expectedStats := []domain.RouteStat{
			{Source: "EGLL", Destination: "KJFK", Count: 100},
			{Source: "KJFK", Destination: "EGLL", Count: 95},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("RouteStats", ctx, limit).
			Return(expectedStats, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).
			Return(nil)

		result, err := svc.GetRouteStats(ctx, limit)

		require.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("limit validation - zero defaults to 10", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		expectedStats := []domain.RouteStat{
			{Source: "EGLL", Destination: "KJFK", Count: 100},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("RouteStats", ctx, 10).
			Return(expectedStats, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).
			Return(nil)

		result, err := svc.GetRouteStats(ctx, 0)

		require.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("limit validation - exceeds max capped to 100", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		expectedStats := []domain.RouteStat{}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("RouteStats", ctx, 100).
			Return(expectedStats, nil)

		mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).
			Return(nil)

		result, err := svc.GetRouteStats(ctx, 500)

		require.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cache hit for route stats", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		limit := 20
		cachedStats := []domain.RouteStat{
			{Source: "EGLL", Destination: "KJFK", Count: 100},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(cachedStats, nil)

		result, err := svc.GetRouteStats(ctx, limit)

		require.NoError(t, err)
		assert.Equal(t, cachedStats, result)
		mockCache.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "RouteStats")
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		svc := NewStatsService(mockRepo, mockCache, logger)

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).
			Return(nil, errors.New("cache miss"))

		mockRepo.On("RouteStats", ctx, 10).
			Return(nil, errors.New("database error"))

		result, err := svc.GetRouteStats(ctx, 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get route stats failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestStatsService_CacheKeyGeneration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svc := &StatsService{logger: logger}

	t.Run("same time range produces same key", func(t *testing.T) {
		now := time.Now()
		timeRange1 := domain.TimeWindow{
			Start: now.Add(-24 * time.Hour),
			End:   now,
		}
		timeRange2 := domain.TimeWindow{
			Start: now.Add(-24 * time.Hour),
			End:   now,
		}

		key1 := svc.buildCacheKey(timeRange1)
		key2 := svc.buildCacheKey(timeRange2)

		assert.Equal(t, key1, key2)
	})

	t.Run("different time ranges produce different keys", func(t *testing.T) {
		now := time.Now()
		timeRange1 := domain.TimeWindow{
			Start: now.Add(-24 * time.Hour),
			End:   now,
		}
		timeRange2 := domain.TimeWindow{
			Start: now.Add(-48 * time.Hour),
			End:   now,
		}

		key1 := svc.buildCacheKey(timeRange1)
		key2 := svc.buildCacheKey(timeRange2)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("cache key format", func(t *testing.T) {
		timeRange := domain.TimeWindow{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		}

		key := svc.buildCacheKey(timeRange)

		// Should start with "stats:" prefix
		assert.Contains(t, key, "stats:")
	})
}
