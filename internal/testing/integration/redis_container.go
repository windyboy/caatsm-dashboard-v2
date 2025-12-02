package integration

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer wraps a Valkey test container (protocol-compatible with Redis)
type RedisContainer struct {
	container testcontainers.Container
	addr      string
}

// NewRedisContainer creates and starts a Valkey test container
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	redisContainer, err := rediscontainer.Run(ctx, "valkey/valkey:9-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start valkey container: %w", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get valkey endpoint: %w", err)
	}

	return &RedisContainer{
		container: redisContainer,
		addr:      endpoint,
	}, nil
}

// Addr returns the Redis address
func (c *RedisContainer) Addr() string {
	return c.addr
}

// Client creates a redis.Client from the container
func (c *RedisContainer) Client() *redisclient.Client {
	return redisclient.NewClient(&redisclient.Options{
		Addr: c.addr,
	})
}

// Close stops and removes the container
func (c *RedisContainer) Close(ctx context.Context) error {
	if c.container != nil {
		return c.container.Terminate(ctx)
	}
	return nil
}

// Terminate is an alias for Close for backward compatibility
func (c *RedisContainer) Terminate(ctx context.Context) error {
	return c.Close(ctx)
}
