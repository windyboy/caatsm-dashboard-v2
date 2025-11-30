package observability

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	// correlationCounter is an atomically incremented counter used to ensure
	// uniqueness in fallback correlation ID generation under high concurrency.
	correlationCounter uint64
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

			// Update request with enriched context
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

// generateCorrelationID creates a unique correlation ID.
// It uses crypto/rand for secure random generation, with a time-based fallback
// if the random source is unavailable. The fallback combines timestamp, process ID,
// and an atomic counter (plus any partial random bytes) to guarantee uniqueness
// under high concurrency.
func generateCorrelationID() string {
	b := make([]byte, 8)
	n, err := rand.Read(b)
	if err != nil {
		// Fallback: combine timestamp, PID, atomic counter, and partial bytes
		// to ensure uniqueness under high concurrency
		timestamp := time.Now().UnixNano()
		pid := uint64(os.Getpid())
		counter := atomic.AddUint64(&correlationCounter, 1)
		if n > 0 {
			// Include partial random bytes if available
			return fmt.Sprintf("%016x-%08x-%016x-%s", timestamp, pid, counter, hex.EncodeToString(b[:n]))
		}
		// Deterministic concatenation: timestamp + pid + counter
		return fmt.Sprintf("%016x-%08x-%016x", timestamp, pid, counter)
	}
	return hex.EncodeToString(b)
}
