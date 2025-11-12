package repository

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/models"
)

// TelegramStore encapsulates persistence for telegram records.
type TelegramStore interface {
	Save(ctx context.Context, telegram *models.Telegram) error
	BulkSave(ctx context.Context, telegrams []*models.Telegram) error
	Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error)
}

// AnalyticsStore provides aggregated metrics queries.
type AnalyticsStore interface {
	TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]models.RouteStat, error)
}

// SearchIndex abstracts operations against the Meilisearch index.
type SearchIndex interface {
	Index(ctx context.Context, telegram *models.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*models.Telegram) error
	Delete(ctx context.Context, messageID string) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (interface{}, error)
}

// StreamConsumer processes live telegram events from NATS JetStream.
type StreamConsumer interface {
	Start(ctx context.Context) error
	Close() error
}
