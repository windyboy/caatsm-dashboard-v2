package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
)

func TestDashboardService_Search(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)
	mockSearch := new(mockSearchIndex)

	searchSvc := NewSearchService(mockRepo, cacheMock, mockSearch, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	filters := domain.SearchFilters{
		Query:      "test",
		Pagination: domain.Pagination{Limit: 10},
	}
	expectedResult := &domain.SearchResult{Total: 5}

	cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
	mockRepo.On("Search", ctx, filters).Return(expectedResult, nil)
	cacheMock.On("Set", ctx, mock.AnythingOfType("string"), expectedResult).Return(nil)

	result, err := dashboardSvc.Search(ctx, filters)

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	cacheMock.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}

func TestDashboardService_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)

	searchSvc := NewSearchService(mockRepo, cacheMock, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	timeRange := domain.TimeWindow{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	expectedStats := &domain.TrafficSummary{TotalMessages: 100}

	cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
	mockRepo.On("TrafficSummary", ctx, timeRange).Return(expectedStats, nil)
	cacheMock.On("Set", ctx, mock.AnythingOfType("string"), expectedStats).Return(nil)
	result, err := dashboardSvc.GetStats(ctx, timeRange)

	require.NoError(t, err)
	assert.Equal(t, expectedStats, result)
	cacheMock.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetRealtimeInfo(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)

	searchSvc := NewSearchService(mockRepo, cacheMock, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	// Update realtime info
	realtimeMgr.UpdateConnections(10)
	realtimeMgr.UpdateMessageRate(5.5)

	info := dashboardSvc.GetRealtimeInfo()

	assert.NotNil(t, info)
	assert.Equal(t, 10, info.ActiveConnections)
	assert.Equal(t, 5.5, info.MessagesPerSecond)
}

func TestDashboardService_GetDashboardData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)

	searchSvc := NewSearchService(mockRepo, cacheMock, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	t.Run("with search and stats", func(t *testing.T) {
		req := &DashboardRequest{
			SearchFilters: domain.SearchFilters{
				Query: "test",
			},
			TimeRange: domain.TimeWindow{
				Start: time.Now().Add(-24 * time.Hour),
				End:   time.Now(),
			},
			UserID: "user123",
		}

		searchResult := &domain.SearchResult{Total: 5}
		statsResult := &domain.TrafficSummary{TotalMessages: 100}

		cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss")).Times(2)
		mockRepo.On("Search", ctx, req.SearchFilters).Return(searchResult, nil)
		cacheMock.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)
		mockRepo.On("TrafficSummary", ctx, req.TimeRange).Return(statsResult, nil)
		cacheMock.On("Set", ctx, mock.AnythingOfType("string"), statsResult).Return(nil)

		result, err := dashboardSvc.GetDashboardData(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, searchResult, result.Search)
		assert.Equal(t, statsResult, result.Stats)
		assert.NotNil(t, result.Realtime)
		assert.NotZero(t, result.Timestamp)
		cacheMock.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("without search query", func(t *testing.T) {
		req := &DashboardRequest{
			TimeRange: domain.TimeWindow{
				Start: time.Now().Add(-24 * time.Hour),
				End:   time.Now(),
			},
			UserID: "user123",
		}

		statsResult := &domain.TrafficSummary{TotalMessages: 100}

		cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("TrafficSummary", ctx, req.TimeRange).Return(statsResult, nil)
		cacheMock.On("Set", ctx, mock.AnythingOfType("string"), statsResult).Return(nil)

		result, err := dashboardSvc.GetDashboardData(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Search)
		assert.Equal(t, statsResult, result.Stats)
		assert.NotNil(t, result.Realtime)
		cacheMock.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("search failure", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepoFail := new(mockRepository)
		cacheMockFail := new(mockCache)

		searchSvcFail := NewSearchService(mockRepoFail, cacheMockFail, nil, nil, logger)
		statsSvcFail := NewStatsService(mockRepoFail, cacheMockFail, logger)
		exportSvcFail := NewExportService(searchSvcFail, mockRepoFail, logger)
		realtimeMgrFail := NewRealtimeManager()

		dashboardSvcFail := NewDashboardService(searchSvcFail, statsSvcFail, exportSvcFail, realtimeMgrFail, mockRepoFail, cacheMockFail, time.Hour, logger)

		req := &DashboardRequest{
			SearchFilters: domain.SearchFilters{
				Query: "test",
			},
			TimeRange: domain.TimeWindow{
				Start: time.Now().Add(-24 * time.Hour),
				End:   time.Now(),
			},
			UserID: "user123",
		}

		cacheMockFail.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepoFail.On("Search", ctx, req.SearchFilters).Return(nil, errors.New("search failed"))

		result, err := dashboardSvcFail.GetDashboardData(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search failed")
		cacheMockFail.AssertExpectations(t)
		mockRepoFail.AssertExpectations(t)
	})

	t.Run("stats failure", func(t *testing.T) {
		req := &DashboardRequest{
			TimeRange: domain.TimeWindow{
				Start: time.Now().Add(-24 * time.Hour),
				End:   time.Now(),
			},
			UserID: "user123",
		}

		cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("TrafficSummary", ctx, req.TimeRange).Return(nil, errors.New("stats failed"))

		result, err := dashboardSvc.GetDashboardData(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "stats failed")
		cacheMock.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestDashboardService_StreamInitialData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)

	searchSvc := NewSearchService(mockRepo, cacheMock, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	t.Run("successful stream", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)

		statsResult := &domain.TrafficSummary{
			TotalMessages: 100,
			ByPriority:    map[int]int64{1: 50, 2: 30, 3: 20},
			ByType:        map[string]int64{"METAR": 40, "TAF": 60},
		}

		searchResult := &domain.SearchResult{
			Telegrams: []domain.Telegram{
				{MessageID: "msg1", Type: "METAR"},
				{MessageID: "msg2", Type: "TAF"},
			},
		}

		cacheMock.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("TrafficSummary", ctx, mock.AnythingOfType("domain.TimeWindow")).Return(statsResult, nil)
		cacheMock.On("Set", ctx, mock.AnythingOfType("string"), statsResult).Return(nil)
		mockRepo.On("Search", ctx, mock.AnythingOfType("domain.SearchFilters")).Return(searchResult, nil)

		err := dashboardSvc.StreamInitialData(ctx, ch)

		require.NoError(t, err)

		// Should receive 4 messages: stats-total, stats-priority, stats-type, message, message
		assert.Len(t, ch, 5)

		messages := make([]ports.WSMessage, 0, 5)
		close(ch)
		for msg := range ch {
			messages = append(messages, msg)
		}

		// Verify message types
		expectedTypes := []string{"stats-total", "stats-priority", "stats-type", "message", "message"}
		for i, msg := range messages {
			assert.Equal(t, expectedTypes[i], msg.Type)
		}

		cacheMock.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("stats failure", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepoFail := new(mockRepository)
		cacheMockFail := new(mockCache)

		searchSvcFail := NewSearchService(mockRepoFail, cacheMockFail, nil, nil, logger)
		statsSvcFail := NewStatsService(mockRepoFail, cacheMockFail, logger)
		exportSvcFail := NewExportService(searchSvcFail, mockRepoFail, logger)
		realtimeMgrFail := NewRealtimeManager()

		dashboardSvcFail := NewDashboardService(searchSvcFail, statsSvcFail, exportSvcFail, realtimeMgrFail, mockRepoFail, cacheMockFail, time.Hour, logger)

		ch := make(chan ports.WSMessage, 10)

		cacheMockFail.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepoFail.On("TrafficSummary", ctx, mock.AnythingOfType("domain.TimeWindow")).Return(nil, errors.New("stats failed"))

		err := dashboardSvcFail.StreamInitialData(ctx, ch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get stats")
		mockRepoFail.AssertExpectations(t)
		cacheMockFail.AssertExpectations(t)
	})

	t.Run("search failure", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepoFail := new(mockRepository)
		cacheMockFail := new(mockCache)

		searchSvcFail := NewSearchService(mockRepoFail, cacheMockFail, nil, nil, logger)
		statsSvcFail := NewStatsService(mockRepoFail, cacheMockFail, logger)
		exportSvcFail := NewExportService(searchSvcFail, mockRepoFail, logger)
		realtimeMgrFail := NewRealtimeManager()

		dashboardSvcFail := NewDashboardService(searchSvcFail, statsSvcFail, exportSvcFail, realtimeMgrFail, mockRepoFail, cacheMockFail, time.Hour, logger)

		ch := make(chan ports.WSMessage, 10)

		statsResult := &domain.TrafficSummary{TotalMessages: 100}

		cacheMockFail.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepoFail.On("TrafficSummary", ctx, mock.AnythingOfType("domain.TimeWindow")).Return(statsResult, nil)
		cacheMockFail.On("Set", ctx, mock.AnythingOfType("string"), statsResult).Return(nil)
		mockRepoFail.On("Search", ctx, mock.AnythingOfType("domain.SearchFilters")).Return(nil, errors.New("search failed"))

		err := dashboardSvcFail.StreamInitialData(ctx, ch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get recent messages")
		mockRepoFail.AssertExpectations(t)
		cacheMockFail.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		err := dashboardSvc.StreamInitialData(cancelCtx, ch)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestDashboardService_HandleEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	mockRepo := new(mockRepository)
	cacheMock := new(mockCache)

	searchSvc := NewSearchService(mockRepo, cacheMock, nil, nil, logger)
	statsSvc := NewStatsService(mockRepo, cacheMock, logger)
	exportSvc := NewExportService(searchSvc, mockRepo, logger)
	realtimeMgr := NewRealtimeManager()

	dashboardSvc := NewDashboardService(searchSvc, statsSvc, exportSvc, realtimeMgr, mockRepo, cacheMock, time.Hour, logger)

	t.Run("successful event handling", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)

		telegram := &domain.Telegram{
			MessageID:    "msg123",
			Type:         "METAR",
			Priority:     1,
			FlightNumber: "BA123",
			Source:       "EGLL",
			Destination:  "KJFK",
			Content:      "Test message",
		}

		event := domain.TelegramReceived{Telegram: telegram}
		eventBytes, _ := json.Marshal(event)

		// Mock cache operations
		cacheMock.On("Incr", ctx, "ws:stats:total:90d", int64(1)).Return(int64(101), nil)
		cacheMock.On("HIncrBy", ctx, "ws:stats:priority:90d", "1", int64(1)).Return(int64(51), nil)
		cacheMock.On("HIncrBy", ctx, "ws:stats:type:90d", "METAR", int64(1)).Return(int64(41), nil)

		// Mock updated stats retrieval
		cacheMock.On("GetInt64", ctx, "ws:stats:total:90d").Return(int64(101), nil)
		cacheMock.On("HGetAll", ctx, "ws:stats:priority:90d").Return(map[string]string{"1": "51", "2": "30"}, nil)
		cacheMock.On("HGetAll", ctx, "ws:stats:type:90d").Return(map[string]string{"METAR": "41", "TAF": "60"}, nil)

		err := dashboardSvc.HandleEvent(ctx, eventBytes, ch)

		require.NoError(t, err)

		// Should receive 4 messages: message, stats-total, stats-priority, stats-type
		assert.Len(t, ch, 4)

		messages := make([]ports.WSMessage, 0, 4)
		close(ch)
		for msg := range ch {
			messages = append(messages, msg)
		}

		// Verify message types
		expectedTypes := []string{"message", "stats-total", "stats-priority", "stats-type"}
		for i, msg := range messages {
			assert.Equal(t, expectedTypes[i], msg.Type)
		}

		// Verify message data
		assert.Equal(t, telegram, messages[0].Data)
		assert.Equal(t, map[string]interface{}{"total": int64(101)}, messages[1].Data)

		cacheMock.AssertExpectations(t)
	})

	t.Run("invalid event JSON", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)

		err := dashboardSvc.HandleEvent(ctx, []byte("invalid json"), ch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse event")
	})

	t.Run("nil telegram", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)

		event := domain.TelegramReceived{Telegram: nil}
		eventBytes, _ := json.Marshal(event)

		err := dashboardSvc.HandleEvent(ctx, eventBytes, ch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "telegram is nil")
	})

	t.Run("cache update failure", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)

		telegram := &domain.Telegram{MessageID: "msg123", Type: "METAR", Priority: 1}
		event := domain.TelegramReceived{Telegram: telegram}
		eventBytes, _ := json.Marshal(event)

		// Mock cache failure
		cacheMock.On("Incr", ctx, "ws:stats:total:90d", int64(1)).Return(int64(0), errors.New("cache error"))

		// Still expect message to be sent despite cache failure
		cacheMock.On("GetInt64", ctx, "ws:stats:total:90d").Return(int64(100), nil)
		cacheMock.On("HGetAll", ctx, "ws:stats:priority:90d").Return(map[string]string{}, nil)
		cacheMock.On("HGetAll", ctx, "ws:stats:type:90d").Return(map[string]string{}, nil)

		err := dashboardSvc.HandleEvent(ctx, eventBytes, ch)

		require.NoError(t, err) // Should not fail due to cache error

		// Should still receive messages
		assert.Len(t, ch, 4)

		cacheMock.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ch := make(chan ports.WSMessage, 10)
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		telegram := &domain.Telegram{MessageID: "msg123", Type: "METAR", Priority: 1}
		event := domain.TelegramReceived{Telegram: telegram}
		eventBytes, _ := json.Marshal(event)

		err := dashboardSvc.HandleEvent(cancelCtx, eventBytes, ch)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
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
