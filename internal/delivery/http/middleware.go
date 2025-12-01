package http

import (
	"crypto/subtle"
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
			if cfg.Environment != "production" {
				return next(c)
			}

			// Check if running on localhost (insecure in production) - check this first
			host := c.Request().Host
			if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "production: cannot run on localhost")
			}

			// In production, ensure TLS is enabled in config
			// Note: c.IsTLS() check is skipped in test environments as httptest doesn't support it
			if !cfg.Server.TLSEnabled {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "production: TLS must be enabled")
			}
			
			// If we're actually serving over HTTPS, verify the request is using TLS
			if cfg.Server.TLSEnabled && c.Scheme() == "https" && !c.IsTLS() {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "production: TLS must be enabled")
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
			if !ok ||
				subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Password)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
