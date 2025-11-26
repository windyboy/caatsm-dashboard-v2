package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
)

// RegisterWebSocketOnly registers only the WebSocket route.
// This allows using new transport handlers for HTTP endpoints while keeping legacy WebSocket handler.
func RegisterWebSocketOnly(e *echo.Echo, h *Handler, container *app.Container, broadcaster *EventBroadcaster) {
	h.container = container
	h.broadcaster = broadcaster
	e.GET("/ws", h.WebSocket)
}

