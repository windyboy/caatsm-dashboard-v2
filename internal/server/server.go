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
	"github.com/windy/caatsm-dashboard/internal/handlers"
	"github.com/windy/caatsm-dashboard/internal/observability"
	transporthttp "github.com/windy/caatsm-dashboard/internal/transport/http"
	"go.uber.org/zap"
)

// Server wraps echo.Echo with configuration and shared dependencies.
type Server struct {
	e         *echo.Echo
	cfg       *config.AppConfig
	logger    *zap.Logger
	container *app.Container
	metrics   *observability.MetricsExporter
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
	broadcaster := handlers.NewEventBroadcaster(redisCli, logger)
	go broadcaster.StartRedisListener()
	logger.Info("started Redis listener goroutine")

	s := &Server{
		e:         e,
		cfg:       cfg,
		logger:    logger,
		container: container,
		metrics:   metricsExporter,
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes(broadcaster *handlers.EventBroadcaster) {
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
	
	// Use new transport layer handlers (clean architecture)
	transportHandlers := s.createTransportHandlers()
	if transportHandlers != nil {
		transporthttp.Register(s.e, transportHandlers)
		s.logger.Info("using new transport layer handlers")
		
		// Register WebSocket handler separately (using legacy handler)
		// TODO: Migrate WebSocket to transport layer in future
		wsH := &handlers.Handler{}
		handlers.RegisterWebSocketOnly(s.e, wsH, s.container, broadcaster)
	} else {
		// Fallback to legacy handlers if new handlers fail to initialize
		s.logger.Warn("falling back to legacy handlers")
		handlers.Register(s.e, s.container, broadcaster)
	}

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

// createTransportHandlers creates the new transport layer handlers.
func (s *Server) createTransportHandlers() *transporthttp.Handler {
	// Ensure we have V2 services available
	if s.container.SearchServiceV2 == nil || s.container.StatsServiceV2 == nil || s.container.ExportServiceV2 == nil {
		s.logger.Warn("V2 services not available, cannot create transport handlers")
		return nil
	}

	// Create health check adapter
	healthAdapter := transporthttp.NewContainerHealthAdapter(s.container)

	// Create new transport handlers
	return transporthttp.NewHandler(
		s.container.SearchServiceV2,
		s.container.StatsServiceV2,
		s.container.ExportServiceV2,
		healthAdapter,
	)
}
