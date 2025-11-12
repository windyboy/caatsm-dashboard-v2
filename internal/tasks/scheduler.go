package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Scheduler coordinates background tasks such as aggregation and cleanup.
type Scheduler struct {
	logger *zap.Logger
	ticker *time.Ticker
}

// NewScheduler returns a new scheduler instance with the supplied interval.
func NewScheduler(interval time.Duration, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		logger: logger,
		ticker: time.NewTicker(interval),
	}
}

// Run starts the scheduler loop. It is currently a placeholder for future jobs.
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info("scheduler started")
	defer s.Stop()
	defer s.logger.Info("scheduler stopped")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.ticker.C:
			s.logger.Debug("scheduler tick", zap.Time("time", time.Now()))
		}
	}
}

// Stop releases scheduler resources.
func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
