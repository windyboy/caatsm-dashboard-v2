package health

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ComponentStatus represents the status of a component.
type ComponentStatus string

const (
	StatusOK       ComponentStatus = "ok"
	StatusError    ComponentStatus = "error"
	StatusDegraded ComponentStatus = "degraded"
	StatusUnknown  ComponentStatus = "unknown"
)

// ComponentHealth represents the health status of a component.
type ComponentHealth struct {
	Status      ComponentStatus `json:"status"`
	Message     string          `json:"message,omitempty"`
	LastChecked time.Time       `json:"last_checked,omitempty"`
	Latency     time.Duration   `json:"latency_ms,omitempty"`
}

// HealthResult contains the health status of all components and overall system.
type HealthResult struct {
	Status      string                      `json:"status"` // "ok", "degraded", or "error"
	Timestamp   time.Time                   `json:"timestamp"`
	PostgreSQL  ComponentHealth            `json:"postgresql"`
	Meilisearch ComponentHealth            `json:"meilisearch"`
	Redis       ComponentHealth            `json:"redis"`
	Degraded    []string                   `json:"degraded,omitempty"` // List of degraded component names
}

// Service provides health checking capabilities.
type Service struct {
	pool     *pgxpool.Pool
	meiliSvc meilisearch.ServiceManager
	redisCli redis.UniversalClient
	logger   *zap.Logger
	timeout  time.Duration
}

// NewService creates a new health service.
func NewService(
	pool *pgxpool.Pool,
	meiliSvc meilisearch.ServiceManager,
	redisCli redis.UniversalClient,
	logger *zap.Logger,
) *Service {
	return &Service{
		pool:     pool,
		meiliSvc: meiliSvc,
		redisCli: redisCli,
		logger:   logger,
		timeout:  2 * time.Second,
	}
}

// Check performs a comprehensive health check of all components.
func (s *Service) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Status:    string(StatusOK),
		Timestamp: time.Now(),
		Degraded:  []string{},
	}

	// Check PostgreSQL
	pgHealth := s.checkPostgreSQL(ctx)
	result.PostgreSQL = pgHealth
	if pgHealth.Status == StatusError {
		result.Status = string(StatusDegraded)
		result.Degraded = append(result.Degraded, "postgresql")
	}

	// Check Meilisearch
	meiliHealth := s.checkMeilisearch(ctx)
	result.Meilisearch = meiliHealth
	if meiliHealth.Status == StatusError {
		if result.Status == string(StatusOK) {
			result.Status = string(StatusDegraded)
		}
		result.Degraded = append(result.Degraded, "meilisearch")
	}

	// Check Redis
	redisHealth := s.checkRedis(ctx)
	result.Redis = redisHealth
	if redisHealth.Status == StatusError {
		if result.Status == string(StatusOK) {
			result.Status = string(StatusDegraded)
		}
		result.Degraded = append(result.Degraded, "redis")
	}

	// If all components are down, system is in error state
	if len(result.Degraded) >= 3 {
		result.Status = "error"
	}

	return result
}

// checkPostgreSQL checks PostgreSQL/TimescaleDB health.
func (s *Service) checkPostgreSQL(ctx context.Context) ComponentHealth {
	start := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := s.pool.Ping(checkCtx)
	latency := time.Since(start)

	health := ComponentHealth{
		LastChecked: time.Now(),
		Latency:     latency,
	}

	if err != nil {
		health.Status = StatusError
		health.Message = err.Error()
		s.logger.Warn("postgresql health check failed", zap.Error(err), zap.Duration("latency", latency))
		return health
	}

	// Check latency - if too slow, mark as degraded
	if latency > 1*time.Second {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("high latency: %v", latency)
		s.logger.Warn("postgresql health check slow", zap.Duration("latency", latency))
		return health
	}

	health.Status = StatusOK
	return health
}

// checkMeilisearch checks Meilisearch health.
func (s *Service) checkMeilisearch(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastChecked: time.Now(),
	}

	// Meilisearch Health() doesn't accept context, but we can timeout with a channel
	healthCh := make(chan error, 1)
	go func() {
		healthResp, err := s.meiliSvc.Health()
		if err != nil {
			healthCh <- err
			return
		}
		if healthResp.Status != "available" {
			healthCh <- fmt.Errorf("meilisearch status: %s", healthResp.Status)
			return
		}
		healthCh <- nil
	}()

	select {
	case err := <-healthCh:
		health.Latency = time.Since(start)
		if err != nil {
			health.Status = StatusError
			health.Message = err.Error()
			s.logger.Warn("meilisearch health check failed", zap.Error(err), zap.Duration("latency", health.Latency))
			return health
		}

		// Check latency
		if health.Latency > 1*time.Second {
			health.Status = StatusDegraded
			health.Message = fmt.Sprintf("high latency: %v", health.Latency)
			s.logger.Warn("meilisearch health check slow", zap.Duration("latency", health.Latency))
			return health
		}

		health.Status = StatusOK
		return health
	case <-ctx.Done():
		health.Latency = time.Since(start)
		health.Status = StatusError
		health.Message = "health check timeout"
		s.logger.Warn("meilisearch health check timeout", zap.Duration("latency", health.Latency))
		return health
	case <-time.After(s.timeout):
		health.Latency = time.Since(start)
		health.Status = StatusError
		health.Message = "health check timeout"
		s.logger.Warn("meilisearch health check timeout", zap.Duration("latency", health.Latency))
		return health
	}
}

// checkRedis checks Redis/Valkey health.
func (s *Service) checkRedis(ctx context.Context) ComponentHealth {
	start := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := s.redisCli.Ping(checkCtx).Err()
	latency := time.Since(start)

	health := ComponentHealth{
		LastChecked: time.Now(),
		Latency:     latency,
	}

	if err != nil {
		health.Status = StatusError
		health.Message = err.Error()
		s.logger.Warn("redis health check failed", zap.Error(err), zap.Duration("latency", latency))
		return health
	}

	// Check latency
	if latency > 500*time.Millisecond {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("high latency: %v", latency)
		s.logger.Warn("redis health check slow", zap.Duration("latency", latency))
		return health
	}

	health.Status = StatusOK
	return health
}

// IsHealthy returns true if the system is in a healthy state.
func (s *Service) IsHealthy(ctx context.Context) bool {
	result := s.Check(ctx)
	return result.Status == string(StatusOK)
}

// IsDegraded returns true if the system is degraded but still operational.
func (s *Service) IsDegraded(ctx context.Context) bool {
	result := s.Check(ctx)
	return result.Status == string(StatusDegraded)
}

// GetDegradedComponents returns a list of degraded component names.
func (s *Service) GetDegradedComponents(ctx context.Context) []string {
	result := s.Check(ctx)
	return result.Degraded
}

