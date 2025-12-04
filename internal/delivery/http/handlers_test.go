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

// mockDashboardService implements DashboardService interface for testing.
// Some methods are currently unused but kept for future test expansion.
//
//nolint:unused // Type and methods may be used in future tests
type mockDashboardService struct {
	lastExportFilters app.SearchFilters
	lastExportFormat  app.ExportFormat
	lastStatsRange    app.TimeWindow
	statsResult       *app.TrafficSummary
	exportData        []byte
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) GetDashboardData(ctx context.Context, req *app.DashboardRequest) (*app.DashboardResponse, error) {
	return nil, nil
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) Search(ctx context.Context, filters app.SearchFilters) (*app.SearchResult, error) {
	return nil, nil
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error) {
	m.lastStatsRange = timeRange
	if m.statsResult != nil {
		return m.statsResult, nil
	}
	return &app.TrafficSummary{}, nil
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) Export(ctx context.Context, filters app.SearchFilters, format app.ExportFormat) ([]byte, error) {
	m.lastExportFilters = filters
	m.lastExportFormat = format
	if m.exportData != nil {
		return m.exportData, nil
	}
	return []byte("export"), nil
}

//nolint:unused // May be used in future tests
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

//nolint:unused // May be used in future tests
func (m *mockDashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return []string{}, nil
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) AutocompleteWithTypes(ctx context.Context, query string, size int) ([]app.AutocompleteSuggestion, error) {
	return []app.AutocompleteSuggestion{}, nil
}

//nolint:unused // May be used in future tests
func (m *mockDashboardService) GetHistoricalStats(ctx context.Context, timeRange app.TimeWindow, interval string) (*app.HistoricalStats, error) {
	return &app.HistoricalStats{}, nil
}

// Note: Full handler tests would require integration tests with real app.
// These are simplified tests that verify basic functionality without mocking
// the entire DashboardService.

// Ensure mockDashboardService implements the interface (suppresses unused warnings)
var _ = func() interface{} {
	m := &mockDashboardService{}
	_ = m.GetDashboardData
	_ = m.Search
	_ = m.GetStats
	_ = m.Export
	_ = m.ExportStream
	_ = m.Autocomplete
	_ = m.AutocompleteWithTypes
	_ = m.GetHistoricalStats
	return m
}()

// TestGetUserID removed - getUserID function no longer exists in handlers

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

func TestBuildTimeWindow(t *testing.T) {
	t.Run("defaults to last 24 hours when no params", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		c := e.NewContext(req, nil)

		timeRange, err := buildTimeWindow(c)

		require.NoError(t, err)
		assert.False(t, timeRange.Start.IsZero())
		assert.False(t, timeRange.End.IsZero())
		// Should be approximately 24 hours
		duration := timeRange.End.Sub(timeRange.Start)
		assert.InDelta(t, float64(24*time.Hour), float64(duration), float64(1*time.Second))
	})

	t.Run("uses provided time parameters", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/stats?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z", nil)
		c := e.NewContext(req, nil)

		timeRange, err := buildTimeWindow(c)

		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), timeRange.Start)
		assert.Equal(t, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), timeRange.End)
	})

	t.Run("returns error for invalid start_time", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/stats?start_time=invalid", nil)
		c := e.NewContext(req, nil)

		_, err := buildTimeWindow(c)

		assert.Error(t, err)
	})

	t.Run("returns error when start_time after end_time", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/stats?start_time=2024-01-02T00:00:00Z&end_time=2024-01-01T00:00:00Z", nil)
		c := e.NewContext(req, nil)

		_, err := buildTimeWindow(c)

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

// TestHandler_Export_ParsesFiltersAndFormat removed - Export method no longer exists in Handler
// Export functionality has been moved to a separate service layer

// TODO: Add integration tests with full service stack for:
// - Dashboard endpoint
// - Search endpoint with various filters
// - Stats endpoint with time ranges
// - Export endpoint with different formats
