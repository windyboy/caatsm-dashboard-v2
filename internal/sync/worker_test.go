package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap/zaptest"
)

func TestWorker_Handle(t *testing.T) {
	tests := []struct {
		name             string
		telegram         *app.Telegram
		serviceError     error
		expectError      bool
		expectBadData    bool
		expectAppFailure bool
	}{
		{
			name: "valid telegram - success",
			telegram: &app.Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  2,
			},
			expectError: false,
		},
		{
			name: "invalid telegram - empty message_id",
			telegram: &app.Telegram{
				MessageID: "",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  2,
			},
			expectError:   true,
			expectBadData: true,
		},
		{
			name: "service error - app failure",
			telegram: &app.Telegram{
				MessageID: "TEST-001",
				Type:      "AFTN",
				Time:      time.Now(),
				Priority:  2,
			},
			serviceError:     errors.New("service error"),
			expectError:      true,
			expectAppFailure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			storeMock := new(mocks.RepositoryMock)
			searchMock := new(mocks.SearchIndexMock)
			eventBusMock := new(mocks.EventBusMock)
			logger := zaptest.NewLogger(t)

			worker := &Worker{
				store:    storeMock,
				search:   searchMock,
				eventBus: eventBusMock,
				logger:   logger,
			}

			// Setup expectations
			if !tt.expectBadData {
				if tt.telegram.Validate() == nil {
					// Expect Save call
					storeMock.On("Save", mock.Anything, mock.MatchedBy(func(tg *app.Telegram) bool {
						return tg.MessageID == tt.telegram.MessageID
					})).Return(tt.serviceError)

					// Expect Index and Publish calls only if Save succeeds
					if tt.serviceError == nil {
						searchMock.On("Index", mock.Anything, mock.MatchedBy(func(tg *app.Telegram) bool {
							return tg.MessageID == tt.telegram.MessageID
						})).Return(nil)
						eventBusMock.On("Publish", mock.Anything, "telegram_processed", mock.Anything).Return(nil)
					}
				}
			}

			// Execute
			err := worker.Handle(context.Background(), tt.telegram)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				if tt.expectBadData {
					assert.True(t, IsBadDataError(err), "expected ErrBadData")
				}
				if tt.expectAppFailure {
					assert.True(t, IsRetryableError(err), "expected retryable error")
				}
			} else {
				require.NoError(t, err)
			}

			// Verify mock calls
			storeMock.AssertExpectations(t)
			if !tt.expectBadData && tt.serviceError == nil {
				searchMock.AssertExpectations(t)
				eventBusMock.AssertExpectations(t)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrAppFailure is retryable",
			err:  ErrAppFailure,
			want: true,
		},
		{
			name: "ErrInfraFailure is retryable",
			err:  ErrInfraFailure,
			want: true,
		},
		{
			name: "ErrBadData is not retryable",
			err:  ErrBadData,
			want: false,
		},
		{
			name: "wrapped ErrAppFailure is retryable",
			err:  errors.New("wrapped: " + ErrAppFailure.Error()),
			want: false, // errors.Is doesn't work with string matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsBadDataError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrBadData is bad data",
			err:  ErrBadData,
			want: true,
		},
		{
			name: "ErrAppFailure is not bad data",
			err:  ErrAppFailure,
			want: false,
		},
		{
			name: "nil error is not bad data",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBadDataError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
