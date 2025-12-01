package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_DeleteMultiple(t *testing.T) {
	// Create a mock Redis client for testing
	// In a real test, you'd use a test Redis instance or a mock
	// For now, we'll test the logic with a real client if available, or skip
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection - skip test if Redis is not available
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping DeleteMultiple test")
	}

	store := NewValkeyStore(client, 5*time.Minute)

	// Clean up any existing test keys
	_ = client.Del(ctx, "test:key1", "test:key2", "test:key3").Err()

	// Set some test keys
	err := store.Set(ctx, "test:key1", "value1")
	require.NoError(t, err)
	err = store.Set(ctx, "test:key2", "value2")
	require.NoError(t, err)
	err = store.Set(ctx, "test:key3", "value3")
	require.NoError(t, err)

	// Verify keys exist
	val1, err := store.Get(ctx, "test:key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val1)

	// Delete multiple keys
	err = store.DeleteMultiple(ctx, "test:key1", "test:key2", "test:key3")
	require.NoError(t, err)

	// Verify keys are deleted
	val1, err = store.Get(ctx, "test:key1")
	require.NoError(t, err)
	assert.Nil(t, val1)

	val2, err := store.Get(ctx, "test:key2")
	require.NoError(t, err)
	assert.Nil(t, val2)

	val3, err := store.Get(ctx, "test:key3")
	require.NoError(t, err)
	assert.Nil(t, val3)

	// Test with empty keys slice
	err = store.DeleteMultiple(ctx)
	require.NoError(t, err)

	// Test with non-existent keys (should not error)
	err = store.DeleteMultiple(ctx, "test:nonexistent1", "test:nonexistent2")
	require.NoError(t, err)

	// Clean up
	_ = client.Del(ctx, "test:key1", "test:key2", "test:key3").Err()
}

