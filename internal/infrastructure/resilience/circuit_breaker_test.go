package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	if cb.State() != StateClosed {
		t.Errorf("expected initial state to be closed, got %s", cb.State())
	}

	// Trigger failures to open the circuit
	testErr := errors.New("test failure")
	for i := 0; i < 3; i++ {
		err := cb.Call(context.Background(), func(ctx context.Context) error {
			return testErr
		})
		if !errors.Is(err, testErr) {
			t.Errorf("expected test error, got %v", err)
		}
	}

	if cb.State() != StateOpen {
		t.Errorf("expected state to be open after %d failures, got %s", 3, cb.State())
	}

	// Verify next call is rejected with ErrCircuitOpen
	err := cb.Call(context.Background(), func(ctx context.Context) error {
		t.Error("function should not be called when circuit is open")
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  2,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Open the circuit
	testErr := errors.New("test failure")
	for i := 0; i < 2; i++ {
		_ = cb.Call(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected state to be open, got %s", cb.State())
	}

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open
	called := false
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if !called {
		t.Error("function should be called in half-open state")
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state to be closed after successful half-open call, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenSuccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  3,
		Logger:       logger,
	})

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Successful call in half-open should close the circuit
	err := cb.Call(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state to be closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  3,
		Logger:       logger,
	})

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Failed call in half-open should reopen the circuit
	testErr := errors.New("still failing")
	err := cb.Call(context.Background(), func(ctx context.Context) error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("expected test error, got %v", err)
	}

	if cb.State() != StateOpen {
		t.Errorf("expected state to be open, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenMaxAttempts(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Wait for reset timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Use beforeCall directly to test max attempts without transitioning state
	// This simulates concurrent calls in half-open state
	cb.mu.Lock()
	cb.state = StateHalfOpen
	cb.halfOpenCount = 0
	cb.mu.Unlock()

	// First attempt - should succeed
	err := cb.beforeCall()
	if err != nil {
		t.Errorf("first half-open attempt should succeed, got %v", err)
	}

	// Second attempt - should succeed
	err = cb.beforeCall()
	if err != nil {
		t.Errorf("second half-open attempt should succeed, got %v", err)
	}

	// Third attempt - should be rejected
	err = cb.beforeCall()
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen when half-open attempts exceeded, got %v", err)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 1 * time.Minute,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if cb.State() != StateOpen {
		t.Fatalf("expected state to be open, got %s", cb.State())
	}

	// Manual reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("expected state to be closed after reset, got %s", cb.State())
	}

	// Should accept calls now
	called := false
	err := cb.Call(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("function should be called after reset")
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Successful call
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return nil
	})

	// Failed calls
	testErr := errors.New("failure")
	for i := 0; i < 3; i++ {
		_ = cb.Call(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	// Rejected call
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return nil
	})

	stats := cb.GetStats()

	if stats.State != StateOpen {
		t.Errorf("expected state to be open, got %s", stats.State)
	}

	if stats.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", stats.SuccessCount)
	}

	if stats.FailureCount != 3 {
		t.Errorf("expected 3 failures, got %d", stats.FailureCount)
	}

	if stats.RejectCount != 1 {
		t.Errorf("expected 1 rejection, got %d", stats.RejectCount)
	}

	if stats.Failures != 3 {
		t.Errorf("expected 3 consecutive failures, got %d", stats.Failures)
	}
}

func TestCircuitBreaker_IsOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	if cb.IsOpen() {
		t.Error("circuit should not be open initially")
	}

	if !cb.IsClosed() {
		t.Error("circuit should be closed initially")
	}

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if !cb.IsOpen() {
		t.Error("circuit should be open after failure")
	}

	if cb.IsClosed() {
		t.Error("circuit should not be closed after failure")
	}
}

func TestCircuitBreaker_ErrCircuitOpen_ErrorsIs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test-circuit",
		MaxFailures:  1,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Open the circuit
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	// Call when open should return ErrCircuitOpen
	err := cb.Call(context.Background(), func(ctx context.Context) error {
		return nil
	})

	// Verify errors.Is works
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected errors.Is(err, ErrCircuitOpen) to be true, got false. Error: %v", err)
	}
}

func TestCircuitBreaker_ClosedState_SuccessResetsFailures(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Two failures
	testErr := errors.New("failure")
	for i := 0; i < 2; i++ {
		_ = cb.Call(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	stats := cb.GetStats()
	if stats.Failures != 2 {
		t.Errorf("expected 2 consecutive failures, got %d", stats.Failures)
	}

	// Success should reset failures
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return nil
	})

	stats = cb.GetStats()
	if stats.Failures != 0 {
		t.Errorf("expected failures to be reset after success, got %d", stats.Failures)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state to remain closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := DefaultConfig("test-circuit", logger)

	if cfg.Name != "test-circuit" {
		t.Errorf("expected name 'test-circuit', got '%s'", cfg.Name)
	}

	if cfg.MaxFailures != 5 {
		t.Errorf("expected MaxFailures 5, got %d", cfg.MaxFailures)
	}

	if cfg.ResetTimeout != 60*time.Second {
		t.Errorf("expected ResetTimeout 60s, got %v", cfg.ResetTimeout)
	}

	if cfg.HalfOpenMax != 3 {
		t.Errorf("expected HalfOpenMax 3, got %d", cfg.HalfOpenMax)
	}

	if cfg.Logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestCircuitBreaker_NilLogger(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  1,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       nil, // Nil logger should be replaced with nop
	})

	// Should not panic
	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})

	if cb.logger == nil {
		t.Error("expected logger to be initialized")
	}
}

func TestCircuitBreaker_PanicRecovery(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Test that panic is recorded and re-raised
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to be re-raised")
		}
	}()

	_ = cb.Call(context.Background(), func(ctx context.Context) error {
		panic("test panic")
	})

	// Should not reach here
	t.Error("should not reach this point after panic")
}

func TestCircuitBreaker_PanicRecorded(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  2,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Trigger panic twice to open circuit
	for i := 0; i < 2; i++ {
		func() {
			defer func() {
				_ = recover()
			}()
			_ = cb.Call(context.Background(), func(ctx context.Context) error {
				panic("failure")
			})
		}()
	}

	// Circuit should be open due to recorded panic failures
	if cb.State() != StateOpen {
		t.Errorf("expected circuit to be open after panics, got %s", cb.State())
	}

	stats := cb.GetStats()
	if stats.FailureCount != 2 {
		t.Errorf("expected 2 failures from panics, got %d", stats.FailureCount)
	}
}

func TestCircuitBreaker_ContextPropagation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Test context is passed to function
	type ctxKey string
	key := ctxKey("test-key")
	value := "test-value"
	ctx := context.WithValue(context.Background(), key, value)

	receivedValue := ""
	err := cb.Call(ctx, func(ctx context.Context) error {
		receivedValue = ctx.Value(key).(string)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if receivedValue != value {
		t.Errorf("expected context value %s, got %s", value, receivedValue)
	}
}

func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Test that canceled context is propagated
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := cb.Call(ctx, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return errors.New("context not cancelled")
		}
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}

	// Context cancellation should count as failure
	if cb.State() == StateOpen {
		t.Error("context cancellation should not immediately open circuit")
	}
}

func TestCircuitBreaker_ContextTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(Config{
		Name:         "test",
		MaxFailures:  3,
		ResetTimeout: 100 * time.Millisecond,
		HalfOpenMax:  2,
		Logger:       logger,
	})

	// Test that timeout context is propagated
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := cb.Call(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded error, got %v", err)
	}
}

