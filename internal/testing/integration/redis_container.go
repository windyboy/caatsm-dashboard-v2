package integration

import (
	"context"
	"fmt"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer wraps a Redis test container
type RedisContainer struct {
	container testcontainers.Container
	addr      string
}

// NewRedisContainer creates and starts a Redis test container
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	redisContainer, err := rediscontainer.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis endpoint: %w", err)
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

// Terminate stops and removes the container
func (c *RedisContainer) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}

