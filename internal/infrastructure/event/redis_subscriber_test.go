package event

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
)

// TestRedisSubscriber_Subscribe tests the Redis Pub/Sub subscriber
func TestRedisSubscriber_Subscribe(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Pub/Sub subscriber test")
	}

	logger := zaptest.NewLogger(t)
	subscriber := NewRedisSubscriber(client, logger)

	testChannel := "test:subscriber:channel"
	handlerCalled := make(chan bool, 1)
	var receivedEvent app.Event

	handler := func(ctx context.Context, event app.Event) error {
		receivedEvent = event
		handlerCalled <- true
		return nil
	}

	// Start subscription in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- subscriber.Subscribe(ctx, testChannel, handler)
	}()

	// Wait a bit for subscription to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish a test event
	event := &domain.TelegramPersisted{
		Telegram: &app.Telegram{
			MessageID:    "TEST-SUB-001",
			Type:         "AFTN",
			Time:         time.Now(),
			FlightNumber: "CA100",
			Priority:     1,
		},
		At: time.Now(),
	}

	// Publish event data
	eventData := map[string]interface{}{
		"type": event.EventType(),
		"data": map[string]interface{}{
			"Telegram": event.Telegram,
			"At":       event.At,
		},
		"occurred_at": event.OccurredAt().Format(time.RFC3339),
	}

	eventBytes, err := json.Marshal(eventData)
	require.NoError(t, err)

	err = client.Publish(ctx, testChannel, eventBytes).Err()
	require.NoError(t, err)

	// Wait for handler to be called
	select {
	case <-handlerCalled:
		require.NotNil(t, receivedEvent)
		assert.Equal(t, "telegram.persisted", receivedEvent.EventType())
	case <-time.After(2 * time.Second):
		t.Fatal("handler was not called within timeout")
	}

	// Cancel context to stop subscription
	cancel()

	// Wait for subscription to stop
	select {
	case err := <-errCh:
		// Context cancellation is expected
		if err != nil && err != context.Canceled {
			require.NoError(t, err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("subscription did not stop in time")
	}
}

// TestRedisSubscriber_Subscribe_InvalidMessage tests handling of invalid messages
func TestRedisSubscriber_Subscribe_InvalidMessage(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Pub/Sub subscriber test")
	}

	logger := zaptest.NewLogger(t)
	subscriber := NewRedisSubscriber(client, logger)

	testChannel := "test:subscriber:invalid"
	handlerCalled := false

	handler := func(ctx context.Context, event app.Event) error {
		handlerCalled = true
		return nil
	}

	// Start subscription in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- subscriber.Subscribe(ctx, testChannel, handler)
	}()

	// Wait a bit for subscription to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish invalid message (not JSON)
	err := client.Publish(ctx, testChannel, "invalid json").Err()
	require.NoError(t, err)

	// Wait a bit - handler should not be called for invalid messages
	time.Sleep(500 * time.Millisecond)
	assert.False(t, handlerCalled, "handler should not be called for invalid messages")

	// Cancel context
	cancel()

	// Wait for subscription to stop
	select {
	case <-errCh:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("subscription did not stop in time")
	}
}

// TestRedisSubscriber_Subscribe_Error tests error handling
func TestRedisSubscriber_Subscribe_Error(t *testing.T) {
	// Use invalid Redis connection
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid address
	})
	logger := zaptest.NewLogger(t)
	subscriber := NewRedisSubscriber(client, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	handler := func(ctx context.Context, event app.Event) error {
		return nil
	}

	err := subscriber.Subscribe(ctx, "test:channel", handler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscribe")
}

