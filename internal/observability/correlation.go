package observability

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// CorrelationIDMiddleware enhances requests with correlation IDs that flow through all layers.
func CorrelationIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			
			// Get or generate correlation ID
			correlationID := req.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = generateCorrelationID()
			}

			// Set in response headers
			c.Response().Header().Set("X-Correlation-ID", correlationID)

			// Add to context for use throughout the request lifecycle
			ctx := WithCorrelationID(req.Context(), correlationID)
			
			// Also use request ID as correlation ID if available
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID != "" {
				ctx = WithRequestID(ctx, requestID)
			}

			// Update request with enriched context
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

// generateCorrelationID creates a unique correlation ID.
// It uses crypto/rand for secure random generation, with a time-based fallback
// if the random source is unavailable.
func generateCorrelationID() string {
	b := make([]byte, 8)
	n, err := rand.Read(b)
	if err != nil {
		// Log the error for observability
		logger, _ := zap.NewProduction()
		if logger != nil {
			logger.Error("crypto/rand failed, using time-based fallback for correlation ID",
				zap.Error(err),
				zap.Int("bytes_read", n))
			defer func() { _ = logger.Sync() }()
		}

		// Fallback: use time-based ID to ensure uniqueness
		// Combine timestamp with any partial bytes read
		timestamp := time.Now().UnixNano()
		if n > 0 {
			// Mix partial random bytes with timestamp for better uniqueness
			return fmt.Sprintf("%016x-%s", timestamp, hex.EncodeToString(b[:n]))
		}
		// Pure time-based fallback if no bytes were read
		return fmt.Sprintf("%016x", timestamp)
	}
	return hex.EncodeToString(b)
}

// GetCorrelationID extracts correlation ID from Echo context.
func GetCorrelationID(c echo.Context) string {
	if correlationID := c.Request().Header.Get("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}
	if correlationID := c.Response().Header().Get("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}
	return CorrelationIDFromContext(c.Request().Context())
}

