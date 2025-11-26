package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// AppConfig contains all configuration knobs for the CAATSM Dashboard.
type AppConfig struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	Database    DatabaseConfig `mapstructure:"database"`
	Meilisearch SearchConfig   `mapstructure:"meilisearch"`
	Redis       RedisConfig    `mapstructure:"redis"`
	NATS        NATSConfig     `mapstructure:"nats"`
	Metrics     MetricsConfig  `mapstructure:"metrics"`
	Auth        AuthConfig     `mapstructure:"auth"`
	Tracing     TracingConfig  `mapstructure:"tracing"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLSEnabled   bool          `mapstructure:"tls_enabled"`
	TLSCertFile  string        `mapstructure:"tls_cert_file"`
	TLSKeyFile   string        `mapstructure:"tls_key_file"`
}

type LoggerConfig struct {
	Level            string `mapstructure:"level"`
	Format           string `mapstructure:"format"`
	EnableStacktrace bool   `mapstructure:"enable_stacktrace"`
	RedactPII        bool   `mapstructure:"redact_pii"`
}

type TracingConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	ServiceName   string  `mapstructure:"service_name"`
	OTLPEndpoint  string  `mapstructure:"otlp_endpoint"`
	SamplingRatio float64 `mapstructure:"sampling_ratio"`
}

type DatabaseConfig struct {
	DSN                   string        `mapstructure:"dsn"`
	MaxOpenConnections    int           `mapstructure:"max_open_connections"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
}

type SearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
	Index  string `mapstructure:"index"`
}

type RedisConfig struct {
	Addr     string        `mapstructure:"addr"`
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TLS      bool          `mapstructure:"tls"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type NATSConfig struct {
	URL            string        `mapstructure:"url"`
	Stream         string        `mapstructure:"stream"`
	Consumer       string        `mapstructure:"consumer"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type AuthConfig struct {
	EnableBasic bool   `mapstructure:"enable_basic"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	JWTSecret   string `mapstructure:"jwt_secret"`
}

// Validate performs sanity checks on configuration values before bootstrap.
func (c *AppConfig) Validate() error {
	var errs []error

	if c.Server.Port <= 0 {
		errs = append(errs, errors.New("server.port must be greater than 0"))
	}

	if c.Database.DSN == "" {
		errs = append(errs, errors.New("database.dsn is required"))
	}

	if c.Meilisearch.Host == "" {
		errs = append(errs, errors.New("meilisearch.host is required"))
	}
	if c.Meilisearch.APIKey == "" {
		errs = append(errs, errors.New("meilisearch.api_key is required"))
	}

	if c.NATS.URL == "" {
		errs = append(errs, errors.New("nats.url is required"))
	}

	if c.Metrics.Enabled && c.Metrics.Path == "" {
		errs = append(errs, errors.New("metrics.path is required when metrics.enabled is true"))
	}

	if c.Auth.EnableBasic {
		if c.Auth.Username == "" || c.Auth.Password == "" {
			errs = append(errs, errors.New("auth username/password required when basic auth enabled"))
		}
	}

	// Production-specific validations
	if c.Environment == "production" {
		// Authentication must be enabled in production
		if !c.Auth.EnableBasic && c.Auth.JWTSecret == "" {
			errs = append(errs, errors.New("production: authentication must be enabled (basic or JWT)"))
		}

		// Meilisearch API key cannot be default value
		if c.Meilisearch.APIKey == "masterKey" {
			errs = append(errs, errors.New("production: meilisearch.api_key cannot be default 'masterKey'"))
		}

		// Database DSN cannot contain default credentials
		if strings.Contains(c.Database.DSN, "caatsm:caatsm@") {
			errs = append(errs, errors.New("production: database cannot use default credentials 'caatsm:caatsm'"))
		}

		// JWT secret must be changed from default
		if c.Auth.JWTSecret == "dev-secret-change-in-production" {
			errs = append(errs, errors.New("production: jwt_secret must be changed from default value"))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("config validation failed: %w", errors.Join(errs...))
}
