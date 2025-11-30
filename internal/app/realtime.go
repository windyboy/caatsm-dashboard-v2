package app

import (
	"sync"
	"time"
)

// RealtimeManager manages real-time dashboard information
type RealtimeManager struct {
	activeConnections int
	messagesPerSecond float64
	startTime         time.Time
	mu                sync.RWMutex
}

// NewRealtimeManager creates a new realtime manager
func NewRealtimeManager() *RealtimeManager {
	return &RealtimeManager{
		startTime: time.Now(),
	}
}

// GetInfo returns current real-time information
func (rm *RealtimeManager) GetInfo() *RealtimeInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return &RealtimeInfo{
		ActiveConnections: rm.activeConnections,
		MessagesPerSecond: rm.messagesPerSecond,
		Uptime:            time.Since(rm.startTime).String(),
	}
}

// UpdateConnections updates the active connection count
func (rm *RealtimeManager) UpdateConnections(count int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.activeConnections = count
}

// UpdateMessageRate updates the messages per second rate
func (rm *RealtimeManager) UpdateMessageRate(rate float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.messagesPerSecond = rate
}
