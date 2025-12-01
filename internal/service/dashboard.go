package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// Type aliases for app types
type (
	DashboardRequest  = app.DashboardRequest
	DashboardResponse = app.DashboardResponse
	RealtimeInfo      = app.RealtimeInfo
)

// DashboardService provides unified dashboard functionality
type DashboardService struct {
	searchSvc *SearchService
	statsSvc  *StatsService
	exportSvc *ExportService
	realtime  *RealtimeManager
	repo      Repository
	cache     Cache
	statsTTL  time.Duration
	logger    *zap.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	searchSvc *SearchService,
	statsSvc *StatsService,
	exportSvc *ExportService,
	realtime *RealtimeManager,
	repo Repository,
	cache Cache,
	statsTTL time.Duration,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		searchSvc: searchSvc,
		statsSvc:  statsSvc,
		exportSvc: exportSvc,
		realtime:  realtime,
		repo:      repo,
		cache:     cache,
		statsTTL:  statsTTL,
		logger:    logger,
	}
}

// GetDashboardData retrieves all dashboard data in a single call
func (ds *DashboardService) GetDashboardData(ctx context.Context, req *DashboardRequest) (*DashboardResponse, error) {
	ds.logger.Debug("fetching dashboard data",
		zap.String("user_id", req.UserID),
		zap.Time("start_time", req.TimeRange.Start),
		zap.Time("end_time", req.TimeRange.End),
	)

	searchResult, statsResult, realtimeResult, err := ds.getDashboardDataHelper(ctx, req)
	if err != nil {
		return nil, err
	}

	response := &DashboardResponse{
		Search:    searchResult,
		Stats:     statsResult,
		Realtime:  realtimeResult,
		Timestamp: time.Now(),
	}

	ds.logger.Info("dashboard data fetched successfully",
		zap.Bool("has_search", searchResult != nil),
		zap.Bool("has_stats", statsResult != nil),
		zap.Int("active_connections", func() int {
			if realtimeResult != nil {
				return realtimeResult.ActiveConnections
			}
			return 0
		}()),
	)

	return response, nil
}

// getDashboardDataHelper fetches dashboard data sequentially
func (ds *DashboardService) getDashboardDataHelper(ctx context.Context, req *DashboardRequest) (*domain.SearchResult, *domain.TrafficSummary, *RealtimeInfo, error) {
	var searchResult *domain.SearchResult
	if req.SearchFilters.Query != "" || len(req.SearchFilters.Types) > 0 {
		result, err := ds.searchSvc.Search(ctx, req.SearchFilters)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("search failed: %w", err)
		}
		searchResult = result
	}

	stats, err := ds.statsSvc.GetStats(ctx, req.TimeRange)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stats failed: %w", err)
	}

	realtime := ds.realtime.GetInfo()
	return searchResult, stats, realtime, nil
}

// Search performs a search operation
func (ds *DashboardService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	return ds.searchSvc.Search(ctx, filters)
}

// GetStats retrieves traffic statistics
func (ds *DashboardService) GetStats(ctx context.Context, timeRange domain.TimeWindow) (*domain.TrafficSummary, error) {
	return ds.statsSvc.GetStats(ctx, timeRange)
}

// Export exports data in the specified format
func (ds *DashboardService) Export(ctx context.Context, filters domain.SearchFilters, format domain.ExportFormat) ([]byte, error) {
	return ds.exportSvc.Export(ctx, filters, format)
}

// ExportStream returns channels for streaming export (for large datasets)
func (ds *DashboardService) ExportStream(ctx context.Context, filters domain.SearchFilters) (<-chan *domain.Telegram, <-chan error, error) {
	return ds.exportSvc.ExportStream(ctx, filters)
}

// GetRealtimeInfo returns real-time dashboard information
func (ds *DashboardService) GetRealtimeInfo() *RealtimeInfo {
	return ds.realtime.GetInfo()
}

// Autocomplete provides search suggestions
func (ds *DashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return ds.searchSvc.Autocomplete(ctx, query, size)
}

// AutocompleteWithTypes provides search suggestions with type information
func (ds *DashboardService) AutocompleteWithTypes(ctx context.Context, query string, size int) ([]app.AutocompleteSuggestion, error) {
	suggestions, err := ds.searchSvc.AutocompleteWithTypes(ctx, query, size)
	if err != nil {
		return nil, err
	}
	// Convert service.AutocompleteSuggestion to app.AutocompleteSuggestion
	result := make([]app.AutocompleteSuggestion, len(suggestions))
	for i, s := range suggestions {
		result[i] = app.AutocompleteSuggestion{
			Value: s.Value,
			Type:  s.Type,
			Label: s.Label,
		}
	}
	return result, nil
}

const (
	Days90           = 90
	RecentLimit      = 20
	statsTotalKey    = "ws:stats:total:90d"
	statsPriorityKey = "ws:stats:priority:90d"
	statsTypeKey     = "ws:stats:type:90d"
)

// StreamInitialData implements ports.DashboardPort.
// It sends initial stats (total/priority/type) and recent 50 messages (batch).
// Messages sent to ch in order: stats-total, stats-priority, stats-type, then messages (oldest first).
func (ds *DashboardService) StreamInitialData(ctx context.Context, ch chan<- app.WSMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Create time window for last 90 days
	window := domain.TimeWindow{
		Start: time.Now().Add(-Days90 * 24 * time.Hour),
		End:   time.Now(),
	}

	// Get statistics for the time window
	stats, err := ds.statsSvc.GetStats(ctx, window)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	// Initialize stats cache with current values to ensure realtime updates
	// increment from correct baseline
	if err := ds.initializeStatsCache(ctx, stats); err != nil {
		return fmt.Errorf("initialize stats cache: %w", err)
	}

	// Send stats-total
	select {
	case ch <- WSMessage{
		Type: "stats-total",
		Data: map[string]interface{}{
			"total": stats.TotalMessages,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Send stats-priority
	select {
	case ch <- WSMessage{
		Type: "stats-priority",
		Data: map[string]interface{}{
			"byPriority": stats.ByPriority,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Send stats-type
	select {
	case ch <- WSMessage{
		Type: "stats-type",
		Data: map[string]interface{}{
			"byType": stats.ByType,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Get recent messages (order by time DESC, then reverse to oldest first)
	filters := domain.SearchFilters{
		Pagination: domain.Pagination{
			Limit:  RecentLimit,
			SortBy: "time",
			Order:  "desc",
		},
	}

	searchResult, err := ds.repo.Search(ctx, filters)
	if err != nil {
		return fmt.Errorf("get recent messages: %w", err)
	}

	// Send messages in reverse order (oldest first)
	if searchResult != nil && len(searchResult.Telegrams) > 0 {
		// Reverse the slice to get oldest first
		for i := len(searchResult.Telegrams) - 1; i >= 0; i-- {
			select {
			case ch <- WSMessage{
				Type: "message",
				Data: searchResult.Telegrams[i],
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// HandleEvent implements ports.DashboardPort.
// It processes a realtime event: parse telegram, increment stats cache atomically,
// send message, then updated stats (total/priority/type).
func (ds *DashboardService) HandleEvent(ctx context.Context, event []byte, ch chan<- WSMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Parse the event as TelegramReceived
	var telegramReceived TelegramReceived
	if err := json.Unmarshal(event, &telegramReceived); err != nil {
		return fmt.Errorf("parse event: %w", err)
	}

	telegram := telegramReceived.Telegram
	if telegram == nil {
		return fmt.Errorf("telegram is nil in event")
	}

	// Atomically update stats cache
	if err := ds.updateStatsCache(ctx, telegram); err != nil {
		ds.logger.Error("failed to update stats cache", zap.Error(err))
		// Continue processing even if cache update fails
	}

	// Send message to channel
	select {
	case ch <- WSMessage{
		Type: "message",
		Data: telegram,
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Send updated stats
	if err := ds.sendUpdatedStats(ctx, ch); err != nil {
		return fmt.Errorf("send updated stats: %w", err)
	}

	return nil
}

// updateStatsCache atomically increments the stats counters in cache
func (ds *DashboardService) updateStatsCache(ctx context.Context, telegram *domain.Telegram) error {
	// Increment total messages
	if _, err := ds.cache.Incr(ctx, statsTotalKey, 1); err != nil {
		return fmt.Errorf("incr total: %w", err)
	}

	// Increment by priority (use string key for hash field)
	priorityField := fmt.Sprintf("%d", telegram.Priority)
	if _, err := ds.cache.HIncrBy(ctx, statsPriorityKey, priorityField, 1); err != nil {
		return fmt.Errorf("hincr priority: %w", err)
	}

	// Increment by type
	if _, err := ds.cache.HIncrBy(ctx, statsTypeKey, telegram.Type, 1); err != nil {
		return fmt.Errorf("hincr type: %w", err)
	}

	return nil
}

// sendUpdatedStats sends the current stats from cache to the channel
func (ds *DashboardService) sendUpdatedStats(ctx context.Context, ch chan<- WSMessage) error {
	// Get total
	total, err := ds.cache.GetInt64(ctx, statsTotalKey)
	if err != nil {
		return fmt.Errorf("get total: %w", err)
	}

	// Send stats-total
	select {
	case ch <- WSMessage{
		Type: "stats-total",
		Data: map[string]interface{}{
			"total": total,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Get by priority
	byPriority, err := ds.cache.HGetAll(ctx, statsPriorityKey)
	if err != nil {
		return fmt.Errorf("get priority stats: %w", err)
	}

	// Convert string map to int map
	priorityMap := make(map[int]int64)
	for k, v := range byPriority {
		priority, err := strconv.Atoi(k)
		if err != nil {
			ds.logger.Warn("failed to parse priority key from cache",
				zap.String("raw_key", k),
				zap.String("raw_value", v),
				zap.Error(err),
			)
			continue
		}
		count, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ds.logger.Warn("failed to parse priority count from cache",
				zap.String("raw_key", k),
				zap.String("raw_value", v),
				zap.Error(err),
			)
			continue
		}
		priorityMap[priority] = count
	}

	// Send stats-priority
	select {
	case ch <- WSMessage{
		Type: "stats-priority",
		Data: map[string]interface{}{
			"byPriority": priorityMap,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Get by type
	byType, err := ds.cache.HGetAll(ctx, statsTypeKey)
	if err != nil {
		return fmt.Errorf("get type stats: %w", err)
	}

	// Convert string values to int64
	typeMap := make(map[string]int64)
	for k, v := range byType {
		count, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ds.logger.Warn("failed to parse type count from cache",
				zap.String("raw_key", k),
				zap.String("raw_value", v),
				zap.Error(err),
			)
			continue
		}
		typeMap[k] = count
	}

	// Send stats-type
	select {
	case ch <- WSMessage{
		Type: "stats-type",
		Data: map[string]interface{}{
			"byType": typeMap,
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// initializeStatsCache seeds the cache with initial stats values.
// It deletes existing keys first to ensure clean state, then sets the total,
// priority, and type counts using the same cache keys/structure that
// sendUpdatedStats expects. Respects TTL semantics via cache implementation.
func (ds *DashboardService) initializeStatsCache(ctx context.Context, stats *domain.TrafficSummary) error {
	// Delete existing keys to ensure clean state before initializing
	if err := ds.cache.Delete(ctx, statsTotalKey); err != nil {
		return fmt.Errorf("delete total key: %w", err)
	}
	if err := ds.cache.Delete(ctx, statsPriorityKey); err != nil {
		return fmt.Errorf("delete priority key: %w", err)
	}
	if err := ds.cache.Delete(ctx, statsTypeKey); err != nil {
		return fmt.Errorf("delete type key: %w", err)
	}

	// Set total count (Incr creates key if it doesn't exist)
	if _, err := ds.cache.Incr(ctx, statsTotalKey, stats.TotalMessages); err != nil {
		return fmt.Errorf("set total count: %w", err)
	}

	// Set priority counts (HIncrBy creates hash and fields if they don't exist)
	for priority, count := range stats.ByPriority {
		priorityField := fmt.Sprintf("%d", priority)
		if _, err := ds.cache.HIncrBy(ctx, statsPriorityKey, priorityField, count); err != nil {
			return fmt.Errorf("set priority count for %d: %w", priority, err)
		}
	}

	// Set type counts (HIncrBy creates hash and fields if they don't exist)
	for msgType, count := range stats.ByType {
		if _, err := ds.cache.HIncrBy(ctx, statsTypeKey, msgType, count); err != nil {
			return fmt.Errorf("set type count for %s: %w", msgType, err)
		}
	}

	return nil
}
