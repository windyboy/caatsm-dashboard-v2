package ws

import "time"

// TimeoutsConfig holds configurable timeout values for WebSocket connections.
type TimeoutsConfig struct {
	// WriteWait is the maximum time to wait for a message write to complete
	WriteWait time.Duration

	// PongWait is the maximum time to wait for a pong response
	PongWait time.Duration

	// PingPeriod is the interval between ping messages (should be < PongWait)
	PingPeriod time.Duration

	// MaxMessageSize is the maximum message size allowed from peer (in bytes)
	MaxMessageSize int64
}

// DefaultTimeouts returns default timeout configuration.
func DefaultTimeouts() TimeoutsConfig {
	pongWait := 60 * time.Second
	return TimeoutsConfig{
		WriteWait:      10 * time.Second,
		PongWait:       pongWait,
		PingPeriod:     (pongWait * 9) / 10,
		MaxMessageSize: 512 * 1024, // 512KB
	}
}

// Validate ensures timeout values are reasonable.
func (tc *TimeoutsConfig) Validate() error {
	if tc.WriteWait <= 0 {
		return ErrInvalidConfig{Field: "WriteWait", Reason: "must be positive"}
	}
	if tc.PongWait <= 0 {
		return ErrInvalidConfig{Field: "PongWait", Reason: "must be positive"}
	}
	if tc.PingPeriod <= 0 {
		return ErrInvalidConfig{Field: "PingPeriod", Reason: "must be positive"}
	}
	if tc.PingPeriod >= tc.PongWait {
		return ErrInvalidConfig{Field: "PingPeriod", Reason: "must be less than PongWait"}
	}
	if tc.MaxMessageSize <= 0 {
		return ErrInvalidConfig{Field: "MaxMessageSize", Reason: "must be positive"}
	}
	return nil
}

// ErrInvalidConfig represents a configuration validation error.
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid config: " + e.Field + " - " + e.Reason
}
