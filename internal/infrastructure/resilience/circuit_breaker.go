package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// State represents the circuit breaker state.
type State string

const (
	StateClosed   State = "closed"   // Normal operation
	StateOpen     State = "open"     // Failing, rejecting requests
	StateHalfOpen State = "half_open" // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern for resilient service calls.
type CircuitBreaker struct {
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenMax   int // Max attempts in half-open state before closing

	mu            sync.RWMutex
	state         State
	failures      int
	lastFailure   time.Time
	halfOpenCount int

	logger *zap.Logger

	// Metrics
	successCount int64
	failureCount int64
	rejectCount  int64
}

// Config holds configuration for a circuit breaker.
type Config struct {
	Name         string
	MaxFailures  int           // Number of failures before opening
	ResetTimeout time.Duration // Time before attempting to close
	HalfOpenMax  int           // Max attempts in half-open state
	Logger       *zap.Logger
}

// DefaultConfig returns a default circuit breaker configuration.
func DefaultConfig(name string, logger *zap.Logger) Config {
	return Config{
		Name:         name,
		MaxFailures:  5,
		ResetTimeout: 60 * time.Second,
		HalfOpenMax:  3,
		Logger:       logger,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	return &CircuitBreaker{
		name:         cfg.Name,
		maxFailures:  cfg.MaxFailures,
		resetTimeout: cfg.ResetTimeout,
		halfOpenMax:  cfg.HalfOpenMax,
		state:        StateClosed,
		logger:       cfg.Logger,
	}
}

// Call executes a function through the circuit breaker.
// Returns the function result or an error if the circuit is open.
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	if err := cb.beforeCall(); err != nil {
		return err
	}

	err := fn()
	cb.afterCall(err)

	return err
}

// beforeCall checks if the circuit breaker should allow the call.
func (cb *CircuitBreaker) beforeCall() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			cb.logger.Info("circuit breaker entering half-open state",
				zap.String("name", cb.name),
			)
		} else {
			cb.rejectCount++
			return fmt.Errorf("circuit breaker %s is open", cb.name)
		}
	}

	// In half-open state, limit attempts
	if cb.state == StateHalfOpen {
		if cb.halfOpenCount >= cb.halfOpenMax {
			cb.rejectCount++
			return fmt.Errorf("circuit breaker %s half-open attempts exceeded", cb.name)
		}
		cb.halfOpenCount++
	}

	return nil
}

// afterCall records the result and updates state accordingly.
func (cb *CircuitBreaker) afterCall(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed call.
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()
	cb.failureCount++

	if cb.state == StateHalfOpen {
		// Failed in half-open, go back to open
		cb.state = StateOpen
		cb.halfOpenCount = 0
		cb.logger.Warn("circuit breaker failed in half-open, reopening",
			zap.String("name", cb.name),
		)
		return
	}

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.logger.Error("circuit breaker opened",
			zap.String("name", cb.name),
			zap.Int("failures", cb.failures),
		)
	}
}

// onSuccess handles a successful call.
func (cb *CircuitBreaker) onSuccess() {
	cb.successCount++

	if cb.state == StateHalfOpen {
		// Success in half-open, close the circuit
		cb.state = StateClosed
		cb.failures = 0
		cb.halfOpenCount = 0
		cb.logger.Info("circuit breaker closed after successful half-open test",
			zap.String("name", cb.name),
		)
	} else if cb.state == StateClosed {
		// Reset failure count on success
		cb.failures = 0
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsOpen returns true if the circuit breaker is open.
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// IsClosed returns true if the circuit breaker is closed.
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenCount = 0
	cb.lastFailure = time.Time{}

	cb.logger.Info("circuit breaker manually reset",
		zap.String("name", cb.name),
	)
}

// Stats returns circuit breaker statistics.
type Stats struct {
	State        State
	Failures     int
	SuccessCount int64
	FailureCount int64
	RejectCount  int64
}

// GetStats returns current circuit breaker statistics.
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		State:        cb.state,
		Failures:     cb.failures,
		SuccessCount: cb.successCount,
		FailureCount: cb.failureCount,
		RejectCount:  cb.rejectCount,
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

