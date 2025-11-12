package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/auth"
	"github.com/windy/caatsm-dashboard/internal/handlers"
	"github.com/windy/caatsm-dashboard/internal/observability"
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

	container, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	metricsExporter := observability.NewMetricsExporter()

	s := &Server{
		e:         e,
		cfg:       cfg,
		logger:    logger,
		container: container,
		metrics:   metricsExporter,
	}

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

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      s.e,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
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

func (s *Server) registerRoutes() {
	s.e.Use(auth.Middleware(s.cfg.Auth))

	// Serve static files (CSS, favicon, etc.) - register before other routes
	s.e.Static("/css", "public/css")
	s.e.File("/favicon.ico", "public/favicon.ico")

	if s.cfg.Metrics.Enabled && s.metrics != nil {
		s.e.GET(s.cfg.Metrics.Path, s.metrics.Handler())
	}
	handlers.Register(s.e, s.container)
}
