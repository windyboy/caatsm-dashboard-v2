package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/models"
	oldpostgres "github.com/windy/caatsm-dashboard/internal/repository/postgres"
)

// Store implements persistence operations using PostgreSQL.
// This is the new infrastructure layer implementation.
// It wraps the old postgres store and adapts types.
type Store struct {
	// Delegate to old store (models.Telegram is aliased to persistence.Telegram, so compatible)
	oldStore *oldpostgres.Store
}

// New creates a new PostgreSQL store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{
		oldStore: oldpostgres.New(pool),
	}
}

// Save persists a single telegram.
// Since models.Telegram is a type alias for persistence.Telegram, types are identical.
func (s *Store) Save(ctx context.Context, telegram *persistence.Telegram) error {
	// Type alias means *persistence.Telegram and *models.Telegram are the same type
	return s.oldStore.Save(ctx, (*models.Telegram)(telegram))
}

// BulkSave persists multiple telegrams.
func (s *Store) BulkSave(ctx context.Context, telegrams []*persistence.Telegram) error {
	// Convert slice - since types are aliased, we can convert element-by-element
	modelTelegrams := make([]*models.Telegram, len(telegrams))
	for i := range telegrams {
		modelTelegrams[i] = (*models.Telegram)(telegrams[i])
	}
	return s.oldStore.BulkSave(ctx, modelTelegrams)
}

// Search performs a structured query over telegram records.
func (s *Store) Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error) {
	// Convert filter - create new struct with same values
	modelFilter := models.SearchFilter{
		Query:       filter.Query,
		Type:        filter.Type,
		Source:      filter.Source,
		Destination: filter.Destination,
		Priority:    filter.Priority,
		TimeRange:   models.TimeWindow(filter.TimeRange),
		Page:        models.Pagination(filter.Page),
	}
	
	result, err := s.oldStore.Search(ctx, modelFilter)
	if err != nil {
		return nil, err
	}
	
	// Convert result
	telegrams := make([]persistence.Telegram, len(result.Telegrams))
	for i := range result.Telegrams {
		telegrams[i] = persistence.Telegram(result.Telegrams[i])
	}
	
	return &persistence.SearchResult{
		Telegrams: telegrams,
		Total:     result.Total,
		Page:      persistence.Pagination(result.Page),
	}, nil
}

// TrafficSummary returns aggregated data for dashboards.
func (s *Store) TrafficSummary(ctx context.Context, window persistence.TimeWindow) (*persistence.TrafficSummary, error) {
	modelWindow := models.TimeWindow(window)
	result, err := s.oldStore.TrafficSummary(ctx, modelWindow)
	if err != nil {
		return nil, err
	}
	
	return &persistence.TrafficSummary{
		TotalMessages: result.TotalMessages,
		ByType:        result.ByType,
		ByPriority:    result.ByPriority,
	}, nil
}

// RouteStats returns top routes.
func (s *Store) RouteStats(ctx context.Context, limit int) ([]persistence.RouteStat, error) {
	modelStats, err := s.oldStore.RouteStats(ctx, limit)
	if err != nil {
		return nil, err
	}
	
	stats := make([]persistence.RouteStat, len(modelStats))
	for i := range modelStats {
		stats[i] = persistence.RouteStat(modelStats[i])
	}
	return stats, nil
}

