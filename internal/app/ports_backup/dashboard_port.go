package ports

import "context"

// WSMessage is the message format for WebSocket communication from app layer.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// DashboardPort defines the interface for dashboard real-time operations.
// Implementations are in app/services, used by delivery/ws.
type DashboardPort interface {
	// StreamInitialData sends initial stats (total/priority/type) and recent 50 messages (batch).
	// Messages sent to ch in order: stats-total, stats-priority, stats-type, then messages (oldest first).
	StreamInitialData(ctx context.Context, ch chan<- WSMessage) error

	// HandleEvent processes a realtime event: parse telegram, increment stats cache atomically,
	// send message, then updated stats (total/priority/type).
	HandleEvent(ctx context.Context, event []byte, ch chan<- WSMessage) error
}
