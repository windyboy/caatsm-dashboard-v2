package app

import (
	"context"
	"fmt"
	"net/http"
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

// NewMeilisearchClient initialises a Meilisearch service manager with sensible defaults.
func NewMeilisearchClient(cfg config.SearchConfig) (meilisearchClient.ServiceManager, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("meilisearch host is empty")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	opts := []meilisearchClient.Option{
		meilisearchClient.WithCustomClient(httpClient),
	}
	if cfg.APIKey != "" {
		opts = append(opts, meilisearchClient.WithAPIKey(cfg.APIKey))
	}

	manager := meilisearchClient.New(cfg.Host, opts...)
	return manager, nil
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

	// Initialize database connection
	pool, err := NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	// Initialize Meilisearch client
	meiliSvc, err := NewMeilisearchClient(cfg.Meilisearch)
	if err != nil {
		return nil, fmt.Errorf("create meilisearch client: %w", err)
	}

	// Initialize Redis client
	redisCli, err := NewValkeyClient(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	// Initialize event bus for real-time updates
	eventBus := NewRedisEventBus(redisCli, "stats:update")

	// Initialize infrastructure implementations
	// TODO: Move infrastructure creation to server layer to avoid circular imports
	var store Repository
	var meiliIndex SearchIndex

	// Ensure Meilisearch index is set up
	if err := meiliIndex.EnsureIndex(ctx); err != nil {
		logger.Warn("failed to ensure meilisearch index", zap.Error(err))
	}

	// Initialize cache
	cacheStore := NewValkeyStore(redisCli, 5*time.Minute)

	// Create adapter to bridge EventBus to EventPublisher
	eventPublisher := NewEventPublisherAdapter(eventBus)

	// Set infrastructure ports
	container.Repo = store         // implements ports.Repository
	container.Cache = cacheStore   // implements ports.Cache
	container.Search = meiliIndex  // implements ports.SearchIndex
	container.Pub = eventPublisher // implements ports.EventPublisher
	container.EventBus = eventBus  // EventBus for real-time updates

	searchService := NewSearchService(store, cacheStore, meiliIndex, eventPublisher, logger)

	// Initialize application services
	container.DashboardService = NewDashboardService(
		searchService,
		NewStatsService(store, cacheStore, logger),
		NewExportService(
			searchService,
			store, // Pass repository for streaming
			logger,
		),
		NewRealtimeManager(),
		store,
		cacheStore,
		5*time.Minute, // statsTTL
		logger,
	)

	// Store client references for health checks
	container.pool = pool
	container.meiliSvc = meiliSvc
	container.redisCli = redisCli

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

	// Meilisearch Health() doesn't take context
	healthResp, err := c.meiliSvc.Health()
	if err != nil {
		result.Meilisearch = ComponentHealth{
			Status:  "error",
			Message: err.Error(),
		}
		result.Status = "degraded"
	} else if healthResp.Status != "available" {
		result.Meilisearch = ComponentHealth{
			Status:  "error",
			Message: fmt.Sprintf("meilisearch status: %s", healthResp.Status),
		}
		result.Status = "degraded"
	} else {
		result.Meilisearch = ComponentHealth{Status: "ok"}
	}

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

	return result
}

// RedisClient returns the Redis client for external use.
func (c *Container) RedisClient() redis.UniversalClient {
	return c.redisCli
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

	// Close Redis client
	if c.redisCli != nil {
		c.Logger.Info("closing Redis client")
		if err := c.redisCli.Close(); err != nil {
			c.Logger.Warn("Redis close error", zap.Error(err))
			// if firstErr == nil {
			firstErr = fmt.Errorf("redis close: %w", err)
			// }
		} else {
			c.Logger.Info("Redis client closed")
		}
	}

	// Note: Meilisearch client doesn't have an explicit Close() method

	return firstErr
}
