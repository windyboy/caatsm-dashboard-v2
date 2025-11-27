package services

import (
	"context"
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// DashboardRequest represents a request for dashboard data
type DashboardRequest struct {
	SearchFilters domain.SearchFilters
	TimeRange     domain.TimeWindow
	UserID        string
}

// DashboardResponse contains all dashboard data
type DashboardResponse struct {
	Search    *domain.SearchResult   `json:"search,omitempty"`
	Stats     *domain.TrafficSummary `json:"stats"`
	Realtime  *RealtimeInfo          `json:"realtime"`
	Timestamp time.Time              `json:"timestamp"`
}

// RealtimeInfo contains real-time dashboard information
type RealtimeInfo struct {
	ActiveConnections int     `json:"active_connections"`
	MessagesPerSecond float64 `json:"messages_per_second"`
	Uptime            string  `json:"uptime"`
}

// DashboardService provides unified dashboard functionality
type DashboardService struct {
	searchSvc *SearchService
	statsSvc  *StatsService
	exportSvc *ExportService
	realtime  *RealtimeManager
	logger    *zap.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	searchSvc *SearchService,
	statsSvc *StatsService,
	exportSvc *ExportService,
	realtime *RealtimeManager,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		searchSvc: searchSvc,
		statsSvc:  statsSvc,
		exportSvc: exportSvc,
		realtime:  realtime,
		logger:    logger,
	}
}

// GetDashboardData retrieves all dashboard data in a single call
func (ds *DashboardService) GetDashboardData(ctx context.Context, req *DashboardRequest) (*DashboardResponse, error) {
	ds.logger.Info("fetching dashboard data",
		zap.String("user_id", req.UserID),
		zap.Time("start_time", req.TimeRange.Start),
		zap.Time("end_time", req.TimeRange.End),
	)

	// Result types that bundle data with errors
	type searchResultOrError struct {
		result *domain.SearchResult
		err    error
	}
	type statsResultOrError struct {
		result *domain.TrafficSummary
		err    error
	}
	type realtimeResultOrError struct {
		result *RealtimeInfo
		err    error
	}

	// Parallel data fetching for better performance
	searchChan := make(chan searchResultOrError, 1)
	statsChan := make(chan statsResultOrError, 1)
	realtimeChan := make(chan realtimeResultOrError, 1)

	// Fetch search results if requested
	if req.SearchFilters.Query != "" || len(req.SearchFilters.Types) > 0 {
		go func() {
			result, err := ds.searchSvc.Search(ctx, req.SearchFilters)
			searchChan <- searchResultOrError{result: result, err: err}
		}()
	} else {
		searchChan <- searchResultOrError{result: nil, err: nil}
	}

	// Fetch statistics
	go func() {
		stats, err := ds.statsSvc.GetStats(ctx, req.TimeRange)
		statsChan <- statsResultOrError{result: stats, err: err}
	}()

	// Fetch realtime info
	go func() {
		info := ds.realtime.GetInfo()
		realtimeChan <- realtimeResultOrError{result: info, err: nil}
	}()

	// Collect results
	var searchResult *domain.SearchResult
	var statsResult *domain.TrafficSummary
	var realtimeResult *RealtimeInfo

	// Collect search results
	select {
	case sr := <-searchChan:
		if sr.err != nil {
			return nil, fmt.Errorf("search failed: %w", sr.err)
		}
		searchResult = sr.result
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Collect stats results
	select {
	case st := <-statsChan:
		if st.err != nil {
			return nil, fmt.Errorf("stats failed: %w", st.err)
		}
		statsResult = st.result
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Collect realtime results
	select {
	case rt := <-realtimeChan:
		if rt.err != nil {
			return nil, fmt.Errorf("realtime failed: %w", rt.err)
		}
		realtimeResult = rt.result
	case <-ctx.Done():
		return nil, ctx.Err()
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

// GetRealtimeInfo returns real-time dashboard information
func (ds *DashboardService) GetRealtimeInfo() *RealtimeInfo {
	return ds.realtime.GetInfo()
}

// Autocomplete provides search suggestions
func (ds *DashboardService) Autocomplete(ctx context.Context, query string, size int) ([]string, error) {
	return ds.searchSvc.Autocomplete(ctx, query, size)
}
