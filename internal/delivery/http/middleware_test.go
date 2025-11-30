package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/config"
)

func TestRateLimitMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(RateLimitMiddleware())

	t.Run("allows requests within rate limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		middleware := RateLimitMiddleware()
		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rate limit configuration", func(t *testing.T) {
		middleware := RateLimitMiddleware()
		assert.NotNil(t, middleware)
	})
}

func TestProductionConfigGuardMiddleware(t *testing.T) {
	t.Run("allows non-production environment", func(t *testing.T) {
		cfg := config.AppConfig{
			Environment: "development",
			Server: config.ServerConfig{
				TLSEnabled: false,
				Host:       "localhost",
			},
		}

		middleware := ProductionConfigGuardMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "localhost:3000"
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("blocks production without TLS", func(t *testing.T) {
		cfg := config.AppConfig{
			Environment: "production",
			Server: config.ServerConfig{
				TLSEnabled: false,
				Host:       "example.com",
			},
		}

		middleware := ProductionConfigGuardMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "example.com"
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, httpErr.Code)
		assert.Contains(t, httpErr.Message.(string), "TLS must be enabled")
	})

	t.Run("blocks production on localhost", func(t *testing.T) {
		cfg := config.AppConfig{
			Environment: "production",
			Server: config.ServerConfig{
				TLSEnabled: true,
				Host:       "example.com",
			},
		}

		middleware := ProductionConfigGuardMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "localhost:3000"
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, httpErr.Code)
		assert.Contains(t, httpErr.Message.(string), "cannot run on localhost")
	})

	t.Run("allows production with valid config", func(t *testing.T) {
		cfg := config.AppConfig{
			Environment: "production",
			Server: config.ServerConfig{
				TLSEnabled: true,
				Host:       "example.com",
			},
		}

		middleware := ProductionConfigGuardMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "example.com"
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestBasicAuthMiddleware(t *testing.T) {
	t.Run("skips auth when disabled", func(t *testing.T) {
		cfg := config.AuthConfig{
			EnableBasic: false,
		}

		middleware := BasicAuthMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("requires auth when enabled", func(t *testing.T) {
		cfg := config.AuthConfig{
			EnableBasic: true,
			Username:    "admin",
			Password:    "secret",
		}

		middleware := BasicAuthMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
		assert.Contains(t, httpErr.Message.(string), "unauthorized")
	})

	t.Run("allows valid credentials", func(t *testing.T) {
		cfg := config.AuthConfig{
			EnableBasic: true,
			Username:    "admin",
			Password:    "secret",
		}

		middleware := BasicAuthMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetBasicAuth("admin", "secret")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects invalid credentials", func(t *testing.T) {
		cfg := config.AuthConfig{
			EnableBasic: true,
			Username:    "admin",
			Password:    "secret",
		}

		middleware := BasicAuthMiddleware(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetBasicAuth("admin", "wrong")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}

		wrappedHandler := middleware(handler)
		err := wrappedHandler(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})
}
