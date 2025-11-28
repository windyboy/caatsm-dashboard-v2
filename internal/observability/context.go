package observability

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
)

// WithCorrelationID adds a correlation ID to the context.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// CorrelationIDFromContext extracts the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// LoggerFromContext creates a logger enriched with context values.
func LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger {
	fields := []zap.Field{}

	if correlationID := CorrelationIDFromContext(ctx); correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	if len(fields) > 0 {
		return base.With(fields...)
	}

	return base
}
