package repository

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
)

// TelegramStore encapsulates persistence for telegram records.
type TelegramStore interface {
	Save(ctx context.Context, telegram *domain.Telegram) error
	BulkSave(ctx context.Context, telegrams []*domain.Telegram) error
	Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error)
}

// AnalyticsStore provides aggregated metrics queries.
type AnalyticsStore interface {
	TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error)
}

// SearchIndex abstracts operations against the Meilisearch index.
type SearchIndex interface {
	Index(ctx context.Context, telegram *domain.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error
	Delete(ctx context.Context, messageID string) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error)
}

// StreamConsumer processes live telegram events from NATS JetStream.
type StreamConsumer interface {
	Start(ctx context.Context) error
	Close() error
}
