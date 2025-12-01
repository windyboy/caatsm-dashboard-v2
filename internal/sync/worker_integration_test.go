//go:build integration

package sync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/event"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/search"
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
	store := persistence.NewPostgresStore(env.Pool)
	searchIndex := search.NewMeilisearchIndex(env.Meili, "telegrams-test")

	// Initialize event bus
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")

	// Create worker
	logger := zaptest.NewLogger(t)
	worker := NewWorker(nil, store, searchIndex, eventBus, logger)

	// Create test telegram
	telegram := testhelpers.NewTelegram("INTEGRATION-TEST-001")

	// Execute
	err = worker.Handle(ctx, telegram)
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
	invalidTelegram := &app.Telegram{
		MessageID: "", // Invalid: empty message_id
		Type:      "AFTN",
		Time:      time.Now(),
		Priority:  2,
	}

	store := persistence.NewPostgresStore(env.Pool)
	searchIndex := search.NewMeilisearchIndex(env.Meili, "telegrams-test")
	eventBus := event.NewRedisEventBus(env.Redis.Client(), "stats:update")
	logger := zaptest.NewLogger(t)
	worker := NewWorker(nil, store, searchIndex, eventBus, logger)

	err = worker.Handle(ctx, invalidTelegram)
	require.Error(t, err)
	require.True(t, IsBadDataError(err), "expected ErrBadData for invalid telegram")
}
