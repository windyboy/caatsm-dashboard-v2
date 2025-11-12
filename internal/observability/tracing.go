package observability

import (
	"context"
)

// SetupTracing initialises OpenTelemetry exporters and returns a shutdown function.
// Currently a placeholder until full tracing pipeline is implemented.
func SetupTracing(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	return func(ctx context.Context) error {
		return nil
	}, nil
}
