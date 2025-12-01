package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/windy/caatsm-dashboard/config"
	"go.uber.org/zap"
)

func TestContainer_HealthCheck_NilClients(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AppConfig{}
	logger := zap.NewNop()

	container, err := New(ctx, cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	// Health check should not panic with nil clients
	result := container.HealthCheckInternal(ctx)

	// Should return degraded status with error messages for nil clients
	assert.Equal(t, "degraded", result.Status)
	assert.Equal(t, "error", result.PostgreSQL.Status)
	assert.Contains(t, result.PostgreSQL.Message, "not initialized")
	assert.Equal(t, "not_configured", result.Meilisearch.Status)
	assert.Equal(t, "not_configured", result.Redis.Status)
}

func TestContainer_IsReady(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AppConfig{}
	logger := zap.NewNop()

	container, err := New(ctx, cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	// Container should not be ready after New() (infrastructure not set)
	assert.False(t, container.IsReady())

	// After setting infrastructure clients, IsReady() should still return false
	// because Repo, Cache, Search, and DashboardService are still nil
	container.SetInfrastructureClients(nil, nil, nil)
	assert.False(t, container.IsReady())
}

func TestContainer_MustBeReady(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AppConfig{}
	logger := zap.NewNop()

	container, err := New(ctx, cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	// MustBeReady should panic if container is not ready
	assert.Panics(t, func() {
		container.MustBeReady()
	})
}

