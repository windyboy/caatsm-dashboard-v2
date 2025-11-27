package ws

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Default client buffer size
	defaultClientBufferSize = 100

	// Maximum time to wait for message write
	writeTimeout = 10 * time.Second
)

// Client represents a single WebSocket client connection with backpressure control.
type Client struct {
	conn        *websocket.Conn
	send        chan []byte
	hub         *Hub
	remoteAddr  string
	connectedAt time.Time
	logger      *zap.Logger
	mu          sync.Mutex
	isClosed    bool
}

// NewClient creates a new client connection.
func NewClient(conn *websocket.Conn, hub *Hub, logger *zap.Logger) *Client {
	return &Client{
		conn:        conn,
		send:        make(chan []byte, defaultClientBufferSize),
		hub:         hub,
		remoteAddr:  conn.RemoteAddr().String(),
		connectedAt: time.Now(),
		logger:      logger,
	}
}

// RemoteAddr returns the client's remote address.
func (c *Client) RemoteAddr() string {
	return c.remoteAddr
}

// ConnectedAt returns when the client connected.
func (c *Client) ConnectedAt() time.Time {
	return c.connectedAt
}

// SendMessage sends a message to the client's send buffer.
// Returns false if the buffer is full (backpressure detected).
func (c *Client) SendMessage(msg []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return false
	}

	select {
	case c.send <- msg:
		return true
	default:
		// Buffer is full - backpressure detected
		return false
	}
}

// SendMessageJSON sends a JSON message to the client.
func (c *Client) SendMessageJSON(msg *Message) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.Warn("failed to marshal message", zap.Error(err))
		return false
	}
	return c.SendMessage(data)
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// It handles inbound messages, triggers client unregister/cleanup on error or context cancellation.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	// Configure connection
	if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		c.logger.Warn("failed to set read deadline", zap.Error(err))
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			c.logger.Warn("failed to set read deadline in pong handler", zap.Error(err))
		}
		return nil
	})

	// Read loop (mainly for pong messages and context cancellation)
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("read pump context cancelled", zap.String("remote_addr", c.remoteAddr))
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Warn("websocket read error", zap.Error(err))
				}
				return
			}
		}
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
// It sends periodic pings to keep the connection alive, and closes the connection
// and unregisters the client when done.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("write pump context cancelled", zap.String("remote_addr", c.remoteAddr))
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				c.logger.Warn("failed to set write deadline", zap.Error(err))
			}
			if !ok {
				// Hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Non-blocking write with timeout
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Warn("websocket write error", zap.Error(err))
				return
			}

			// Track successful write for metrics
			c.hub.metrics.messagesSentInc()

		case <-ticker.C:
			// Send ping to keep connection alive
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Warn("failed to send ping", zap.Error(err))
				return
			}
		}
	}
}

// Close closes the client connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return
	}

	c.isClosed = true
	close(c.send)
}

// IsSlow detects if the client is slow based on buffer fill level.
func (c *Client) IsSlow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If buffer is more than 80% full, consider it slow
	bufferCapacity := cap(c.send)
	bufferLength := len(c.send)

	return bufferLength > (bufferCapacity * 80 / 100)
}

// BufferFillLevel returns the buffer fill level as a percentage.
func (c *Client) BufferFillLevel() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cap(c.send) == 0 {
		return 0
	}

	return (len(c.send) * 100) / cap(c.send)
}

// GetIP extracts the IP address from the client's remote address.
func (c *Client) GetIP() string {
	host, _, err := net.SplitHostPort(c.remoteAddr)
	if err != nil {
		return c.remoteAddr
	}
	return host
}
