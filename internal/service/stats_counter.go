package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// StatsCounterService maintains real-time statistics counters in Redis.
type StatsCounterService struct {
	cache  Cache
	logger *zap.Logger
}

// NewStatsCounterService creates a new stats counter service.
func NewStatsCounterService(cache Cache, logger *zap.Logger) *StatsCounterService {
	return &StatsCounterService{
		cache:  cache,
		logger: logger,
	}
}

// Increment updates counters for a telegram message.
// It updates total count, type count, route count, and minute-based count for rate calculation.
func (s *StatsCounterService) Increment(ctx context.Context, telegram *app.Telegram) (*domain.StatsIncremented, error) {
	now := time.Now()

	// Update total count for 24h window
	totalKey := "stats:total:24h"
	_, err := s.cache.Incr(ctx, totalKey, 1)
	if err != nil {
		return nil, fmt.Errorf("increment total: %w", err)
	}

	// Update type count
	typeKey := "stats:by_type:24h"
	typeCount := int64(1)
	if telegram.Type != "" {
		_, err = s.cache.HIncrBy(ctx, typeKey, telegram.Type, 1)
		if err != nil {
			s.logger.Warn("failed to increment type count",
				zap.String("type", telegram.Type),
				zap.Error(err))
		}
	}

	// Update route count if route is available
	routeKey := "stats:routes:24h"
	var routeString string
	byType := make(map[string]int64)
	if telegram.Type != "" {
		byType[telegram.Type] = typeCount
	}

	if telegram.Source != "" && telegram.Destination != "" {
		routeString = telegram.Source + "→" + telegram.Destination
		routeHashKey := routeString
		_, err = s.cache.HIncrBy(ctx, routeKey, routeHashKey, 1)
		if err != nil {
			s.logger.Warn("failed to increment route count",
				zap.String("route", routeHashKey),
				zap.Error(err))
		}
	}

	// Update minute-based counter for rate calculation (keep for 2 minutes)
	minuteKey := fmt.Sprintf("stats:minute:%d", now.Unix()/60)
	minuteCount, err := s.cache.Incr(ctx, minuteKey, 1)
	if err != nil {
		s.logger.Warn("failed to increment minute count",
			zap.String("minute_key", minuteKey),
			zap.Error(err))
	} else {
		// Set TTL to 2 minutes for minute counters
		// Note: The cache.Incr already sets TTL, but we want to ensure it's 2 minutes
		// This is handled by the cache implementation's TTL setting
		s.logger.Debug("minute counter incremented",
			zap.String("minute_key", minuteKey),
			zap.Int64("count", minuteCount))
	}

	// Create incremental stats event
	event := &domain.StatsIncremented{
		Total:     1,
		ByType:    byType,
		Route:     routeString,
		Timestamp: now,
	}

	return event, nil
}

// GetCurrentSnapshot retrieves current statistics snapshot from Redis counters.
func (s *StatsCounterService) GetCurrentSnapshot(ctx context.Context) (*domain.TrafficSummary, error) {
	now := time.Now()
	windowStart := now.Add(-24 * time.Hour)

	// Get total count
	totalKey := "stats:total:24h"
	total, err := s.cache.GetInt64(ctx, totalKey)
	if err != nil {
		return nil, fmt.Errorf("get total count: %w", err)
	}

	// Get type distribution
	typeKey := "stats:by_type:24h"
	typeHash, err := s.cache.HGetAll(ctx, typeKey)
	if err != nil {
		return nil, fmt.Errorf("get type distribution: %w", err)
	}

	byType := make(map[string]int64)
	for typeName, countStr := range typeHash {
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			s.logger.Warn("failed to parse type count",
				zap.String("type", typeName),
				zap.String("count", countStr),
				zap.Error(err))
			continue
		}
		byType[typeName] = count
	}

	// Get active routes count
	routeKey := "stats:routes:24h"
	routeHash, err := s.cache.HGetAll(ctx, routeKey)
	if err != nil {
		return nil, fmt.Errorf("get routes: %w", err)
	}

	activeRoutes := int64(len(routeHash))

	// Calculate messages per second from last 60 seconds
	messagesPerSec, err := s.getMessagesPerSec(ctx, now)
	if err != nil {
		s.logger.Debug("failed to calculate messages per sec",
			zap.Error(err))
		messagesPerSec = 0.0
	}

	// Determine time window identifier
	windowDuration := now.Sub(windowStart)
	var timeWindow string
	if windowDuration <= time.Hour {
		timeWindow = "last_1h"
	} else if windowDuration <= 24*time.Hour {
		timeWindow = "last_24h"
	} else {
		timeWindow = "all_time"
	}

	return &domain.TrafficSummary{
		TotalMessages:  total,
		ByType:         byType,
		ActiveRoutes:   activeRoutes,
		MessagesPerSec: messagesPerSec,
		TimeWindow:     timeWindow,
	}, nil
}

// getMessagesPerSec calculates messages per second based on minute counters from the last 60 seconds.
func (s *StatsCounterService) getMessagesPerSec(ctx context.Context, now time.Time) (float64, error) {
	// Get counters from the last minute (current minute)
	currentMinute := now.Unix() / 60
	currentMinuteKey := fmt.Sprintf("stats:minute:%d", currentMinute)

	currentCount, err := s.cache.GetInt64(ctx, currentMinuteKey)
	if err != nil {
		return 0.0, fmt.Errorf("get current minute count: %w", err)
	}

	// Calculate how many seconds have elapsed in the current minute
	secondsInCurrentMinute := now.Unix() % 60
	if secondsInCurrentMinute == 0 {
		secondsInCurrentMinute = 60 // If exactly at minute boundary, use full minute
	}

	// Calculate rate based on current minute's count and elapsed seconds
	// This gives us an approximate rate
	var rate float64
	if secondsInCurrentMinute > 0 {
		rate = float64(currentCount) / float64(secondsInCurrentMinute)
	}

	return rate, nil
}

