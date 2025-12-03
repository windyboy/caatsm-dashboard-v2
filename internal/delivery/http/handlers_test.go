package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap/zaptest"
)

type mockDashboardService struct {
	lastExportFilters app.SearchFilters
	lastExportFormat  app.ExportFormat
	lastStatsRange    app.TimeWindow
	statsResult       *app.TrafficSummary
	exportData        []byte
}

func (m *mockDashboardService) GetDashboardData(ctx context.Context, req *app.DashboardRequest) (*app.DashboardResponse, error) {
	return nil, nil
}

func (m *mockDashboardService) Search(ctx context.Context, filters app.SearchFilters) (*app.SearchResult, error) {
	return nil, nil
}

func (m *mockDashboardService) GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error) {
	m.lastStatsRange = timeRange
	if m.statsResult != nil {
		return m.statsResult, nil
	}
	return &app.TrafficSummary{}, nil
}

func (m *mockDashboardService) Export(ctx context.Context, filters app.SearchFilters, format app.ExportFormat) ([]byte, error) {
	m.lastExportFilters = filters
	m.lastExportFormat = format
	if m.exportData != nil {
		return m.exportData, nil
	}
	return []byte("export"), nil
}

func (m *mockDashboardService) ExportStream(ctx context.Context, filters app.SearchFilters) (<-chan *app.Telegram, <-chan error, error) {
	m.lastExportFilters = filters
	telegramCh := make(chan *app.Telegram)
	errCh := make(chan error)
	go func() {
		defer close(telegramCh)
		defer close(errCh)
	}()
	return telegramCh, errCh, nil
}

func (m *mockDashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return []string{}, nil
}

func (m *mockDashboardService) AutocompleteWithTypes(ctx context.Context, query string, size int) ([]app.AutocompleteSuggestion, error) {
	return []app.AutocompleteSuggestion{}, nil
}

func (m *mockDashboardService) GetHistoricalStats(ctx context.Context, timeRange app.TimeWindow, interval string) (*app.HistoricalStats, error) {
	return &app.HistoricalStats{}, nil
}

// Note: Full handler tests would require integration tests with real app.
// These are simplified tests that verify basic functionality without mocking
// the entire DashboardService.

func TestGetUserID(t *testing.T) {
	t.Run("valid user_id", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set("user_id", "test-user-123")

		userID, err := getUserID(c)

		require.NoError(t, err)
		assert.Equal(t, "test-user-123", userID)
	})

	t.Run("missing user_id", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)

		userID, err := getUserID(c)

		assert.Error(t, err)
		assert.Empty(t, userID)
	})

	t.Run("empty user_id", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set("user_id", "")

		userID, err := getUserID(c)

		assert.Error(t, err)
		assert.Empty(t, userID)
	})

	t.Run("wrong type", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set("user_id", 12345)

		userID, err := getUserID(c)

		assert.Error(t, err)
		assert.Empty(t, userID)
	})
}

func TestParseTime(t *testing.T) {
	t.Run("valid RFC3339 time", func(t *testing.T) {
		timeStr := "2025-11-27T10:00:00Z"
		result, err := parseTime(timeStr)

		require.NoError(t, err)
		assert.Equal(t, 2025, result.Year())
		assert.Equal(t, time.Month(11), result.Month())
		assert.Equal(t, 27, result.Day())
	})

	t.Run("invalid time format", func(t *testing.T) {
		timeStr := "invalid-time"
		_, err := parseTime(timeStr)

		assert.Error(t, err)
	})
}

func TestHandler_Health(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("health check returns ok", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create handler with nil service (health check doesn't use it)
		handler := &Handler{
			dashboardSvc: nil,
			logger:       logger,
		}

		err := handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
		assert.Contains(t, rec.Body.String(), "dashboard")
	})
}

func TestHandler_Export_ParsesFiltersAndFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockSvc := &mockDashboardService{
		exportData: []byte("data"),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/export?query=test&type=AFTN&source=HND&destination=LAX&priority=2&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z&format=csv", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := &Handler{
		dashboardSvc: mockSvc,
		logger:       logger,
	}

	err := handler.Export(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// CSV format now uses ExportStream, so lastExportFormat won't be set
	// Verify that ExportStream was called with correct filters
	assert.Equal(t, "test", mockSvc.lastExportFilters.Query)
	assert.Equal(t, []string{"AFTN"}, mockSvc.lastExportFilters.Types)
	assert.Equal(t, []string{"HND"}, mockSvc.lastExportFilters.Sources)
	assert.Equal(t, []string{"LAX"}, mockSvc.lastExportFilters.Destinations)
	assert.Equal(t, []int{2}, mockSvc.lastExportFilters.Priorities)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), mockSvc.lastExportFilters.TimeRange.Start)
	assert.Equal(t, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), mockSvc.lastExportFilters.TimeRange.End)
	assert.Equal(t, "text/csv", rec.Header().Get(echo.HeaderContentType))
	assert.Equal(t, "attachment; filename=telegrams.csv", rec.Header().Get("Content-Disposition"))
}

// TODO: Add integration tests with full service stack for:
// - Dashboard endpoint
// - Search endpoint with various filters
// - Stats endpoint with time ranges
// - Export endpoint with different formats
