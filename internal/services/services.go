package services

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/models"
)

// SearchService defines operations for querying telemetry messages.
type SearchService interface {
	Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error)
	Autocomplete(ctx context.Context, term string, size int) ([]string, error)
}

// StatsService exposes aggregated insights for dashboards.
type StatsService interface {
	TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error)
	TopRoutes(ctx context.Context, limit int) ([]models.RouteStat, error)
}

// ExportService handles exporting search results into different formats.
type ExportService interface {
	Export(ctx context.Context, filter models.SearchFilter, format models.ExportFormat) ([]byte, error)
}
