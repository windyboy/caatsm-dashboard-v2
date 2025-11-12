package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store provides light caching utilities for search results and sessions.
type Store struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// New returns a new cache store with the given ttl.
func New(client redis.UniversalClient, ttl time.Duration) *Store {
	return &Store{client: client, ttl: ttl}
}

// Set stores a value for a given key.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	return s.client.Set(ctx, key, value, s.ttl).Err()
}

// Get fetches a key.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	return s.client.Get(ctx, key).Bytes()
}
