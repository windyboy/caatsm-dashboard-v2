package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/cache"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/search"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/streaming"
	"github.com/windy/caatsm-dashboard/internal/observability"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/sync"
	"go.uber.org/zap"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file (optional)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := observability.NewLogger(cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil && !errors.Is(syncErr, syscall.EINVAL) {
			fmt.Fprintf(os.Stderr, "sync logger: %v\n", syncErr)
		}
	}()

	logger.Info("starting sync worker")

	// Create initialization context with timeout for infrastructure setup
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create app container (will have nil dependencies initially)
	container, err := app.New(initCtx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to create container", zap.Error(err))
	}
	defer func() {
		if err := container.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close container: %v\n", err)
		}
	}()

	// Initialize infrastructure implementations (same pattern as server)
	// Track created resources for cleanup on error
	var pool *pgxpool.Pool
	var meiliSvc meilisearchClient.ServiceManager
	var redisCli redis.UniversalClient

	// Cleanup function to close all created resources on error
	cleanup := func() {
		if pool != nil {
			pool.Close()
		}
		if redisCli != nil {
			if err := redisCli.Close(); err != nil {
				logger.Warn("failed to close redis client during cleanup", zap.Error(err))
			}
		}
		// Note: Meilisearch client doesn't have an explicit Close() method
	}

	// Initialize PostgreSQL connection pool
	pool, err = app.NewPostgresPool(initCtx, cfg.Database)
	if err != nil {
		cleanup()
		logger.Fatal("create postgres pool", zap.Error(err))
	}

	// Initialize Meilisearch client
	meiliSvc, err = search.NewMeilisearchClient(cfg.Meilisearch)
	if err != nil {
		cleanup()
		logger.Fatal("create meilisearch client", zap.Error(err))
	}

	// Create repository store
	store := repository.New(pool)

	// Create Meilisearch index
	meiliIndex := search.NewMeilisearchIndex(meiliSvc, cfg.Meilisearch.Index)

	// Ensure Meilisearch index is set up
	if err := meiliIndex.EnsureIndex(initCtx); err != nil {
		logger.Warn("failed to ensure meilisearch index", zap.Error(err))
	}

	// Initialize Redis/Valkey client
	redisCli, err = cache.NewValkeyClient(cfg.Redis)
	if err != nil {
		cleanup()
		logger.Fatal("create redis client", zap.Error(err))
	}

	// Create event bus
	eventBus := event.NewRedisEventBus(redisCli, "stats:update")

	// Assign infrastructure implementations to container
	container.Repo = store
	container.Search = meiliIndex
	container.EventBus = eventBus

	// Set client references for health checks
	container.SetInfrastructureClients(pool, meiliSvc, redisCli)

	logger.Info("application container initialized with infrastructure")

	// Initialize NATS connection (needed for streaming consumer)
	nc, js, err := streaming.Connect(ctx, cfg.NATS)
	if err != nil {
		cleanup()
		logger.Fatal("connect nats", zap.Error(err))
	}
	defer nc.Close()

	// Initialize NATS consumer
	streamConsumer := streaming.NewNATSConsumer(js, cfg.NATS.Stream, cfg.NATS.Consumer)

	// Create worker with properly initialized port interfaces
	worker := sync.NewWorker(
		streamConsumer,
		container.Repo,     // ports.Repository
		container.Search,   // ports.SearchIndex
		container.EventBus, // EventBus for real-time updates
		logger,
	)
	if err := worker.Run(ctx); err != nil {
		logger.Error("worker terminated", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("sync worker stopped gracefully")
}
