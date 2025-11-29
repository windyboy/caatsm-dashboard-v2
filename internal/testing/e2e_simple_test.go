//go:build integration

package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/testing/integration"
)

func TestE2E_BasicPipelineWithContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := integration.NewPostgresContainer(ctx)
	require.NoError(t, err, "failed to start postgres container")
	defer pgContainer.Close(ctx)

	pool, err := pgContainer.Pool(ctx)
	require.NoError(t, err, "failed to create postgres pool")
	defer pool.Close()

	// Create schema
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS telegrams (
			message_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			time TIMESTAMPTZ NOT NULL,
			flight_number TEXT NOT NULL,
			source TEXT NOT NULL,
			destination TEXT NOT NULL,
			priority INTEGER NOT NULL,
			content TEXT,
			raw_data TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "failed to create schema")

	// Initialize repository
	telegramStore := persistence.NewPostgresStore(pool)

	// Create test telegram
	telegram := domain.Telegram{
		MessageID:    "E2E-SIMPLE-001",
		Type:         "AFTN",
		Time:         time.Now(),
		FlightNumber: "CA1111",
		Source:       "ZBAA",
		Destination:  "ZSPD",
		Priority:     2,
		Content:      "E2E simple test content",
		RawData:      "RAW DATA",
	}

	// Save telegram
	err = telegramStore.SaveTelegram(ctx, telegram)
	require.NoError(t, err, "failed to save telegram")

	t.Log("Telegram saved successfully")

	// Query back
	filter := persistence.SearchFilter{
		Page: persistence.Pagination{
			Limit:  10,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	result, err := telegramStore.Search(ctx, filter)
	require.NoError(t, err, "failed to search telegrams")
	require.Greater(t, len(result.Telegrams), 0, "should have at least one telegram")

	// Verify the telegram
	found := false
	for _, tg := range result.Telegrams {
		if tg.MessageID == "E2E-SIMPLE-001" {
			found = true
			assert.Equal(t, "AFTN", tg.Type)
			assert.Equal(t, "CA1111", tg.FlightNumber)
			assert.Equal(t, "ZBAA", tg.Source)
			assert.Equal(t, "ZSPD", tg.Destination)
			assert.Equal(t, 2, tg.Priority)
			break
		}
	}
	assert.True(t, found, "telegram should be found in search results")

	t.Log("Telegram retrieved successfully")

	// Test stats
	statsResult, err := telegramStore.TrafficSummary(ctx, persistence.TimeWindow{})
	require.NoError(t, err, "stats should work")
	assert.Greater(t, statsResult.TotalMessages, int64(0), "stats should show at least one message")

	t.Log("E2E basic pipeline test passed")
}

func TestE2E_LargeDatasetHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := integration.NewPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Close(ctx)

	pool, err := pgContainer.Pool(ctx)
	require.NoError(t, err)
	defer pool.Close()

	// Create schema
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS telegrams (
			message_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			time TIMESTAMPTZ NOT NULL,
			flight_number TEXT NOT NULL,
			source TEXT NOT NULL,
			destination TEXT NOT NULL,
			priority INTEGER NOT NULL,
			content TEXT,
			raw_data TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	telegramStore := persistence.NewPostgresStore(pool)

	// Create 100 test telegrams
	count := 100
	for i := 0; i < count; i++ {
		telegram := NewTelegram(fmt.Sprintf("LOAD-TEST-%03d", i))
		err := telegramStore.SaveTelegram(ctx, telegram)
		require.NoError(t, err, "failed to save telegram %d", i)
	}

	t.Logf("Created %d test telegrams", count)

	// Query with pagination
	filter := persistence.SearchFilter{
		Page: persistence.Pagination{
			Limit:  50,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	result, err := telegramStore.Search(ctx, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Telegrams), 50, "should return at least 50 telegrams")
	assert.GreaterOrEqual(t, result.Total, int64(count), "total should be at least count")

	t.Log("Large dataset handling test passed")
}

func NewTelegram(messageID string) domain.Telegram {
	return domain.Telegram{
		MessageID:    messageID,
		Type:         "AFTN",
		Time:         time.Now(),
		FlightNumber: "CA0000",
		Source:       "ZBAA",
		Destination:  "ZSPD",
		Priority:     1,
		Content:      "Test telegram content",
		RawData:      "RAW",
	}
}
