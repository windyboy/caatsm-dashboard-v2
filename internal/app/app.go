package app

import (
	"context"
	"fmt"
	"time"

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
		meiliSvc,
		cfg.Meilisearch.Index,
	)

	statsService := services.NewStatsService(store, logger)
	exportService := services.NewExportService(searchService, logger)

	container.SearchService = searchService
	container.StatsService = statsService
	container.ExportService = exportService

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
