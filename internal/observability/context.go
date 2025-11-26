package observability

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	requestIDKey     contextKey = "request_id"
	layerKey         contextKey = "layer"
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

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithLayer adds a layer identifier to the context for tracking layer transitions.
func WithLayer(ctx context.Context, layer string) context.Context {
	return context.WithValue(ctx, layerKey, layer)
}

// LayerFromContext extracts the layer from context.
func LayerFromContext(ctx context.Context) string {
	if layer, ok := ctx.Value(layerKey).(string); ok {
		return layer
	}
	return ""
}

// LoggerFromContext creates a logger enriched with context values.
func LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger {
	fields := []zap.Field{}

	if correlationID := CorrelationIDFromContext(ctx); correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	if requestID := RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if layer := LayerFromContext(ctx); layer != "" {
		fields = append(fields, zap.String("layer", layer))
	}

	if len(fields) > 0 {
		return base.With(fields...)
	}

	return base
}

