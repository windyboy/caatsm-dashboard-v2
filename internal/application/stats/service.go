package stats

import (
	"context"
	"time"

	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"go.uber.org/zap"
)

const (
	DefaultTimeWindow = 24 * time.Hour
	DefaultStatsLimit = 10
)

// Service implements statistics use cases.
type Service struct {
	store  repository.AnalyticsStore
	logger *zap.Logger
}

// NewService creates a new stats service.
func NewService(store repository.AnalyticsStore, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// TrafficSummary returns aggregated traffic metrics.
// If no time window is provided, defaults to the last 24 hours.
func (s *Service) TrafficSummary(ctx context.Context, window persistence.TimeWindow) (*persistence.TrafficSummary, error) {
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
func (s *Service) TopRoutes(ctx context.Context, limit int) ([]persistence.RouteStat, error) {
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

