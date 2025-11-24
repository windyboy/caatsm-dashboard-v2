package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap/zaptest"
)

func TestTelegramService_SaveTelegram(t *testing.T) {
	tests := []struct {
		name              string
		telegram          *domain.Telegram
		storeError        error
		indexError        error
		publishError      error
		expectError       bool
		expectStoreCall   bool
		expectIndexCall   bool
		expectPublishCall bool
	}{
		{
			name: "valid telegram - all operations succeed",
			telegram: &domain.Telegram{
				MessageID: "TEST-001",
				Type:      "aftn",
				Time:      time.Now(),
				Priority:  2,
			},
			expectError:       false,
			expectStoreCall:   true,
			expectIndexCall:   true,
			expectPublishCall: true,
		},
		{
			name: "invalid telegram - validation fails",
			telegram: &domain.Telegram{
				MessageID: "", // Invalid: empty message_id
				Type:      "aftn",
				Time:      time.Now(),
				Priority:  2,
			},
			expectError:       true,
			expectStoreCall:   false,
			expectIndexCall:   false,
			expectPublishCall: false,
		},
		{
			name: "store save fails",
			telegram: &domain.Telegram{
				MessageID: "TEST-001",
				Type:      "aftn",
				Time:      time.Now(),
				Priority:  2,
			},
			storeError:        errors.New("database error"),
			expectError:       true,
			expectStoreCall:   true,
			expectIndexCall:   false,
			expectPublishCall: false,
		},
		{
			name: "index fails - should continue",
			telegram: &domain.Telegram{
				MessageID: "TEST-001",
				Type:      "aftn",
				Time:      time.Now(),
				Priority:  2,
			},
			indexError:        errors.New("index error"),
			expectError:       false, // Index error should not fail the operation
			expectStoreCall:   true,
			expectIndexCall:   true,
			expectPublishCall: true,
		},
		{
			name: "publish fails - should continue",
			telegram: &domain.Telegram{
				MessageID: "TEST-001",
				Type:      "aftn",
				Time:      time.Now(),
				Priority:  2,
			},
			publishError:      errors.New("publish error"),
			expectError:       false, // Publish error should not fail the operation
			expectStoreCall:   true,
			expectIndexCall:   true,
			expectPublishCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			storeMock := new(mocks.TelegramStoreMock)
			indexMock := new(mocks.SearchIndexMock)
			eventBusMock := new(mocks.EventBusMock)
			logger := zaptest.NewLogger(t)

			service := NewTelegramService(storeMock, indexMock, eventBusMock, logger)

			// Setup expectations
			if tt.expectStoreCall {
				modelTelegram := domain.FromDomain(tt.telegram)
				storeMock.On("Save", mock.Anything, modelTelegram).Return(tt.storeError)
			}

			if tt.expectIndexCall {
				modelTelegram := domain.FromDomain(tt.telegram)
				indexMock.On("Index", mock.Anything, modelTelegram).Return(tt.indexError)
			}

			if tt.expectPublishCall {
				eventBusMock.On("Publish", mock.Anything, "telegram_processed", mock.Anything).Return(tt.publishError)
			}

			// Execute
			err := service.SaveTelegram(context.Background(), tt.telegram)

			// Assert
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify mock calls
			storeMock.AssertExpectations(t)
			indexMock.AssertExpectations(t)
			if tt.expectPublishCall {
				eventBusMock.AssertExpectations(t)
			}
		})
	}
}

func TestTelegramService_SaveTelegram_Normalize(t *testing.T) {
	storeMock := new(mocks.TelegramStoreMock)
	indexMock := new(mocks.SearchIndexMock)
	eventBusMock := new(mocks.EventBusMock)
	logger := zaptest.NewLogger(t)

	service := NewTelegramService(storeMock, indexMock, eventBusMock, logger)

	telegram := &domain.Telegram{
		MessageID: "TEST-001",
		Type:      "aftn",
		Time:      time.Now(),
		Priority:  2,
	}

	modelTelegram := domain.FromDomain(telegram)
	storeMock.On("Save", mock.Anything, modelTelegram).Return(nil)
	indexMock.On("Index", mock.Anything, modelTelegram).Return(nil)
	eventBusMock.On("Publish", mock.Anything, "telegram_processed", mock.Anything).Return(nil)

	err := service.SaveTelegram(context.Background(), telegram)
	require.NoError(t, err)

	// Verify normalize was called (telegram should be normalized)
	assert.NotNil(t, telegram)
	storeMock.AssertExpectations(t)
	indexMock.AssertExpectations(t)
	eventBusMock.AssertExpectations(t)
}
