package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateType(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		ttype string
		want  bool
	}{
		{"valid aftn", "aftn", true},
		{"valid sita", "sita", true},
		{"valid acars", "acars", true},
		{"valid cpdlc", "cpdlc", true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidateType(tt.ttype))
		})
	}
}

func TestValidator_ValidatePriority(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		priority int
		want     bool
	}{
		{"priority 1", 1, true},
		{"priority 2", 2, true},
		{"priority 3", 3, true},
		{"priority 0", 0, false},
		{"priority 4", 4, false},
		{"negative priority", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidatePriority(tt.priority))
		})
	}
}

func TestValidator_ValidateICAOCode(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{"valid ICAO", "ZBAA", true},
		{"valid ICAO lowercase", "zbaa", false},
		{"too short", "ZBA", false},
		{"too long", "ZBAAA", false},
		{"contains numbers", "ZB11", false},
		{"contains special chars", "ZB-A", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidateICAOCode(tt.code))
		})
	}
}

func TestSearchFilters_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filters SearchFilters
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filters",
			filters: SearchFilters{
				Query: "test",
				Pagination: Pagination{
					Limit:  50,
					Offset: 0,
					SortBy: "time",
					Order:  "desc",
				},
			},
			wantErr: false,
		},
		{
			name: "negative limit",
			filters: SearchFilters{
				Pagination: Pagination{
					Limit:  -1,
					Offset: 0,
				},
			},
			wantErr: true,
			errMsg:  "limit",
		},
		{
			name: "negative offset",
			filters: SearchFilters{
				Pagination: Pagination{
					Limit:  50,
					Offset: -1,
				},
			},
			wantErr: true,
			errMsg:  "offset",
		},
		{
			name: "invalid sort field",
			filters: SearchFilters{
				Pagination: Pagination{
					Limit:  50,
					Offset: 0,
					SortBy: "invalid_field",
				},
			},
			wantErr: true,
			errMsg:  "sort_by",
		},
		{
			name: "invalid order",
			filters: SearchFilters{
				Pagination: Pagination{
					Limit:  50,
					Offset: 0,
					Order:  "invalid",
				},
			},
			wantErr: true,
			errMsg:  "order",
		},
		{
			name: "time range start after end",
			filters: SearchFilters{
				TimeRange: TimeWindow{
					Start: time.Now().Add(24 * time.Hour),
					End:   time.Now(),
				},
				Pagination: DefaultPagination(),
			},
			wantErr: true,
			errMsg:  "start must be before end",
		},
		{
			name: "time range too large",
			filters: SearchFilters{
				TimeRange: TimeWindow{
					Start: time.Now().Add(-180 * 24 * time.Hour), // 180 days ago
					End:   time.Now(),
				},
				Pagination: DefaultPagination(),
			},
			wantErr: true,
			errMsg:  "cannot exceed",
		},
		{
			name: "time range at maximum limit (90 days)",
			filters: SearchFilters{
				TimeRange: TimeWindow{
					Start: time.Now().Add(-90*24*time.Hour + time.Minute), // Slightly less than 90 days to avoid floating point issues
					End:   time.Now(),
				},
				Pagination: DefaultPagination(),
			},
			wantErr: false,
		},
		{
			name: "time range within limit (30 days)",
			filters: SearchFilters{
				TimeRange: TimeWindow{
					Start: time.Now().Add(-30 * 24 * time.Hour),
					End:   time.Now(),
				},
				Pagination: DefaultPagination(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filters.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
