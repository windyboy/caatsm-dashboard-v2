//go:build integration

package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	pgstore "github.com/windy/caatsm-dashboard/internal/repository/postgres"
	testhelpers "github.com/windy/caatsm-dashboard/internal/testing"
	"go.uber.org/zap/zaptest"
)

func TestTelegramService_Integration_SaveTelegram(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment
	env, err := testhelpers.SetupTestEnv(ctx)
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	// Initialize repositories
	store := pgstore.New(env.Pool)

	// Initialize event bus
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")

	// Initialize application service
	logger := zaptest.NewLogger(t)
	telegramService := NewTelegramService(store, nil, eventBus, logger)

	// Create test telegram
	telegram := testhelpers.NewTelegram("INTEGRATION-TEST-001")
	domainTelegram := domain.ToDomain(telegram)

	// Execute
	err = telegramService.SaveTelegram(ctx, domainTelegram)
	require.NoError(t, err)

	// Verify: Check that telegram was saved to database
	result, err := store.Search(ctx, testhelpers.NewSearchFilter())
	require.NoError(t, err)
	require.Greater(t, result.Total, int64(0), "telegram should be saved to database")

	// Verify the saved telegram
	found := false
	for _, tg := range result.Telegrams {
		if tg.MessageID == "INTEGRATION-TEST-001" {
			found = true
			require.Equal(t, telegram.Type, tg.Type)
			require.Equal(t, telegram.FlightNumber, tg.FlightNumber)
			break
		}
	}
	require.True(t, found, "telegram should be found in search results")
}

func TestTelegramService_Integration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment
	env, err := testhelpers.SetupTestEnv(ctx)
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	store := pgstore.New(env.Pool)
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")
	logger := zaptest.NewLogger(t)
	telegramService := NewTelegramService(store, nil, eventBus, logger)

	// Test invalid telegram (should return validation error)
	invalidTelegram := &domain.Telegram{
		MessageID: "", // Invalid: empty message_id
		Type:      "AFTN",
		Time:      testhelpers.NewTimeWindowLast24h().Start, // Use time window start as test time
		Priority:  2,
	}

	err = telegramService.SaveTelegram(ctx, invalidTelegram)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validate telegram")
}

func TestTelegramService_Integration_MultipleTelegrams(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment
	env, err := testhelpers.SetupTestEnv(ctx)
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	// Clean database before test
	err = testhelpers.CleanDatabase(ctx, env.Pool)
	require.NoError(t, err)

	store := pgstore.New(env.Pool)
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")
	logger := zaptest.NewLogger(t)
	telegramService := NewTelegramService(store, nil, eventBus, logger)

	// Save multiple telegrams
	telegrams := testhelpers.NewTelegrams(5)
	for _, tg := range telegrams {
		domainTelegram := domain.ToDomain(tg)
		err := telegramService.SaveTelegram(ctx, domainTelegram)
		require.NoError(t, err)
	}

	// Verify all telegrams were saved
	result, err := store.Search(ctx, testhelpers.NewSearchFilter())
	require.NoError(t, err)
	require.Equal(t, int64(5), result.Total, "all 5 telegrams should be saved")
}
