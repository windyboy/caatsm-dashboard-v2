package cache

import (
	"context"
	"encoding/json"
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
func (s *Store) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, s.ttl).Err()
}

// Get fetches a key.
func (s *Store) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var value interface{}
	err = json.Unmarshal(data, &value)
	return value, err
}

// Delete removes a key from the cache.
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
