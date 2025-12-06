package http

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/windy/caatsm-dashboard/config"
)

// JWTMiddleware returns an Echo middleware for JWT authentication
func JWTMiddleware(cfg config.AuthConfig) echo.MiddlewareFunc {
	if cfg.JWTSecret == "" {
		// If no JWT secret configured, skip JWT validation
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(auth, "Bearer ")
			if tokenString == auth {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "token is not valid")
			}

			// Store claims in context for later use
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user", claims)
			}

			return next(c)
		}
	}
}

// CSRFMiddleware returns an Echo middleware for CSRF protection
func CSRFMiddleware() echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		ContextKey:     "csrf",
		CookieName:     "csrf",
		CookiePath:     "/",
		CookieHTTPOnly: false, // Allow JavaScript access for SPA
		CookieSameSite: http.SameSiteStrictMode,
	})
}

// BasicAuthMiddleware returns an Echo middleware for Basic authentication
func BasicAuthMiddleware(cfg config.AuthConfig) echo.MiddlewareFunc {
	if !cfg.EnableBasic {
		// If basic auth is disabled, skip authentication
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	// Use Echo's built-in BasicAuth middleware
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == cfg.Username && password == cfg.Password {
			return true, nil
		}
		return false, nil
	})
}

// RateLimitMiddleware returns an Echo middleware for rate limiting
func RateLimitMiddleware() echo.MiddlewareFunc {
	// Use Echo's built-in rate limiter with memory store
	// Default: 10 requests per second
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10))
}

// ProductionConfigGuardMiddleware returns an Echo middleware that enforces production security requirements
func ProductionConfigGuardMiddleware(cfg config.AppConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only enforce in production environment
			if cfg.Environment != "production" {
				return next(c)
			}

			// Check TLS requirement
			if !cfg.Server.TLSEnabled {
				return echo.NewHTTPError(
					http.StatusServiceUnavailable,
					"TLS must be enabled in production environment",
				)
			}

			// Check that we're not running on localhost in production
			host := c.Request().Host
			if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
				return echo.NewHTTPError(
					http.StatusServiceUnavailable,
					"production server cannot run on localhost",
				)
			}

			return next(c)
		}
	}
}
