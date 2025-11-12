package observability

import (
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

			fields := []zap.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", res.Status),
				zap.Duration("latency", latency),
			}
			if requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("request completed with error", fields...)
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
