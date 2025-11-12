package services

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

type noopSearchService struct{}

// NewNoopSearchService returns a placeholder implementation of SearchService.
func NewNoopSearchService() SearchService {
	return &noopSearchService{}
}

func (n *noopSearchService) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
	return nil, repository.ErrNotImplemented
}

func (n *noopSearchService) Autocomplete(ctx context.Context, term string, size int) ([]string, error) {
	return nil, repository.ErrNotImplemented
}

type noopStatsService struct{}

// NewNoopStatsService returns a placeholder implementation of StatsService.
func NewNoopStatsService() StatsService {
	return &noopStatsService{}
}

func (n *noopStatsService) TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error) {
	return nil, repository.ErrNotImplemented
}

func (n *noopStatsService) TopRoutes(ctx context.Context, limit int) ([]models.RouteStat, error) {
	return nil, repository.ErrNotImplemented
}

type noopExportService struct{}

// NewNoopExportService returns a placeholder implementation of ExportService.
func NewNoopExportService() ExportService {
	return &noopExportService{}
}

func (n *noopExportService) Export(ctx context.Context, filter models.SearchFilter, format models.ExportFormat) ([]byte, error) {
	return nil, repository.ErrNotImplemented
}
