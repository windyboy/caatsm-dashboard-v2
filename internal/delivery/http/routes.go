package http

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterRoutes registers all HTTP routes
// If handler is nil, a new handler will be created from dashboardSvc
func RegisterRoutes(e *echo.Echo, dashboardSvc DashboardService, logger *zap.Logger, handler *Handler) {
	if handler == nil {
		handler = NewHandler(dashboardSvc, logger)
	}

	api := e.Group("/api")
	api.GET("/search", handler.Search)
	api.GET("/stats", handler.Stats)
	api.GET("/stats/historical", handler.HistoricalStats)
	api.GET("/health", handler.Health)
}

