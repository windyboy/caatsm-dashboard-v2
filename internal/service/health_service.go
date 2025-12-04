package service

import (
	"context"
	"sync"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// HealthService periodically checks health status and publishes updates.
type HealthService struct {
	healthChecker  HealthCheckService
	eventPublisher app.EventPublisher
	logger         *zap.Logger
	lastHealth     app.HealthCheckResult
	lastHealthMu   sync.RWMutex
	checkInterval  time.Duration
}

// HealthCheckService interface for checking health status
type HealthCheckService interface {
	HealthCheck(ctx context.Context) app.HealthCheckResult
}

// NewHealthService creates a new health service.
func NewHealthService(
	healthChecker HealthCheckService,
	eventPublisher app.EventPublisher,
	logger *zap.Logger,
) *HealthService {
	return &HealthService{
		healthChecker:  healthChecker,
		eventPublisher: eventPublisher,
		logger:         logger,
		checkInterval:  30 * time.Second,
	}
}

// Start starts the health service as a background task.
func (s *HealthService) Start(ctx context.Context) error {
	// Perform initial health check
	s.performHealthCheck(ctx)

	// Start periodic health checks
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("health service stopped", zap.Error(ctx.Err()))
			return nil
		case <-ticker.C:
			s.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck checks health status and publishes event if changed.
func (s *HealthService) performHealthCheck(ctx context.Context) {
	health := s.healthChecker.HealthCheck(ctx)

	// Check if health status has changed
	s.lastHealthMu.RLock()
	lastStatus := s.lastHealth.Status
	s.lastHealthMu.RUnlock()

	// Only publish if status changed
	if lastStatus == "" || lastStatus != health.Status {
		event := domain.HealthUpdated{
			Health: health,
			At:     time.Now(),
		}

		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			s.logger.Warn("failed to publish health update event",
				zap.Error(err),
			)
		} else {
			s.logger.Debug("health update event published",
				zap.String("status", health.Status),
			)
		}

		// Update last known health
		s.lastHealthMu.Lock()
		s.lastHealth = health
		s.lastHealthMu.Unlock()
	}
}

