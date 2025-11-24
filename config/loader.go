package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Load constructs AppConfig by merging configuration file and environment variables.
// When cfgPath is empty the loader searches for config.toml in ./config.
// Environment variables are loaded from .env.local (if exists) and .env files.
func Load(cfgPath string) (*AppConfig, error) {
	// Load .env.local first (local overrides), then .env (shared defaults)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()

	v := viper.New()
	v.SetEnvPrefix("CAATSM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
	} else {
		v.AddConfigPath("./config")
		v.SetConfigName("config")
		v.SetConfigType("toml")
	}

	if err := v.ReadInConfig(); err != nil {
		// Ignore missing config file and only rely on env vars/defaults.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := new(AppConfig)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad wraps Load and panics when configuration can't be constructed.
func MustLoad(cfgPath string) *AppConfig {
	cfg, err := Load(cfgPath)
	if err != nil {
		panic(err)
	}
	return cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 3002)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.tls_enabled", false)
	v.SetDefault("server.tls_cert_file", "")
	v.SetDefault("server.tls_key_file", "")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.enable_stacktrace", false)

	v.SetDefault("database.dsn", "postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable")
	v.SetDefault("database.max_open_connections", 25)
	v.SetDefault("database.max_idle_connections", 10)
	v.SetDefault("database.connection_max_lifetime", "30m")

	v.SetDefault("meilisearch.host", "http://localhost:7700")
	v.SetDefault("meilisearch.index", "telegrams")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.username", "")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.tls", false)
	v.SetDefault("redis.timeout", "5s")

	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.stream", "telegrams")
	v.SetDefault("nats.consumer", "dashboard-sync")
	v.SetDefault("nats.connect_timeout", "5s")

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")

	v.SetDefault("auth.enable_basic", false)
	v.SetDefault("auth.username", "")
	v.SetDefault("auth.password", "")
	v.SetDefault("auth.jwt_secret", "")
}
