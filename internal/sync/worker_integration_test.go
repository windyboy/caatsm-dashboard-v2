//go:build integration

package sync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/application"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	pgstore "github.com/windy/caatsm-dashboard/internal/repository/postgres"
	testhelpers "github.com/windy/caatsm-dashboard/internal/testing"
	"go.uber.org/zap/zaptest"
)

func TestWorker_Integration_SaveTelegram(t *testing.T) {
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
	// For integration test, we'll use a nil index (or create a mock index)
	telegramService := application.NewTelegramService(store, nil, eventBus, logger)

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
}

func TestWorker_Integration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Setup test environment
	env, err := testhelpers.SetupTestEnv(ctx)
	require.NoError(t, err)
	defer env.Cleanup(ctx)

	// Test invalid telegram (should return ErrBadData)
	invalidTelegram := &domain.Telegram{
		MessageID: "", // Invalid: empty message_id
		Type:      "aftn",
		Time:      time.Now(),
		Priority:  2,
	}

	store := pgstore.New(env.Pool)
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")
	logger := zaptest.NewLogger(t)
	telegramService := application.NewTelegramService(store, nil, eventBus, logger)

	err = telegramService.SaveTelegram(ctx, invalidTelegram)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validate telegram")
}

