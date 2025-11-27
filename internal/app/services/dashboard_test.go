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

func TestDashboardService_Search(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	mockCache := new(mockCache)
	mockSearch := new(mockSearchIndex)

	searchSvc := NewSearchService(mockRepo, mockCache, mockSearch, nil, logger)
	statsSvc := NewStatsService(mockRepo, mockCache, logger)
	exportSvc := NewExportService(searchSvc, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, logger)

	filters := domain.SearchFilters{
		Query:      "test",
		Pagination: domain.Pagination{Limit: 10},
	}
	expectedResult := &domain.SearchResult{Total: 5}

	mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
	mockRepo.On("Search", ctx, filters).Return(expectedResult, nil)
	mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedResult).Return(nil)

	result, err := dashboardSvc.Search(ctx, filters)

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}

func TestDashboardService_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	mockCache := new(mockCache)

	searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, mockCache, logger)
	exportSvc := NewExportService(searchSvc, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, logger)

	timeRange := domain.TimeWindow{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	expectedStats := &domain.TrafficSummary{TotalMessages: 100}

	mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
	mockRepo.On("TrafficSummary", ctx, timeRange).Return(expectedStats, nil)
	mockCache.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).Return(nil)
	result, err := dashboardSvc.GetStats(ctx, timeRange)

	require.NoError(t, err)
	assert.Equal(t, expectedStats, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetRealtimeInfo(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockRepo := new(mockRepository)
	mockCache := new(mockCache)

	searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, mockCache, logger)
	exportSvc := NewExportService(searchSvc, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, logger)

	// Update realtime info
	realtimeMgr.UpdateConnections(10)
	realtimeMgr.UpdateMessageRate(5.5)

	info := dashboardSvc.GetRealtimeInfo()

	assert.NotNil(t, info)
	assert.Equal(t, 10, info.ActiveConnections)
	assert.Equal(t, 5.5, info.MessagesPerSecond)
}

func TestRealtimeManager(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		mgr := NewRealtimeManager()
		info := mgr.GetInfo()

		assert.Equal(t, 0, info.ActiveConnections)
		assert.Equal(t, 0.0, info.MessagesPerSecond)
		assert.NotEmpty(t, info.Uptime)
	})

	t.Run("update connections", func(t *testing.T) {
		mgr := NewRealtimeManager()
		mgr.UpdateConnections(10)

		info := mgr.GetInfo()
		assert.Equal(t, 10, info.ActiveConnections)
	})

	t.Run("update message rate", func(t *testing.T) {
		mgr := NewRealtimeManager()
		mgr.UpdateMessageRate(5.5)

		info := mgr.GetInfo()
		assert.Equal(t, 5.5, info.MessagesPerSecond)
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		mgr := NewRealtimeManager()
		done := make(chan bool, 2)

		// Writer goroutine
		go func() {
			for i := 0; i < 100; i++ {
				mgr.UpdateConnections(i)
				mgr.UpdateMessageRate(float64(i))
			}
			done <- true
		}()

		// Reader goroutine
		go func() {
			for i := 0; i < 100; i++ {
				_ = mgr.GetInfo()
			}
			done <- true
		}()

		<-done
		<-done
		// Test passes if no race condition
	})
}
