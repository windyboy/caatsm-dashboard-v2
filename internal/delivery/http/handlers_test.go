package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// Note: Full handler tests would require integration tests with real services.
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

// TODO: Add integration tests with full service stack for:
// - Dashboard endpoint
// - Search endpoint with various filters
// - Stats endpoint with time ranges
// - Export endpoint with different formats
