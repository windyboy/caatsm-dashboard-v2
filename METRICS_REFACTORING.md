# WebSocket Hub Metrics Refactoring

## Overview

Refactored the `HubMetrics` struct to use a dedicated Prometheus registry per instance instead of the global registry, preventing duplicate registration panics when creating multiple hub instances.

## Changes Made

### 1. Core Metrics Refactoring (`internal/infrastructure/ws/metrics.go`)

**Before:**
- Used `prometheus.MustRegister()` with the global registry
- Could panic if metrics were registered more than once (e.g., in tests or multiple hub instances)

**After:**
- Each `HubMetrics` instance now has its own dedicated `prometheus.Registry`
- Uses `registry.MustRegister()` on the instance-specific registry
- Added `Registry()` method to expose the registry for HTTP exposition
- Prevents duplicate registration panics completely

**Key Changes:**
```go
type HubMetrics struct {
    // ... existing fields ...
    
    // Dedicated registry to avoid duplicate registration panics
    registry *prometheus.Registry
    
    enabled bool
}

func NewHubMetrics(enabled bool) *HubMetrics {
    if enabled {
        // Create a dedicated registry for this hub instance
        m.registry = prometheus.NewRegistry()
        
        // ... create metrics ...
        
        // Register on the dedicated registry instead of the global registry
        m.registry.MustRegister(
            m.messagesSent,
            m.messagesBroadcast,
            m.messagesDropped,
            m.connectionRejected,
        )
    }
    return m
}

// Registry returns the dedicated Prometheus registry for this hub's metrics.
func (m *HubMetrics) Registry() *prometheus.Registry {
    return m.registry
}
```

### 2. Domain Type Converters (`internal/domain/converter.go`)

Added conversion functions between persistence and domain types to fix type mismatches in the transport layer:

- `SearchFilterToDomain()` - Converts `persistence.SearchFilter` to `domain.SearchFilter`
- `TimeWindowToDomain()` - Converts `persistence.TimeWindow` to `domain.TimeWindow`
- `ExportFormatToDomain()` - Converts `persistence.ExportFormat` to `domain.ExportFormat`

### 3. Transport Layer Fixes

Fixed type conversion issues in HTTP and WebSocket handlers:
- `internal/transport/http/handlers.go` - Added domain type conversions for search, stats, and export operations
- `internal/transport/ws/handler.go` - Fixed syntax errors and method calls
- `internal/transport/ws/handler_simple.go` - Fixed TimeWindow type mismatches

### 4. Test Coverage (`internal/infrastructure/ws/metrics_test.go`)

Added comprehensive tests to verify:
- ✅ Dedicated registry is created per instance
- ✅ Registry isolation (multiple instances have separate registries)
- ✅ Metrics can be gathered from the registry
- ✅ Metrics work correctly when enabled
- ✅ Metrics are disabled when requested
- ✅ No duplicate registration panics

## Benefits

1. **No More Registration Panics**: Multiple hub instances can coexist without conflicting metric registrations
2. **Testability**: Tests can create multiple hub instances without cleanup between tests
3. **Isolation**: Each hub's metrics are isolated, allowing for better monitoring of individual instances
4. **Flexibility**: Metrics can be exposed via different HTTP endpoints using `promhttp.HandlerFor(metrics.Registry(), ...)`
5. **Follows Best Practices**: Aligns with the existing `MetricsExporter` pattern used in `internal/observability/prometheus.go`

## Usage

### Creating Hub Metrics

```go
// Enable metrics with dedicated registry
metrics := ws.NewHubMetrics(true)

// Access the registry for HTTP exposition
registry := metrics.Registry()
handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
```

### Multiple Instances

```go
// Safe to create multiple instances without panics
hub1Metrics := ws.NewHubMetrics(true)
hub2Metrics := ws.NewHubMetrics(true)

// Each has its own registry
registry1 := hub1Metrics.Registry()
registry2 := hub2Metrics.Registry()
```

## Architecture Compliance

This refactoring maintains clean architecture principles:
- ✅ Infrastructure layer components are isolated
- ✅ No dependencies on global state
- ✅ Follows dependency injection pattern
- ✅ Consistent with existing observability patterns

## Testing

Run the metrics tests:
```bash
go test ./internal/infrastructure/ws/... -v -run TestHubMetrics
go test ./internal/infrastructure/ws/... -v -run TestNewHubMetrics
```

All tests pass successfully.

