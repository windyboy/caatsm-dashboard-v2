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
	// StreamSearch streams search results for large exports (avoids loading all data into memory).
	//
	// Channel Lifecycle:
	//   - The implementation (producer) is responsible for closing both channels.
	//   - Both channels are closed once all work is finished, whether due to successful completion,
	//     an error, or context cancellation.
	//   - Callers must never close these channels.
	//
	// Error Semantics:
	//   - The error channel emits a single terminal error (if any) and then is closed.
	//   - If no error occurs, the error channel is closed without sending any value.
	//   - Once an error is sent, no further telegrams will be sent on the telegram channel.
	//
	// Context Cancellation:
	//   - Implementations must monitor ctx.Done() and stop work promptly when the context is cancelled.
	//   - On cancellation, implementations must:
	//     1. Stop producing telegrams immediately
	//     2. Send ctx.Err() to the error channel (if not already closed)
	//     3. Close both channels
	//   - Implementations should use ctx-aware select statements to avoid blocking on channel sends
	//     when the context is done.
	//
	// Consumption Requirements:
	//   - Callers must consume both channels concurrently to avoid deadlocks.
	//   - The telegram channel may be buffered (implementation-dependent), but callers should not
	//     rely on buffering and must consume both channels together.
	//   - Use a select statement to read from both channels:
	//     select {
	//     case telegram, ok := <-telegramCh:
	//       if !ok { /* channel closed, check errCh */ }
	//     case err, ok := <-errCh:
	//       if !ok { /* channel closed, check telegramCh */ }
	//       if err != nil { /* handle error */ }
	//     }
	//
	// Implementation Guidance:
	//   - Use a goroutine to produce results asynchronously.
	//   - Always use defer to close both channels on all exit paths.
	//   - Use select statements with ctx.Done() when sending to channels to avoid blocking:
	//     select {
	//     case <-ctx.Done():
	//       errCh <- ctx.Err()
	//       return
	//     case telegramCh <- telegram:
	//     }
	//   - Check ctx.Done() before expensive operations (e.g., database queries).
	//   - Ensure the error channel has sufficient buffering (at least 1) to prevent blocking
	//     when sending terminal errors.
	//   - Avoid goroutine leaks by ensuring all code paths eventually close channels and return.
	StreamSearch(ctx context.Context, filters domain.SearchFilters) (<-chan *domain.Telegram, <-chan error)

	// Statistics operations
	TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error)

	// Persistence operations
	Save(ctx context.Context, telegram *domain.Telegram) error
	BulkSave(ctx context.Context, telegrams []any) error
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
