package http

import (
	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
)

// ContainerHealthAdapter adapts app.Container health check to the transport layer interface.
type ContainerHealthAdapter struct {
	container *app.Container
}

// NewContainerHealthAdapter creates a new health check adapter.
func NewContainerHealthAdapter(container *app.Container) *ContainerHealthAdapter {
	return &ContainerHealthAdapter{container: container}
}

// HealthCheck implements the HealthChecker interface for transport layer.
func (a *ContainerHealthAdapter) HealthCheck(ctx echo.Context) HealthCheckResult {
	result := a.container.HealthCheckInternal(ctx.Request().Context())
	
	return HealthCheckResult{
		Status:      result.Status,
		PostgreSQL:  ComponentHealth{Status: result.PostgreSQL.Status, Message: result.PostgreSQL.Message},
		Meilisearch: ComponentHealth{Status: result.Meilisearch.Status, Message: result.Meilisearch.Message},
		Redis:       ComponentHealth{Status: result.Redis.Status, Message: result.Redis.Message},
	}
}

