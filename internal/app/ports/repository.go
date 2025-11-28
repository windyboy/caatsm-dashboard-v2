package ports

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
)

// Repository defines the interface for data access operations.
// This is the primary port for the application layer.
type Repository interface {
	// Search operations
	Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error)

	// Statistics operations
	TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error)

	// Persistence operations
	Save(ctx context.Context, telegram *domain.Telegram) error
	BulkSave(ctx context.Context, telegrams []*domain.Telegram) error
}

// Cache defines the interface for caching operations.
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}

// SearchIndex defines the interface for search operations.
type SearchIndex interface {
	Index(ctx context.Context, telegram *domain.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (interface{}, error)
}

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}

// StreamConsumer processes live telegram events from NATS JetStream.
type StreamConsumer interface {
	Start(ctx context.Context) error
	Close() error
}
