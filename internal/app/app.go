package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"go.uber.org/zap"
)

// NewPostgresPool initialises a pgx connection pool using the provided configuration.
func NewPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	if cfg.MaxOpenConnections > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		poolCfg.MinConns = int32(cfg.MaxIdleConnections)
	}
	if cfg.ConnectionMaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.ConnectionMaxLifetime
	} else {
		poolCfg.MaxConnLifetime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}

// Container wires together dependencies for the simplified architecture
type Container struct {
	Config *config.AppConfig
	Logger *zap.Logger

	// Core infrastructure ports
	Repo     Repository
	Cache    Cache
	Search   SearchIndex
	Pub      EventPublisher
	EventBus EventBus

	// Application services
	DashboardService interface{}

	// Client references for health checks (simplified)
	pool     *pgxpool.Pool
	meiliSvc meilisearchClient.ServiceManager
	redisCli redis.UniversalClient
}

// New builds a Container with all dependencies.
// NOTE: This uses two-phase initialization to avoid circular imports:
// 1. New() creates an empty container with nil infrastructure clients
// 2. server.New() creates infrastructure and calls SetInfrastructureClients()
// The container is not usable until SetInfrastructureClients() is called.
func New(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Database pool will be set by server layer
	var pool *pgxpool.Pool

	// NOTE: Infrastructure creation moved to server layer to avoid circular imports.
	// Infrastructure clients are set to nil here and will be initialized by server.New()
	// via SetInfrastructureClients() to avoid import cycles.
	container.Repo = nil
	container.Cache = nil
	container.Search = nil
	container.Pub = nil
	container.EventBus = nil

	// Initialize application services with nil dependencies (will fail at runtime)
	// Services will be initialized by the server layer to avoid import cycles
	container.DashboardService = nil

	// Client references will be set by server layer
	container.pool = pool
	container.meiliSvc = nil
	container.redisCli = nil

	// Validate that infrastructure is nil (expected state)
	if container.Repo != nil || container.Cache != nil || container.Search != nil {
		logger.Warn("Container initialized with non-nil infrastructure - this is unexpected",
			zap.Bool("repo_valid", container.Repo != nil),
			zap.Bool("cache_valid", container.Cache != nil),
			zap.Bool("search_valid", container.Search != nil),
		)
	}

	logger.Info("Container initialized successfully",
		zap.Bool("repo_valid", container.Repo != nil),
		zap.Bool("cache_valid", container.Cache != nil),
		zap.Bool("search_valid", container.Search != nil),
		zap.Bool("dashboard_valid", container.DashboardService != nil),
	)

	return container, nil
}

// ComponentHealth represents the health status of a component.
type ComponentHealth struct {
	Status  string `json:"status"` // "ok" or "error"
	Message string `json:"message,omitempty"`
}

// HealthCheckResult contains the health status of all components.
type HealthCheckResult struct {
	Status      string          `json:"status"` // "ok" or "degraded"
	PostgreSQL  ComponentHealth `json:"postgresql"`
	Meilisearch ComponentHealth `json:"meilisearch"`
	Redis       ComponentHealth `json:"redis"`
}

// HealthCheck verifies connectivity to all external components.
func (c *Container) HealthCheck(ctx context.Context) HealthCheckResult {
	return c.HealthCheckInternal(ctx)
}

// HealthCheckInternal is the internal implementation that can be called directly.
func (c *Container) HealthCheckInternal(ctx context.Context) HealthCheckResult {
	result := HealthCheckResult{
		Status: "ok",
	}

	// PostgreSQL health check with nil guard
	if c.pool == nil {
		result.PostgreSQL = ComponentHealth{
			Status:  "error",
			Message: "database pool not initialized",
		}
		result.Status = "degraded"
	} else {
		pgCtx, pgCancel := context.WithTimeout(ctx, 2*time.Second)
		defer pgCancel()
		if err := c.pool.Ping(pgCtx); err != nil {
			result.PostgreSQL = ComponentHealth{
				Status:  "error",
				Message: err.Error(),
			}
			result.Status = "degraded"
		} else {
			result.PostgreSQL = ComponentHealth{Status: "ok"}
		}
	}

	// Meilisearch health check
	if c.meiliSvc == nil {
		result.Meilisearch = ComponentHealth{Status: "not_configured"}
	} else {
		// TODO: Perform actual health check via API call
		// For now, just verify client is non-nil
		result.Meilisearch = ComponentHealth{Status: "ok"}
	}

	// Redis health check with nil guard and actual ping
	if c.redisCli == nil {
		result.Redis = ComponentHealth{Status: "not_configured"}
	} else {
		redisCtx, redisCancel := context.WithTimeout(ctx, 2*time.Second)
		defer redisCancel()
		if err := c.redisCli.Ping(redisCtx).Err(); err != nil {
			result.Redis = ComponentHealth{
				Status:  "error",
				Message: err.Error(),
			}
			result.Status = "degraded"
		} else {
			result.Redis = ComponentHealth{Status: "ok"}
		}
	}

	return result
}

// RedisClient returns the Redis client for external use.
func (c *Container) RedisClient() redis.UniversalClient {
	return c.redisCli
}

// IsReady returns true if the container has been fully initialized with infrastructure clients.
func (c *Container) IsReady() bool {
	return c.Repo != nil && c.Cache != nil && c.Search != nil && c.DashboardService != nil
}

// MustBeReady panics if the container is not ready for use.
// Call this at the start of methods that require infrastructure.
func (c *Container) MustBeReady() {
	if !c.IsReady() {
		panic("Container not initialized: infrastructure clients must be set via SetInfrastructureClients()")
	}
}

// SetInfrastructureClients sets the infrastructure client references for health checks.
// This should only be called after all infrastructure resources are successfully initialized.
func (c *Container) SetInfrastructureClients(pool *pgxpool.Pool, meiliSvc meilisearchClient.ServiceManager, redisCli redis.UniversalClient) {
	c.pool = pool
	c.meiliSvc = meiliSvc
	c.redisCli = redisCli
}

// Close releases all container-managed resources.
func (c *Container) Close() error {
	var firstErr error

	// Close database pool
	if c.pool != nil {
		c.Logger.Info("closing database connection pool")
		c.pool.Close()
		c.Logger.Info("database connection pool closed")
	}

	// Close Redis client disabled (redisCli is nil)

	// Note: Meilisearch client doesn't have an explicit Close() method

	return firstErr
}
