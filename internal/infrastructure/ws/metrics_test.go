package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHubMetrics(t *testing.T) {
	t.Run("metrics enabled creates dedicated registry", func(t *testing.T) {
		metrics := NewHubMetrics(true)
		
		require.NotNil(t, metrics)
		require.NotNil(t, metrics.registry, "should create a dedicated registry")
		require.NotNil(t, metrics.messagesSent, "should create messagesSent counter")
		require.NotNil(t, metrics.messagesBroadcast, "should create messagesBroadcast counter")
		require.NotNil(t, metrics.messagesDropped, "should create messagesDropped counter")
		require.NotNil(t, metrics.connectionRejected, "should create connectionRejected counter")
		
		// Verify Registry() method returns the registry
		registry := metrics.Registry()
		assert.Equal(t, metrics.registry, registry, "Registry() should return the internal registry")
	})
	
	t.Run("metrics disabled does not create registry", func(t *testing.T) {
		metrics := NewHubMetrics(false)
		
		require.NotNil(t, metrics)
		assert.Nil(t, metrics.registry, "should not create registry when disabled")
		assert.Nil(t, metrics.Registry(), "Registry() should return nil when disabled")
	})
	
	t.Run("multiple instances have separate registries", func(t *testing.T) {
		metrics1 := NewHubMetrics(true)
		metrics2 := NewHubMetrics(true)
		
		require.NotNil(t, metrics1.registry)
		require.NotNil(t, metrics2.registry)
		
		// Verify they are different registry instances
		assert.NotEqual(t, metrics1.registry, metrics2.registry, 
			"each hub should have its own dedicated registry to avoid duplicate registration panics")
	})
	
	t.Run("metrics can be gathered from registry", func(t *testing.T) {
		metrics := NewHubMetrics(true)
		
		// Increment some metrics
		metrics.activeConnectionsInc()
		metrics.messagesSentInc()
		metrics.totalConnectionsInc()
		
		// Verify we can gather metrics from the registry
		metricFamilies, err := metrics.registry.Gather()
		require.NoError(t, err)
		require.NotEmpty(t, metricFamilies, "registry should contain registered metrics")
		
		// Verify the expected metrics are present
		metricNames := make(map[string]bool)
		for _, mf := range metricFamilies {
			metricNames[mf.GetName()] = true
		}
		
		assert.True(t, metricNames["websocket_messages_sent_total"], "should have messagesSent metric")
		assert.True(t, metricNames["websocket_messages_broadcast_total"], "should have messagesBroadcast metric")
		assert.True(t, metricNames["websocket_messages_dropped_total"], "should have messagesDropped metric")
		assert.True(t, metricNames["websocket_connections_rejected_total"], "should have connectionRejected metric")
	})
}

func TestHubMetrics_Counters(t *testing.T) {
	t.Run("active connections counter", func(t *testing.T) {
		metrics := NewHubMetrics(true)
		
		assert.Equal(t, int64(0), metrics.ActiveConnections())
		
		metrics.activeConnectionsInc()
		assert.Equal(t, int64(1), metrics.ActiveConnections())
		
		metrics.activeConnectionsInc()
		assert.Equal(t, int64(2), metrics.ActiveConnections())
		
		metrics.activeConnectionsDec()
		assert.Equal(t, int64(1), metrics.ActiveConnections())
	})
	
	t.Run("total connections counter", func(t *testing.T) {
		metrics := NewHubMetrics(true)
		
		assert.Equal(t, int64(0), metrics.TotalConnections())
		
		metrics.totalConnectionsInc()
		assert.Equal(t, int64(1), metrics.TotalConnections())
		
		metrics.totalConnectionsInc()
		assert.Equal(t, int64(2), metrics.TotalConnections())
	})
	
	t.Run("disabled metrics return zero", func(t *testing.T) {
		metrics := NewHubMetrics(false)
		
		metrics.activeConnectionsInc()
		metrics.totalConnectionsInc()
		
		assert.Equal(t, int64(0), metrics.ActiveConnections())
		assert.Equal(t, int64(0), metrics.TotalConnections())
	})
}

