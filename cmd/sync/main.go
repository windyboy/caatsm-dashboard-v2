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
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/observability"
	natsClient "github.com/windy/caatsm-dashboard/internal/platform/nats"
	natsrepo "github.com/windy/caatsm-dashboard/internal/repository/nats"
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

	// Create app container with all dependencies
	container, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to create container", zap.Error(err))
	}
	defer container.Close()
	logger.Info("application container initialized")

	// Initialize NATS connection (needed for streaming consumer)
	nc, js, err := natsClient.Connect(ctx, cfg.NATS)
	if err != nil {
		logger.Fatal("connect nats", zap.Error(err))
	}
	defer nc.Close()

	// Initialize NATS consumer
	streamConsumer := natsrepo.New(js, cfg.NATS.Stream, cfg.NATS.Consumer)

	// Create worker with new architecture dependencies
	worker := sync.NewWorker(
		streamConsumer,
		container.TelegramStore, // repository.TelegramStore
		container.SearchIndex,   // repository.SearchIndex
		container.EventBus,      // event.EventBus
		logger,
	)
	if err := worker.Run(ctx); err != nil {
		logger.Error("worker terminated", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("sync worker stopped gracefully")
}
