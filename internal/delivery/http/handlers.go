package http

import (
	"context"
	stdErrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap"
)

const (
	DefaultLimit  = 50
	MaxSearchLimit = 1000
)

type DashboardService interface {
	Search(ctx context.Context, filters app.SearchFilters) (*app.SearchResult, error)
	GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error)
	GetHistoricalStats(ctx context.Context, timeRange app.TimeWindow, interval string) (*app.HistoricalStats, error)
}

type AdminService interface {
	Reindex(ctx context.Context, from, to time.Time) (interface{}, error)
}

type HealthCheckService interface {
	HealthCheck(ctx context.Context) app.HealthCheckResult
}

type Handler struct {
	dashboardSvc DashboardService
	adminSvc     AdminService
	healthSvc    HealthCheckService
	logger       *zap.Logger
}

func NewHandler(dashboardSvc DashboardService, logger *zap.Logger) *Handler {
	return &Handler{
		dashboardSvc: dashboardSvc,
		logger:       logger,
	}
}

func (h *Handler) SetAdminService(adminSvc AdminService) {
	h.adminSvc = adminSvc
}

func (h *Handler) SetHealthCheckService(healthSvc HealthCheckService) {
	h.healthSvc = healthSvc
}

func (h *Handler) Search(c echo.Context) error {
	filters := parseSearchFilters(c)
	filters.Pagination = parsePagination(c)
	timeRange, err := parseTimeRange(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}
	filters.TimeRange = timeRange

	result, err := h.dashboardSvc.Search(c.Request().Context(), filters)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		return handleError(c, err)
	}

	telegrams := result.Telegrams
	if telegrams == nil {
		telegrams = []app.Telegram{}
	}

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

func (h *Handler) HistoricalStats(c echo.Context) error {
	timeRange, err := buildTimeWindow(c)
	if err != nil {
		h.logger.Error("invalid time range", zap.Error(err))
		return handleError(c, app.ErrInvalidInput)
	}

	interval := c.QueryParam("interval")
	if interval == "" {
		interval = "hour"
	}
	if interval != "hour" && interval != "day" {
		h.logger.Error("invalid interval", zap.String("interval", interval))
		return handleError(c, app.ErrInvalidInput)
	}

	stats, err := h.dashboardSvc.GetHistoricalStats(c.Request().Context(), timeRange, interval)
	if err != nil {
		h.logger.Error("historical stats retrieval failed", zap.Error(err))
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	if h.healthSvc != nil {
		result := h.healthSvc.HealthCheck(ctx)
		status := http.StatusOK
		if result.Status == "degraded" || result.Status == "unhealthy" {
			status = http.StatusServiceUnavailable
		}
		return c.JSON(status, result)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "dashboard",
	})
}

func (h *Handler) Reindex(c echo.Context) error {
	if h.adminSvc == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "admin service not available",
		})
	}

	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")
	if fromStr == "" || toStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from and to parameters are required (format: YYYY-MM-DD)",
		})
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid from date format: %v", err),
		})
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid to date format: %v", err),
		})
	}

	if from.After(to) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from date must be before to date",
		})
	}

	to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	jobID, err := h.adminSvc.Reindex(c.Request().Context(), from, to)
	if err != nil {
		h.logger.Error("reindex failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, jobID)
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

	// If no time parameters provided, default to last 24 hours
	if timeRange.Start.IsZero() && timeRange.End.IsZero() {
		now := time.Now()
		timeRange.End = now
		timeRange.Start = now.Add(-24 * time.Hour)
	}

	if !timeRange.Start.IsZero() && !timeRange.End.IsZero() && timeRange.Start.After(timeRange.End) {
		return app.TimeWindow{}, fmt.Errorf("start_time must be before end_time")
	}

	return timeRange, nil
}

func handleError(c echo.Context, err error) error {
	var apiErr app.APIError

	if e, ok := err.(app.APIError); ok {
		apiErr = e
	} else {
		switch {
		case stdErrors.Is(err, app.ErrNotFound):
			apiErr = app.ErrNotFound
		case stdErrors.Is(err, app.ErrInvalidInput):
			apiErr = app.ErrInvalidInput
		case stdErrors.Is(err, app.ErrUnauthorized):
			apiErr = app.ErrUnauthorized
		default:
			var invalidTelegram app.ErrInvalidTelegram
			var invalidFilter app.ErrInvalidFilter
			if stdErrors.As(err, &invalidTelegram) {
				apiErr = invalidTelegram.ToAPIError()
			} else if stdErrors.As(err, &invalidFilter) {
				apiErr = invalidFilter.ToAPIError()
			} else {
				apiErr = app.ErrInternalError
			}
		}
	}

	status := app.GetHTTPStatus(apiErr.Code)

	return c.JSON(status, apiErr)
}

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

func parseTimeRange(c echo.Context) (app.TimeWindow, error) {
	return buildTimeWindow(c)
}
