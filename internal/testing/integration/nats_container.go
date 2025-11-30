package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NATSContainer wraps a NATS testcontainer
type NATSContainer struct {
	container testcontainers.Container
	URI       string
}

// NewNATSContainer creates and starts a NATS container with JetStream enabled
func NewNATSContainer(ctx context.Context) (*NATSContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nats:2.12.2-alpine",
		ExposedPorts: []string{"4222/tcp"},
		Cmd:          []string{"-js"}, // Enable JetStream
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("4222/tcp"),
			wait.ForLog("Server is ready"),
		).WithDeadline(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start NATS container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "4222")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	uri := fmt.Sprintf("nats://%s:%s", host, port.Port())

	return &NATSContainer{
		container: container,
		URI:       uri,
	}, nil
}

// Connect creates a NATS connection to the container.
// Note: This method does not support context-based cancellation as nats.Connect
// does not accept a context parameter. Use nats.Conn.Close() to terminate the connection.
func (c *NATSContainer) Connect() (*nats.Conn, error) {
	conn, err := nats.Connect(c.URI,
		nats.Timeout(5*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return conn, nil
}

// Close terminates the container
func (c *NATSContainer) Close(ctx context.Context) error {
	if c.container != nil {
		return c.container.Terminate(ctx)
	}
	return nil
}
