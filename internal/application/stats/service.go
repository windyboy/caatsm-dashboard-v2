package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/domain"
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
func (s *Service) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	// If no time window is provided, default to last 24 hours
	if window.Start.IsZero() && window.End.IsZero() {
		window.End = time.Now()
		window.Start = window.End.Add(-DefaultTimeWindow)
	}

	// Validate time window doesn't exceed max range
	if !window.Start.IsZero() && !window.End.IsZero() {
		duration := window.End.Sub(window.Start)
		maxDuration := 90 * 24 * time.Hour
		if duration > maxDuration {
			return nil, fmt.Errorf("time window cannot exceed 90 days")
		}
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
func (s *Service) TopRoutes(ctx context.Context, limit int) ([]domain.RouteStat, error) {
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

