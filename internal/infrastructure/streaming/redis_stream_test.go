package streaming

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"go.uber.org/zap/zaptest"
)

// TestRedisStreamPublisher_Publish tests the Redis Streams publisher
func TestRedisStreamPublisher_Publish(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Streams publisher test")
	}

	logger := zaptest.NewLogger(t)
	publisher := NewRedisStreamPublisher(client, logger)

	// Clean up test stream
	testStream := "test:stream:publish"
	_ = client.Del(ctx, testStream).Err()

	telegram := &app.Telegram{
		MessageID:    "TEST-STREAM-001",
		Type:         "AFTN",
		Time:         time.Now(),
		FlightNumber: "CA100",
		Source:       "HND",
		Destination:  "SFO",
		Priority:     1,
		Content:      "Test message for stream",
	}

	// Publish message
	err := publisher.Publish(ctx, testStream, telegram)
	require.NoError(t, err)

	// Verify message was published
	messages, err := client.XRange(ctx, testStream, "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, messages, 1)

	// Verify message content
	msg := messages[0]
	telegramData, ok := msg.Values["telegram"].(string)
	require.True(t, ok, "telegram field should be a string")

	var publishedTelegram app.Telegram
	err = json.Unmarshal([]byte(telegramData), &publishedTelegram)
	require.NoError(t, err)
	assert.Equal(t, telegram.MessageID, publishedTelegram.MessageID)
	assert.Equal(t, telegram.Type, publishedTelegram.Type)
	assert.Equal(t, telegram.FlightNumber, publishedTelegram.FlightNumber)

	// Clean up
	_ = client.Del(ctx, testStream).Err()
}

// TestRedisStreamPublisher_Publish_Error tests error handling
func TestRedisStreamPublisher_Publish_Error(t *testing.T) {
	// Use invalid Redis connection to test error handling
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid address
	})
	logger := zaptest.NewLogger(t)
	publisher := NewRedisStreamPublisher(client, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	telegram := &app.Telegram{
		MessageID: "TEST-001",
		Type:      "AFTN",
		Time:      time.Now(),
		Priority:  1,
	}

	err := publisher.Publish(ctx, "test:stream", telegram)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xadd")
}

// TestRedisStreamConsumer_SetHandler tests handler setting
func TestRedisStreamConsumer_SetHandler(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Streams consumer test")
	}

	logger := zaptest.NewLogger(t)
	consumer := NewRedisStreamConsumer(
		client,
		"test:stream:consumer",
		"test-group",
		"test-consumer",
		logger,
	)

	handler := func(ctx context.Context, telegram *app.Telegram) error {
		return nil
	}

	consumer.SetHandler(handler)
	// Handler is set, we can't directly verify it, but we can check it doesn't panic
	assert.NotNil(t, consumer)
}

// TestRedisStreamConsumer_Start_NoHandler tests that Start fails without handler
func TestRedisStreamConsumer_Start_NoHandler(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Streams consumer test")
	}

	logger := zaptest.NewLogger(t)
	consumer := NewRedisStreamConsumer(
		client,
		"test:stream:consumer",
		"test-group",
		"test-consumer",
		logger,
	)

	// Don't set handler - should fail
	err := consumer.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler not set")
}

// TestRedisStreamConsumer_Close tests consumer close
func TestRedisStreamConsumer_Close(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping Redis Streams consumer test")
	}

	logger := zaptest.NewLogger(t)
	consumer := NewRedisStreamConsumer(
		client,
		"test:stream:close",
		"test-group",
		"test-consumer",
		logger,
	)

	// Close should not error even if not started
	err := consumer.Close()
	require.NoError(t, err)
}

