package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/config"
)

// Middleware returns an Echo middleware enforcing basic authentication when enabled.
func Middleware(cfg config.AuthConfig) echo.MiddlewareFunc {
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

// ErrUnauthorized is returned when authentication fails.
var ErrUnauthorized = errors.New("unauthorized")
