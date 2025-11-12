package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/views/pages"
)

// Handler bundles view and API handlers.
type Handler struct {
	container *app.Container
}

// Register attaches all HTTP routes to the provided Echo instance.
func Register(e *echo.Echo, c *app.Container) {
	h := &Handler{container: c}

	e.GET("/", render(pages.Dashboard()))
	e.GET("/search", render(pages.Search()))

	api := e.Group("/api")
	api.POST("/search", h.Search)
	api.GET("/autocomplete", h.Autocomplete)
	api.GET("/stats/total", h.StatsTotal)
	api.GET("/stats/priority", h.StatsPriority)
	api.GET("/export", h.Export)
	api.GET("/stream", h.Stream)
	api.GET("/health", h.Health)
}

func (h *Handler) Search(ctx echo.Context) error {
	filter := models.SearchFilter{
		Query: ctx.FormValue("query"),
		Page: models.Pagination{
			Limit:  50,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	// Parse pagination
	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Page.Limit = limit
		}
	}
	if offsetStr := ctx.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
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
		if priority, err := strconv.Atoi(priorityStr); err == nil {
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
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": err.Error()})
		}
		return err
	}
	return ctx.JSON(http.StatusOK, result)
}

func (h *Handler) Autocomplete(ctx echo.Context) error {
	sizeStr := ctx.QueryParam("size")
	size, _ := strconv.Atoi(sizeStr)
	if size <= 0 {
		size = 5
	}

	suggestions, err := h.container.SearchService.Autocomplete(ctx.Request().Context(), ctx.QueryParam("term"), size)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": err.Error()})
		}
		return err
	}

	return ctx.JSON(http.StatusOK, suggestions)
}

func (h *Handler) StatsTotal(ctx echo.Context) error {
	window := models.TimeWindow{
		End: time.Now(),
	}
	// Default to last 24 hours
	window.Start = window.End.Add(-24 * time.Hour)

	summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": err.Error()})
		}
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"total": summary.TotalMessages,
	})
}

func (h *Handler) StatsPriority(ctx echo.Context) error {
	window := models.TimeWindow{
		End: time.Now(),
	}
	// Default to last 24 hours
	window.Start = window.End.Add(-24 * time.Hour)

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": err.Error()})
		}
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"by_priority": stats.ByPriority,
		"by_type":     stats.ByType,
	})
}

func (h *Handler) Export(ctx echo.Context) error {
	format := models.ExportFormat(ctx.QueryParam("format"))
	if format == "" {
		format = models.ExportFormatCSV
	}

	filter := models.SearchFilter{
		Query: ctx.QueryParam("query"),
		Page: models.Pagination{
			Limit:  10000,
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

func (h *Handler) Stream(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": repository.ErrNotImplemented.Error()})
}

func (h *Handler) Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func render(component templ.Component) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		return component.Render(ctx.Request().Context(), ctx.Response().Writer)
	}
}
