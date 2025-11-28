package valkey

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
)

// NewClient creates a Valkey/Redis client based on configuration.
func NewClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
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

// New returns a new cache store with the given ttl.
func New(client redis.UniversalClient, ttl time.Duration) *Store {
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
