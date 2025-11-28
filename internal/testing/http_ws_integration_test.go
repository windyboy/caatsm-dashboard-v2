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
	"github.com/windy/caatsm-dashboard/internal/app/services"
	deliveryhttp "github.com/windy/caatsm-dashboard/internal/delivery/http"
	deliveryws "github.com/windy/caatsm-dashboard/internal/delivery/ws"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"go.uber.org/zap/zaptest"
)

// Integration test that exercises both HTTP and WebSocket transports with the same data,
// mirroring how the frontend consumes the API and live stream.
func TestHTTPAndWebSocketIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	telegrams := []domain.Telegram{
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

	statsSummary := domain.TrafficSummary{
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
	wsHandler := deliveryws.NewHandler(
		hub,
		logger,
		&stubStatsService{summary: toPersistenceSummary(statsSummary)},
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
		Telegrams []domain.Telegram `json:"telegrams"`
		Total     int64             `json:"total"`
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
	telegrams []domain.Telegram
	stats     domain.TrafficSummary
}

func (s *stubDashboardService) GetDashboardData(ctx context.Context, req *services.DashboardRequest) (*services.DashboardResponse, error) {
	return &services.DashboardResponse{
		Search: &domain.SearchResult{
			Telegrams: s.telegrams,
			Total:     int64(len(s.telegrams)),
			Page:      domain.DefaultPagination(),
		},
		Stats: &s.stats,
	}, nil
}

func (s *stubDashboardService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	return &domain.SearchResult{
		Telegrams: s.telegrams,
		Total:     int64(len(s.telegrams)),
		Page:      filters.Pagination,
	}, nil
}

func (s *stubDashboardService) GetStats(ctx context.Context, timeRange domain.TimeWindow) (*domain.TrafficSummary, error) {
	return &s.stats, nil
}

func (s *stubDashboardService) Export(ctx context.Context, filters domain.SearchFilters, format domain.ExportFormat) ([]byte, error) {
	return []byte("export"), nil
}

func (s *stubDashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return []string{"INT-001", "INT-002"}, nil
}

type stubStatsService struct {
	summary *persistence.TrafficSummary
}

func (s *stubStatsService) TrafficSummary(ctx context.Context, window interface{}) (interface{}, error) {
	return s.summary, nil
}

type stubQueryService struct {
	telegrams []domain.Telegram
}

func (s *stubQueryService) Recent(ctx context.Context, limit int) ([]*domain.Telegram, error) {
	result := make([]*domain.Telegram, 0, len(s.telegrams))
	for i := range s.telegrams {
		result = append(result, &s.telegrams[i])
	}
	return result, nil
}

func toPersistenceSummary(summary domain.TrafficSummary) *persistence.TrafficSummary {
	return &persistence.TrafficSummary{
		TotalMessages: summary.TotalMessages,
		ByType:        summary.ByType,
		ByPriority:    summary.ByPriority,
	}
}
