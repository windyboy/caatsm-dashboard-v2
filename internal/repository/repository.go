package repository

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
)

// TelegramStore encapsulates persistence for telegram records.
type TelegramStore interface {
	Save(ctx context.Context, telegram *persistence.Telegram) error
	BulkSave(ctx context.Context, telegrams []*persistence.Telegram) error
	Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error)
}

// AnalyticsStore provides aggregated metrics queries.
type AnalyticsStore interface {
	TrafficSummary(ctx context.Context, window persistence.TimeWindow) (*persistence.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]persistence.RouteStat, error)
}

// SearchIndex abstracts operations against the Meilisearch index.
type SearchIndex interface {
	Index(ctx context.Context, telegram *persistence.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*persistence.Telegram) error
	Delete(ctx context.Context, messageID string) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error)
}

// StreamConsumer processes live telegram events from NATS JetStream.
type StreamConsumer interface {
	Start(ctx context.Context) error
	Close() error
}
