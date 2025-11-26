package tracing

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerName = "github.com/windy/caatsm-dashboard"
)

// Middleware returns an Echo middleware that traces HTTP requests
func Middleware(enabled bool) echo.MiddlewareFunc {
	if !enabled {
		// Return no-op middleware if tracing is disabled
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	tracer := otel.Tracer(TracerName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			// Extract trace context from incoming request
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

		// Start span
		spanName := req.Method + " " + c.Path()
		
		// Derive scheme from TLS state
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(req.Method),
				semconv.HTTPRoute(c.Path()),
				semconv.HTTPScheme(scheme),
				semconv.HTTPTarget(req.URL.Path),
				semconv.HTTPURL(scheme+"://"+req.Host+req.RequestURI),
				semconv.HTTPUserAgent(req.UserAgent()),
				semconv.NetHostName(req.Host),
			),
		)
			defer span.End()

			// Update request context
			c.SetRequest(req.WithContext(ctx))

			// Inject trace context into response headers
			propagator.Inject(ctx, propagation.HeaderCarrier(c.Response().Header()))

			// Call next handler
			err := next(c)

			// Set span status based on response
			status := c.Response().Status
			span.SetAttributes(semconv.HTTPStatusCode(status))

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else if status >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		}
	}
}

// StartSpan is a helper to start a new span in application code
func StartSpan(ctx echo.Context, name string, opts ...trace.SpanStartOption) (echo.Context, trace.Span) {
	tracer := otel.Tracer(TracerName)
	reqCtx := ctx.Request().Context()
	spanCtx, span := tracer.Start(reqCtx, name, opts...)
	ctx.SetRequest(ctx.Request().WithContext(spanCtx))
	return ctx, span
}

// AddAttribute adds an attribute to the current span
func AddAttribute(ctx echo.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx.Request().Context())
	
	var attr attribute.KeyValue
	switch v := value.(type) {
	case string:
		attr = attribute.String(key, v)
	case int:
		attr = attribute.Int(key, v)
	case int64:
		attr = attribute.Int64(key, v)
	case bool:
		attr = attribute.Bool(key, v)
	default:
		attr = attribute.String(key, fmt.Sprint(v))
	}
	
	span.SetAttributes(attr)
}

