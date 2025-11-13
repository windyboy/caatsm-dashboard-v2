package handlers

import (
	"bytes"
	"encoding/json"
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
	"github.com/windy/caatsm-dashboard/views/components"
	"github.com/windy/caatsm-dashboard/views/pages"
	"go.uber.org/zap"
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

	e.GET("/", render(pages.Dashboard()))
	e.GET("/search", render(pages.Search()))

	// HTMX action routes
	actions := e.Group("/actions")
	actions.POST("/search", h.Search)
	actions.GET("/autocomplete", h.Autocomplete)
	actions.GET("/stats/total", h.StatsTotal)
	actions.GET("/stats/priority", h.StatsPriority)
	actions.GET("/stats/type", h.StatsType)

	// API routes (non-HTMX)
	api := e.Group("/api")
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
			return render(components.ErrorMessage("Search not available"))(ctx)
		}
		return err
	}

	return render(components.SearchResults(result.Telegrams, result.Total))(ctx)
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

	return render(components.Autocomplete(suggestions))(ctx)
}

func (h *Handler) StatsTotal(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return render(components.ErrorMessage("Not available"))(ctx)
		}
		return err
	}
	return render(components.StatsTotal(summary.TotalMessages))(ctx)
}

func (h *Handler) StatsPriority(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return render(components.ErrorMessage("Not available"))(ctx)
		}
		return err
	}

	return render(components.StatsPriority(stats.ByPriority))(ctx)
}

func (h *Handler) StatsType(ctx echo.Context) error {
	// Use empty TimeWindow - service layer will apply default (last 24 hours)
	window := models.TimeWindow{}

	stats, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err != nil {
		if errors.Is(err, repository.ErrNotImplemented) {
			return render(components.ErrorMessage("Not available"))(ctx)
		}
		return err
	}

	return render(components.StatsType(stats.ByType))(ctx)
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

// Stream provides a Server-Sent Events (SSE) endpoint for real-time stats updates.
func (h *Handler) Stream(ctx echo.Context) error {
	// Add panic recovery to prevent connection issues
	defer func() {
		if r := recover(); r != nil {
			h.container.Logger.Error("panic in SSE stream handler",
				zap.Any("panic", r),
				zap.String("remote_addr", ctx.RealIP()))
		}
	}()

	ctx.Response().Header().Set("Content-Type", "text/event-stream")
	ctx.Response().Header().Set("Cache-Control", "no-cache")
	ctx.Response().Header().Set("Connection", "keep-alive")
	ctx.Response().Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Subscribe to events
	clientChan := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(clientChan)

	h.container.Logger.Info("SSE connection established", zap.String("remote_addr", ctx.RealIP()))

	writeSSE := func(eventType, data string) error {
		select {
		case <-ctx.Request().Context().Done():
			return ctx.Request().Context().Err()
		default:
		}

		if ctx.Response().Committed {
			return fmt.Errorf("response already committed, connection may be closed")
		}

		_, err := fmt.Fprintf(ctx.Response().Writer, "event: %s\ndata: %s\n\n", eventType, data)
		if err != nil {
			return fmt.Errorf("write SSE data: %w", err)
		}

		ctx.Response().Flush()
		return nil
	}

	formatSSEData := func(html string) string {
		cleaned := strings.ReplaceAll(html, "\r\n", "\n")
		cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.ReplaceAll(cleaned, "\n", " ")
		for strings.Contains(cleaned, "  ") {
			cleaned = strings.ReplaceAll(cleaned, "  ", " ")
		}
		return cleaned
	}

	// Send initial stats
	window := models.TimeWindow{}
	summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
	if err == nil {
		totalHTML := components.StatsTotal(summary.TotalMessages)
		var buf bytes.Buffer
		if err := totalHTML.Render(ctx.Request().Context(), &buf); err == nil {
			data := formatSSEData(buf.String())
			if err := writeSSE("stats-total", data); err != nil {
				h.container.Logger.Warn("failed to write initial stats-total", zap.Error(err))
				// Don't return on initial stats failure - continue with connection
				if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
					return err
				}
			}
		}

		priorityHTML := components.StatsPriority(summary.ByPriority)
		buf.Reset()
		if err := priorityHTML.Render(ctx.Request().Context(), &buf); err == nil {
			data := formatSSEData(buf.String())
			if err := writeSSE("stats-priority", data); err != nil {
				h.container.Logger.Warn("failed to write initial stats-priority", zap.Error(err))
				// Don't return on initial stats failure - continue with connection
				if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
					return err
				}
			}
		}

		// Send type breakdown update
		typeHTML := components.StatsType(summary.ByType)
		buf.Reset()
		if err := typeHTML.Render(ctx.Request().Context(), &buf); err == nil {
			data := formatSSEData(buf.String())
			if err := writeSSE("stats-type", data); err != nil {
				h.container.Logger.Warn("failed to write initial stats-type", zap.Error(err))
				// Don't return on initial stats failure - continue with connection
				if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
					return err
				}
			}
		}
	}

	// Send heartbeat every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen for events
	for {
		select {
		case <-ctx.Request().Context().Done():
			h.container.Logger.Info("SSE connection closed by client")
			return nil
		case <-ticker.C:
			// Send heartbeat (SSE comment format)
			select {
			case <-ctx.Request().Context().Done():
				return nil
			default:
				if ctx.Response().Committed {
					return nil
				}
				if _, err := fmt.Fprintf(ctx.Response().Writer, ": heartbeat\n\n"); err != nil {
					h.container.Logger.Warn("failed to write heartbeat", zap.Error(err))
					return err
				}
				ctx.Response().Flush()
			}
		case event, ok := <-clientChan:
			if !ok {
				h.container.Logger.Info("SSE client channel closed")
				return nil
			}

			h.container.Logger.Info("received event from broadcaster", zap.ByteString("event", event))

			// Parse event JSON to determine event type
			var eventData map[string]interface{}
			if err := json.Unmarshal(event, &eventData); err != nil {
				h.container.Logger.Warn("failed to parse event", zap.Error(err), zap.ByteString("raw", event))
				continue
			}

			eventType, _ := eventData["type"].(string)
			h.container.Logger.Info("parsed event type", zap.String("type", eventType))

			// If it's a telegram_processed event, send the message to live stream
			if eventType == "telegram_processed" {
				// Extract telegram from event
				telegramData, ok := eventData["telegram"].(map[string]interface{})
				if ok {
					var telegram models.Telegram
					telegramBytes, _ := json.Marshal(telegramData)
					if err := json.Unmarshal(telegramBytes, &telegram); err == nil {
						// Render message item component
						messageHTML := components.MessageItem(telegram)
						var buf bytes.Buffer
						if err := messageHTML.Render(ctx.Request().Context(), &buf); err == nil {
							data := formatSSEData(buf.String())
							// Log the data being sent for debugging
							h.container.Logger.Debug("rendered message HTML",
								zap.String("message_id", telegram.MessageID),
								zap.Int("html_length", len(data)))
							if err := writeSSE("message", data); err != nil {
								h.container.Logger.Warn("failed to write message", zap.Error(err))
								// Don't return here - continue processing other events
								// Only return if context is done or connection is truly closed
								if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
									return err
								}
								// Continue to stats update even if message write failed
							}
							h.container.Logger.Info("sent message event to client",
								zap.String("message_id", telegram.MessageID),
								zap.String("type", telegram.Type),
								zap.Int("data_length", len(data)))
						} else {
							h.container.Logger.Warn("failed to render message", zap.Error(err))
						}
					} else {
						h.container.Logger.Warn("failed to unmarshal telegram", zap.Error(err))
					}
				} else {
					h.container.Logger.Warn("telegram data not found in event")
				}
			}

			// Always update stats when we receive any event
			window := models.TimeWindow{}
			summary, err := h.container.StatsService.TrafficSummary(ctx.Request().Context(), window)
			if err == nil {
				// Send total count update
				totalHTML := components.StatsTotal(summary.TotalMessages)
				var buf bytes.Buffer
				if err := totalHTML.Render(ctx.Request().Context(), &buf); err == nil {
					data := formatSSEData(buf.String())
					if err := writeSSE("stats-total", data); err != nil {
						h.container.Logger.Warn("failed to write stats-total update", zap.Error(err))
						// Don't return here - continue processing
						if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
							return err
						}
						// Continue to try sending priority stats
					}
					h.container.Logger.Debug("sent stats-total event", zap.Int64("total", summary.TotalMessages))
				}

				// Send priority breakdown update
				priorityHTML := components.StatsPriority(summary.ByPriority)
				buf.Reset()
				if err := priorityHTML.Render(ctx.Request().Context(), &buf); err == nil {
					data := formatSSEData(buf.String())
					if err := writeSSE("stats-priority", data); err != nil {
						h.container.Logger.Warn("failed to write stats-priority update", zap.Error(err))
						// Don't return here - continue processing other events
						if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
							return err
						}
						// Continue processing
					}
					h.container.Logger.Debug("sent stats-priority event")
				}

				// Send type breakdown update
				typeHTML := components.StatsType(summary.ByType)
				buf.Reset()
				if err := typeHTML.Render(ctx.Request().Context(), &buf); err == nil {
					data := formatSSEData(buf.String())
					if err := writeSSE("stats-type", data); err != nil {
						h.container.Logger.Warn("failed to write stats-type update", zap.Error(err))
						// Don't return here - continue processing other events
						if ctx.Request().Context().Err() != nil || ctx.Response().Committed {
							return err
						}
						// Continue processing
					}
					h.container.Logger.Debug("sent stats-type event")
				}
			} else {
				h.container.Logger.Warn("failed to get traffic summary", zap.Error(err))
			}
		}
	}
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
