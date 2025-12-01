package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"

	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app"
)

// NewValkeyClient creates a Valkey/Redis client based on configuration.
func NewValkeyClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
	options := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeAuto,
		},
	}

	if cfg.Timeout > 0 {
		options.DialTimeout = cfg.Timeout
		options.ReadTimeout = cfg.Timeout
		options.WriteTimeout = cfg.Timeout
	} else {
		options.DialTimeout = 5 * time.Second
		options.ReadTimeout = 5 * time.Second
		options.WriteTimeout = 5 * time.Second
	}

	if cfg.TLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping valkey/redis: %w", err)
	}

	return client, nil
}

// Store provides caching utilities using Valkey/Redis.
type Store struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewValkeyStore returns a new cache store with the given ttl.
func NewValkeyStore(client redis.UniversalClient, ttl time.Duration) *Store {
	return &Store{client: client, ttl: ttl}
}

// Ensure Store implements app.Cache
var _ app.Cache = (*Store)(nil)

// Set stores a value for a given key.
func (s *Store) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, s.ttl).Err()
}

// Get retrieves a JSON-encoded value from the cache and returns it as a decoded interface{}.
// It returns nil when the key is not present.
func (s *Store) Get(ctx context.Context, key string) (interface{}, error) {

	data, err := s.client.Get(ctx, key).Bytes()

	if err != nil {

		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cache key %q: %w", key, err)

	}

	if len(data) == 0 {
		return nil, nil
	}

	var value interface{}

	if err := json.Unmarshal(data, &value); err != nil {

		return nil, fmt.Errorf("decode cache key %q: %w", key, err)
	}

	return value, nil
}

// Delete removes a key from the cache.
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// DeleteMultiple removes multiple keys from the cache atomically.
func (s *Store) DeleteMultiple(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

// Incr increments the integer value stored at key by delta. Creates key if not exists.
func (s *Store) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	result, err := s.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, err
	}
	// Ensure key has TTL (only sets if no TTL exists)
	if err := s.client.Expire(ctx, key, s.ttl).Err(); err != nil {
		// Log or return error based on desired semantics
		return result, fmt.Errorf("set ttl for key %q: %w", key, err)
	}
	return result, nil
}

// HIncrBy increments the integer value stored at the specified hash field by delta.
func (s *Store) HIncrBy(ctx context.Context, key string, field string, delta int64) (int64, error) {
	result, err := s.client.HIncrBy(ctx, key, field, delta).Result()
	if err != nil {
		return 0, err
	}
	// Ensure key has TTL (only sets if no TTL exists)
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return result, err
	}
	if ttl == -1 {
		if err := s.client.Expire(ctx, key, s.ttl).Err(); err != nil {
			return result, err
		}
	}
	return result, nil
}

// GetInt64 fetches the string value stored at key and parses it into an int64.
// A missing key results in a zero value without an error.
func (s *Store) GetInt64(ctx context.Context, key string) (int64, error) {

	str, err := s.client.Get(ctx, key).Result()

	if err != nil {

		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, fmt.Errorf("get int cache key %q: %w", key, err)

	}

	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("parse cache key %q: %w", key, err)
	}
	return i, nil
}

// HGetAll returns all field/value pairs for the given hash key as a map.
// It yields an empty map when the hash does not exist.
func (s *Store) HGetAll(ctx context.Context, key string) (map[string]string, error) {

	result, err := s.client.HGetAll(ctx, key).Result()

	if err != nil {
		return nil, fmt.Errorf("hgetall cache key %q: %w", key, err)
	}
	if len(result) == 0 {
		return map[string]string{}, nil
	}
	return result, nil
}
