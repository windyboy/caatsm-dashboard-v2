package observability

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestLogger emits structured logs for every HTTP request/response pair.
func RequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()

			requestID := res.Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = req.Header.Get(echo.HeaderXRequestID)
			}

			// Get the actual status code - if there's an error, Echo may not have set it yet
			status := res.Status
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					status = httpErr.Code
				} else if status == 0 || status == http.StatusOK {
					status = http.StatusInternalServerError
				}
			}

			fields := []zap.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
			}
			if requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				// Only log as error for 5xx status codes, warn for 4xx
				if status >= 500 {
					logger.Error("request completed with error", fields...)
				} else if status >= 400 {
					logger.Warn("request completed with client error", fields...)
				} else {
					logger.Info("request completed with error", fields...)
				}
				return err
			}

			logger.Info("request completed", fields...)
			return nil
		}
	}
}

// Correlation enriches request context with correlation identifiers.
func Correlation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Header.Get(echo.HeaderXRequestID) == "" {
				req.Header.Set(echo.HeaderXRequestID, c.Response().Header().Get(echo.HeaderXRequestID))
			}
			return next(c)
		}
	}
}
