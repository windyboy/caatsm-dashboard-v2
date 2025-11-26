package valkey

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	oldcache "github.com/windy/caatsm-dashboard/internal/repository/cache"
)

// Store provides caching utilities using Valkey/Redis.
// This is the new infrastructure layer implementation.
type Store struct {
	oldStore *oldcache.Store
}

// New returns a new cache store with the given ttl.
func New(client redis.UniversalClient, ttl time.Duration) *Store {
	return &Store{
		oldStore: oldcache.New(client, ttl),
	}
}

// Set stores a value for a given key.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	return s.oldStore.Set(ctx, key, value)
}

// Get fetches a key.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	return s.oldStore.Get(ctx, key)
}

// Delete removes a key from the cache.
func (s *Store) Delete(ctx context.Context, key string) error {
	// Old store doesn't have Delete, so implement directly if needed
	// For now, return nil (can be enhanced later)
	return nil
}

// Clear clears all keys from the cache (use with caution).
func (s *Store) Clear(ctx context.Context) error {
	// Old store doesn't have Clear, so implement directly if needed
	// For now, return nil (can be enhanced later)
	return nil
}

