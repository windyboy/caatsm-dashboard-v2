package service

import (
	"context"
	"time"

	"github.com/windy/caatsm-dashboard/internal/domain"
)

// DashboardService provides unified dashboard functionality
type DashboardService struct {
	searchSvc *SearchService
	statsSvc  *StatsService
	realtime *RealtimeManager
	repo     Repository
	cache    Cache
	statsTTL time.Duration
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	searchSvc *SearchService,
	statsSvc *StatsService,
	realtime *RealtimeManager,
	repo Repository,
	cache Cache,
	statsTTL time.Duration,
) *DashboardService {
	return &DashboardService{
		searchSvc: searchSvc,
		statsSvc:  statsSvc,
		realtime:  realtime,
		repo:      repo,
		cache:     cache,
		statsTTL:  statsTTL,
	}
}

// Search performs a search operation
func (ds *DashboardService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	return ds.searchSvc.Search(ctx, filters)
}

// GetStats retrieves traffic statistics
func (ds *DashboardService) GetStats(ctx context.Context, timeRange domain.TimeWindow) (*domain.TrafficSummary, error) {
	return ds.statsSvc.GetStats(ctx, timeRange)
}

// GetHistoricalStats retrieves time-series statistics
func (ds *DashboardService) GetHistoricalStats(ctx context.Context, timeRange domain.TimeWindow, interval string) (*domain.HistoricalStats, error) {
	return ds.statsSvc.GetHistoricalStats(ctx, timeRange, interval)
}
