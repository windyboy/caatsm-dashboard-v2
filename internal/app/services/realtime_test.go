package services

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRealtimeManager_InitialState(t *testing.T) {
	mgr := NewRealtimeManager()

	info := mgr.GetInfo()

	assert.NotNil(t, info)
	assert.Equal(t, 0, info.ActiveConnections)
	assert.Equal(t, 0.0, info.MessagesPerSecond)
	assert.NotEmpty(t, info.Uptime)
}

func TestRealtimeManager_UpdateConnections(t *testing.T) {
	mgr := NewRealtimeManager()

	tests := []struct {
		name  string
		count int
	}{
		{"set to 10", 10},
		{"set to 0", 0},
		{"set to 1000", 1000},
		{"set to negative", -1}, // Should still work, though semantically odd
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr.UpdateConnections(tt.count)
			info := mgr.GetInfo()
			assert.Equal(t, tt.count, info.ActiveConnections)
		})
	}
}

func TestRealtimeManager_UpdateMessageRate(t *testing.T) {
	mgr := NewRealtimeManager()

	tests := []struct {
		name string
		rate float64
	}{
		{"zero rate", 0.0},
		{"normal rate", 5.5},
		{"high rate", 1000.5},
		{"fractional rate", 0.123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr.UpdateMessageRate(tt.rate)
			info := mgr.GetInfo()
			assert.Equal(t, tt.rate, info.MessagesPerSecond)
		})
	}
}

func TestRealtimeManager_UptimeIncreases(t *testing.T) {
	mgr := NewRealtimeManager()

	info1 := mgr.GetInfo()
	uptime1 := info1.Uptime

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	info2 := mgr.GetInfo()
	uptime2 := info2.Uptime

	// Uptime strings should be different
	assert.NotEqual(t, uptime1, uptime2)
}

func TestRealtimeManager_ConcurrentWrites(t *testing.T) {
	mgr := NewRealtimeManager()
	iterations := 100

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Update connections
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			mgr.UpdateConnections(i)
		}
	}()

	// Goroutine 2: Update message rate
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			mgr.UpdateMessageRate(float64(i) * 1.5)
		}
	}()

	wg.Wait()

	// Should complete without panics or data races
	info := mgr.GetInfo()
	assert.NotNil(t, info)
}

func TestRealtimeManager_ConcurrentReadsAndWrites(t *testing.T) {
	mgr := NewRealtimeManager()
	iterations := 100

	var wg sync.WaitGroup
	wg.Add(3)

	// Channel to collect read results for validation
	type readResult struct {
		info   *RealtimeInfo
		uptime string
	}
	resultCh := make(chan readResult, iterations)

	// Goroutine 1: Update connections
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			mgr.UpdateConnections(i)
		}
	}()

	// Goroutine 2: Update message rate
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			mgr.UpdateMessageRate(float64(i))
		}
	}()

	// Goroutine 3: Concurrent reads
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			info := mgr.GetInfo()
			if info == nil {
				resultCh <- readResult{info: nil, uptime: ""}
				continue
			}
			resultCh <- readResult{info: info, uptime: info.Uptime}
		}
		close(resultCh)
	}()

	wg.Wait()

	// Validate all read results on main goroutine
	readCount := 0
	for result := range resultCh {
		assert.NotNil(t, result.info)
		assert.NotEmpty(t, result.uptime)
		readCount++
	}
	assert.Equal(t, iterations, readCount, "should have collected all read results")

	// Final read should work
	finalInfo := mgr.GetInfo()
	assert.NotNil(t, finalInfo)
}

func TestRealtimeManager_MultipleReaders(t *testing.T) {
	mgr := NewRealtimeManager()
	mgr.UpdateConnections(42)
	mgr.UpdateMessageRate(10.5)

	numReaders := 10
	readsPerReader := 100
	var wg sync.WaitGroup
	wg.Add(numReaders)

	// Multiple concurrent readers
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < readsPerReader; j++ {
				info := mgr.GetInfo()
				if info == nil {
					t.Errorf("reader %d iteration %d: got nil info", readerID, j)
					return
				}
				if info.ActiveConnections < 0 || info.MessagesPerSecond < 0.0 {
					t.Errorf("reader %d iteration %d: got negative values (connections=%d, rate=%.2f)",
						readerID, j, info.ActiveConnections, info.MessagesPerSecond)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestRealtimeManager_SequentialUpdates(t *testing.T) {
	mgr := NewRealtimeManager()

	// Update connections
	mgr.UpdateConnections(10)
	info := mgr.GetInfo()
	assert.Equal(t, 10, info.ActiveConnections)
	assert.Equal(t, 0.0, info.MessagesPerSecond) // Should be unchanged

	// Update message rate
	mgr.UpdateMessageRate(5.5)
	info = mgr.GetInfo()
	assert.Equal(t, 10, info.ActiveConnections) // Should be unchanged
	assert.Equal(t, 5.5, info.MessagesPerSecond)

	// Update connections again
	mgr.UpdateConnections(20)
	info = mgr.GetInfo()
	assert.Equal(t, 20, info.ActiveConnections)
	assert.Equal(t, 5.5, info.MessagesPerSecond) // Should be unchanged

	// Update message rate again
	mgr.UpdateMessageRate(15.0)
	info = mgr.GetInfo()
	assert.Equal(t, 20, info.ActiveConnections) // Should be unchanged
	assert.Equal(t, 15.0, info.MessagesPerSecond)
}
