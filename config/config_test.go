package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppConfig_Validate_BasicValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  AppConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid development config",
			config: AppConfig{
				Environment: "development",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://user:pass@localhost:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "test-key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: AppConfig{
				Server: ServerConfig{
					Port: 0,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
			},
			wantErr: true,
			errMsg:  "server.port must be greater than 0",
		},
		{
			name: "missing database DSN",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
			},
			wantErr: true,
			errMsg:  "database.dsn is required",
		},
		{
			name: "missing meilisearch host",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
			},
			wantErr: true,
			errMsg:  "meilisearch.host is required",
		},
		{
			name: "missing meilisearch API key",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
			},
			wantErr: true,
			errMsg:  "meilisearch.api_key is required",
		},
		{
			name: "missing NATS URL",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "",
				},
			},
			wantErr: true,
			errMsg:  "nats.url is required",
		},
		{
			name: "metrics enabled without path",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Metrics: MetricsConfig{
					Enabled: true,
					Path:    "",
				},
			},
			wantErr: true,
			errMsg:  "metrics.path is required",
		},
		{
			name: "basic auth enabled without credentials",
			config: AppConfig{
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://localhost",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "key",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "",
					Password:    "",
				},
			},
			wantErr: true,
			errMsg:  "auth username/password required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_Validate_ProductionDefaults(t *testing.T) {
	tests := []struct {
		name    string
		config  AppConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "production with proper auth",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:secure_pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-production-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure-password",
				},
			},
			wantErr: false,
		},
		{
			name: "production without auth",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: false,
					JWTSecret:   "",
				},
			},
			wantErr: true,
			errMsg:  "production: authentication must be enabled",
		},
		{
			name: "production with default meilisearch key",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "masterKey",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "password",
				},
			},
			wantErr: true,
			errMsg:  "production: meilisearch.api_key cannot be default 'masterKey'",
		},
		{
			name: "production with default database credentials",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://caatsm:caatsm@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "password",
				},
			},
			wantErr: true,
			errMsg:  "production: database cannot use default credentials",
		},
		{
			name: "production with default JWT secret",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: false,
					JWTSecret:   "dev-secret-change-in-production",
				},
			},
			wantErr: true,
			errMsg:  "production: jwt_secret must be changed from default",
		},
		{
			name: "production with JWT auth",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:secure_pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-production-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: false,
					JWTSecret:   "secure-jwt-secret-production",
				},
			},
			wantErr: false,
		},
		{
			name: "development with default credentials (should pass)",
			config: AppConfig{
				Environment: "development",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					DSN: "postgres://caatsm:caatsm@localhost:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://localhost:7700",
					APIKey: "masterKey",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Auth: AuthConfig{
					EnableBasic: false,
					JWTSecret:   "dev-secret-change-in-production",
				},
			},
			wantErr: false,
		},
		{
			name: "production with URL-encoded password containing default username (should pass)",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					// Different user with password containing "caatsm" - should pass
					DSN: "postgres://produser:pass%40caatsm@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: false,
		},
		{
			name: "production with key-value DSN format with default credentials",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port: 3000,
				},
				Database: DatabaseConfig{
					// Key-value format with default credentials - should fail
					DSN: "host=prod-db port=5432 user=caatsm password=caatsm dbname=db sslmode=disable",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: database cannot use default credentials",
		},
		{
			name: "production with different username, same password (should pass)",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://produser:caatsm@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: false,
		},
		{
			name: "production with same username, different password (should pass)",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://caatsm:securepass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: false,
		},
		{
			name: "production without TLS",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: false,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: TLS must be enabled",
		},
		{
			name: "production with localhost server host",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					Host:       "localhost",
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: server host cannot be localhost",
		},
		{
			name: "production with 0.0.0.0 server host",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					Host:       "0.0.0.0",
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Logger: LoggerConfig{
					Level: "info",
				},
				Tracing: TracingConfig{
					Enabled: true,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: server host cannot be 0.0.0.0 unless server.allow_bind_all is set to true",
		},
		{
			name: "production with default Redis password",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: redis password must be set",
		},
		{
			name: "production with dev Redis password",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "dev-password",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: redis password must be set and not default",
		},
		{
			name: "production with default NATS URL",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: nats.url cannot be default localhost",
		},
		{
			name: "production with debug logger level",
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Logger: LoggerConfig{
					Level: "debug",
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: true,
			errMsg:  "production: logger level cannot be 'debug'",
		},
		{
			name: "production without tracing enabled",
			// NOTE: Tracing validation is currently disabled because tracing instrumentation
			// is not yet implemented. See config.go Validate() method comments.
			// Once tracing is implemented, this test should be updated to expect an error.
			config: AppConfig{
				Environment: "production",
				Server: ServerConfig{
					Port:       3000,
					TLSEnabled: true,
				},
				Database: DatabaseConfig{
					DSN: "postgres://prod_user:pass@prod-db:5432/db",
				},
				Meilisearch: SearchConfig{
					Host:   "http://meilisearch:7700",
					APIKey: "secure-key",
				},
				Redis: RedisConfig{
					Addr:     "redis:6379",
					Password: "secure-redis-pass",
				},
				NATS: NATSConfig{
					URL: "nats://nats:4222",
				},
				Logger: LoggerConfig{
					Level: "info",
				},
				Tracing: TracingConfig{
					Enabled: false,
				},
				Auth: AuthConfig{
					EnableBasic: true,
					Username:    "admin",
					Password:    "secure",
				},
			},
			wantErr: false, // Tracing validation is currently disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerConfig(t *testing.T) {
	cfg := ServerConfig{
		Host:         "0.0.0.0",
		Port:         3000,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSEnabled:   false,
	}

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	assert.False(t, cfg.TLSEnabled)
}
