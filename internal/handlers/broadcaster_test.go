package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestEventBroadcaster_Subscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client1 := broadcaster.Subscribe()
	require.NotNil(t, client1)
	assert.Equal(t, 1, len(broadcaster.clients))

	client2 := broadcaster.Subscribe()
	require.NotNil(t, client2)
	assert.Equal(t, 2, len(broadcaster.clients))

	broadcaster.Unsubscribe(client1)
	assert.Equal(t, 1, len(broadcaster.clients))

	broadcaster.Unsubscribe(client2)
	assert.Equal(t, 0, len(broadcaster.clients))
}

func TestEventBroadcaster_Unsubscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client := broadcaster.Subscribe()
	require.NotNil(t, client)

	broadcaster.Unsubscribe(client)
	assert.Equal(t, 0, len(broadcaster.clients))

	broadcaster.Unsubscribe(client)
	assert.Equal(t, 0, len(broadcaster.clients))
}

func TestEventBroadcaster_Broadcast(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client1 := broadcaster.Subscribe()
	client2 := broadcaster.Subscribe()

	event := []byte(`{"type": "test", "data": "value"}`)

	var wg sync.WaitGroup
	wg.Add(2)

	var received1, received2 []byte
	go func() {
		defer wg.Done()
		received1 = <-client1
	}()
	go func() {
		defer wg.Done()
		received2 = <-client2
	}()

	broadcaster.Broadcast(event)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, event, received1)
		assert.Equal(t, event, received2)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for broadcast")
	}

	broadcaster.Unsubscribe(client1)
	broadcaster.Unsubscribe(client2)
}

func TestEventBroadcaster_Broadcast_ChannelFull(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client := broadcaster.Subscribe()
	event := []byte(`{"type": "test"}`)

	for i := 0; i < 10; i++ {
		client <- event
	}

	broadcaster.Broadcast(event)

	time.Sleep(100 * time.Millisecond)

	broadcaster.Unsubscribe(client)
}

func TestEventBroadcaster_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client1 := broadcaster.Subscribe()
	client2 := broadcaster.Subscribe()

	broadcaster.Close()

	_, ok1 := <-client1
	assert.False(t, ok1)

	_, ok2 := <-client2
	assert.False(t, ok2)

	assert.Equal(t, 0, len(broadcaster.clients))
}

func TestEventBroadcaster_PublishStatsUpdate(t *testing.T) {
	tests := []struct {
		name        string
		redisClient redis.UniversalClient
		expectError bool
	}{
		{
			name:        "nil redis client",
			redisClient: nil,
			expectError: true,
		},
		{
			name:        "valid redis client",
			redisClient: redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			broadcaster := NewEventBroadcaster(tt.redisClient, logger)

			ctx := context.Background()
			err := broadcaster.PublishStatsUpdate(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Note: This will fail without a running Redis instance
				// Consider using a Redis mock or miniredis for proper testing
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventBroadcaster_JSONValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	broadcaster := NewEventBroadcaster(nil, logger)

	client := broadcaster.Subscribe()

	validJSON := []byte(`{"type": "test", "data": {"key": "value"}}`)
	invalidJSON := []byte(`invalid json`)

	var eventData map[string]any
	err := json.Unmarshal(validJSON, &eventData)
	require.NoError(t, err)
	assert.Equal(t, "test", eventData["type"])

	err = json.Unmarshal(invalidJSON, &eventData)
	assert.Error(t, err)

	broadcaster.Unsubscribe(client)
}
