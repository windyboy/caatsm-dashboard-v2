package http

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app/services"
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
			"suggestions": []services.AutocompleteSuggestion{},
		})
	}

	size := 5 // default
	if sizeStr := c.QueryParam("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 50 {
			size = s
		}
	}

	suggestions, err := h.dashboardSvc.AutocompleteWithTypes(c.Request().Context(), query, size)
	if err != nil {
		h.logger.Error("autocomplete failed", zap.Error(err))
		return handleError(c, err)
	}

	return c.JSON(200, map[string]interface{}{
		"suggestions": suggestions,
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
