package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/windy/caatsm-dashboard/config"
)

// RateLimitMiddleware returns an Echo middleware enforcing rate limiting (10 req/s).
func RateLimitMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 10},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusTooManyRequests,
				Message:  "rate limit exceeded",
				Internal: err,
			}
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusTooManyRequests,
				Message:  "rate limit exceeded",
				Internal: err,
			}
		},
	})
}

// ProductionConfigGuardMiddleware returns an Echo middleware that enforces production config guards.
func ProductionConfigGuardMiddleware(cfg config.AppConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// In production, ensure TLS is enabled
			if cfg.Environment == "production" && !cfg.Server.TLSEnabled {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "production: TLS must be enabled")
			}

			// In production, ensure secure headers are set
			if cfg.Environment == "production" {
				// Check if running on localhost (insecure in production)
				host := c.Request().Host
				if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
					return echo.NewHTTPError(http.StatusServiceUnavailable, "production: cannot run on localhost")
				}
			}

			return next(c)
		}
	}
}

// BasicAuthMiddleware returns an Echo middleware enforcing basic authentication when enabled.
func BasicAuthMiddleware(cfg config.AuthConfig) echo.MiddlewareFunc {
	if !cfg.EnableBasic {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username, password, ok := c.Request().BasicAuth()
			if !ok || username != cfg.Username || password != cfg.Password {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
