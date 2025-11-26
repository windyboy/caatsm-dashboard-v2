package observability

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/labstack/echo/v4"
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
func generateCorrelationID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
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

