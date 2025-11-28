package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/windy/caatsm-dashboard/config"
)

// NewClient creates a Valkey/Redis client based on configuration.
// Uses github.com/redis/go-redis/v9 which is protocol-compatible with both Redis and Valkey.
func NewClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
	options := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		// Use auto mode for maintenance notifications: enables server-driven
		// maintenance notifications when supported, gracefully falls back when not.
		// This preserves graceful failover behavior while maintaining compatibility.
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
