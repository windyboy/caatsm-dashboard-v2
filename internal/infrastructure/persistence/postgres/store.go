package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/windy/caatsm-dashboard/internal/domain"
	oldpostgres "github.com/windy/caatsm-dashboard/internal/repository/postgres"
)

// Store implements persistence operations using PostgreSQL.
// This is the new infrastructure layer implementation.
// It wraps the old postgres store and adapts types.
type Store struct {
	// Delegate to old store (persistence.Telegram is aliased to persistence.Telegram, so compatible)
	oldStore *oldpostgres.Store
}

// New creates a new PostgreSQL store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{
		oldStore: oldpostgres.New(pool),
	}
}

// Save persists a single telegram.
func (s *Store) Save(ctx context.Context, telegram *domain.Telegram) error {
	return s.oldStore.Save(ctx, telegram)
}

// BulkSave persists multiple telegrams.
func (s *Store) BulkSave(ctx context.Context, telegrams []*domain.Telegram) error {
	return s.oldStore.BulkSave(ctx, telegrams)
}

// Search performs a structured query over telegram records.
func (s *Store) Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error) {
	return s.oldStore.Search(ctx, filter)
}

// FindByID retrieves a telegram by its message ID.
func (s *Store) FindByID(ctx context.Context, messageID string) (*domain.Telegram, error) {
	// Use Search with message_id filter as a workaround until FindByID is implemented in old store
	filter := domain.SearchFilters{
		Query: messageID,
		Pagination: domain.Pagination{
			Limit:  1,
			Offset: 0,
		},
	}
	result, err := s.Search(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(result.Telegrams) == 0 {
		return nil, nil // or return a NotFound error
	}
	return &result.Telegrams[0], nil
}

// TrafficSummary returns aggregated data for dashboards.
func (s *Store) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	return s.oldStore.TrafficSummary(ctx, window)
}

// RouteStats returns top routes.
func (s *Store) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	return s.oldStore.RouteStats(ctx, limit)
}
