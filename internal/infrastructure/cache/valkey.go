package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
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

// Ensure Store implements ports.Cache
var _ ports.Cache = (*Store)(nil)

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

// Incr increments the integer value stored at key by delta. Creates key if not exists.
func (s *Store) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	return s.client.IncrBy(ctx, key, delta).Result()
}

// HIncrBy increments the integer value stored at the specified hash field by delta.
func (s *Store) HIncrBy(ctx context.Context, key string, field string, delta int64) (int64, error) {
	return s.client.HIncrBy(ctx, key, field, delta).Result()
}

// GetInt64 retrieves the value at key parsed as int64.
func (s *Store) GetInt64(ctx context.Context, key string) (int64, error) {
	str, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err
}

// HGetAll retrieves all fields and values from the hash stored at key.
func (s *Store) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.client.HGetAll(ctx, key).Result()
}
