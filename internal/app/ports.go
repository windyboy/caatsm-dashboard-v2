package app

import "context"

// Repository defines the interface for data access operations.
// This is the primary port for the application layer.
type Repository interface {
	// Search operations
	Search(ctx context.Context, filters SearchFilters) (*SearchResult, error)
	// StreamSearch streams search results for large exports.
	// Returns two channels: telegrams and errors. Both must be consumed concurrently.
	// Producer closes both channels when done. Consumer must not close channels.
	StreamSearch(ctx context.Context, filters SearchFilters) (<-chan *Telegram, <-chan error)

	// Statistics operations
	TrafficSummary(ctx context.Context, window TimeWindow) (*TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]RouteStat, error)

	// Persistence operations
	Save(ctx context.Context, telegram *Telegram) error
	BulkSave(ctx context.Context, telegrams []any) error

	// Convenience methods for simple operations
	FindByID(ctx context.Context, messageID string) (*Telegram, error)
	Delete(ctx context.Context, messageID string) error
}

// Cache defines the interface for caching operations.
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string, delta int64) (int64, error)
	HIncrBy(ctx context.Context, key string, field string, delta int64) (int64, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

// AutocompleteHit represents a single hit from an autocomplete search.
type AutocompleteHit struct {
	FlightNumber string
	MessageID    string
	Source       string
	Destination  string
}

// AutocompleteResponse represents the response from an autocomplete search.
type AutocompleteResponse struct {
	Hits []AutocompleteHit
}

// SearchIndex defines the interface for search operations.
type SearchIndex interface {
	EnsureIndex(ctx context.Context) error
	Index(ctx context.Context, telegram *Telegram) error
	BulkIndex(ctx context.Context, telegrams []*Telegram) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (*AutocompleteResponse, error)
}

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

// StreamConsumer processes live telegram events from NATS JetStream.
type StreamConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

// WSMessage is the message format for WebSocket communication from app layer.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
