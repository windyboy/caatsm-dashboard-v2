package observability

import (
	"errors"
	"net/http"
	"strings"
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

			// Add correlation ID if available
			if correlationID := CorrelationIDFromContext(req.Context()); correlationID != "" {
				fields = append(fields, zap.String("correlation_id", correlationID))
			}

			// Add request ID if available
			if requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			if err != nil {
				// Check if the response is hijacked (WebSocket upgrade)
				// After hijacking, we shouldn't try to write errors to the response
				isHijacked := false

				// Check if ResponseWriter implements http.Hijacker
				if _, ok := res.Writer.(http.Hijacker); ok {
					// Check if this is a WebSocket upgrade request
					upgrade := req.Header.Get("Upgrade")
					connection := req.Header.Get("Connection")
					if strings.EqualFold(upgrade, "websocket") &&
						strings.Contains(strings.ToLower(connection), "upgrade") {
						isHijacked = true
					}
				}

				// Check if error is http.ErrHijacked (actual hijack occurred)
				if errors.Is(err, http.ErrHijacked) {
					isHijacked = true
				}

				if isHijacked {
					// Response is hijacked (WebSocket connection)
					// Log the error but don't return it to avoid Echo trying to write to hijacked connection
					fields = append(fields, zap.Error(err))
					logger.Warn("request completed with error (hijacked connection)", fields...)
					return nil
				}

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
