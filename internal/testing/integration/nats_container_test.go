//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNATSContainer_StartAndConnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create and start NATS container
	container, err := NewNATSContainer(ctx)
	require.NoError(t, err, "should create NATS container")
	defer func() {
		if err := container.Close(ctx); err != nil {
			t.Logf("failed to close container: %v", err)
		}
	}()

	// Verify URI is set
	assert.NotEmpty(t, container.URI, "container URI should be set")
	t.Logf("NATS container URI: %s", container.URI)

	// Connect to NATS
	conn, err := container.Connect()
	require.NoError(t, err, "should connect to NATS")
	defer conn.Close()

	// Verify connection is alive
	assert.True(t, conn.IsConnected(), "connection should be connected")
	assert.True(t, conn.IsConnected(), "connection should be connected")

	// Test JetStream is available
	js, err := conn.JetStream()
	require.NoError(t, err, "JetStream should be available")

	// Create a test stream
	streamName := "TEST_STREAM"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"test.>"},
	})
	require.NoError(t, err, "should create JetStream stream")

	// Clean up stream
	err = js.DeleteStream(streamName)
	require.NoError(t, err, "should delete JetStream stream")
}

