package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/application"
	"github.com/windy/caatsm-dashboard/internal/application/export"
	"github.com/windy/caatsm-dashboard/internal/application/health"
	"github.com/windy/caatsm-dashboard/internal/application/search"
	"github.com/windy/caatsm-dashboard/internal/application/stats"
	meiliClient "github.com/windy/caatsm-dashboard/internal/platform/meili"
	postgresClient "github.com/windy/caatsm-dashboard/internal/platform/postgres"
	redisClient "github.com/windy/caatsm-dashboard/internal/platform/redis"
	"github.com/windy/caatsm-dashboard/internal/repository/cache"
	meiliRepo "github.com/windy/caatsm-dashboard/internal/repository/meili"
	pgstore "github.com/windy/caatsm-dashboard/internal/repository/postgres"
	"go.uber.org/zap"
)

// Container wires together dependencies required by handlers and background workers.
type Container struct {
	Config       *config.AppConfig
	Logger       *zap.Logger
	QueryService application.QueryService

	// Application layer services (clean architecture)
	SearchServiceV2 *search.Service
	StatsServiceV2  *stats.Service
	ExportServiceV2 *export.Service
	HealthService   *health.Service
	PolicyManager   *health.PolicyManager

	// Client references for health checks
	pool     *pgxpool.Pool
	meiliSvc meilisearch.ServiceManager
	redisCli redis.UniversalClient
}

// New builds a Container from the provided options.
func New(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger, opts ...Option) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize database connection
	pool, err := postgresClient.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	// Initialize Meilisearch client
	meiliSvc, err := meiliClient.NewClient(cfg.Meilisearch)
	if err != nil {
		return nil, fmt.Errorf("create meilisearch client: %w", err)
	}

	// Initialize Redis client
	redisCli, err := redisClient.NewClient(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	// Initialize repositories
	store := pgstore.New(pool)
	meiliIndex := meiliRepo.New(meiliSvc, cfg.Meilisearch.Index)

	// Ensure Meilisearch index is set up
	if err := meiliIndex.EnsureIndex(ctx); err != nil {
		logger.Warn("failed to ensure meilisearch index", zap.Error(err))
	}

	// Initialize cache
	cacheStore := cache.New(redisCli, 5*time.Minute)

	// Initialize QueryService
	queryService := application.NewQueryService(store, meiliIndex)

	// Initialize application layer services (clean architecture)
	searchServiceV2 := search.NewService(
		meiliIndex, // repository.SearchIndex - compatible
		store,      // repository.TelegramStore - compatible (implements both TelegramStore and AnalyticsStore)
		cacheStore, // cache.Store
		logger,
	)

	statsServiceV2 := stats.NewService(
		store, // repository.AnalyticsStore - store implements this
		logger,
	)

	exportServiceV2 := export.NewService(
		searchServiceV2, // Uses the new search service
		logger,
	)

	// Set services
	container.QueryService = queryService
	container.SearchServiceV2 = searchServiceV2
	container.StatsServiceV2 = statsServiceV2
	container.ExportServiceV2 = exportServiceV2

	// Initialize health service
	healthService := health.NewService(pool, meiliSvc, redisCli, logger)
	container.HealthService = healthService

	// Initialize policy manager
	policyManager := health.NewPolicyManager(logger)
	container.PolicyManager = policyManager

	// Store client references for health checks
	container.pool = pool
	container.meiliSvc = meiliSvc
	container.redisCli = redisCli

	// Apply custom options
	for _, opt := range opts {
		if err := opt(ctx, container); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return container, nil
}

// Option represents a functional option for configuring the container.
type Option func(context.Context, *Container) error

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
	// Use new health service if available, otherwise fall back to old implementation
	if c.HealthService != nil {
		healthResult := c.HealthService.Check(ctx)
		return HealthCheckResult{
			Status:      healthResult.Status,
			PostgreSQL:  ComponentHealth{Status: string(healthResult.PostgreSQL.Status), Message: healthResult.PostgreSQL.Message},
			Meilisearch: ComponentHealth{Status: string(healthResult.Meilisearch.Status), Message: healthResult.Meilisearch.Message},
			Redis:       ComponentHealth{Status: string(healthResult.Redis.Status), Message: healthResult.Redis.Message},
		}
	}

	// Fallback to old implementation
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
