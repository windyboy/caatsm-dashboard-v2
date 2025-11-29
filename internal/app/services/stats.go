package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// StatsService handles statistics operations
type StatsService struct {
	repo   ports.Repository
	cache  ports.Cache
	logger *zap.Logger
}

// NewStatsService creates a new stats service
func NewStatsService(repo ports.Repository, cache ports.Cache, logger *zap.Logger) *StatsService {
	return &StatsService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// GetStats retrieves traffic statistics for the given time window
func (s *StatsService) GetStats(ctx context.Context, timeRange domain.TimeWindow) (*domain.TrafficSummary, error) {
	cacheKey := s.buildCacheKey(timeRange)

	// Try cache first
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			statsBytes, _ := json.Marshal(cached)
			var stats domain.TrafficSummary
			if err := json.Unmarshal(statsBytes, &stats); err == nil {
				s.logger.Debug("stats cache hit", zap.String("key", cacheKey))
				return &stats, nil
			}
		}
	}

	// Fetch from repository
	stats, err := s.repo.TrafficSummary(ctx, timeRange)
	if err != nil {
		s.logger.Error("failed to get stats",
			zap.Error(err),
			zap.Time("start", timeRange.Start),
			zap.Time("end", timeRange.End),
		)
		return nil, fmt.Errorf("get stats failed: %w", err)
	}

	// Cache result
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, stats); err != nil {
			s.logger.Warn("failed to cache stats", zap.Error(err))
		}
	}

	s.logger.Info("stats retrieved",
		zap.Int64("total_messages", stats.TotalMessages),
		zap.Int("by_type_count", len(stats.ByType)),
		zap.Int("by_priority_count", len(stats.ByPriority)),
	)

	return stats, nil
}

// GetRouteStats retrieves top routes statistics
func (s *StatsService) GetRouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	cacheKey := fmt.Sprintf("route_stats:%d", limit)

	// Try cache first
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			if stats, ok := cached.([]domain.RouteStat); ok {
				s.logger.Debug("route stats cache hit", zap.String("key", cacheKey))
				return stats, nil
			}
			// Type assertion failed - log warning and fetch fresh data
			s.logger.Warn("cache type mismatch",
				zap.String("key", cacheKey),
				zap.String("expected_type", "[]domain.RouteStat"),
				zap.String("actual_type", fmt.Sprintf("%T", cached)),
			)
		}
	}

	// Fetch from repository
	stats, err := s.repo.RouteStats(ctx, limit)
	if err != nil {
		s.logger.Error("failed to get route stats", zap.Error(err), zap.Int("limit", limit))
		return nil, fmt.Errorf("get route stats failed: %w", err)
	}

	// Cache result
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, stats); err != nil {
			s.logger.Warn("failed to cache route stats", zap.Error(err))
		}
	}

	return stats, nil
}

// buildCacheKey creates a cache key for stats
func (s *StatsService) buildCacheKey(timeRange domain.TimeWindow) string {
	return fmt.Sprintf("stats:%d:%d",
		timeRange.Start.Unix(),
		timeRange.End.Unix(),
	)
}
