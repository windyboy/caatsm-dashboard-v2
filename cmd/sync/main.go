package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/application"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/observability"
	meiliClient "github.com/windy/caatsm-dashboard/internal/platform/meili"
	natsClient "github.com/windy/caatsm-dashboard/internal/platform/nats"
	postgresClient "github.com/windy/caatsm-dashboard/internal/platform/postgres"
	redisClient "github.com/windy/caatsm-dashboard/internal/platform/redis"
	"github.com/windy/caatsm-dashboard/internal/repository/meili"
	natsrepo "github.com/windy/caatsm-dashboard/internal/repository/nats"
	pgstore "github.com/windy/caatsm-dashboard/internal/repository/postgres"
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

	pool, err := postgresClient.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("connect postgres", zap.Error(err))
	}
	defer pool.Close()

	meiliSvc, err := meiliClient.NewClient(cfg.Meilisearch)
	if err != nil {
		logger.Fatal("connect meilisearch", zap.Error(err))
	}

	redisCli, err := redisClient.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal("connect redis", zap.Error(err))
	}
	defer redisCli.Close()

	nc, js, err := natsClient.Connect(ctx, cfg.NATS)
	if err != nil {
		logger.Fatal("connect nats", zap.Error(err))
	}
	defer nc.Close()

	// Initialize repositories
	store := pgstore.New(pool)
	index := meili.New(meiliSvc, cfg.Meilisearch.Index)

	// Initialize event bus
	eventBus := event.NewRedisEventBus(redisCli, "stats:update")

	// Initialize application service
	telegramService := application.NewTelegramService(store, index, eventBus, logger)

	// Initialize NATS consumer
	streamConsumer := natsrepo.New(js, cfg.NATS.Stream, cfg.NATS.Consumer)

	// Create worker with application service
	worker := sync.NewWorker(streamConsumer, telegramService, logger)
	if err := worker.Run(ctx); err != nil {
		logger.Error("worker terminated", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("sync worker stopped gracefully")
}
