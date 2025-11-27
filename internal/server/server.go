package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/auth"
	deliveryhttp "github.com/windy/caatsm-dashboard/internal/delivery/http"
	deliveryws "github.com/windy/caatsm-dashboard/internal/delivery/ws"
	"github.com/windy/caatsm-dashboard/internal/observability"
	"go.uber.org/zap"
)

// Server wraps echo.Echo with configuration and shared dependencies.
type Server struct {
	e           *echo.Echo
	cfg         *config.AppConfig
	logger      *zap.Logger
	container   *app.Container
	metrics     *observability.MetricsExporter
	broadcaster *deliveryws.EventBroadcaster
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

	e.Use(observability.Correlation())
	e.Use(observability.RequestLogger(logger))

	// Create initialization context with timeout for app setup
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	container, err := app.New(initCtx, cfg, logger)
	if err != nil {
		return nil, err
	}

	// Only create metrics exporter if metrics are enabled
	var metricsExporter *observability.MetricsExporter
	if cfg.Metrics.Enabled {
		metricsExporter = observability.NewMetricsExporter()
		logger.Info("metrics enabled", zap.String("path", cfg.Metrics.Path))
	} else {
		logger.Info("metrics disabled")
	}

	// Create event broadcaster for WebSocket
	redisCli := container.RedisClient()
	if redisCli == nil {
		logger.Warn("redis client is nil, WebSocket real-time updates will not work")
	} else {
		logger.Info("redis client available for WebSocket")
	}
	broadcaster := deliveryws.NewEventBroadcaster(redisCli, logger)
	go broadcaster.StartRedisListener()
	logger.Info("started Redis listener goroutine")

	s := &Server{
		e:           e,
		cfg:         cfg,
		logger:      logger,
		container:   container,
		metrics:     metricsExporter,
		broadcaster: broadcaster,
	}

	s.registerRoutes(broadcaster)
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

	// Step 2: Cancel worker contexts and wait for goroutines to finish
	s.logger.Info("step 2: stopping worker goroutines (broadcaster)")
	if s.broadcaster != nil {
		s.broadcaster.Close() // Cancels context and waits for goroutine via WaitGroup
		s.logger.Info("broadcaster stopped successfully")
	}

	// Step 3: Flush metrics (before closing downstream connections)
	if s.metrics != nil {
		s.logger.Info("step 3: flushing metrics")
		// Metrics are pulled via Prometheus, no explicit flush needed
		s.logger.Info("metrics flushed (pull-based, no action required)")
	}

	// Step 4: Close container-managed resources (DB pool, Redis, etc.)
	// This should be done after workers complete to avoid connection errors
	s.logger.Info("step 4: closing container resources (DB pool, Redis)")
	if err := s.container.Close(); err != nil {
		s.logger.Error("container close error", zap.Error(err))
		shutdownErrors = append(shutdownErrors, fmt.Errorf("container close: %w", err))
	} else {
		s.logger.Info("container resources closed successfully")
	}

	// Step 5: Report shutdown completion
	if len(shutdownErrors) > 0 {
		s.logger.Error("graceful shutdown completed with errors",
			zap.Int("error_count", len(shutdownErrors)))
		// Return the first error (most critical)
		return shutdownErrors[0]
	}

	s.logger.Info("graceful shutdown completed successfully")
	return nil
}

func (s *Server) registerRoutes(broadcaster *deliveryws.EventBroadcaster) {
	s.e.Use(auth.Middleware(s.cfg.Auth))

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
	deliveryhttp.RegisterRoutes(s.e, s.container.DashboardService, s.logger)
	s.logger.Info("registered HTTP handlers in delivery layer")

	// Register WebSocket handler using transport layer with configured allowed origins
	wsHandler := deliveryws.NewSimpleHandler(s.container, broadcaster, s.cfg.WebSocket.AllowedOrigins)
	wsHandler.Register(s.e)
	s.logger.Info("registered WebSocket handler in transport layer",
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
