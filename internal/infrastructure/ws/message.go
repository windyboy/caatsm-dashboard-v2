package ws

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// StatsData represents the unified stats message data (full snapshot)
type StatsData struct {
	Total         int64            `json:"total"`
	ByType        map[string]int64 `json:"byType"`
	ActiveRoutes int64            `json:"activeRoutes"`
	MessagesPerSec float64        `json:"messagesPerSec"`
	TimeWindow    string           `json:"timeWindow"`
}

// StatsDeltaData represents incremental statistics updates
type StatsDeltaData struct {
	Total      int64            `json:"total"`
	ByType     map[string]int64 `json:"byType"`
	Route      string           `json:"route,omitempty"`
	Timestamp  string           `json:"timestamp"`
	TimeWindow string           `json:"timeWindow,omitempty"`
}

// HealthData represents health status information
type HealthData struct {
	Status      string          `json:"status"`
	Version     string          `json:"version,omitempty"`
	Uptime      string          `json:"uptime,omitempty"`
	Timestamp   string          `json:"timestamp"`
	PostgreSQL  interface{}     `json:"postgresql,omitempty"`
	Meilisearch interface{}     `json:"meilisearch,omitempty"`
	Redis       interface{}     `json:"redis,omitempty"`
	NATS        interface{}     `json:"nats,omitempty"`
}

// Message types
const (
	MessageTypeStats     = "stats"       // Full stats snapshot
	MessageTypeStatsDelta = "stats_delta" // Incremental stats update
	MessageTypeStatsFull  = "stats_full"  // Full stats (triggered intelligently)
	MessageTypeHealth     = "health"      // Health status update
	MessageTypeMessage    = "message"     // New telegram message
)
