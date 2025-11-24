package integration

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MeiliContainer wraps a Meilisearch test container
type MeiliContainer struct {
	container testcontainers.Container
	host      string
	port      string
}

// NewMeiliContainer creates and starts a Meilisearch test container
func NewMeiliContainer(ctx context.Context) (*MeiliContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "getmeili/meilisearch:v1.7",
		ExposedPorts: []string{"7700/tcp"},
		Env: map[string]string{
			"MEILI_MASTER_KEY": "testmasterkey",
		},
		WaitingFor: wait.ForHTTP("/health").WithPort("7700/tcp").WithStartupTimeout(30),
	}

	meiliContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start meilisearch container: %w", err)
	}

	host, err := meiliContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get meilisearch host: %w", err)
	}

	port, err := meiliContainer.MappedPort(ctx, "7700")
	if err != nil {
		return nil, fmt.Errorf("failed to get meilisearch port: %w", err)
	}

	return &MeiliContainer{
		container: meiliContainer,
		host:      host,
		port:      port.Port(),
	}, nil
}

// Host returns the Meilisearch host
func (c *MeiliContainer) Host() string {
	return c.host
}

// Port returns the Meilisearch port
func (c *MeiliContainer) Port() string {
	return c.port
}

// URL returns the full Meilisearch URL
func (c *MeiliContainer) URL() string {
	return fmt.Sprintf("http://%s:%s", c.host, c.port)
}

// MasterKey returns the master key for the test container
func (c *MeiliContainer) MasterKey() string {
	return "testmasterkey"
}

// Terminate stops and removes the container
func (c *MeiliContainer) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}
