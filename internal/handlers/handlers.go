package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

const (
	maxQueryLength      = 500
	maxLimit            = 1000
	maxOffset           = 100000
	maxAutocompleteSize = 50
	maxExportLimit      = 10000
)

// Handler bundles view and API handlers.
type Handler struct {
	container   *app.Container
	broadcaster *EventBroadcaster
}

// Register attaches all HTTP routes to the provided Echo instance.
func Register(e *echo.Echo, c *app.Container, broadcaster *EventBroadcaster) {
	h := &Handler{
		container:   c,
		broadcaster: broadcaster,
	}

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

	// WebSocket route
	e.GET("/ws", h.WebSocket)

	// Root route - helpful message for development
	e.GET("/", h.Root)
}

func (h *Handler) Search(ctx echo.Context) error {
	filter := models.SearchFilter{
		Query: sanitizeQuery(ctx.FormValue("query")),
		Page: models.Pagination{
			Limit:  DefaultPageLimit,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	if len(filter.Query) > maxQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "query too long")
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
	if order := ctx.QueryParam("order"); order != "" {
		filter.Page.Order = order
	}

	// Parse filters
	if typeStr := ctx.QueryParam("type"); typeStr != "" {
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

	result, err := h.container.SearchService.Search(ctx.Request().Context(), filter)
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

func (h *Handler) Autocomplete(ctx echo.Context) error {
	sizeStr := ctx.QueryParam("size")
	size, _ := strconv.Atoi(sizeStr)
	if size <= 0 {
		size = DefaultAutocompleteSize
	}
	if size > maxAutocompleteSize {
		size = maxAutocompleteSize
	}

	suggestions, err := h.container.SearchService.Autocomplete(ctx.Request().Context(), ctx.QueryParam("term"), size)
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

func (h *Handler) StatsTotal(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
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

func (h *Handler) StatsPriority(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Stats not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"byPriority": stats.ByPriority,
	})
}

func (h *Handler) StatsType(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Stats not available"})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"byType": stats.ByType,
	})
}

func (h *Handler) Export(ctx echo.Context) error {
	// Require authentication for export
	if !h.container.Config.Auth.EnableBasic {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required for export")
	}

	format := models.ExportFormat(ctx.QueryParam("format"))
	if format == "" {
		format = models.ExportFormatCSV
	}

	filter := models.SearchFilter{
		Query: ctx.QueryParam("query"),
		Page: models.Pagination{
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

	payload, err := h.container.ExportService.Export(ctx.Request().Context(), filter, format)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": err.Error()})
		}
		return err
	}

	contentType := "application/octet-stream"
	switch format {
	case models.ExportFormatCSV:
		contentType = "text/csv"
	case models.ExportFormatExcel:
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case models.ExportFormatPDF:
		contentType = "application/pdf"
	}

	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=telegrams.%s", format))
	return ctx.Blob(http.StatusOK, contentType, payload)
}

func (h *Handler) Health(ctx echo.Context) error {
	health := h.container.HealthCheck(ctx.Request().Context())

	// Return 200 if all components are healthy, 503 if degraded
	statusCode := http.StatusOK
	if health.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	return ctx.JSON(statusCode, health)
}

// Root provides information about the API and how to access the frontend
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
	// Remove control characters and limit length
	q = strings.TrimSpace(q)
	if len(q) > maxQueryLength {
		q = q[:maxQueryLength]
	}
	return q
}
