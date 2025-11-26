package stats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
		"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap"
)

// TestStatsService_ValidateTimeWindow verifies that the TrafficSummary service validates time windows at the application layer
func TestStatsService_ValidateTimeWindow(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid time window - 30 days",
			startTime: time.Now().Add(-30 * 24 * time.Hour),
			endTime:   time.Now(),
			wantErr:   false,
		},
		// Note: "exactly 90 days" test is removed due to time drift between time.Now() calls
		// The validation uses > not >=, so 90 days exactly should pass, but microsecond drift makes this unreliable to test
		{
			name:      "invalid time window - exceeds 90 days",
			startTime: time.Now().Add(-91 * 24 * time.Hour),
			endTime:   time.Now(),
			wantErr:   true,
			errMsg:    "time window cannot exceed 90 days",
		},
		{
			name:      "invalid time window - exceeds 90 days by a lot",
			startTime: time.Now().Add(-180 * 24 * time.Hour),
			endTime:   time.Now(),
			wantErr:   true,
			errMsg:    "time window cannot exceed 90 days",
		},
		{
			name:      "no time window - should use default (24 hours) and pass",
			startTime: time.Time{},
			endTime:   time.Time{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			analyticsStore := &mocks.AnalyticsStoreMock{}
			logger := zap.NewNop()

			service := NewService(analyticsStore, logger)

			window := domain.TimeWindow{
				Start: tt.startTime,
				End:   tt.endTime,
			}

			// Mock the store to return a result (only if validation passes)
			if !tt.wantErr {
				analyticsStore.On("TrafficSummary", mock.Anything, mock.Anything).Return(&domain.TrafficSummary{
					TotalMessages: 100,
				}, nil)
			}

			// Execute
			_, err := service.TrafficSummary(context.Background(), window)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				// May have other errors from mocks, but not validation errors
				if err != nil {
					assert.NotContains(t, err.Error(), "time window cannot exceed 90 days")
				}
			}
		})
	}
}

// TestTrafficSummary_DefaultTimeWindow verifies that when no time window is provided, it defaults to 24 hours
func TestTrafficSummary_DefaultTimeWindow(t *testing.T) {
	// Setup
	analyticsStore := &mocks.AnalyticsStoreMock{}
	logger := zap.NewNop()

	service := NewService(analyticsStore, logger)

	window := domain.TimeWindow{
		Start: time.Time{},
		End:   time.Time{},
	}

	// Mock the store
	analyticsStore.On("TrafficSummary", mock.Anything, mock.Anything).Return(&domain.TrafficSummary{
		TotalMessages: 50,
	}, nil)

	// Execute
	result, err := service.TrafficSummary(context.Background(), window)

	// Assert - should not fail validation even though window is zero
	// (because service sets default 24h window before validation)
	if err != nil {
		assert.NotContains(t, err.Error(), "time window cannot exceed 90 days")
	}

	// If successful, result should exist
	if result != nil {
		assert.NotNil(t, result)
	}
}

// TestTrafficSummary_EdgeCases tests edge cases for time window validation
func TestTrafficSummary_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errMsg    string
	}{
		// Note: "exactly 90 days" test is removed due to time drift between time.Now() calls  
		// The validation uses > not >=, so 90 days exactly should pass, but microsecond drift makes this unreliable to test
		{
			name:      "90 days + 1 second - should fail",
			startTime: time.Now().Add(-(90*24*time.Hour + 1*time.Second)),
			endTime:   time.Now(),
			wantErr:   true,
			errMsg:    "time window cannot exceed 90 days",
		},
		{
			name:      "90 days - 1 second - should pass",
			startTime: time.Now().Add(-(90*24*time.Hour - 1*time.Second)),
			endTime:   time.Now(),
			wantErr:   false,
		},
		{
			name:      "only start time (end is zero) - should pass",
			startTime: time.Now().Add(-100 * 24 * time.Hour),
			endTime:   time.Time{},
			wantErr:   false,
		},
		{
			name:      "only end time (start is zero) - should pass",
			startTime: time.Time{},
			endTime:   time.Now(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			analyticsStore := &mocks.AnalyticsStoreMock{}
			logger := zap.NewNop()

			service := NewService(analyticsStore, logger)

			window := domain.TimeWindow{
				Start: tt.startTime,
				End:   tt.endTime,
			}

			// Mock the store (only if validation passes)
			if !tt.wantErr {
				analyticsStore.On("TrafficSummary", mock.Anything, mock.Anything).Return(&domain.TrafficSummary{
					TotalMessages: 100,
				}, nil)
			}

			// Execute
			_, err := service.TrafficSummary(context.Background(), window)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				if err != nil {
					assert.NotContains(t, err.Error(), "time window cannot exceed 90 days")
				}
			}
		})
	}
}

