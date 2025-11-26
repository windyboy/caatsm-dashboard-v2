package ports

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
)

// TelegramRepository defines persistence operations for telegrams.
// This is a port interface that the application layer depends on.
// Infrastructure implementations (PostgreSQL, etc.) implement this interface.
type TelegramRepository interface {
	Save(ctx context.Context, telegram *persistence.Telegram) error
	BulkSave(ctx context.Context, telegrams []*persistence.Telegram) error
	Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error)
	FindByID(ctx context.Context, messageID string) (*persistence.Telegram, error)
}

// SearchIndex defines indexing operations for search engines.
// Implementations include Meilisearch, Elasticsearch, etc.
type SearchIndex interface {
	Index(ctx context.Context, telegram *persistence.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*persistence.Telegram) error
	Delete(ctx context.Context, messageID string) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error)
}

// Cache defines caching operations.
// Implementations include Redis/Valkey, in-memory, etc.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// EventBus defines event publishing operations.
// Implementations include Redis Pub/Sub, NATS, Kafka, etc.
type EventBus interface {
	Publish(ctx context.Context, topic string, data map[string]any) error
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
}

// AnalyticsRepository defines analytics and aggregation operations.
type AnalyticsRepository interface {
	TrafficSummary(ctx context.Context, window persistence.TimeWindow) (*persistence.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]persistence.RouteStat, error)
}

// DomainEventBus defines domain event publishing operations.
// This is separate from the general EventBus to handle domain events specifically.
type DomainEventBus interface {
	PublishDomainEvent(ctx context.Context, event domain.Event) error
}

