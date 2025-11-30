package http

import (
	"context"
	stdErrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

// Pagination and limit constants
const (
	// DefaultLimit is the default number of records returned when limit is not specified
	DefaultLimit = 50
	// MaxSearchLimit is the maximum number of records that can be returned in a search/listing query
	MaxSearchLimit = 1000
)

// DashboardService defines the application boundary used by this transport layer.
type DashboardService interface {
	GetDashboardData(ctx context.Context, req *app.DashboardRequest) (*app.DashboardResponse, error)
	Search(ctx context.Context, filters app.SearchFilters) (*app.SearchResult, error)
	GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error)
	Export(ctx context.Context, filters app.SearchFilters, format app.ExportFormat) ([]byte, error)
	ExportStream(ctx context.Context, filters app.SearchFilters) (<-chan *app.Telegram, <-chan error, error)
	Autocomplete(ctx context.Context, query string, size int) ([]string, error)
	AutocompleteWithTypes(ctx context.Context, query string, size int) ([]app.AutocompleteSuggestion, error)
}

// Handler handles HTTP requests for the dashboard
type Handler struct {
	dashboardSvc DashboardService
	logger       *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(dashboardSvc DashboardService, logger *zap.Logger) *Handler {
	return &Handler{
		dashboardSvc: dashboardSvc,
		logger:       logger,
	}
}

// Dashboard handles GET /api/dashboard - unified dashboard data endpoint
func (h *Handler) Dashboard(c echo.Context) error {
	// Extract user ID from context
	userID, err := getUserID(c)
	if err != nil {
		h.logger.Error("failed to get user_id from context", zap.Error(err))
		return handleError(c, err)
	}

	// Parse query parameters
	req := &app.DashboardRequest{
		UserID: userID,
	}

	// Parse search filters
	if query := c.QueryParam("query"); query != "" {
		req.SearchFilters = app.SearchFilters{
			Query: query,
			Pagination: app.Pagination{
				Limit:  DefaultLimit,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
		}

		// Parse pagination
		if limitStr := c.QueryParam("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= MaxSearchLimit {
				req.SearchFilters.Pagination.Limit = limit
			}
		}
		if offsetStr := c.QueryParam("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				req.SearchFilters.Pagination.Offset = offset
			}
		}
	}

	// Parse time range for stats
	timeRange, err := buildTimeWindow(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}
	req.TimeRange = timeRange

	// Get dashboard data
	result, err := h.dashboardSvc.GetDashboardData(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

// Search handles POST/GET /api/search - search telegrams
func (h *Handler) Search(c echo.Context) error {
	filters := parseSearchFilters(c)
	filters.Pagination = parsePagination(c)
	timeRange, err := parseTimeRange(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}
	filters.TimeRange = timeRange

	// Execute search
	result, err := h.dashboardSvc.Search(c.Request().Context(), filters)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		return handleError(c, err)
	}

	// Ensure telegrams is always a non-nil slice (empty array instead of nil)
	telegrams := result.Telegrams
	if telegrams == nil {
		telegrams = []app.Telegram{}
	}

	// Convert page object to have proper JSON field names
	page := map[string]interface{}{
		"limit":   result.Page.Limit,
		"offset":  result.Page.Offset,
		"sort_by": result.Page.SortBy,
		"order":   result.Page.Order,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"telegrams": telegrams,
		"total":     result.Total,
		"page":      page,
	})
}

// Stats handles GET /api/stats - statistics endpoints
func (h *Handler) Stats(c echo.Context) error {
	timeRange, err := buildTimeWindow(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}

	stats, err := h.dashboardSvc.GetStats(c.Request().Context(), timeRange)
	if err != nil {
		h.logger.Error("stats retrieval failed", zap.Error(err))
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, stats)
}

// Export handles GET /api/export - data export
func (h *Handler) Export(c echo.Context) error {
	filters := parseSearchFilters(c)
	filters.Pagination = app.Pagination{
		Limit:  app.MaxExportRecords,
		Offset: 0,
	}
	timeRange, err := parseTimeRange(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}
	filters.TimeRange = timeRange

	formatStr := strings.ToLower(c.QueryParam("format"))
	if formatStr == "" {
		formatStr = "csv"
	}

	var format app.ExportFormat
	switch formatStr {
	case "csv":
		format = app.ExportFormatCSV
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid format. supported: csv",
		})
	}

	// Use streaming export for CSV (supports large datasets)
	if format == app.ExportFormatCSV {
		stream, errCh, err := h.dashboardSvc.ExportStream(c.Request().Context(), filters)
		if err != nil {
			h.logger.Error("export stream failed", zap.Error(err))
			return handleError(c, err)
		}
		return StreamCSV(c, stream, errCh)
	}

	// Fallback to non-streaming export for other formats or small datasets
	data, err := h.dashboardSvc.Export(c.Request().Context(), filters, format)
	if err != nil {
		h.logger.Error("export failed", zap.Error(err))
		return handleError(c, err)
	}

	// Set appropriate headers
	contentType := "application/octet-stream"
	filename := "telegrams." + formatStr

	switch format {
	case app.ExportFormatCSV:
		contentType = "text/csv"
	}

	c.Response().Header().Set(echo.HeaderContentType, contentType)
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)

	return c.Blob(http.StatusOK, contentType, data)
}

// Health handles GET /api/health - health check
func (h *Handler) Health(c echo.Context) error {
	// For now, simple health check
	// In production, this would check all dependencies
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "dashboard",
	})
}

// Helper functions

// getUserID safely extracts the user_id from the echo context.
// Returns an error if user_id is missing or not a string.
func getUserID(c echo.Context) (string, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return "", app.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", app.ErrUnauthorized
	}

	if userIDStr == "" {
		return "", app.ErrUnauthorized
	}

	return userIDStr, nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

func buildTimeWindow(c echo.Context) (app.TimeWindow, error) {
	timeRange := app.TimeWindow{}

	if startStr := c.QueryParam("start_time"); startStr != "" {
		start, err := parseTime(startStr)
		if err != nil {
			return app.TimeWindow{}, fmt.Errorf("invalid start_time: %w", err)
		}
		timeRange.Start = start
	}
	if endStr := c.QueryParam("end_time"); endStr != "" {
		end, err := parseTime(endStr)
		if err != nil {
			return app.TimeWindow{}, fmt.Errorf("invalid end_time: %w", err)
		}
		timeRange.End = end
	}

	if !timeRange.Start.IsZero() && !timeRange.End.IsZero() && timeRange.Start.After(timeRange.End) {
		return app.TimeWindow{}, fmt.Errorf("start_time must be before end_time")
	}

	return timeRange, nil
}

// handleError handles API errors and returns standardized error responses
func handleError(c echo.Context, err error) error {
	var apiErr app.APIError

	// Check if it's already an APIError
	if e, ok := err.(app.APIError); ok {
		apiErr = e
	} else {
		// Handle legacy errors and convert them
		switch {
		case stdErrors.Is(err, app.ErrNotFound):
			apiErr = app.ErrNotFound
		case stdErrors.Is(err, app.ErrInvalidInput):
			apiErr = app.ErrInvalidInput
		case stdErrors.Is(err, app.ErrUnauthorized):
			apiErr = app.ErrUnauthorized
		default:
			// Check for custom domain errors
			if invalidTelegram, ok := err.(app.ErrInvalidTelegram); ok {
				apiErr = invalidTelegram.ToAPIError()
			} else if invalidFilter, ok := err.(app.ErrInvalidFilter); ok {
				apiErr = invalidFilter.ToAPIError()
			} else {
				// Default to internal error for unknown errors
				apiErr = app.ErrInternalError
			}
		}
	}

	// Get HTTP status from error code
	status := app.GetHTTPStatus(apiErr.Code)

	return c.JSON(status, apiErr)
}

// parseSearchFilters parses search filter parameters from the request context
func parseSearchFilters(c echo.Context) app.SearchFilters {
	filters := app.SearchFilters{}

	if query := c.QueryParam("query"); query != "" {
		filters.Query = query
	}
	if types := c.QueryParam("type"); types != "" {
		filters.Types = []string{types}
	}
	if source := c.QueryParam("source"); source != "" {
		filters.Sources = []string{source}
	}
	if dest := c.QueryParam("destination"); dest != "" {
		filters.Destinations = []string{dest}
	}
	if priorityStr := c.QueryParam("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			filters.Priorities = []int{priority}
		}
	}

	return filters
}

// parsePagination parses pagination parameters from the request context
func parsePagination(c echo.Context) app.Pagination {
	pagination := app.Pagination{
		Limit:  DefaultLimit,
		Offset: 0,
		SortBy: "time",
		Order:  "desc",
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= MaxSearchLimit {
			pagination.Limit = limit
		}
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			pagination.Offset = offset
		}
	}
	if sortBy := c.QueryParam("sort_by"); sortBy != "" {
		pagination.SortBy = sortBy
	}
	if order := c.QueryParam("order"); order != "" {
		pagination.Order = order
	}

	return pagination
}

// parseTimeRange parses time range parameters from the request context
func parseTimeRange(c echo.Context) (app.TimeWindow, error) {
	return buildTimeWindow(c)
}
