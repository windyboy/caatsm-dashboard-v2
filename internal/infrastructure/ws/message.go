package ws

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// StatsData represents the unified stats message data
type StatsData struct {
	Total      int64            `json:"total"`
	ByPriority map[int]int64    `json:"byPriority"`
	ByType     map[string]int64 `json:"byType"`
}

// Message types
const (
	MessageTypeStats   = "stats"
	MessageTypeMessage = "message"
)
