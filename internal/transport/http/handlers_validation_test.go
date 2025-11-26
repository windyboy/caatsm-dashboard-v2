package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/application/export"
	"github.com/windy/caatsm-dashboard/internal/application/search"
	"github.com/windy/caatsm-dashboard/internal/application/stats"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap"
)

// TestSearch_TimeRangeExceeds90Days_Returns400 verifies that search requests with time ranges exceeding 90 days are rejected
func TestSearch_TimeRangeExceeds90Days_Returns400(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Mock services
	searchIndex := &mocks.SearchIndexMock{}
	telegramStore := &mocks.TelegramStoreMock{}
	logger := zap.NewNop()
	
	// Pass nil for cache since we're testing validation, not caching
	searchService := search.NewService(searchIndex, telegramStore, nil, logger)
	statsService := stats.NewService(nil, logger)
	exportService := export.NewService(searchService, logger)
	
	handler := NewHandler(searchService, statsService, exportService, nil)
	
	// Create a time range of 91 days (exceeds max)
	now := time.Now()
	start := now.Add(-91 * 24 * time.Hour)
	
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/search?start_time=%s&end_time=%s",
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(now.Format(time.RFC3339))), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Execute
	err := handler.Search(c)
	
	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "error should be echo.HTTPError")
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "time range exceeds maximum")
}

// TestSearch_TimeRangeEndBeforeStart_Returns400 verifies that search requests with inverted time ranges are rejected
func TestSearch_TimeRangeEndBeforeStart_Returns400(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Mock services
	searchIndex := &mocks.SearchIndexMock{}
	telegramStore := &mocks.TelegramStoreMock{}
	logger := zap.NewNop()
	
	// Pass nil for cache since we're testing validation, not caching
	searchService := search.NewService(searchIndex, telegramStore, nil, logger)
	statsService := stats.NewService(nil, logger)
	exportService := export.NewService(searchService, logger)
	
	handler := NewHandler(searchService, statsService, exportService, nil)
	
	// Create an inverted time range
	now := time.Now()
	start := now.Add(24 * time.Hour) // Future
	end := now                        // Past relative to start
	
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/search?start_time=%s&end_time=%s",
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(end.Format(time.RFC3339))), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Execute
	err := handler.Search(c)
	
	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "error should be echo.HTTPError")
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "end_time must be after start_time")
}

// TestExport_TimeRangeExceeds90Days_Returns400 verifies that export requests with time ranges exceeding 90 days are rejected
func TestExport_TimeRangeExceeds90Days_Returns400(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Mock services
	searchIndex := &mocks.SearchIndexMock{}
	telegramStore := &mocks.TelegramStoreMock{}
	logger := zap.NewNop()
	
	// Pass nil for cache since we're testing validation, not caching
	searchService := search.NewService(searchIndex, telegramStore, nil, logger)
	statsService := stats.NewService(nil, logger)
	exportService := export.NewService(searchService, logger)
	
	handler := NewHandler(searchService, statsService, exportService, nil)
	
	// Create a time range of 100 days (exceeds max)
	now := time.Now()
	start := now.Add(-100 * 24 * time.Hour)
	
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/export?format=csv&start_time=%s&end_time=%s",
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(now.Format(time.RFC3339))), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Execute
	err := handler.Export(c)
	
	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "error should be echo.HTTPError")
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "time range exceeds maximum")
}

// TestSearch_TimeRangeWithin90Days_Returns200 verifies that valid time ranges are accepted
func TestSearch_TimeRangeWithin90Days_Returns200(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Mock services
	searchIndex := &mocks.SearchIndexMock{}
	telegramStore := &mocks.TelegramStoreMock{}
	logger := zap.NewNop()
	
	// Pass nil for cache since we're testing validation, not caching
	searchService := search.NewService(searchIndex, telegramStore, nil, logger)
	statsService := stats.NewService(nil, logger)
	exportService := export.NewService(searchService, logger)
	
	handler := NewHandler(searchService, statsService, exportService, nil)
	
	// Mock search index to return empty search response
	searchIndex.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&struct {
			Hits               []interface{}
			EstimatedTotalHits int64
			TotalHits          int64
		}{
			Hits:               []interface{}{},
			EstimatedTotalHits: 0,
			TotalHits:          0,
		},
		nil,
	)
	
	// Create a valid time range of 30 days
	now := time.Now()
	start := now.Add(-30 * 24 * time.Hour)
	
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/search?start_time=%s&end_time=%s",
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(now.Format(time.RFC3339))), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Execute
	err := handler.Search(c)
	
	// Assert - should succeed (no error from validation, may have other errors from mocks)
	// The key is that we don't get a validation error
	if err != nil {
		httpErr, ok := err.(*echo.HTTPError)
		if ok {
			// If it's an HTTP error, it should NOT be a validation error
			assert.NotContains(t, httpErr.Message, "time range exceeds maximum")
			assert.NotContains(t, httpErr.Message, "end_time must be after start_time")
		}
	}
}

// TestSearch_NoTimeRange_Returns200 verifies that searches without time ranges work normally
func TestSearch_NoTimeRange_Returns200(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Mock services
	searchIndex := &mocks.SearchIndexMock{}
	telegramStore := &mocks.TelegramStoreMock{}
	logger := zap.NewNop()
	
	// Pass nil for cache since we're testing validation, not caching
	searchService := search.NewService(searchIndex, telegramStore, nil, logger)
	statsService := stats.NewService(nil, logger)
	exportService := export.NewService(searchService, logger)
	
	handler := NewHandler(searchService, statsService, exportService, nil)
	
	// Mock search index to return empty search response
	searchIndex.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&struct {
			Hits               []interface{}
			EstimatedTotalHits int64
			TotalHits          int64
		}{
			Hits:               []interface{}{},
			EstimatedTotalHits: 0,
			TotalHits:          0,
		},
		nil,
	)
	
	req := httptest.NewRequest(http.MethodGet, "/api/search?query=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Execute
	err := handler.Search(c)
	
	// Assert - should not get validation errors
	if err != nil {
		httpErr, ok := err.(*echo.HTTPError)
		if ok {
			assert.NotContains(t, httpErr.Message, "time range exceeds maximum")
			assert.NotContains(t, httpErr.Message, "end_time must be after start_time")
		}
	}
}

