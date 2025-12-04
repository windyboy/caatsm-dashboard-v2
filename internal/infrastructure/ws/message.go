package ws

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// StatsData represents the unified stats message data
type StatsData struct {
	Total         int64            `json:"total"`
	ByType        map[string]int64 `json:"byType"`
	ActiveRoutes int64            `json:"activeRoutes"`
	MessagesPerSec float64        `json:"messagesPerSec"`
	TimeWindow    string           `json:"timeWindow"`
}

// Message types
const (
	MessageTypeStats   = "stats"
	MessageTypeMessage = "message"
)
