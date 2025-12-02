package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/windy/caatsm-dashboard/config"
)

// InitTracer initializes OpenTelemetry tracing.
// Returns a shutdown function that should be called on application exit.
func InitTracer(ctx context.Context, cfg config.TracingConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return no-op shutdown function if tracing is disabled
		return func(context.Context) error { return nil }, nil
	}

	// Validate configuration
	if cfg.OTLPEndpoint == "" {
		return nil, fmt.Errorf("tracing enabled but otlp_endpoint not configured")
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "caatsm-dashboard" // Default service name
	}
	if cfg.SamplingRatio <= 0 || cfg.SamplingRatio > 1 {
		cfg.SamplingRatio = 1.0 // Default to always sample in development
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"), // TODO: Use actual version from build
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Return shutdown function
	return tp.Shutdown, nil
}

// Tracer returns a tracer for the given name.
// This is a convenience function that wraps otel.Tracer.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

