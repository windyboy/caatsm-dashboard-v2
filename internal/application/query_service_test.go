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
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/testing/mocks"
)

func TestQueryService_Search(t *testing.T) {
	tests := []struct {
		name        string
		filter      models.SearchFilter
		indexError  error
		indexResult interface{}
		expectError bool
	}{
		{
			name: "valid search",
			filter: models.SearchFilter{
				Query: "test",
				Page: models.Pagination{
					Limit:  10,
					Offset: 0,
					SortBy: "time",
					Order:  "desc",
				},
			},
			expectError: false,
		},
		{
			name: "search with filters",
			filter: models.SearchFilter{
				Query: "test",
				Type:  []string{"aftn"},
				Page: models.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			expectError: false,
		},
		{
			name: "index error",
			filter: models.SearchFilter{
				Query: "test",
				Page: models.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			indexError:  errors.New("index error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeMock := new(mocks.TelegramStoreMock)
			indexMock := new(mocks.SearchIndexMock)

			service := NewQueryService(storeMock, indexMock)

			// Setup expectations
			filterStr := buildFilterString(tt.filter)
			var sort []string
			if tt.filter.Page.SortBy != "" {
				sortBy := tt.filter.Page.SortBy
				if tt.filter.Page.Order == "ASC" {
					sortBy += ":asc"
				} else {
					sortBy += ":desc"
				}
				sort = []string{sortBy}
			}

			indexMock.On("Search", mock.Anything, tt.filter.Query, filterStr,
				int64(tt.filter.Page.Limit), int64(tt.filter.Page.Offset), sort).
				Return(tt.indexResult, tt.indexError)

			// Execute
			result, err := service.Search(context.Background(), tt.filter)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.filter.Page, result.Page)
			}

			indexMock.AssertExpectations(t)
		})
	}
}

func TestQueryService_Recent(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		storeError  error
		storeResult *models.SearchResult
		expectError bool
		expectCount int
	}{
		{
			name:  "valid recent query",
			limit: 10,
			storeResult: &models.SearchResult{
				Telegrams: []models.Telegram{
					{MessageID: "TEST-001", Type: "aftn", Time: time.Now(), Priority: 2},
					{MessageID: "TEST-002", Type: "sita", Time: time.Now(), Priority: 2},
				},
				Total: 2,
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "store error",
			limit:       10,
			storeError:  errors.New("database error"),
			expectError: true,
		},
		{
			name:  "empty result",
			limit: 10,
			storeResult: &models.SearchResult{
				Telegrams: []models.Telegram{},
				Total:     0,
			},
			expectError: false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeMock := new(mocks.TelegramStoreMock)
			indexMock := new(mocks.SearchIndexMock)

			service := NewQueryService(storeMock, indexMock)

			// Setup expectations
			expectedFilter := models.SearchFilter{
				Page: models.Pagination{
					Limit:  tt.limit,
					Offset: 0,
					SortBy: "time",
					Order:  "desc",
				},
			}

			storeMock.On("Search", mock.Anything, expectedFilter).Return(tt.storeResult, tt.storeError)

			// Execute
			result, err := service.Recent(context.Background(), tt.limit)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectCount)
				for _, tg := range result {
					assert.IsType(t, &domain.Telegram{}, tg)
				}
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestBuildFilterString(t *testing.T) {
	tests := []struct {
		name   string
		filter models.SearchFilter
		want   string
	}{
		{
			name:   "empty filter",
			filter: models.SearchFilter{},
			want:   "",
		},
		{
			name: "type filter",
			filter: models.SearchFilter{
				Type: []string{"aftn", "sita"},
			},
			want: `(type = "aftn" OR type = "sita")`,
		},
		{
			name: "source filter",
			filter: models.SearchFilter{
				Source: []string{"ZBAA", "ZSPD"},
			},
			want: `(source = "ZBAA" OR source = "ZSPD")`,
		},
		{
			name: "priority filter",
			filter: models.SearchFilter{
				Priority: []int{1, 2},
			},
			want: `(priority = 1 OR priority = 2)`,
		},
		{
			name: "multiple filters",
			filter: models.SearchFilter{
				Type:     []string{"aftn"},
				Priority: []int{1},
			},
			want: `(type = "aftn") AND (priority = 1)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFilterString(tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}
