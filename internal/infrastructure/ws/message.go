package ws

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Message types
const (
	MessageTypeStatsTotal    = "stats-total"
	MessageTypeStatsPriority = "stats-priority"
	MessageTypeStatsType     = "stats-type"
	MessageTypeMessage       = "message"
)
