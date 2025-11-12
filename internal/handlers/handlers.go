package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
			Limit:  DefaultPageLimit,
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
			return ctx.HTML(http.StatusNotImplemented, "<p class=\"text-slate-500\">Search not available</p>")
		}
		return err
	}

	// Build HTML response
	var html strings.Builder
	html.WriteString(fmt.Sprintf("<p class=\"text-sm text-slate-400 mb-4\">Found %d results</p>", result.Total))
	if len(result.Telegrams) > 0 {
		html.WriteString("<div class=\"space-y-2\">")
		for _, t := range result.Telegrams {
			html.WriteString(fmt.Sprintf(`
			<div class="rounded border border-slate-800 p-4">
				<div class="flex justify-between mb-2">
					<span class="font-medium">%s</span>
					<span class="text-xs text-slate-500">%s</span>
				</div>
				<p class="text-sm text-slate-300">%s</p>
				<div class="mt-2 flex gap-2 text-xs text-slate-400">
					<span>Type: %s</span>
					<span>Priority: %d</span>
					<span>%s → %s</span>
				</div>
			</div>`,
				t.MessageID,
				t.Time.Format("2006-01-02 15:04:05"),
				t.Content,
				t.Type,
				t.Priority,
				t.Source,
				t.Destination))
		}
		html.WriteString("</div>")
	} else {
		html.WriteString("<p class=\"text-slate-500\">No results found</p>")
	}

	return ctx.HTML(http.StatusOK, html.String())
}

func (h *Handler) Autocomplete(ctx echo.Context) error {
	sizeStr := ctx.QueryParam("size")
	size, _ := strconv.Atoi(sizeStr)
	if size <= 0 {
		size = DefaultAutocompleteSize
	}

	suggestions, err := h.container.SearchService.Autocomplete(ctx.Request().Context(), ctx.QueryParam("term"), size)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.HTML(http.StatusNotImplemented, "")
		}
		return err
	}

	var html strings.Builder
	for _, suggestion := range suggestions {
		html.WriteString(fmt.Sprintf("<div class=\"cursor-pointer hover:text-slate-200 p-1\">%s</div>", suggestion))
	}

	return ctx.HTML(http.StatusOK, html.String())
}

func (h *Handler) StatsTotal(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.HTML(http.StatusNotImplemented, "<span class=\"text-slate-500\">Not available</span>")
		}
		return err
	}
	return ctx.HTML(http.StatusOK, fmt.Sprintf("%d", summary.TotalMessages))
}

func (h *Handler) StatsPriority(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return ctx.HTML(http.StatusNotImplemented, "<p class=\"text-slate-500\">Not available</p>")
		}
		return err
	}

	// Build HTML for priority breakdown
	var html strings.Builder
	if len(stats.ByPriority) > 0 {
		for priority, count := range stats.ByPriority {
			html.WriteString(fmt.Sprintf(
				"<div class=\"flex justify-between\"><span>Priority %d</span><span class=\"font-medium\">%d</span></div>",
				priority, count))
		}
	} else {
		html.WriteString("<p class=\"text-slate-500\">No data</p>")
	}

	return ctx.HTML(http.StatusOK, html.String())
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

// Stream provides a Server-Sent Events (SSE) endpoint for real-time telegram updates.
//
// Status: Placeholder implementation
//
// Intended implementation:
//   - Connect to NATS JetStream consumer (see internal/repository/nats/consumer.go)
//   - Subscribe to telegram events stream
//   - Forward events to client via SSE format
//   - Handle client disconnections gracefully
//   - Implement heartbeat/ping messages to keep connection alive
//
// Architecture notes:
//   - NATS consumer is already implemented in cmd/sync/main.go for background processing
//   - This endpoint would provide real-time updates to web dashboard
//   - Consider using the same StreamConsumer interface used by sync worker
func (h *Handler) Stream(ctx echo.Context) error {
	ctx.Response().Header().Set("Content-Type", "text/event-stream")
	ctx.Response().Header().Set("Cache-Control", "no-cache")
	ctx.Response().Header().Set("Connection", "keep-alive")

	// Placeholder: Return not implemented message
	// TODO: Implement NATS consumer integration for real-time streaming
	fmt.Fprintf(ctx.Response().Writer, "data: <p class=\"text-slate-500\">Stream not yet implemented</p>\n\n")
	ctx.Response().Flush()

	return nil
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

func render(component templ.Component) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		return component.Render(ctx.Request().Context(), ctx.Response().Writer)
	}
}
