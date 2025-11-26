package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/application/export"
	"github.com/windy/caatsm-dashboard/internal/application/search"
	"github.com/windy/caatsm-dashboard/internal/application/stats"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

// Handler bundles HTTP handlers using the new application services.
type Handler struct {
	searchService *search.Service
	statsService  *stats.Service
	exportService *export.Service
	healthCheck   HealthChecker
}

// HealthChecker defines the interface for health checks.
type HealthChecker interface {
	HealthCheck(ctx echo.Context) HealthCheckResult
}

// HealthCheckResult contains health status.
type HealthCheckResult struct {
	Status      string          `json:"status"`
	PostgreSQL  ComponentHealth `json:"postgresql"`
	Meilisearch ComponentHealth `json:"meilisearch"`
	Redis       ComponentHealth `json:"redis"`
}

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewHandler creates a new HTTP handler.
func NewHandler(
	searchService *search.Service,
	statsService *stats.Service,
	exportService *export.Service,
	healthCheck HealthChecker,
) *Handler {
	return &Handler{
		searchService: searchService,
		statsService:  statsService,
		exportService: exportService,
		healthCheck:   healthCheck,
	}
}

// Register attaches all HTTP routes to the provided Echo instance.
func Register(e *echo.Echo, h *Handler) {
	// HTMX action routes
	actions := e.Group("/actions")
	actions.POST("/search", h.Search)
	actions.GET("/autocomplete", h.Autocomplete)
	actions.GET("/stats/total", h.StatsTotal)
	actions.GET("/stats/priority", h.StatsPriority)
	actions.GET("/stats/type", h.StatsType)

	// API routes (JSON)
	api := e.Group("/api")
	api.POST("/search", h.Search)
	api.GET("/search", h.Search)
	api.GET("/autocomplete", h.Autocomplete)
	api.GET("/stats/total", h.StatsTotal)
	api.GET("/stats/priority", h.StatsPriority)
	api.GET("/stats/type", h.StatsType)
	api.GET("/export", h.Export)
	api.GET("/health", h.Health)

	// Root route
	e.GET("/", h.Root)
}

// Search godoc
// @Summary Search telegrams
// @Description Search aviation telegrams with full-text search and filters
// @Tags search
// @Accept json
// @Produce json
// @Param query query string false "Search query text"
// @Param limit query int false "Number of results per page (max 1000)" default(50)
// @Param offset query int false "Pagination offset (max 100000)" default(0)
// @Param sort_by query string false "Sort field (time, priority, message_id, type, flight_number, source, destination)" default(time)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Param type query string false "Filter by telegram type"
// @Param source query string false "Filter by source airport (ICAO code)"
// @Param destination query string false "Filter by destination airport (ICAO code)"
// @Param priority query int false "Filter by priority (1-3)"
// @Param start_time query string false "Filter start time (RFC3339 format)"
// @Param end_time query string false "Filter end time (RFC3339 format, max 90 days range)"
// @Success 200 {object} map[string]interface{} "Search results with telegrams array, total count, and pagination"
// @Failure 400 {object} map[string]string "Bad request (invalid parameters)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /search [get]
func (h *Handler) Search(ctx echo.Context) error {
	filter := persistence.SearchFilter{
		Query: sanitizeQuery(ctx.QueryParam("query")),
		Page: persistence.Pagination{
			Limit:  DefaultPageLimit,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	// Parse pagination
	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > maxLimit {
				limit = maxLimit
			}
			filter.Page.Limit = limit
		}
	}
	if offsetStr := ctx.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			if offset > maxOffset {
				return echo.NewHTTPError(http.StatusBadRequest, "offset too large")
			}
			filter.Page.Offset = offset
		}
	}
	if sortBy := ctx.QueryParam("sort_by"); sortBy != "" {
		filter.Page.SortBy = sortBy
	}
	if sortBy := ctx.QueryParam("sort_by"); sortBy != "" {
		allowedSortFields := map[string]bool{
			"time": true, "priority": true, "message_id": true,
			"type": true, "flight_number": true, "source": true, "destination": true,
		}
		if !allowedSortFields[sortBy] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort_by field")
		}
		filter.Page.SortBy = sortBy
	}
	if order := ctx.QueryParam("order"); order != "" {
		if order != "asc" && order != "desc" {
			return echo.NewHTTPError(http.StatusBadRequest, "order must be 'asc' or 'desc'")
		}
		filter.Page.Order = order
	}
		filter.Type = []string{typeStr}
	}
	if srcStr := ctx.QueryParam("source"); srcStr != "" {
		filter.Source = []string{srcStr}
	}
	if dstStr := ctx.QueryParam("destination"); dstStr != "" {
		filter.Destination = []string{dstStr}
	}
	if priorityStr := ctx.QueryParam("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil && priority >= 0 {
			filter.Priority = []int{priority}
		}
	}

	// Parse time range
	if startStr := ctx.QueryParam("start_time"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.TimeRange.Start = start
		}
	}
	if endStr := ctx.QueryParam("end_time"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.TimeRange.End = end
		}
	}

	// Validate time range
	if !filter.TimeRange.Start.IsZero() && !filter.TimeRange.End.IsZero() {
		duration := filter.TimeRange.End.Sub(filter.TimeRange.Start)
		maxDuration := time.Duration(maxTimeRangeDays) * 24 * time.Hour
		if duration > maxDuration {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Sprintf("time range exceeds maximum of %d days", maxTimeRangeDays),
			)
		}
		if filter.TimeRange.End.Before(filter.TimeRange.Start) {
			return echo.NewHTTPError(http.StatusBadRequest, "end_time must be after start_time")
		}
	}

	// Convert persistence filter to domain filter
	domainFilter := domain.SearchFilterToDomain(filter)
	
	result, err := h.searchService.Search(ctx.Request().Context(), domainFilter)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Search not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"telegrams": result.Telegrams,
		"total":     result.Total,
		"page":      result.Page,
	})
}

// Autocomplete godoc
// @Summary Get autocomplete suggestions
// @Description Get search term suggestions for autocomplete
// @Tags search
// @Accept json
// @Produce json
// @Param term query string true "Search term"
// @Param size query int false "Number of suggestions (max 50)" default(5)
// @Success 200 {object} map[string]interface{} "Autocomplete suggestions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /autocomplete [get]
func (h *Handler) Autocomplete(ctx echo.Context) error {
	sizeStr := ctx.QueryParam("size")
	size, _ := strconv.Atoi(sizeStr)
	if size <= 0 {
		size = DefaultAutocompleteSize
	}
	if size > maxAutocompleteSize {
		size = maxAutocompleteSize
	}

	suggestions, err := h.searchService.Autocomplete(ctx.Request().Context(), ctx.QueryParam("term"), size)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Autocomplete not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"suggestions": suggestions,
	})
}

// StatsTotal godoc
// @Summary Get total telegram count
// @Description Get total number of telegrams in the system
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Total telegram count"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /stats/total [get]
func (h *Handler) StatsTotal(ctx echo.Context) error {
	window := persistence.TimeWindow{}

	// Convert to domain TimeWindow
	domainWindow := domain.TimeWindowToDomain(window)
	summary, err := h.statsService.TrafficSummary(ctx.Request().Context(), domainWindow)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Stats not available"})
		}
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"total": summary.TotalMessages,
	})
}

// StatsPriority godoc
// @Summary Get telegram statistics by priority
// @Description Get breakdown of telegrams by priority level
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Telegram count by priority"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /stats/priority [get]
func (h *Handler) StatsPriority(ctx echo.Context) error {
	window := persistence.TimeWindow{}

	// Convert to domain TimeWindow
	domainWindow := domain.TimeWindowToDomain(window)
	summary, err := h.statsService.TrafficSummary(ctx.Request().Context(), domainWindow)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Stats not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"byPriority": summary.ByPriority,
	})
}

// StatsType godoc
// @Summary Get telegram statistics by type
// @Description Get breakdown of telegrams by message type (AFTN, SITA, ACARS, CPDLC)
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Telegram count by type"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /stats/type [get]
func (h *Handler) StatsType(ctx echo.Context) error {
	window := persistence.TimeWindow{}

	// Convert to domain TimeWindow
	domainWindow := domain.TimeWindowToDomain(window)
	summary, err := h.statsService.TrafficSummary(ctx.Request().Context(), domainWindow)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Stats not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"byType": summary.ByType,
	})
}

	format := persistence.ExportFormat(ctx.QueryParam("format"))
	if format == "" {
		format = persistence.ExportFormatCSV
	}
	if format != persistence.ExportFormatCSV && 
	   format != persistence.ExportFormatExcel && 
	   format != persistence.ExportFormatPDF {
		return echo.NewHTTPError(http.StatusBadRequest, "format must be csv, xlsx, or pdf")
	}
// @Accept json
// @Produce text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf
// @Param format query string false "Export format (csv, xlsx, pdf)" default(csv)
// @Param query query string false "Search query text"
// @Param type query string false "Filter by telegram type"
// @Param source query string false "Filter by source airport (ICAO code)"
// @Param destination query string false "Filter by destination airport (ICAO code)"
// @Success 200 {file} file "Exported file"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /export [get]
func (h *Handler) Export(ctx echo.Context) error {
	format := persistence.ExportFormat(ctx.QueryParam("format"))
	if format == "" {
		format = persistence.ExportFormatCSV
	}

	filter := persistence.SearchFilter{
		Query: ctx.QueryParam("query"),
		Page: persistence.Pagination{
			Limit:  maxExportLimit,
			Offset: 0,
		},
	}

	// Parse filters (same as Search)
	if typeStr := ctx.QueryParam("type"); typeStr != "" {
		filter.Type = []string{typeStr}
	}
	if srcStr := ctx.QueryParam("source"); srcStr != "" {
		filter.Source = []string{srcStr}
	}
	if dstStr := ctx.QueryParam("destination"); dstStr != "" {
		filter.Destination = []string{dstStr}
	}

	// Parse time range
	if startStr := ctx.QueryParam("start_time"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.TimeRange.Start = start
		}
	}
	if endStr := ctx.QueryParam("end_time"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.TimeRange.End = end
		}
	}

	// Validate time range
	if !filter.TimeRange.Start.IsZero() && !filter.TimeRange.End.IsZero() {
		duration := filter.TimeRange.End.Sub(filter.TimeRange.Start)
		maxDuration := time.Duration(maxTimeRangeDays) * 24 * time.Hour
		if duration > maxDuration {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Sprintf("time range exceeds maximum of %d days", maxTimeRangeDays),
			)
		}
		if filter.TimeRange.End.Before(filter.TimeRange.Start) {
			return echo.NewHTTPError(http.StatusBadRequest, "end_time must be after start_time")
		}
	}

	// Determine content type
	contentType := "application/octet-stream"
	switch format {
	case persistence.ExportFormatCSV:
		contentType = "text/csv"
	case persistence.ExportFormatExcel:
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case persistence.ExportFormatPDF:
		contentType = "application/pdf"
	}

	// Set headers for streaming download
	ctx.Response().Header().Set(echo.HeaderContentType, contentType)
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=telegrams.%s", format))
	ctx.Response().WriteHeader(http.StatusOK)

	// Convert persistence types to domain types
	domainFilter := domain.SearchFilterToDomain(filter)
	domainFormat := domain.ExportFormatToDomain(format)
	
	// Stream directly to response writer to avoid memory issues
	if err := h.exportService.ExportStream(ctx.Request().Context(), domainFilter, domainFormat, ctx.Response().Writer); err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			// Note: Can't change status code after WriteHeader, log error instead
			return fmt.Errorf("export not implemented: %w", err)
		}
		return err
	}

	return nil
}

// Health godoc
// @Summary Health check
// @Description Check health status of all system components (PostgreSQL, Meilisearch, Redis, NATS, WebSocket)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "System is healthy"
// @Failure 503 {object} map[string]interface{} "System is degraded"
// @Router /health [get]
func (h *Handler) Health(ctx echo.Context) error {
	if h.healthCheck == nil {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	health := h.healthCheck.HealthCheck(ctx)

	statusCode := http.StatusOK
	if health.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	return ctx.JSON(statusCode, health)
}

func (h *Handler) Root(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"message": "CAATSM Dashboard API",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"api":          "/api",
			"websocket":    "/ws",
			"health":       "/api/health",
			"metrics":      "/metrics",
			"search":       "/api/search",
			"stats":        "/api/stats/total, /api/stats/priority, /api/stats/type",
			"export":       "/api/export",
			"autocomplete": "/api/autocomplete",
		},
		"frontend": map[string]string{
			"development": "http://localhost:5173 (run 'task frontend:dev' or 'cd frontend && deno task dev')",
			"production":  "Build frontend with 'task frontend:build' or 'cd frontend && deno task build'",
		},
		"note": "In development, access the frontend at http://localhost:5173. This backend (port 3002) only serves API endpoints.",
	})
}

func sanitizeQuery(q string) string {
	q = strings.TrimSpace(q)
	if len(q) > maxQueryLength {
		q = q[:maxQueryLength]
	}
	return q
}
