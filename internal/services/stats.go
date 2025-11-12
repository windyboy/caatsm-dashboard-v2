package services

import (
	"context"
	"time"

	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"go.uber.org/zap"
)

// statsService implements StatsService using repository.
type statsService struct {
	store  repository.AnalyticsStore
	logger *zap.Logger
}

// NewStatsService creates a new StatsService implementation.
func NewStatsService(store repository.AnalyticsStore, logger *zap.Logger) StatsService {
	return &statsService{
		store:  store,
		logger: logger,
	}
}

// TrafficSummary returns aggregated traffic metrics.
// If no time window is provided, defaults to the last 24 hours.
func (s *statsService) TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error) {
	// If no time window is provided, default to last 24 hours
	if window.Start.IsZero() && window.End.IsZero() {
		window.End = time.Now()
		window.Start = window.End.Add(-DefaultTimeWindow)
	}

	summary, err := s.store.TrafficSummary(ctx, window)
	if err != nil {
		if err == repository.ErrNotImplemented {
			return nil, err
		}
		return nil, err
	}

	return summary, nil
}

// TopRoutes returns the top routes by message count.
func (s *statsService) TopRoutes(ctx context.Context, limit int) ([]models.RouteStat, error) {
	if limit <= 0 {
		limit = DefaultStatsLimit
	}

	routes, err := s.store.RouteStats(ctx, limit)
	if err != nil {
		if err == repository.ErrNotImplemented {
			return nil, err
		}
		return nil, err
	}

	return routes, nil
}
