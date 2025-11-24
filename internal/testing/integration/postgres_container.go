package integration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a PostgreSQL test container
type PostgresContainer struct {
	container testcontainers.Container
	connStr   string
}

// NewPostgresContainer creates and starts a PostgreSQL test container
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx, "timescale/timescaledb:2.16.0-pg15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PostgresContainer{
		container: pgContainer,
		connStr:   connStr,
	}, nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *PostgresContainer) ConnectionString() string {
	return c.connStr
}

// Pool creates a pgxpool.Pool from the container
func (c *PostgresContainer) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(c.connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	return pool, nil
}

// Terminate stops and removes the container
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}
