//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	deliveryhttp "github.com/windy/caatsm-dashboard/internal/delivery/http"
	deliveryws "github.com/windy/caatsm-dashboard/internal/delivery/ws"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/cache"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/search"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/service"
	"go.uber.org/zap/zaptest"
)

// TestFullDataFlowIntegration tests the complete data flow from HTTP API to database persistence.
// This integration test verifies that API calls properly interact with the database layer.
func TestFullDataFlowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment with real containers
	env, err := SetupTestEnv(ctx)
	if err != nil {
		t.Skipf("skipping integration test: failed to setup test environment: %v", err)
	}
	defer env.Cleanup(ctx)

	// Initialize real infrastructure components
	store := repository.New(env.Pool)
	meiliClient, err := search.NewMeilisearchClient(config.SearchConfig{
		Host:   env.Meili.URL(),
		APIKey: env.Meili.MasterKey(),
	})
	require.NoError(t, err)
	searchIndex := search.NewMeilisearchIndex(meiliClient, "telegrams-test")
	redisClient := env.Redis.Client()
	cacheStore := cache.NewValkeyStore(redisClient, 5*time.Minute)
	eventBus := event.NewRedisEventBus(redisClient, "msg:broadcast")
	eventPublisher := event.NewEventPublisherAdapter(eventBus)
	logger := zaptest.NewLogger(t)

	// Wire application services with real dependencies
	realtimeManager := service.NewRealtimeManager()
	searchSvc := service.NewSearchService(store, cacheStore, searchIndex, eventPublisher, logger)
	statsSvc := service.NewStatsService(store, cacheStore, logger)
	exportSvc := service.NewExportService(searchSvc, store, logger)
	dashboardSvc := service.NewDashboardService(
		searchSvc,
		statsSvc,
		exportSvc,
		realtimeManager,
		store,
		cacheStore,
		5*time.Minute, // statsTTL
		logger,
	)

	// Setup HTTP routes with real service
	e := echo.New()
	deliveryhttp.RegisterRoutes(e, dashboardSvc, zaptest.NewLogger(t), nil)

	// Setup WebSocket route (using stubs for WS-specific services)
	hub := ws.NewHub(ws.DefaultConfig(), zaptest.NewLogger(t))
	defer hub.Close()
	wsHandler := deliveryws.NewHandler(
		hub,
		zaptest.NewLogger(t),
		&stubStatsService{summary: &app.TrafficSummary{}},
		&stubQueryService{telegrams: []app.Telegram{}},
		nil,
		nil,
	)
	e.GET("/ws", wsHandler.HandleWebSocket)

	// Start test server
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	server := httptest.NewUnstartedServer(e)
	server.Listener = listener
	server.Start()
	defer server.Close()

	// Insert test data directly into database
	testTelegram := NewTelegram("FULL-FLOW-001")
	err = store.Save(ctx, testTelegram)
	require.NoError(t, err)

	// Test HTTP search endpoint - verifies API -> service -> repo -> DB query flow
	resp, err := http.Get(server.URL + "/api/search?query=FULL-FLOW")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var searchResp struct {
		Telegrams []app.Telegram `json:"telegrams"`
		Total     int64          `json:"total"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&searchResp))
	assert.Equal(t, int64(1), searchResp.Total)
	assert.Len(t, searchResp.Telegrams, 1)
	assert.Equal(t, "FULL-FLOW-001", searchResp.Telegrams[0].MessageID)

	// Test stats endpoint - verifies stats calculation from DB
	resp, err = http.Get(server.URL + "/api/stats")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var statsResp app.TrafficSummary
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&statsResp))
	assert.Greater(t, statsResp.TotalMessages, int64(0))
}

// TestAPIErrorScenarios tests various error conditions in the API
func TestAPIErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment
	env, err := SetupTestEnv(ctx)
	if err != nil {
		t.Skipf("skipping integration test: failed to setup test environment: %v", err)
	}
	defer env.Cleanup(ctx)

	// Initialize real infrastructure components
	store := repository.New(env.Pool)
	meiliClient, err := search.NewMeilisearchClient(config.SearchConfig{
		Host:   env.Meili.URL(),
		APIKey: env.Meili.MasterKey(),
	})
	require.NoError(t, err)
	searchIndex := search.NewMeilisearchIndex(meiliClient, "telegrams-test")
	redisClient := env.Redis.Client()
	cacheStore := cache.NewValkeyStore(redisClient, 5*time.Minute)
	eventBus := event.NewRedisEventBus(redisClient, "msg:broadcast")
	eventPublisher := event.NewEventPublisherAdapter(eventBus)
	logger := zaptest.NewLogger(t)

	// Wire application services
	realtimeManager := service.NewRealtimeManager()
	searchSvc := service.NewSearchService(store, cacheStore, searchIndex, eventPublisher, logger)
	statsSvc := service.NewStatsService(store, cacheStore, logger)
	exportSvc := service.NewExportService(searchSvc, store, logger)
	dashboardSvc := service.NewDashboardService(
		searchSvc,
		statsSvc,
		exportSvc,
		realtimeManager,
		store,
		cacheStore,
		5*time.Minute, // statsTTL
		logger,
	)

	// Setup HTTP routes
	e := echo.New()
	deliveryhttp.RegisterRoutes(e, dashboardSvc, zaptest.NewLogger(t), nil)

	// Start test server
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	server := httptest.NewUnstartedServer(e)
	server.Listener = listener
	server.Start()
	defer server.Close()

	t.Run("invalid time range exceeds max window", func(t *testing.T) {
		// Time range > 90 days should return 400
		start := "2024-01-01T00:00:00Z"
		end := "2024-04-02T00:00:00Z" // 92 days (exceeds 90-day maximum)
		url := server.URL + "/api/search?start_time=" + start + "&end_time=" + end

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 400 Bad Request for invalid time range
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid message type", func(t *testing.T) {
		// Invalid type should be ignored or return error
		resp, err := http.Get(server.URL + "/api/search?type=INVALID_TYPE")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 200 but with empty results
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var searchResp struct {
			Telegrams []app.Telegram `json:"telegrams"`
			Total     int64          `json:"total"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&searchResp))
		assert.Equal(t, int64(0), searchResp.Total)
	})

	t.Run("invalid priority", func(t *testing.T) {
		// Invalid priority should be ignored
		resp, err := http.Get(server.URL + "/api/search?priority=invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("malformed time format", func(t *testing.T) {
		// Malformed time should return 400
		resp, err := http.Get(server.URL + "/api/search?start_time=invalid-time")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// stubStatsService is a stub implementation for testing
type stubStatsService struct {
	summary *app.TrafficSummary
}

func (s *stubStatsService) TrafficSummary(ctx context.Context, window interface{}) (interface{}, error) {
	return s.summary, nil
}

// stubQueryService is a stub implementation for testing
type stubQueryService struct {
	telegrams []app.Telegram
}

func (s *stubQueryService) Recent(ctx context.Context, limit int) ([]*app.Telegram, error) {
	result := make([]*app.Telegram, 0, len(s.telegrams))
	for i := range s.telegrams {
		result = append(result, &s.telegrams[i])
	}
	return result, nil
}
