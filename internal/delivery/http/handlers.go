package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app/services"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/pkg/errors"
	"go.uber.org/zap"
)

// Pagination and limit constants
const (
	// DefaultLimit is the default number of records returned when limit is not specified
	DefaultLimit = 50
	// MaxSearchLimit is the maximum number of records that can be returned in a search/listing query
	MaxSearchLimit = 1000
	// MaxExportLimit is the maximum number of records that can be exported at once
	MaxExportLimit = 10000
)

// DashboardService defines the application boundary used by this transport layer.
type DashboardService interface {
	GetDashboardData(ctx context.Context, req *services.DashboardRequest) (*services.DashboardResponse, error)
	Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error)
	GetStats(ctx context.Context, timeRange domain.TimeWindow) (*domain.TrafficSummary, error)
	Export(ctx context.Context, filters domain.SearchFilters, format domain.ExportFormat) ([]byte, error)
	Autocomplete(ctx context.Context, query string, size int) ([]string, error)
	AutocompleteWithTypes(ctx context.Context, query string, size int) ([]services.AutocompleteSuggestion, error)
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
	req := &services.DashboardRequest{
		UserID: userID,
	}

	// Parse search filters
	if query := c.QueryParam("query"); query != "" {
		req.SearchFilters = domain.SearchFilters{
			Query: query,
			Pagination: domain.Pagination{
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
	req.TimeRange = buildTimeWindow(c)

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
	filters.TimeRange = parseTimeRange(c)

	// Execute search
	result, err := h.dashboardSvc.Search(c.Request().Context(), filters)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		return handleError(c, err)
	}

	// Ensure telegrams is always a non-nil slice (empty array instead of nil)
	domainTelegrams := result.Telegrams
	if domainTelegrams == nil {
		domainTelegrams = []domain.Telegram{}
	}

	// Convert domain.Telegram to persistence.Telegram for proper JSON serialization
	// This ensures snake_case field names (message_id, flight_number, etc.) are used
	telegrams := make([]persistence.Telegram, len(domainTelegrams))
	for i, dt := range domainTelegrams {
		pt := domain.FromDomain(&dt)
		if pt == nil {
			// This should never happen, but handle gracefully
			h.logger.Warn("failed to convert domain telegram to persistence model",
				zap.String("message_id", dt.MessageID),
				zap.Int("index", i))
			// Create empty persistence model as fallback
			telegrams[i] = persistence.Telegram{}
			continue
		}
		telegrams[i] = *pt
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
	timeRange := buildTimeWindow(c)

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
	filters.Pagination = domain.Pagination{
		Limit:  MaxExportLimit,
		Offset: 0,
	}
	filters.TimeRange = parseTimeRange(c)

	formatStr := strings.ToLower(c.QueryParam("format"))
	if formatStr == "" {
		formatStr = "csv"
	}

	var format domain.ExportFormat
	switch formatStr {
	case "csv":
		format = domain.ExportFormatCSV
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid format. supported: csv",
		})
	}

	// Execute export
	data, err := h.dashboardSvc.Export(c.Request().Context(), filters, format)
	if err != nil {
		h.logger.Error("export failed", zap.Error(err))
		return handleError(c, err)
	}

	// Set appropriate headers
	contentType := "application/octet-stream"
	filename := "telegrams." + formatStr

	switch format {
	case domain.ExportFormatCSV:
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
		return "", errors.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.ErrUnauthorized
	}

	if userIDStr == "" {
		return "", errors.ErrUnauthorized
	}

	return userIDStr, nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

func buildTimeWindow(c echo.Context) domain.TimeWindow {
	timeRange := domain.TimeWindow{}

	if startStr := c.QueryParam("start_time"); startStr != "" {
		if start, err := parseTime(startStr); err == nil {
			timeRange.Start = start
		}
	}
	if endStr := c.QueryParam("end_time"); endStr != "" {
		if end, err := parseTime(endStr); err == nil {
			timeRange.End = end
		}
	}

	return timeRange
}

func handleError(c echo.Context, err error) error {
	switch err {
	case errors.ErrNotFound:
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.ErrInvalidInput:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
	case errors.ErrUnauthorized:
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	default:
		// Note: In a real implementation, we'd need access to logger here
		// For now, just return generic error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// parseSearchFilters parses search filter parameters from the request context
func parseSearchFilters(c echo.Context) domain.SearchFilters {
	filters := domain.SearchFilters{}

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
func parsePagination(c echo.Context) domain.Pagination {
	pagination := domain.Pagination{
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
func parseTimeRange(c echo.Context) domain.TimeWindow {
	return buildTimeWindow(c)
}
