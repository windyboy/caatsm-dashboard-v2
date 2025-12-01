package app

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
)

// Re-export domain functions
func GetHTTPStatus(code string) int {
	return domain.GetHTTPStatus(code)
}

// Type re-exports for backward compatibility with handlers and other packages
type (
	Telegram       = domain.Telegram
	SearchFilters  = domain.SearchFilters
	SearchResult   = domain.SearchResult
	TimeWindow     = domain.TimeWindow
	TrafficSummary = domain.TrafficSummary
	RouteStat      = domain.RouteStat
	ExportFormat   = domain.ExportFormat
	Pagination     = domain.Pagination
	Event          = domain.Event
)

// Re-export domain constants
const (
	MaxExportRecords = domain.MaxExportRecords
	MaxTimeRangeDays = domain.MaxTimeRangeDays
	ExportFormatCSV  = domain.ExportFormatCSV
)

// Re-export domain errors
var (
	ErrInvalidInput   = domain.ErrInvalidInput
	ErrNotFound       = domain.ErrNotFound
	ErrUnauthorized   = domain.ErrUnauthorized
	ErrForbidden      = domain.ErrForbidden
	ErrRateLimit      = domain.ErrRateLimit
	ErrInternalError  = domain.ErrInternalError
	ErrServiceUnavailable = domain.ErrServiceUnavailable
)

// Re-export domain error types
type (
	ErrInvalidTelegram = domain.ErrInvalidTelegram
	ErrInvalidFilter   = domain.ErrInvalidFilter
	APIError           = domain.APIError
)

// Repository defines the interface for data persistence operations.
type Repository interface {
	Save(ctx context.Context, telegram *domain.Telegram) error
	BulkSave(ctx context.Context, telegrams []any) error
	Search(ctx context.Context, filter domain.SearchFilters) (*domain.SearchResult, error)
	StreamSearch(ctx context.Context, filter domain.SearchFilters) (<-chan *domain.Telegram, <-chan error)
	FindByID(ctx context.Context, messageID string) (*domain.Telegram, error)
	TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error)
	RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error)
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
	Index(ctx context.Context, telegram *domain.Telegram) error
	BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error
	Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error)
	SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (*AutocompleteResponse, error)
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

// WSMessage is the message format for WebSocket communication from app layer.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// DashboardRequest represents a request for dashboard data
type DashboardRequest struct {
	SearchFilters SearchFilters
	TimeRange     TimeWindow
	UserID        string
}

// DashboardResponse contains all dashboard data
type DashboardResponse struct {
	Search    *SearchResult   `json:"search,omitempty"`
	Stats     *TrafficSummary `json:"stats"`
	Realtime  interface{}     `json:"realtime"` // Using interface{} to avoid circular dependency
	Timestamp interface{}     `json:"timestamp"` // Using interface{} to avoid import cycle with time
}

// RealtimeInfo contains real-time dashboard information
type RealtimeInfo struct {
	ActiveConnections int     `json:"active_connections"`
	MessagesPerSecond float64 `json:"messages_per_second"`
	Uptime            string  `json:"uptime"`
}

// AutocompleteSuggestion represents a single autocomplete suggestion with its type
type AutocompleteSuggestion struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Label string `json:"label"`
}
