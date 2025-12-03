package app

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/windy/caatsm-dashboard/config"
	"go.uber.org/zap"
)

// NewPostgresPool initialises a pgx connection pool using the provided configuration.
func NewPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	if cfg.MaxOpenConnections > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		poolCfg.MinConns = int32(cfg.MaxIdleConnections)
	}
	if cfg.ConnectionMaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.ConnectionMaxLifetime
	} else {
		poolCfg.MaxConnLifetime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}

// Container wires together dependencies for the simplified architecture
type Container struct {
	Config *config.AppConfig
	Logger *zap.Logger

	// Core infrastructure ports
	Repo              Repository
	Cache             Cache
	Search            SearchIndex
	Pub               EventPublisher
	EventBus          EventBus
	EventSub          EventSubscriber
	StreamPub         StreamPublisher
	RedisStreamCons   RedisStreamConsumer
	WebSocketHub     WebSocketHubPort

	// Application services
	DashboardService interface{}

	// Client references for health checks (simplified)
	pool     *pgxpool.Pool
	meiliSvc meilisearchClient.ServiceManager
	redisCli redis.UniversalClient
	natsConn interface{} // *nats.Conn - using interface{} to avoid import cycle
	
	// Service metadata
	startTime time.Time
	version   string
}

// New builds a Container with all dependencies.
// NOTE: This uses two-phase initialization to avoid circular imports:
// 1. New() creates an empty container with nil infrastructure clients
// 2. server.New() creates infrastructure and calls SetInfrastructureClients()
// The container is not usable until SetInfrastructureClients() is called.
func New(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Database pool will be set by server layer
	var pool *pgxpool.Pool

	// NOTE: Infrastructure creation moved to server layer to avoid circular imports.
	// Infrastructure clients are set to nil here and will be initialized by server.New()
	// via SetInfrastructureClients() to avoid import cycles.
	container.Repo = nil
	container.Cache = nil
	container.Search = nil
	container.Pub = nil
	container.EventBus = nil
	container.EventSub = nil
	container.StreamPub = nil
	container.RedisStreamCons = nil
	container.WebSocketHub = nil

	// Initialize application services with nil dependencies (will fail at runtime)
	// Services will be initialized by the server layer to avoid import cycles
	container.DashboardService = nil

	// Client references will be set by server layer
	container.pool = pool
	container.meiliSvc = nil
	container.redisCli = nil
	container.natsConn = nil
	container.startTime = time.Now()
	container.version = "dev" // Can be set via build flags or config

	// Validate that infrastructure is nil (expected state)
	if container.Repo != nil || container.Cache != nil || container.Search != nil {
		logger.Warn("Container initialized with non-nil infrastructure - this is unexpected",
			zap.Bool("repo_valid", container.Repo != nil),
			zap.Bool("cache_valid", container.Cache != nil),
			zap.Bool("search_valid", container.Search != nil),
		)
	}

	logger.Info("Container initialized successfully",
		zap.Bool("repo_valid", container.Repo != nil),
		zap.Bool("cache_valid", container.Cache != nil),
		zap.Bool("search_valid", container.Search != nil),
		zap.Bool("dashboard_valid", container.DashboardService != nil),
	)

	return container, nil
}

// ComponentHealth represents the health status of a component.
type ComponentHealth struct {
	Status      string        `json:"status"`                // "ok", "error", "not_configured"
	Message     string        `json:"message,omitempty"`      // Error message if status is "error"
	ResponseTime string       `json:"response_time,omitempty"` // Response time in milliseconds
	Details     interface{}   `json:"details,omitempty"`       // Additional component-specific details
}

// DatabasePoolStats contains database connection pool statistics.
type DatabasePoolStats struct {
	TotalConnections     int32 `json:"total_connections"`
	AcquiredConnections  int32 `json:"acquired_connections"`
	IdleConnections      int32 `json:"idle_connections"`
	MaxConnections       int32 `json:"max_connections"`
	ConstructingConns    int32 `json:"constructing_connections"`
}

// HealthCheckResult contains the health status of all components.
type HealthCheckResult struct {
	Status      string          `json:"status"`       // "healthy", "degraded", "unhealthy"
	Version     string          `json:"version,omitempty"`      // Service version
	Uptime      string          `json:"uptime,omitempty"`      // Service uptime in seconds
	Timestamp   string          `json:"timestamp"`    // RFC3339 timestamp
	PostgreSQL  ComponentHealth `json:"postgresql"`
	Meilisearch ComponentHealth `json:"meilisearch"`
	Redis       ComponentHealth `json:"redis"`
	NATS        ComponentHealth `json:"nats"`
}

// HealthCheck verifies connectivity to all external components.
func (c *Container) HealthCheck(ctx context.Context) HealthCheckResult {
	return c.HealthCheckInternal(ctx)
}

// HealthCheckInternal is the internal implementation that can be called directly.
func (c *Container) HealthCheckInternal(ctx context.Context) HealthCheckResult {
	result := HealthCheckResult{
		Status:    "healthy",
		Version:   c.version,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	// Calculate uptime
	if !c.startTime.IsZero() {
		uptime := time.Since(c.startTime)
		result.Uptime = fmt.Sprintf("%.0f", uptime.Seconds())
	}

	errorCount := 0

	// PostgreSQL health check with nil guard
	if c.pool == nil {
		result.PostgreSQL = ComponentHealth{
			Status:  "error",
			Message: "database pool not initialized",
		}
		errorCount++
		if result.Status == "healthy" {
			result.Status = "degraded"
		}
	} else {
		start := time.Now()
		pgCtx, pgCancel := context.WithTimeout(ctx, 2*time.Second)
		defer pgCancel()
		if err := c.pool.Ping(pgCtx); err != nil {
			result.PostgreSQL = ComponentHealth{
				Status:  "error",
				Message: err.Error(),
			}
			errorCount++
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
		} else {
			responseTime := time.Since(start)
			stats := c.pool.Stat()
			result.PostgreSQL = ComponentHealth{
				Status:       "ok",
				ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
				Details: DatabasePoolStats{
					TotalConnections:    stats.TotalConns(),
					AcquiredConnections: stats.AcquiredConns(),
					IdleConnections:     stats.IdleConns(),
					MaxConnections:      stats.MaxConns(),
					ConstructingConns:   stats.ConstructingConns(),
				},
			}
		}
	}

	// Meilisearch health check with actual API call
	if c.meiliSvc == nil {
		result.Meilisearch = ComponentHealth{Status: "not_configured"}
	} else {
		start := time.Now()
		
		// Perform actual health check via API - try to get version info
		// This is a lightweight operation that verifies connectivity
		health, err := c.meiliSvc.Health()
		responseTime := time.Since(start)
		
		if err != nil {
			result.Meilisearch = ComponentHealth{
				Status:       "error",
				Message:      err.Error(),
				ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
			}
			errorCount++
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
		} else if health.Status != "available" {
			result.Meilisearch = ComponentHealth{
				Status:       "error",
				Message:      fmt.Sprintf("meilisearch status: %s", health.Status),
				ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
			}
			errorCount++
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
		} else {
			result.Meilisearch = ComponentHealth{
				Status:       "ok",
				ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
			}
		}
	}

	// Redis health check with nil guard and actual ping
	if c.redisCli == nil {
		result.Redis = ComponentHealth{Status: "not_configured"}
	} else {
		start := time.Now()
		redisCtx, redisCancel := context.WithTimeout(ctx, 2*time.Second)
		defer redisCancel()
		if err := c.redisCli.Ping(redisCtx).Err(); err != nil {
			result.Redis = ComponentHealth{
				Status:       "error",
				Message:      err.Error(),
				ResponseTime: fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds())/1e6),
			}
			errorCount++
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
		} else {
			responseTime := time.Since(start)
			result.Redis = ComponentHealth{
				Status:       "ok",
				ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
			}
		}
	}

	if c.natsConn == nil {
		result.NATS = ComponentHealth{Status: "not_configured"}
	} else {
		start := time.Now()
		
		connValue := reflect.ValueOf(c.natsConn)
		// Handle pointer types - methods are on the pointer, not the value
		if connValue.Kind() == reflect.Ptr {
			if connValue.IsNil() {
				result.NATS = ComponentHealth{
					Status:  "error",
					Message: "NATS connection is nil",
				}
				errorCount++
				if result.Status == "healthy" {
					result.Status = "degraded"
				}
			} else {
				// Don't call Elem() - methods are on the pointer type (*nats.Conn), not the value type
				// Keep connValue as the pointer to access methods
			}
		}
		
		isClosedMethod := connValue.MethodByName("IsClosed")
		isConnectedMethod := connValue.MethodByName("IsConnected")
		statsMethod := connValue.MethodByName("Stats")
		
		if !isClosedMethod.IsValid() || !isConnectedMethod.IsValid() {
			result.NATS = ComponentHealth{
				Status:       "error",
				Message:      fmt.Sprintf("NATS connection type check failed (type: %s)", connValue.Type().String()),
				ResponseTime: fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds())/1e6),
			}
			errorCount++
			if result.Status == "healthy" {
				result.Status = "degraded"
			}
		} else {
			responseTime := time.Since(start)
			
			isClosed := isClosedMethod.Call(nil)[0].Bool()
			isConnected := isConnectedMethod.Call(nil)[0].Bool()
			
			if isClosed {
				result.NATS = ComponentHealth{
					Status:       "error",
					Message:      "NATS connection is closed",
					ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
				}
				errorCount++
				if result.Status == "healthy" {
					result.Status = "degraded"
				}
			} else if !isConnected {
				result.NATS = ComponentHealth{
					Status:       "error",
					Message:      "NATS connection is not connected",
					ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
				}
				errorCount++
				if result.Status == "healthy" {
					result.Status = "degraded"
				}
			} else {
				details := make(map[string]interface{})
				if statsMethod.IsValid() {
					statsValue := statsMethod.Call(nil)[0]
					if inMsgsMethod := statsValue.MethodByName("InMsgs"); inMsgsMethod.IsValid() {
						details["in_msgs"] = inMsgsMethod.Call(nil)[0].Uint()
					}
					if outMsgsMethod := statsValue.MethodByName("OutMsgs"); outMsgsMethod.IsValid() {
						details["out_msgs"] = outMsgsMethod.Call(nil)[0].Uint()
					}
				}
				
				result.NATS = ComponentHealth{
					Status:       "ok",
					ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1e6),
					Details:      details,
				}
			}
		}
	}

	// Final status determination
	if errorCount > 0 && result.Status == "healthy" {
		result.Status = "degraded"
	}
	if errorCount > 1 {
		result.Status = "unhealthy"
	}

	return result
}

// RedisClient returns the Redis client for external use.
func (c *Container) RedisClient() redis.UniversalClient {
	return c.redisCli
}

// IsReady returns true if the container has been fully initialized with infrastructure clients.
func (c *Container) IsReady() bool {
	return c.Repo != nil && c.Cache != nil && c.Search != nil && c.DashboardService != nil
}

// MustBeReady panics if the container is not ready for use.
// Call this at the start of methods that require infrastructure.
func (c *Container) MustBeReady() {
	if !c.IsReady() {
		panic("Container not initialized: infrastructure clients must be set via SetInfrastructureClients()")
	}
}

// SetInfrastructureClients sets the infrastructure client references for health checks.
// This should only be called after all infrastructure resources are successfully initialized.
func (c *Container) SetInfrastructureClients(pool *pgxpool.Pool, meiliSvc meilisearchClient.ServiceManager, redisCli redis.UniversalClient) {
	c.pool = pool
	c.meiliSvc = meiliSvc
	c.redisCli = redisCli
}

// SetNATSConnection sets the NATS connection for health checks.
func (c *Container) SetNATSConnection(natsConn interface{}) {
	c.natsConn = natsConn
}

// SetVersion sets the service version.
func (c *Container) SetVersion(version string) {
	c.version = version
}

// Close releases all container-managed resources.
func (c *Container) Close() error {
	var firstErr error

	// Close database pool
	if c.pool != nil {
		c.Logger.Info("closing database connection pool")
		c.pool.Close()
		c.Logger.Info("database connection pool closed")
	}

	// Close Redis client disabled (redisCli is nil)

	// Note: Meilisearch client doesn't have an explicit Close() method

	return firstErr
}
