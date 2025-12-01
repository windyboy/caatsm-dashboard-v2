//go:build integration

package testing

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	deliveryhttp "github.com/windy/caatsm-dashboard/internal/delivery/http"
	deliveryws "github.com/windy/caatsm-dashboard/internal/delivery/ws"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap/zaptest"
)

// Integration test that exercises both HTTP and WebSocket transports with the same data,
// mirroring how the frontend consumes the API and live stream.
func TestHTTPAndWebSocketIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	telegrams := []app.Telegram{
		{
			MessageID:    "INT-001",
			Type:         "AFTN",
			Time:         time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC),
			FlightNumber: "CA100",
			Source:       "HND",
			Destination:  "SFO",
			Priority:     1,
			Content:      "Integration message 1",
		},
		{
			MessageID:    "INT-002",
			Type:         "SITA",
			Time:         time.Date(2024, 10, 1, 13, 0, 0, 0, time.UTC),
			FlightNumber: "UA200",
			Source:       "SFO",
			Destination:  "LAX",
			Priority:     3,
			Content:      "Integration message 2",
		},
	}

	statsSummary := app.TrafficSummary{
		TotalMessages: int64(len(telegrams)),
		ByType: map[string]int64{
			"AFTN": 1,
			"SITA": 1,
		},
		ByPriority: map[int]int64{
			1: 1,
			3: 1,
		},
	}

	dashboard := &stubDashboardService{
		telegrams: telegrams,
		stats:     statsSummary,
	}

	// HTTP routes with stubbed dashboard service
	e := echo.New()
	deliveryhttp.RegisterRoutes(e, dashboard, logger)

	// WebSocket route using transport handler (no Redis required for this test)
	hub := ws.NewHub(ws.DefaultConfig(), logger)
	defer hub.Close() // Ensure hub is properly shut down
	wsHandler := deliveryws.NewHandler(
		hub,
		logger,
		&stubStatsService{summary: &statsSummary},
		&stubQueryService{telegrams: telegrams},
		nil,
		nil,
	)
	e.GET("/ws", wsHandler.HandleWebSocket)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping integration test: cannot open listener (%v)", err)
	}
	server := httptest.NewUnstartedServer(e)
	server.Listener = listener
	server.Start()
	defer server.Close()

	// HTTP integration: search endpoint returns payload the frontend expects
	resp, err := http.Get(server.URL + "/api/search?query=test")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var searchResp struct {
		Telegrams []app.Telegram `json:"telegrams"`
		Total     int64          `json:"total"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&searchResp))
	assert.Equal(t, int64(len(telegrams)), searchResp.Total)
	assert.Len(t, searchResp.Telegrams, len(telegrams))

	// WebSocket integration: initial stats + recent messages reach the client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	expectedTypes := map[string]bool{
		"stats-total":    false,
		"stats-priority": false,
		"stats-type":     false,
		"message":        false,
	}

	readDeadline := time.After(3 * time.Second)
	for !allReceived(expectedTypes) {
		select {
		case <-readDeadline:
			t.Fatalf("timeout waiting for all message types, got %+v", expectedTypes)
		default:
			_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			_, data, err := conn.ReadMessage()
			if err != nil {
				continue
			}

			var msg struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			require.NoError(t, json.Unmarshal(data, &msg))

			switch msg.Type {
			case "stats-total":
				var payload struct {
					Total int64 `json:"total"`
				}
				require.NoError(t, json.Unmarshal(msg.Data, &payload))
				assert.Equal(t, statsSummary.TotalMessages, payload.Total)
			case "message":
				expectedTypes["message"] = true
			}

			if _, ok := expectedTypes[msg.Type]; ok {
				expectedTypes[msg.Type] = true
			}
		}
	}
}

func allReceived(flags map[string]bool) bool {
	for _, v := range flags {
		if !v {
			return false
		}
	}
	return true
}

type stubDashboardService struct {
	telegrams []app.Telegram
	stats     app.TrafficSummary
}

func (s *stubDashboardService) GetDashboardData(ctx context.Context, req *app.DashboardRequest) (*app.DashboardResponse, error) {
	return &app.DashboardResponse{
		Search: &app.SearchResult{
			Telegrams: s.telegrams,
			Total:     int64(len(s.telegrams)),
			Page:      app.DefaultPagination(),
		},
		Stats: &s.stats,
	}, nil
}

func (s *stubDashboardService) Search(ctx context.Context, filters app.SearchFilters) (*app.SearchResult, error) {
	return &app.SearchResult{
		Telegrams: s.telegrams,
		Total:     int64(len(s.telegrams)),
		Page:      filters.Pagination,
	}, nil
}

func (s *stubDashboardService) GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error) {
	return &s.stats, nil
}

func (s *stubDashboardService) Export(ctx context.Context, filters app.SearchFilters, format app.ExportFormat) ([]byte, error) {
	return []byte("export"), nil
}

func (s *stubDashboardService) ExportStream(ctx context.Context, filters app.SearchFilters) (<-chan *app.Telegram, <-chan error, error) {
	telegramCh := make(chan *app.Telegram)
	errCh := make(chan error)
	go func() {
		defer close(telegramCh)
		defer close(errCh)
		for i := range s.telegrams {
			telegramCh <- &s.telegrams[i]
		}
	}()
	return telegramCh, errCh, nil
}

func (s *stubDashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return []string{"INT-001", "INT-002"}, nil
}

func (s *stubDashboardService) AutocompleteWithTypes(ctx context.Context, query string, size int) ([]app.AutocompleteSuggestion, error) {
	return []app.AutocompleteSuggestion{
		{Value: "INT-001", Type: "message_id", Label: "Message ID"},
		{Value: "INT-002", Type: "message_id", Label: "Message ID"},
	}, nil
}

type stubStatsService struct {
	summary *app.TrafficSummary
}

func (s *stubStatsService) TrafficSummary(ctx context.Context, window interface{}) (interface{}, error) {
	return s.summary, nil
}

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

// TestFullDataFlowIntegration tests the complete data flow from HTTP API to database persistence.
// This integration test verifies that API calls properly interact with the database layer.
func TestFullDataFlowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment with real containers
	env, err := SetupTestEnv(ctx)
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	// Initialize real infrastructure components
	store := persistence.NewPostgresStore(env.Pool)
	searchIndex := search.NewMeilisearchIndex(env.Meili, "telegrams-test")
	cache := cache.NewValkeyCache(env.Redis)
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")

	// Wire application services with real dependencies
	dashboardSvc := services.NewDashboardService(store, searchIndex, cache, eventBus)

	// Setup HTTP routes with real service
	e := echo.New()
	deliveryhttp.RegisterRoutes(e, dashboardSvc, zaptest.NewLogger(t))

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
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	// Initialize real infrastructure components
	store := persistence.NewPostgresStore(env.Pool)
	searchIndex := search.NewMeilisearchIndex(env.Meili, "telegrams-test")
	cache := cache.NewValkeyCache(env.Redis)
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")

	// Wire application services
	dashboardSvc := services.NewDashboardService(store, searchIndex, cache, eventBus)

	// Setup HTTP routes
	e := echo.New()
	deliveryhttp.RegisterRoutes(e, dashboardSvc, zaptest.NewLogger(t))

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
