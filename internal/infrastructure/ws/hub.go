package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// IP-based connection tracking
	ipConnections map[string]int

	// Configuration
	config Config

	// Channels
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Logger
	logger *zap.Logger

	// Metrics
	metrics *HubMetrics
}

// Config holds configuration for the WebSocket hub.
type Config struct {
	// MaxConnectionsPerIP limits connections per IP address (0 = unlimited)
	MaxConnectionsPerIP int

	// ClientBufferSize is the buffer size for each client's send channel
	ClientBufferSize int

	// EnableMetrics enables metrics collection
	EnableMetrics bool
}

// DefaultConfig returns default hub configuration.
func DefaultConfig() Config {
	return Config{
		MaxConnectionsPerIP: 5,
		ClientBufferSize:    100,
		EnableMetrics:       false, // Default to disabled, enable via config
	}
}

// NewHub creates a new WebSocket hub.
func NewHub(config Config, logger *zap.Logger) *Hub {
	if config.ClientBufferSize == 0 {
		config.ClientBufferSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	hub := &Hub{
		clients:       make(map[*Client]bool),
		ipConnections: make(map[string]int),
		config:        config,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte, 256),
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		metrics:       NewHubMetrics(config.EnableMetrics),
	}

	go hub.run()

	return hub
}

// run is the main hub loop that handles client registration and message broadcasting.
func (h *Hub) run() {
	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("hub shutting down")
			return

		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// handleRegister registers a new client with the hub.
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check connection limits per IP
	clientIP := client.GetIP()
	if h.config.MaxConnectionsPerIP > 0 {
		currentConnections := h.ipConnections[clientIP]
		if currentConnections >= h.config.MaxConnectionsPerIP {
			h.logger.Warn("connection limit exceeded for IP",
				zap.String("ip", clientIP),
				zap.Int("limit", h.config.MaxConnectionsPerIP))
			client.Close()
			h.metrics.connectionRejectedInc()
			return
		}
		h.ipConnections[clientIP] = currentConnections + 1
	}

	h.clients[client] = true
	h.metrics.activeConnectionsInc()
	h.metrics.totalConnectionsInc()

	h.logger.Info("client registered",
		zap.String("remote_addr", client.RemoteAddr()),
		zap.String("ip", clientIP),
		zap.Int("total_clients", len(h.clients)))
}

// unregisterClient removes a client from the hub.
// The caller must hold h.mu write lock.
func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	client.Close()

	// Update IP connection count
	clientIP := client.GetIP()
	if h.config.MaxConnectionsPerIP > 0 {
		h.ipConnections[clientIP]--
		if h.ipConnections[clientIP] <= 0 {
			delete(h.ipConnections, clientIP)
		}
	}

	h.metrics.activeConnectionsDec()

	h.logger.Info("client unregistered",
		zap.String("remote_addr", client.RemoteAddr()),
		zap.Int("total_clients", len(h.clients)))
}

// handleUnregister removes a client from the hub.
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unregisterClient(client)
}

// handleBroadcast sends a message to all registered clients.
func (h *Hub) handleBroadcast(message []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	dropped := 0
	success := 0
	var slowClients []*Client

	for _, client := range clients {
		if client.SendMessage(message) {
			success++
		} else {
			dropped++
			// If client is slow and buffer is full, mark for disconnection
			if client.IsSlow() {
				h.logger.Warn("slow client detected, marking for disconnection",
					zap.String("remote_addr", client.RemoteAddr()),
					zap.Int("buffer_fill", client.BufferFillLevel()))
				slowClients = append(slowClients, client)
			}
		}
	}

	h.metrics.messagesBroadcastAdd(float64(success))
	h.metrics.messagesDroppedAdd(float64(dropped))

	if dropped > 0 {
		h.logger.Debug("broadcast complete",
			zap.Int("success", success),
			zap.Int("dropped", dropped))
	}

	// Unregister slow clients synchronously while holding write lock
	// This avoids deadlock since handleBroadcast runs in the run() goroutine
	// and cannot send to the unbuffered unregister channel
	if len(slowClients) > 0 {
		h.mu.Lock()
		for _, client := range slowClients {
			h.unregisterClient(client)
		}
		h.mu.Unlock()
	}
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		// Broadcast channel is full - log warning
		h.logger.Warn("broadcast channel full, message dropped")
		h.metrics.messagesDroppedAdd(1)
	}
}

// BroadcastJSON broadcasts a JSON message to all clients.
func (h *Hub) BroadcastJSON(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	h.Broadcast(data)
	return nil
}

// GetClientCount returns the current number of connected clients.
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetIPConnectionCount returns the number of connections for a given IP.
func (h *Hub) GetIPConnectionCount(ip string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ipConnections[ip]
}

// Context returns the hub's context for lifecycle management.
func (h *Hub) Context() context.Context {
	return h.ctx
}

// Close gracefully shuts down the hub.
func (h *Hub) Close() {
	h.cancel()

	// Wait a bit for messages to drain
	time.Sleep(100 * time.Millisecond)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all clients
	for client := range h.clients {
		client.Close()
	}

	h.clients = make(map[*Client]bool)
	h.ipConnections = make(map[string]int)

	h.logger.Info("hub closed")
}
