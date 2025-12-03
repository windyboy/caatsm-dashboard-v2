package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// StatsService handles statistics operations
type StatsService struct {
	repo   Repository
	cache  Cache
	logger *zap.Logger
}

// NewStatsService creates a new stats service
func NewStatsService(repo Repository, cache Cache, logger *zap.Logger) *StatsService {
	return &StatsService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// GetStats retrieves traffic statistics for the given time window
func (s *StatsService) GetStats(ctx context.Context, timeRange TimeWindow) (*TrafficSummary, error) {
	cacheKey := s.buildCacheKey(timeRange)

	// Try cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			s.logger.Warn("cache get error", zap.String("key", cacheKey), zap.Error(err))
		} else if cached != nil {
			statsBytes, err := json.Marshal(cached)
			if err != nil {
				s.logger.Warn("failed to marshal cached stats", zap.String("key", cacheKey), zap.Error(err))
				_ = s.cache.Delete(ctx, cacheKey)
			} else {
				var stats TrafficSummary
				if err := json.Unmarshal(statsBytes, &stats); err != nil {
					s.logger.Warn("failed to unmarshal cached stats", zap.String("key", cacheKey), zap.Error(err))
					_ = s.cache.Delete(ctx, cacheKey)
				} else {
					s.logger.Debug("stats cache hit", zap.String("key", cacheKey))
					return &stats, nil
				}
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
func (s *StatsService) GetRouteStats(ctx context.Context, limit int) ([]RouteStat, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	cacheKey := fmt.Sprintf("route_stats:%d", limit)

	// Try cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			s.logger.Warn("cache get error", zap.String("key", cacheKey), zap.Error(err))
		} else if cached != nil {
			payload, err := json.Marshal(cached)
			if err != nil {
				s.logger.Warn("failed to marshal cached route stats", zap.String("key", cacheKey), zap.Error(err))
				_ = s.cache.Delete(ctx, cacheKey)
			} else {
				var stats []RouteStat
				if err := json.Unmarshal(payload, &stats); err != nil {
					s.logger.Warn("failed to unmarshal cached route stats", zap.String("key", cacheKey), zap.Error(err))
					_ = s.cache.Delete(ctx, cacheKey)
				} else {
					s.logger.Debug("route stats cache hit", zap.String("key", cacheKey))
					return stats, nil
				}
			}
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

// GetHistoricalStats retrieves time-series statistics for the given time window and interval.
func (s *StatsService) GetHistoricalStats(ctx context.Context, timeRange TimeWindow, interval string) (*domain.HistoricalStats, error) {
	if interval != "hour" && interval != "day" {
		interval = "hour" // default
	}

	cacheKey := s.buildHistoricalCacheKey(timeRange, interval)

	// Try cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			s.logger.Warn("cache get error", zap.String("key", cacheKey), zap.Error(err))
		} else if cached != nil {
			statsBytes, err := json.Marshal(cached)
			if err != nil {
				s.logger.Warn("failed to marshal cached historical stats", zap.String("key", cacheKey), zap.Error(err))
				_ = s.cache.Delete(ctx, cacheKey)
			} else {
				var stats domain.HistoricalStats
				if err := json.Unmarshal(statsBytes, &stats); err != nil {
					s.logger.Warn("failed to unmarshal cached historical stats", zap.String("key", cacheKey), zap.Error(err))
					_ = s.cache.Delete(ctx, cacheKey)
				} else {
					s.logger.Debug("historical stats cache hit", zap.String("key", cacheKey))
					return &stats, nil
				}
			}
		}
	}

	// Fetch from repository
	stats, err := s.repo.HistoricalStats(ctx, timeRange, interval)
	if err != nil {
		s.logger.Error("failed to get historical stats",
			zap.Error(err),
			zap.Time("start", timeRange.Start),
			zap.Time("end", timeRange.End),
			zap.String("interval", interval),
		)
		return nil, fmt.Errorf("get historical stats failed: %w", err)
	}

	// Cache result
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, stats); err != nil {
			s.logger.Warn("failed to cache historical stats", zap.Error(err))
		}
	}

	s.logger.Info("historical stats retrieved",
		zap.String("interval", interval),
		zap.Int("data_points", len(stats.Data)),
	)

	return stats, nil
}

// buildCacheKey creates a cache key for stats
func (s *StatsService) buildCacheKey(timeRange TimeWindow) string {
	return fmt.Sprintf("stats:%d:%d",
		timeRange.Start.Unix(),
		timeRange.End.Unix(),
	)
}

// buildHistoricalCacheKey creates a cache key for historical stats
func (s *StatsService) buildHistoricalCacheKey(timeRange TimeWindow, interval string) string {
	return fmt.Sprintf("historical_stats:%s:%d:%d",
		interval,
		timeRange.Start.Unix(),
		timeRange.End.Unix(),
	)
}
