package health

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// checkNATS checks NATS JetStream connection health.
func (s *Service) checkNATS() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastChecked: time.Now(),
	}

	if s.natsConn == nil {
		health.Status = StatusUnknown
		health.Message = "NATS connection not configured"
		return health
	}

	// Check if connection is connected
	if !s.natsConn.IsConnected() {
		health.Latency = time.Since(start)
		health.Status = StatusError
		health.Message = "NATS connection not established"
		s.logger.Warn("nats health check failed: not connected", zap.Duration("latency", health.Latency))
		return health
	}

	// Check if connection is closed
	if s.natsConn.IsClosed() {
		health.Latency = time.Since(start)
		health.Status = StatusError
		health.Message = "NATS connection closed"
		s.logger.Warn("nats health check failed: connection closed", zap.Duration("latency", health.Latency))
		return health
	}

	// Get stats
	stats := s.natsConn.Stats()
	health.Latency = time.Since(start)

	// Check for reconnects (too many indicates instability)
	if stats.Reconnects > 10 {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("high reconnect count: %d", stats.Reconnects)
		s.logger.Warn("nats health check degraded: high reconnects", 
			zap.Uint64("reconnects", stats.Reconnects),
			zap.Duration("latency", health.Latency))
		return health
	}

	health.Status = StatusOK
	health.Message = fmt.Sprintf("connected, %d in / %d out msgs", stats.InMsgs, stats.OutMsgs)
	return health
}

// checkWebSocket checks WebSocket hub health.
func (s *Service) checkWebSocket() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastChecked: time.Now(),
		Latency:     time.Since(start),
	}

	if s.wsHub == nil {
		health.Status = StatusUnknown
		health.Message = "WebSocket hub not configured"
		return health
	}

	// Get client count
	clientCount := s.wsHub.GetClientCount()
	health.Status = StatusOK
	health.Message = fmt.Sprintf("%d active connections", clientCount)

	// Warn if too many connections (potential DoS or memory issue)
	if clientCount > 1000 {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("high connection count: %d", clientCount)
		s.logger.Warn("websocket health check: high connection count", zap.Int("clients", clientCount))
	}

	return health
}

