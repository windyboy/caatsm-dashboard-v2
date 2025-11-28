package http

import (
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(e *echo.Echo, dashboardSvc DashboardService, logger *zap.Logger) {
	handler := NewHandler(dashboardSvc, logger)

	// API v1 routes
	api := e.Group("/api")

	// Dashboard routes
	api.GET("/dashboard", handler.Dashboard)

	// Search routes
	api.GET("/search", handler.Search)
	api.POST("/search", handler.Search)

	// Stats routes
	api.GET("/stats", handler.Stats)

	// Export routes
	api.GET("/export", handler.Export)

	// Health check
	api.GET("/health", handler.Health)

	// Autocomplete (for frontend)
	api.GET("/autocomplete", handler.Autocomplete)

	// Legacy routes for backward compatibility
	// These can be removed once frontend is updated
	api.GET("/stats/total", handler.StatsTotal)
	api.GET("/stats/priority", handler.StatsPriority)
	api.GET("/stats/type", handler.StatsType)
}

// Autocomplete handles GET /api/autocomplete - search suggestions
func (h *Handler) Autocomplete(c echo.Context) error {
	query := c.QueryParam("term")
	if query == "" {
		return c.JSON(200, map[string]interface{}{
			"suggestions": []string{},
		})
	}

	size := 5 // default
	if sizeStr := c.QueryParam("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 50 {
			size = s
		}
	}

	// Check if client wants structured suggestions with labels
	withTypes := c.QueryParam("with_types") == "true"

	if withTypes {
		// Try to get typed suggestions if the service supports it
		// For now, we'll always return structured format for better UX
		// The frontend will handle both formats
	}

	suggestions, err := h.dashboardSvc.Autocomplete(c.Request().Context(), query, size)
	if err != nil {
		h.logger.Error("autocomplete failed", zap.Error(err))
		return handleError(c, err)
	}

	// Format suggestions with labels for better readability
	formattedSuggestions := make([]map[string]interface{}, 0, len(suggestions))
	for _, sug := range suggestions {
		// Determine type based on value patterns
		sugType := "text"
		label := ""

		// Flight numbers: typically 2-3 letters + 3-4 digits (e.g., CA320, AF413)
		if matched, _ := regexp.MatchString(`^[A-Z]{2,3}\d{3,4}$`, sug); matched {
			sugType = "flight_number"
			label = "Flight"
		} else if matched, _ := regexp.MatchString(`^LIVE-`, sug); matched {
			// Message IDs: start with LIVE-
			sugType = "message_id"
			label = "Message ID"
		} else if matched, _ := regexp.MatchString(`^[A-Z]{4}$`, sug); matched {
			// Airport codes: 4 letters (ICAO)
			sugType = "airport"
			label = "Airport"
		} else if len(sug) >= 3 && len(sug) <= 4 {
			// Short codes might be airports
			sugType = "airport"
			label = "Airport"
		}

		formattedSuggestions = append(formattedSuggestions, map[string]interface{}{
			"value": sug,
			"type":  sugType,
			"label": label,
		})
	}

	return c.JSON(200, map[string]interface{}{
		"suggestions": formattedSuggestions,
	})
}

// Legacy handlers for backward compatibility
func (h *Handler) StatsTotal(c echo.Context) error {
	timeRange := buildTimeWindow(c)
	stats, err := h.dashboardSvc.GetStats(c.Request().Context(), timeRange)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(200, map[string]interface{}{
		"total": stats.TotalMessages,
	})
}

func (h *Handler) StatsPriority(c echo.Context) error {
	timeRange := buildTimeWindow(c)
	stats, err := h.dashboardSvc.GetStats(c.Request().Context(), timeRange)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(200, map[string]interface{}{
		"byPriority": stats.ByPriority,
	})
}

func (h *Handler) StatsType(c echo.Context) error {
	timeRange := buildTimeWindow(c)
	stats, err := h.dashboardSvc.GetStats(c.Request().Context(), timeRange)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(200, map[string]interface{}{
		"byType": stats.ByType,
	})
}
