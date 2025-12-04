package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	// Determine time window identifier
	stats.TimeWindow = s.determineTimeWindow(timeRange)

	// Cache result
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, stats); err != nil {
			s.logger.Warn("failed to cache stats", zap.Error(err))
		}
	}

	s.logger.Info("stats retrieved",
		zap.Int64("total_messages", stats.TotalMessages),
		zap.Int("by_type_count", len(stats.ByType)),
		zap.Int64("active_routes", stats.ActiveRoutes),
		zap.String("time_window", stats.TimeWindow),
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
// For dynamic time windows (like "last 24 hours"), we round down to the nearest minute
// to allow some caching while still being reasonably fresh
func (s *StatsService) buildCacheKey(timeRange TimeWindow) string {
	// Round down to nearest minute for dynamic windows to allow short-term caching
	// This prevents cache misses due to microsecond differences while keeping data fresh
	startUnix := timeRange.Start.Unix()
	endUnix := timeRange.End.Unix()
	
	// Round down to nearest minute (60 seconds)
	startRounded := (startUnix / 60) * 60
	endRounded := (endUnix / 60) * 60
	
	return fmt.Sprintf("stats:%d:%d", startRounded, endRounded)
}

// buildHistoricalCacheKey creates a cache key for historical stats
func (s *StatsService) buildHistoricalCacheKey(timeRange TimeWindow, interval string) string {
	return fmt.Sprintf("historical_stats:%s:%d:%d",
		interval,
		timeRange.Start.Unix(),
		timeRange.End.Unix(),
	)
}

// GetMessagesPerSec calculates the message rate based on the last 1 minute
func (s *StatsService) GetMessagesPerSec(ctx context.Context) (float64, error) {
	now := time.Now()
	windowStart := now.Add(-60 * time.Second)
	window := TimeWindow{
		Start: windowStart,
		End:   now,
	}

	stats, err := s.repo.TrafficSummary(ctx, window)
	if err != nil {
		s.logger.Warn("failed to get messages per sec",
			zap.Error(err),
			zap.Time("window_start", windowStart),
			zap.Time("window_end", now),
		)
		return 0.0, fmt.Errorf("get messages per sec failed: %w", err)
	}

	// Calculate actual time window duration for accurate rate calculation
	actualDuration := now.Sub(windowStart)
	if actualDuration <= 0 {
		actualDuration = 60 * time.Second // Fallback to 60 seconds
	}

	// Calculate rate: messages per second
	var rate float64
	if actualDuration.Seconds() > 0 {
		rate = float64(stats.TotalMessages) / actualDuration.Seconds()
	}

	// Log for debugging (only when there are messages or rate > 0 to avoid log spam)
	if stats.TotalMessages > 0 || rate > 0 {
		s.logger.Debug("messages per sec calculated",
			zap.Int64("total_messages", stats.TotalMessages),
			zap.Duration("duration", actualDuration),
			zap.Float64("rate", rate),
			zap.Time("window_start", windowStart),
			zap.Time("window_end", now),
		)
	}

	return rate, nil
}

// determineTimeWindow determines the time window identifier based on the time range
func (s *StatsService) determineTimeWindow(timeRange TimeWindow) string {
	if timeRange.Start.IsZero() && timeRange.End.IsZero() {
		return "all_time"
	}

	duration := timeRange.End.Sub(timeRange.Start)
	if duration <= time.Hour {
		return "last_1h"
	} else if duration <= 24*time.Hour {
		return "last_24h"
	}
	return "all_time"
}
