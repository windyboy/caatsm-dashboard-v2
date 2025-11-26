package ws

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

// HubMetrics tracks metrics for the WebSocket hub.
type HubMetrics struct {
	activeConnections *atomic.Int64
	totalConnections  *atomic.Int64
	
	// Prometheus metrics
	messagesSent       prometheus.Counter
	messagesBroadcast  prometheus.Counter
	messagesDropped    prometheus.Counter
	connectionRejected prometheus.Counter
	
	// Dedicated registry to avoid duplicate registration panics
	registry *prometheus.Registry
	
	enabled bool
}

// NewHubMetrics creates a new metrics tracker with a dedicated Prometheus registry.
// This avoids duplicate registration panics when creating multiple hub instances.
func NewHubMetrics(enabled bool) *HubMetrics {
	m := &HubMetrics{
		activeConnections: &atomic.Int64{},
		totalConnections:  &atomic.Int64{},
		enabled:           enabled,
	}
	
	if enabled {
		// Create a dedicated registry for this hub instance
		m.registry = prometheus.NewRegistry()
		
		m.messagesSent = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "websocket_messages_sent_total",
			Help: "Total number of WebSocket messages sent",
		})
		
		m.messagesBroadcast = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "websocket_messages_broadcast_total",
			Help: "Total number of WebSocket messages broadcasted",
		})
		
		m.messagesDropped = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "websocket_messages_dropped_total",
			Help: "Total number of WebSocket messages dropped due to backpressure",
		})
		
		m.connectionRejected = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "websocket_connections_rejected_total",
			Help: "Total number of WebSocket connections rejected",
		})
		
		// Register on the dedicated registry instead of the global registry
		m.registry.MustRegister(
			m.messagesSent,
			m.messagesBroadcast,
			m.messagesDropped,
			m.connectionRejected,
		)
	}
	
	return m
}

// ActiveConnections returns the current number of active connections.
func (m *HubMetrics) ActiveConnections() int64 {
	if !m.enabled {
		return 0
	}
	return m.activeConnections.Load()
}

// TotalConnections returns the total number of connections since hub start.
func (m *HubMetrics) TotalConnections() int64 {
	if !m.enabled {
		return 0
	}
	return m.totalConnections.Load()
}

// activeConnectionsInc increments active connections counter.
func (m *HubMetrics) activeConnectionsInc() {
	if m.enabled {
		m.activeConnections.Add(1)
	}
}

// activeConnectionsDec decrements active connections counter.
func (m *HubMetrics) activeConnectionsDec() {
	if m.enabled {
		m.activeConnections.Add(-1)
	}
}

// totalConnectionsInc increments total connections counter.
func (m *HubMetrics) totalConnectionsInc() {
	if m.enabled {
		m.totalConnections.Add(1)
	}
}

// messagesSentInc increments messages sent counter.
func (m *HubMetrics) messagesSentInc() {
	if m.enabled {
		m.messagesSent.Inc()
	}
}

// messagesBroadcastAdd adds a value to messages broadcast counter.
func (m *HubMetrics) messagesBroadcastAdd(val float64) {
	if m.enabled {
		m.messagesBroadcast.Add(val)
	}
}

// messagesDroppedAdd adds a value to messages dropped counter.
func (m *HubMetrics) messagesDroppedAdd(val float64) {
	if m.enabled {
		m.messagesDropped.Add(val)
	}
}

// connectionRejectedInc increments connection rejected counter.
func (m *HubMetrics) connectionRejectedInc() {
	if m.enabled {
		m.connectionRejected.Inc()
	}
}

// Registry returns the dedicated Prometheus registry for this hub's metrics.
// This registry can be used with promhttp.HandlerFor() to expose metrics over HTTP
// or with registry.Gather() for custom metric collection.
// Returns nil if metrics are disabled.
func (m *HubMetrics) Registry() *prometheus.Registry {
	return m.registry
}

