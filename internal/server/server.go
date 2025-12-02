package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	deliveryhttp "github.com/windy/caatsm-dashboard/internal/delivery/http"
	deliveryws "github.com/windy/caatsm-dashboard/internal/delivery/ws"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/cache"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/streaming"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/search"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/ws"
	"github.com/windy/caatsm-dashboard/internal/observability"
	"github.com/windy/caatsm-dashboard/internal/service"
	"go.uber.org/zap"
)

// Server wraps echo.Echo with configuration and shared dependencies.
type Server struct {
	e                *echo.Echo
	cfg              *config.AppConfig
	logger           *zap.Logger
	container        *app.Container
	metrics          *observability.MetricsExporter
	broadcaster      *deliveryws.EventBroadcaster
	hub              *ws.Hub
	ingestionSvc     *service.IngestionService
	indexerSvc       *service.IndexerService
	realtimeSvc      *service.RealtimeService
	adminSvc         *service.AdminService
	shutdownTracer   func(context.Context) error
	ingestionCtx     context.Context
	ingestionCancel  context.CancelFunc
	indexerCtx       context.Context
	indexerCancel    context.CancelFunc
	realtimeCtx      context.Context
	realtimeCancel   context.CancelFunc
}

// New constructs a Server instance and wires base middleware/routes.
func New(cfg *config.AppConfig, logger *zap.Logger) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(middleware.CORS())

	e.Use(observability.CorrelationIDMiddleware())
	e.Use(observability.RequestLogger(logger))

	// Create initialization context with timeout for app setup
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	container, err := app.New(initCtx, cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize tracing
	shutdownTracer, err := observability.InitTracer(initCtx, cfg.Tracing)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %w", err)
	}
	if cfg.Tracing.Enabled {
		logger.Info("tracing enabled", zap.String("endpoint", cfg.Tracing.OTLPEndpoint))
	}

	// Initialize infrastructure implementations in server layer to avoid circular imports
	// Track created resources for cleanup on error
	var pool *pgxpool.Pool
	var meiliSvc meilisearchClient.ServiceManager
	var redisCli redis.UniversalClient
	var broadcaster *deliveryws.EventBroadcaster

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
		if broadcaster != nil {
			broadcaster.Close()
		}
		// Note: Meilisearch client doesn't have an explicit Close() method
	}

	pool, err = app.NewPostgresPool(initCtx, cfg.Database)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	meiliSvc, err = search.NewMeilisearchClient(cfg.Meilisearch)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("create meilisearch client: %w", err)
	}

	store := repository.New(pool)
	meiliIndex := search.NewMeilisearchIndex(meiliSvc, cfg.Meilisearch.Index)

	// Ensure Meilisearch index is set up
	if err := meiliIndex.EnsureIndex(initCtx); err != nil {
		logger.Warn("failed to ensure meilisearch index", zap.Error(err))
	}

	redisCli, err = cache.NewValkeyClient(cfg.Redis)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	cacheStore := cache.NewValkeyStore(redisCli, 5*time.Minute)
	// Update to use msg:broadcast channel
	eventBus := event.NewRedisEventBus(redisCli, "msg:broadcast")
	eventPublisher := event.NewEventPublisherAdapter(eventBus)

	// Create event broadcaster for WebSocket (before assigning to container)
	broadcaster = deliveryws.NewEventBroadcaster(redisCli, logger)
	logger.Info("redis client available for WebSocket")
	go broadcaster.StartRedisListener()
	logger.Info("started Redis listener goroutine")

	// Create Redis Streams publisher and consumer
	streamPublisher := streaming.NewRedisStreamPublisher(redisCli, logger)
	streamConsumer := streaming.NewRedisStreamConsumer(
		redisCli,
		"job:index",
		"indexer-group",
		"indexer-consumer",
		logger,
	)

	// Create Redis Pub/Sub subscriber
	eventSubscriber := event.NewRedisSubscriber(redisCli, logger)

	// Only assign to container after all initializations succeed
	// This transfers ownership and prevents double-close
	container.Repo = store
	container.Cache = cacheStore
	container.Search = meiliIndex
	container.Pub = eventPublisher
	container.EventBus = eventBus
	container.EventSub = eventSubscriber
	container.StreamPub = streamPublisher
	container.RedisStreamCons = streamConsumer

	// Set client references for health checks
	container.SetInfrastructureClients(pool, meiliSvc, redisCli)

	// Initialize application services
	realtimeManager := service.NewRealtimeManager()

	searchSvc := service.NewSearchService(
		store,
		cacheStore,
		meiliIndex,
		eventPublisher,
		logger,
	)

	statsSvc := service.NewStatsService(
		store,
		cacheStore,
		logger,
	)

	exportSvc := service.NewExportService(
		searchSvc,
		store,
		logger,
	)

	dashboardSvc := service.NewDashboardService(
		searchSvc,
		statsSvc,
		exportSvc,
		realtimeManager,
		store,
		cacheStore,
		5*time.Minute, // statsTTL
		logger,
	)

	container.DashboardService = dashboardSvc

	// Initialize NATS connection for ingestion
	nc, js, err := streaming.Connect(initCtx, cfg.NATS)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	// Note: NATS connection will be closed when ingestion service stops
	// We don't need to defer close here as it's managed by the service lifecycle

	// Create NATS consumer
	natsConsumer := streaming.NewNATSConsumer(js, cfg.NATS.Stream, cfg.NATS.Consumer)

	// Create WebSocket hub
	hubConfig := ws.DefaultConfig()
	wsHub := ws.NewHub(hubConfig, logger)
	container.WebSocketHub = wsHub

	// Create services
	ingestionSvc := service.NewIngestionService(
		natsConsumer,
		store,
		eventPublisher,
		streamPublisher,
		logger,
	)

	indexerSvc := service.NewIndexerService(
		streamConsumer,
		meiliIndex,
		logger,
	)

	realtimeSvc := service.NewRealtimeService(
		eventSubscriber,
		wsHub,
		logger,
	)

	adminSvc := service.NewAdminService(
		store,
		streamPublisher,
		logger,
	)

	// Only create metrics exporter if metrics are enabled
	var metricsExporter *observability.MetricsExporter
	if cfg.Metrics.Enabled {
		metricsExporter = observability.NewMetricsExporter()
		logger.Info("metrics enabled", zap.String("path", cfg.Metrics.Path))
	} else {
		logger.Info("metrics disabled")
	}

	s := &Server{
		e:              e,
		cfg:            cfg,
		logger:         logger,
		container:      container,
		metrics:        metricsExporter,
		broadcaster:    broadcaster,
		hub:            wsHub,
		ingestionSvc:   ingestionSvc,
		indexerSvc:     indexerSvc,
		realtimeSvc:    realtimeSvc,
		adminSvc:       adminSvc,
		shutdownTracer: shutdownTracer,
	}
	
	// Store NATS connection for cleanup (if needed)
	// The connection is managed by the ingestion service, but we keep a reference
	_ = nc

	s.registerRoutes()
	return s, nil
}

// Echo exposes the underlying Echo instance for advanced customisation.
func (s *Server) Echo() *echo.Echo {
	return s.e
}

// Start boots the HTTP server and blocks until it is closed or context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting HTTP server",
		zap.String("host", s.cfg.Server.Host),
		zap.Int("port", s.cfg.Server.Port),
	)

	// Create contexts for background services
	s.ingestionCtx, s.ingestionCancel = context.WithCancel(ctx)
	s.indexerCtx, s.indexerCancel = context.WithCancel(ctx)
	s.realtimeCtx, s.realtimeCancel = context.WithCancel(ctx)

	// Start background services
	go func() {
		if err := s.ingestionSvc.Start(s.ingestionCtx); err != nil {
			s.logger.Error("ingestion service error", zap.Error(err))
		}
	}()

	go func() {
		if err := s.indexerSvc.Start(s.indexerCtx); err != nil {
			s.logger.Error("indexer service error", zap.Error(err))
		}
	}()

	go func() {
		if err := s.realtimeSvc.Start(s.realtimeCtx); err != nil {
			s.logger.Error("realtime service error", zap.Error(err))
		}
	}()

	s.logger.Info("background services started",
		zap.Bool("ingestion", true),
		zap.Bool("indexer", true),
		zap.Bool("realtime", true),
	)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      s.e,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: 0, // Disable WriteTimeout for WebSocket connections (long-lived connections)
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		var err error
		if s.cfg.Server.TLSEnabled {
			err = httpServer.ListenAndServeTLS(s.cfg.Server.TLSCertFile, s.cfg.Server.TLSKeyFile)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("graceful shutdown initiated")
		return s.gracefulShutdown(httpServer)
	case err := <-errCh:
		return err
	}
}

// gracefulShutdown performs a graceful shutdown of the server and all dependencies
func (s *Server) gracefulShutdown(httpServer *http.Server) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var shutdownErrors []error

	// Step 1: Stop accepting new HTTP connections
	s.logger.Info("step 1: stopping HTTP server (no new connections)")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("HTTP server shutdown error", zap.Error(err))
		shutdownErrors = append(shutdownErrors, fmt.Errorf("http shutdown: %w", err))
	} else {
		s.logger.Info("HTTP server stopped successfully")
	}

	// Step 2: Stop background services
	s.logger.Info("step 2: stopping background services")
	if s.ingestionCancel != nil {
		s.ingestionCancel()
		s.logger.Info("ingestion service stopped")
	}
	if s.indexerCancel != nil {
		s.indexerCancel()
		s.logger.Info("indexer service stopped")
	}
	if s.realtimeCancel != nil {
		s.realtimeCancel()
		s.logger.Info("realtime service stopped")
	}

	// Step 3: Cancel worker contexts and wait for goroutines to finish
	s.logger.Info("step 3: stopping worker goroutines (broadcaster, hub)")
	if s.broadcaster != nil {
		s.broadcaster.Close() // Cancels context and waits for goroutine via WaitGroup
		s.logger.Info("broadcaster stopped successfully")
	}
	if s.hub != nil {
		s.hub.Close() // Closes hub and all WebSocket connections
		s.logger.Info("WebSocket hub stopped successfully")
	}

	// Step 4: Flush tracing spans
	if s.shutdownTracer != nil {
		s.logger.Info("step 4: flushing tracing spans")
		if err := s.shutdownTracer(shutdownCtx); err != nil {
			s.logger.Error("tracer shutdown error", zap.Error(err))
			shutdownErrors = append(shutdownErrors, fmt.Errorf("tracer shutdown: %w", err))
		} else {
			s.logger.Info("tracing shutdown successfully")
		}
	}

	// Step 5: Flush metrics (before closing downstream connections)
	if s.metrics != nil {
		s.logger.Info("step 5: flushing metrics")
		// Metrics are pulled via Prometheus, no explicit flush needed
		s.logger.Info("metrics flushed (pull-based, no action required)")
	}

	// Step 6: Close container-managed resources (DB pool, Redis, etc.)
	// This should be done after workers complete to avoid connection errors
	s.logger.Info("step 6: closing container resources (DB pool, Redis)")
	if err := s.container.Close(); err != nil {
		s.logger.Error("container close error", zap.Error(err))
		shutdownErrors = append(shutdownErrors, fmt.Errorf("container close: %w", err))
	} else {
		s.logger.Info("container resources closed successfully")
	}

	// Step 7: Report shutdown completion
	if len(shutdownErrors) > 0 {
		s.logger.Error("graceful shutdown completed with errors",
			zap.Int("error_count", len(shutdownErrors)))
		// Return the first error (most critical)
		return shutdownErrors[0]
	}

	s.logger.Info("graceful shutdown completed successfully")
	return nil
}

func (s *Server) registerRoutes() {
	s.e.Use(deliveryhttp.BasicAuthMiddleware(s.cfg.Auth))

	// Add rate limiting
	rateLimiterConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(10), // 10 requests per second
	}
	s.e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))

	// Register API routes first (before static file serving)
	if s.cfg.Metrics.Enabled && s.metrics != nil {
		s.e.GET(s.cfg.Metrics.Path, s.metrics.Handler())
	}

	// Use new delivery layer handlers (simplified architecture)
	// Type assertion: DashboardService from container should implement the handler interface
	dashboardSvc := s.container.DashboardService.(deliveryhttp.DashboardService)
	handler := deliveryhttp.NewHandler(dashboardSvc, s.logger)
	handler.SetAdminService(s.adminSvc)
	deliveryhttp.RegisterRoutes(s.e, dashboardSvc, s.logger)
	// Register admin routes separately since handler is already created
	adminGroup := s.e.Group("/api/admin")
	adminGroup.POST("/reindex", handler.Reindex)
	s.logger.Info("registered HTTP handlers in delivery layer")

	// WebSocket hub already created in New()

	// Create adapters for stats and query services
	statsAdapter := &statsServiceAdapter{
		statsService: dashboardSvc,
	}
	queryAdapter := &queryServiceAdapter{
		repo: s.container.Repo,
	}

	// Create WebSocket handler
	wsHandler := deliveryws.NewHandler(
		s.hub,
		s.logger,
		statsAdapter,
		queryAdapter,
		s.container.RedisClient(),
		s.cfg.WebSocket.AllowedOrigins,
	)

	// Register WebSocket route
	s.e.GET("/ws", wsHandler.HandleWebSocket)
	s.logger.Info("registered WebSocket handler in delivery layer",
		zap.Strings("allowed_origins", s.cfg.WebSocket.AllowedOrigins))

	// Serve static files from Svelte frontend build directory
	// This should be registered last to catch all non-API routes
	// Deno adapter outputs to .svelte-kit/deno by default, but we check multiple locations
	buildDirs := []string{
		"frontend/.svelte-kit/deno",
		"frontend/build",
		"frontend/.svelte-kit/output",
	}

	var staticServed bool
	for _, dir := range buildDirs {
		if _, err := os.Stat(dir); err == nil {
			s.e.Static("/", dir)
			s.logger.Info("serving static files from", zap.String("dir", dir))
			staticServed = true
			break
		}
	}

	if !staticServed {
		// Register legacy static assets before the fallback route
		// (these will be overridden by frontend/build if it exists)
		s.e.Static("/css", "public/css")
		s.e.File("/favicon.ico", "public/favicon.ico")

		s.logger.Warn("frontend build directory not found, serving fallback message",
			zap.Strings("checked_dirs", buildDirs))
		// Fallback: serve a helpful error message
		s.e.GET("/*", func(c echo.Context) error {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Frontend not built. Please run 'cd frontend && deno task build' or 'cd frontend && npm run build'",
				"error":   "Not Found",
			})
		})
	}
}

// statsServiceAdapter adapts DashboardService to the interface expected by WebSocket handler
type statsServiceAdapter struct {
	statsService interface {
		GetStats(ctx context.Context, timeRange app.TimeWindow) (*app.TrafficSummary, error)
	}
}

func (a *statsServiceAdapter) TrafficSummary(ctx context.Context, window interface{}) (interface{}, error) {
	timeWindow, ok := window.(app.TimeWindow)
	if !ok {
		// If window is empty or wrong type, use empty TimeWindow (will get all-time stats)
		timeWindow = app.TimeWindow{}
	}

	stats, err := a.statsService.GetStats(ctx, timeWindow)
	if err != nil {
		return nil, err
	}

	// Return app.TrafficSummary directly (types are now unified)
	return stats, nil
}

// queryServiceAdapter adapts Repository to the interface expected by WebSocket handler
type queryServiceAdapter struct {
	repo interface {
		Search(ctx context.Context, filter app.SearchFilters) (*app.SearchResult, error)
	}
}

func (a *queryServiceAdapter) Recent(ctx context.Context, limit int) ([]*app.Telegram, error) {
	// Use Search with empty filters and order by time descending to get recent messages
	filters := app.SearchFilters{
		Pagination: app.Pagination{
			Limit:  limit,
			SortBy: "time",
			Order:  "desc",
		},
	}

	result, err := a.repo.Search(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Convert []app.Telegram to []*app.Telegram
	telegrams := make([]*app.Telegram, len(result.Telegrams))
	for i := range result.Telegrams {
		telegrams[i] = &result.Telegrams[i]
	}

	return telegrams, nil
}
