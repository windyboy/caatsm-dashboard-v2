package testing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/windy/caatsm-dashboard/internal/testing/integration"
)

// TestEnv provides a complete test environment with all containers
type TestEnv struct {
	Postgres *integration.PostgresContainer
	Redis    *integration.RedisContainer
	Meili    *integration.MeiliContainer
	Pool     *pgxpool.Pool
}

// SetupTestEnv creates and starts all test containers
func SetupTestEnv(ctx context.Context) (*TestEnv, error) {
	pgContainer, err := integration.NewPostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup postgres: %w", err)
	}

	redisContainer, err := integration.NewRedisContainer(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to setup redis: %w", err)
	}

	meiliContainer, err := integration.NewMeiliContainer(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to setup meilisearch: %w", err)
	}

	pool, err := pgContainer.Pool(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		meiliContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Run migrations
	if err := RunMigrations(ctx, pool); err != nil {
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		meiliContainer.Terminate(ctx)
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestEnv{
		Postgres: pgContainer,
		Redis:    redisContainer,
		Meili:    meiliContainer,
		Pool:     pool,
	}, nil
}

// Cleanup stops all containers and closes connections
func (e *TestEnv) Cleanup(ctx context.Context) error {
	var errs []error

	if e.Pool != nil {
		e.Pool.Close()
	}

	if e.Postgres != nil {
		if err := e.Postgres.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate postgres: %w", err))
		}
	}

	if e.Redis != nil {
		if err := e.Redis.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate redis: %w", err))
		}
	}

	if e.Meili != nil {
		if err := e.Meili.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate meilisearch: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// RunMigrations runs database migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create telegrams table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS telegrams (
		message_id VARCHAR(255) PRIMARY KEY,
		type VARCHAR(50) NOT NULL,
		time TIMESTAMPTZ NOT NULL,
		flight_number VARCHAR(20),
		source VARCHAR(10),
		destination VARCHAR(10),
		content TEXT,
		priority INTEGER NOT NULL DEFAULT 3,
		raw_data TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_telegrams_time ON telegrams(time DESC);
	CREATE INDEX IF NOT EXISTS idx_telegrams_type ON telegrams(type);
	CREATE INDEX IF NOT EXISTS idx_telegrams_priority ON telegrams(priority);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

// CleanDatabase removes all data from test tables
func CleanDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE telegrams CASCADE")
	return err
}
