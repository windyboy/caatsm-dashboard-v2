package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFilters_Validate_TimeRangeTooLarge(t *testing.T) {
	tests := []struct {
		name      string
		timeRange TimeWindow
		wantError bool
		errMsg    string
	}{
		{
			name: "valid time range within 90 days",
			timeRange: TimeWindow{
				Start: time.Now().AddDate(0, 0, -30),
				End:   time.Now(),
			},
			wantError: false,
		},
		{
			name: "time range exactly 90 days",
			timeRange: TimeWindow{
				Start: time.Now().Add(-90*24*time.Hour + time.Minute), // Slightly less than 90 days to avoid precision issues
				End:   time.Now(),
			},
			wantError: false,
		},
		{
			name: "time range exceeds 90 days",
			timeRange: TimeWindow{
				Start: time.Now().AddDate(0, 0, -91),
				End:   time.Now(),
			},
			wantError: true,
			errMsg:    "range cannot exceed 90 days",
		},
		{
			name: "time range exceeds 90 days by a lot",
			timeRange: TimeWindow{
				Start: time.Now().AddDate(0, 0, -180),
				End:   time.Now(),
			},
			wantError: true,
			errMsg:    "range cannot exceed 90 days",
		},
		{
			name: "start after end",
			timeRange: TimeWindow{
				Start: time.Now(),
				End:   time.Now().AddDate(0, 0, -1),
			},
			wantError: true,
			errMsg:    "start must be before end",
		},
		{
			name: "zero time range",
			timeRange: TimeWindow{
				Start: time.Time{},
				End:   time.Time{},
			},
			wantError: false,
		},
		{
			name: "only start set",
			timeRange: TimeWindow{
				Start: time.Now().AddDate(0, 0, -30),
				End:   time.Time{},
			},
			wantError: false,
		},
		{
			name: "only end set",
			timeRange: TimeWindow{
				Start: time.Time{},
				End:   time.Now(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := &SearchFilters{
				TimeRange: tt.timeRange,
				Pagination: Pagination{
					Limit:  50,
					Offset: 0,
					SortBy: "time",
					Order:  "desc",
				},
			}

			err := filters.Validate()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchFilters_Validate_Pagination(t *testing.T) {
	tests := []struct {
		name       string
		pagination Pagination
		wantError  bool
		errMsg     string
	}{
		{
			name: "valid pagination",
			pagination: Pagination{
				Limit:  50,
				Offset: 0,
				SortBy: "time",
				Order:  "desc",
			},
			wantError: false,
		},
		{
			name: "negative limit",
			pagination: Pagination{
				Limit:  -1,
				Offset: 0,
			},
			wantError: true,
			errMsg:    "limit",
		},
		{
			name: "negative offset",
			pagination: Pagination{
				Limit:  50,
				Offset: -1,
			},
			wantError: true,
			errMsg:    "offset",
		},
		{
			name: "invalid sort field",
			pagination: Pagination{
				Limit:  50,
				Offset: 0,
				SortBy: "invalid_field",
			},
			wantError: true,
			errMsg:    "invalid sort field",
		},
		{
			name: "valid sort fields",
			pagination: Pagination{
				Limit:  50,
				Offset: 0,
				SortBy: "priority",
				Order:  "asc",
			},
			wantError: false,
		},
		{
			name: "invalid order",
			pagination: Pagination{
				Limit:  50,
				Offset: 0,
				Order:  "invalid",
			},
			wantError: true,
			errMsg:    "must be 'asc' or 'desc'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := &SearchFilters{
				Pagination: tt.pagination,
			}

			err := filters.Validate()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultPagination(t *testing.T) {
	p := DefaultPagination()

	assert.Equal(t, 50, p.Limit)
	assert.Equal(t, 0, p.Offset)
	assert.Equal(t, "time", p.SortBy)
	assert.Equal(t, "desc", p.Order)
}

func TestSearchFilters_TimeRangeValidation(t *testing.T) {
	// Test time range validation at domain level
	filter := SearchFilters{
		TimeRange: TimeWindow{
			Start: time.Now().Add(-180 * 24 * time.Hour), // 180 days ago
			End:   time.Now(),
		},
		Pagination: Pagination{
			Limit:  50,
			Offset: 0,
		},
	}

	err := filter.Validate()
	assert.Error(t, err, "filter validation should fail for > 90 day range")
	assert.Contains(t, err.Error(), "range cannot exceed", "error should mention range limit")
}
