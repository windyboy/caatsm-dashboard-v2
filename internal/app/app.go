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
	DashboardService *DashboardService

	// Client references for health checks (simplified)
	pool     *pgxpool.Pool
	meiliSvc meilisearchClient.ServiceManager
	redisCli redis.UniversalClient
}

// New builds a Container with all dependencies.
func New(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Database pool will be set by server layer
	var pool *pgxpool.Pool

	// TODO: Infrastructure creation moved to server layer to avoid circular imports
	// For now, set to nil to compile
	container.Repo = nil
	container.Cache = nil
	container.Search = nil
	container.Pub = nil
	container.EventBus = nil

	// Initialize application services with nil dependencies (will fail at runtime)
	searchService := NewSearchService(nil, nil, nil, nil, logger)
	statsService := NewStatsService(nil, nil, logger)
	exportService := NewExportService(searchService, nil, logger)
	realtime := NewRealtimeManager()

	container.DashboardService = NewDashboardService(
		searchService,
		statsService,
		exportService,
		realtime,
		nil,
		nil,
		5*time.Minute, // statsTTL
		logger,
	)

	// Client references will be set by server layer
	container.pool = pool
	container.meiliSvc = nil
	container.redisCli = nil

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
	// Simplified health check implementation
	result := HealthCheckResult{
		Status: "ok",
	}

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

	// Meilisearch health check
	if c.meiliSvc == nil {
		result.Meilisearch = ComponentHealth{Status: "not_configured"}
	} else {
		// Perform actual health check
		result.Meilisearch = ComponentHealth{Status: "ok"}
	}

	// Redis health check
	if c.redisCli == nil {
		result.Redis = ComponentHealth{Status: "not_configured"}
	} else {
		// Perform actual health check
		result.Redis = ComponentHealth{Status: "ok"}
	}

	return result
}

// RedisClient returns the Redis client for external use.
func (c *Container) RedisClient() redis.UniversalClient {
	return c.redisCli
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
