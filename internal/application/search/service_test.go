package search

import (
	"context"
	"testing"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
	"go.uber.org/zap"
)

// TestSearchService_ValidateTimeRange verifies that the Search service validates time ranges at the application layer
func TestSearchService_ValidateTimeRange(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "invalid time range - exceeds 90 days",
			startTime: time.Now().Add(-91 * 24 * time.Hour),
			endTime:   time.Now(),
			wantErr:   true,
			errMsg:    "time range cannot exceed 90 days",
		},
		{
			name:      "invalid time range - exceeds 90 days by a lot",
			startTime: time.Now().Add(-180 * 24 * time.Hour),
			endTime:   time.Now(),
			wantErr:   true,
			errMsg:    "time range cannot exceed 90 days",
		},
		{
			name:      "invalid time range - end before start",
			startTime: time.Now(),
			endTime:   time.Now().Add(-1 * time.Hour),
			wantErr:   true,
			errMsg:    "end time must be after start time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			searchIndex := &mocks.SearchIndexMock{}
			telegramStore := &mocks.TelegramStoreMock{}
			logger := zap.NewNop()

			// Pass nil for cache since we're testing validation, not caching
			service := NewService(searchIndex, telegramStore, nil, logger)

			filter := domain.SearchFilter{
				Query: "test",
				TimeRange: domain.TimeWindow{
					Start: tt.startTime,
					End:   tt.endTime,
				},
				Page: domain.Pagination{
					Limit:  50,
					Offset: 0,
				},
			}

			// Execute
			_, err := service.Search(context.Background(), filter)

			// Assert - validation should fail before hitting the mocks
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestValidateSearchFilter tests the validateSearchFilter helper function
func TestValidateSearchFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  domain.SearchFilter
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filter - 30 days",
			filter: domain.SearchFilter{
				TimeRange: domain.TimeWindow{
					Start: time.Now().Add(-30 * 24 * time.Hour),
					End:   time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "valid filter - exactly 90 days",
			filter: func() domain.SearchFilter {
				// Use fixed times to avoid time drift between operations
				end := time.Now()
				start := end.Add(-90 * 24 * time.Hour)
				return domain.SearchFilter{
					TimeRange: domain.TimeWindow{
						Start: start,
						End:   end,
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "invalid filter - 91 days",
			filter: domain.SearchFilter{
				TimeRange: domain.TimeWindow{
					Start: time.Now().Add(-91 * 24 * time.Hour),
					End:   time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "time range cannot exceed 90 days",
		},
		{
			name: "invalid filter - end before start",
			filter: domain.SearchFilter{
				TimeRange: domain.TimeWindow{
					Start: time.Now(),
					End:   time.Now().Add(-1 * time.Hour),
				},
			},
			wantErr: true,
			errMsg:  "end time must be after start time",
		},
		{
			name: "valid filter - no time range",
			filter: domain.SearchFilter{
				TimeRange: domain.TimeWindow{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearchFilter(tt.filter)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSearchService_FilterSanitization verifies that filter values with invalid characters are properly handled
func TestSearchService_FilterSanitization(t *testing.T) {
	tests := []struct {
		name             string
		typeFilters      []string
		sourceFilters    []string
		destFilters      []string
		expectedFilter   string
		expectSearchCall bool
	}{
		{
			name:             "valid filters - all clean",
			typeFilters:      []string{"MVT", "LDM"},
			sourceFilters:    []string{"KJFK", "KLAX"},
			destFilters:      []string{"EGLL", "LFPG"},
			expectedFilter:   `(type = "MVT" OR type = "LDM") AND (source = "KJFK" OR source = "KLAX") AND (destination = "EGLL" OR destination = "LFPG")`,
			expectSearchCall: true,
		},
		{
			name:             "mixed filters - some contain injection attempts",
			typeFilters:      []string{"MVT", "bad=type", "LDM"},
			sourceFilters:    []string{"KJFK", `dangerous"value`, "KLAX"},
			destFilters:      []string{"EGLL", "LFPG"},
			expectedFilter:   `(type = "MVT" OR type = "LDM") AND (source = "KJFK" OR source = "KLAX") AND (destination = "EGLL" OR destination = "LFPG")`,
			expectSearchCall: true,
		},
		{
			name:             "all invalid - should be filtered out",
			typeFilters:      []string{"bad=type", `evil"value`, "test()"},
			sourceFilters:    []string{`dangerous"value`, "evil[]"},
			destFilters:      []string{"bad{}", "evil=val"},
			expectedFilter:   "",
			expectSearchCall: true,
		},
		{
			name:             "empty filters",
			typeFilters:      []string{},
			sourceFilters:    []string{},
			destFilters:      []string{},
			expectedFilter:   "",
			expectSearchCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			searchIndex := &mocks.SearchIndexMock{}
			telegramStore := &mocks.TelegramStoreMock{}
			logger := zap.NewNop()

			service := NewService(searchIndex, telegramStore, nil, logger)

			// Configure mock to capture and validate the filter parameter
			var capturedFilter string
			searchIndex.On("Search",
				context.Background(),
				"test",
				mock.MatchedBy(func(filter string) bool {
					capturedFilter = filter
					return true
				}),
				int64(50),
				int64(0),
				[]string(nil),
			).Return(&meilisearch.SearchResponse{
				Hits:               meilisearch.Hits{},
				EstimatedTotalHits: 0,
				TotalHits:          0,
			}, nil).Once()

			filter := domain.SearchFilter{
				Query:       "test",
				Type:        tt.typeFilters,
				Source:      tt.sourceFilters,
				Destination: tt.destFilters,
				Page: domain.Pagination{
					Limit:  50,
					Offset: 0,
				},
			}

			// Execute
			result, err := service.Search(context.Background(), filter)

			// Assert
			require.NoError(t, err)
			assert.NotNil(t, result)
			if tt.expectSearchCall {
				searchIndex.AssertExpectations(t)
				assert.Equal(t, tt.expectedFilter, capturedFilter, "filter string should match expected after sanitization")
			}
		})
	}
}


