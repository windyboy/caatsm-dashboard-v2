package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/windy/caatsm-dashboard/config"
)

// Connect creates a NATS connection and JetStream context.
func Connect(ctx context.Context, cfg config.NATSConfig) (*nats.Conn, nats.JetStreamContext, error) {
	opts := []nats.Option{
		nats.Name("caatsm-dashboard"),
		nats.Timeout(cfg.ConnectTimeout),
		nats.RetryOnFailedConnect(true),
	}

	if cfg.ConnectTimeout <= 0 {
		opts = append(opts, nats.Timeout(5*time.Second))
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, nil, fmt.Errorf("jetstream context: %w", err)
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			_ = conn.Drain()
		}()
	}

	return conn, js, nil
}
