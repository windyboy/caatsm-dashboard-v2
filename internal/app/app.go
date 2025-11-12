package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	meiliClient "github.com/windy/caatsm-dashboard/internal/platform/meili"
	postgresClient "github.com/windy/caatsm-dashboard/internal/platform/postgres"
	redisClient "github.com/windy/caatsm-dashboard/internal/platform/redis"
	"github.com/windy/caatsm-dashboard/internal/repository/cache"
	meiliRepo "github.com/windy/caatsm-dashboard/internal/repository/meili"
	pgstore "github.com/windy/caatsm-dashboard/internal/repository/postgres"
	"github.com/windy/caatsm-dashboard/internal/services"
	"go.uber.org/zap"
)

// Container wires together dependencies required by handlers and background workers.
type Container struct {
	Config        *config.AppConfig
	Logger        *zap.Logger
	SearchService services.SearchService
	StatsService  services.StatsService
	ExportService services.ExportService

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

	// Initialize services
	searchService := services.NewSearchService(
		meiliIndex,
		store,
		cacheStore,
		logger,
	)

	statsService := services.NewStatsService(store, logger)
	exportService := services.NewExportService(searchService, logger)

	container.SearchService = searchService
	container.StatsService = statsService
	container.ExportService = exportService

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
	result := HealthCheckResult{
		Status: "ok",
	}

	// Check PostgreSQL
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.pool.Ping(healthCtx); err != nil {
		result.PostgreSQL = ComponentHealth{
			Status:  "error",
			Message: err.Error(),
		}
		result.Status = "degraded"
	} else {
		result.PostgreSQL = ComponentHealth{Status: "ok"}
	}

	// Check Meilisearch
	healthCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	// Meilisearch Health() doesn't take context, but we use timeout context for cancellation
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

	// Check Redis
	healthCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.redisCli.Ping(healthCtx).Err(); err != nil {
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
