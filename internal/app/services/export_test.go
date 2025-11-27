package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
)

func TestExportService_Export(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("CSV export with valid data", func(t *testing.T) {
		ctx := context.Background()

		// Create real services with mocked dependencies
		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit: 100,
			},
		}

		searchResult := &domain.SearchResult{
			Total: 2,
			Telegrams: []domain.Telegram{
				{
					MessageID:    "MSG-001",
					Type:         "METAR",
					Time:         time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC),
					FlightNumber: "BA123",
					Source:       "EGLL",
					Destination:  "KJFK",
					Priority:     1,
					Content:      "Test message 1",
				},
				{
					MessageID:    "MSG-002",
					Type:         "TAF",
					Time:         time.Date(2025, 11, 27, 11, 0, 0, 0, time.UTC),
					FlightNumber: "AA456",
					Source:       "KJFK",
					Destination:  "EGLL",
					Priority:     2,
					Content:      "Test message 2",
				},
			},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("Search", ctx, filters).Return(searchResult, nil)
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormatCSV)

		require.NoError(t, err)
		assert.NotNil(t, data)

		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)

		// Verify CSV content
		reader := csv.NewReader(bytes.NewReader(data))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Should have header + 2 data rows
		assert.Len(t, records, 3)

		// Verify header
		expectedHeader := []string{
			"message_id", "type", "time", "flight_number",
			"source", "destination", "priority", "content",
		}
		assert.Equal(t, expectedHeader, records[0])

		// Verify first data row
		assert.Equal(t, "MSG-001", records[1][0])
		assert.Equal(t, "METAR", records[1][1])
		assert.Equal(t, "BA123", records[1][3])

		// Verify second data row
		assert.Equal(t, "MSG-002", records[2][0])
		assert.Equal(t, "TAF", records[2][1])
		assert.Equal(t, "AA456", records[2][3])
	})

	t.Run("CSV export with empty result", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "nonexistent",
			Pagination: domain.Pagination{
				Limit: 100,
			},
		}

		searchResult := &domain.SearchResult{
			Total:     0,
			Telegrams: []domain.Telegram{},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("Search", ctx, filters).Return(searchResult, nil)
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormatCSV)

		require.NoError(t, err)
		assert.NotNil(t, data)

		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)

		// Verify CSV content - should have only header
		reader := csv.NewReader(bytes.NewReader(data))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 1) // Only header
	})

	t.Run("CSV export with special characters", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit: 100,
			},
		}

		searchResult := &domain.SearchResult{
			Total: 1,
			Telegrams: []domain.Telegram{
				{
					MessageID:    "MSG-001",
					Type:         "METAR",
					Time:         time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC),
					FlightNumber: "BA123",
					Source:       "EGLL",
					Destination:  "KJFK",
					Priority:     1,
					Content:      "Message with, comma and \"quotes\"",
				},
			},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("Search", ctx, filters).Return(searchResult, nil)
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormatCSV)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)

		// CSV library should properly escape special characters
		reader := csv.NewReader(bytes.NewReader(data))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 2)
		// Content should be properly parsed despite special chars
		assert.Equal(t, "Message with, comma and \"quotes\"", records[1][7])
	})

	t.Run("Excel export returns placeholder", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit: 100,
			},
		}

		searchResult := &domain.SearchResult{
			Total:     1,
			Telegrams: []domain.Telegram{{MessageID: "MSG-001"}},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("Search", ctx, filters).Return(searchResult, nil)
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormatExcel)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)

		assert.Contains(t, string(data), "Excel export not yet implemented")
	})

	t.Run("PDF export returns placeholder", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "test",
			Pagination: domain.Pagination{
				Limit: 100,
			},
		}

		searchResult := &domain.SearchResult{
			Total:     1,
			Telegrams: []domain.Telegram{{MessageID: "MSG-001"}},
		}

		mockCache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("cache miss"))
		mockRepo.On("Search", ctx, filters).Return(searchResult, nil)
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), searchResult).Return(nil)

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormatPDF)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)

		assert.Contains(t, string(data), "PDF export not yet implemented")
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, logger)

		filters := domain.SearchFilters{
			Query: "test",
		}

		data, err := exportSvc.Export(ctx, filters, domain.ExportFormat("invalid"))

		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "unsupported export format")
	})
}

func TestExportService_ValidateExportFilters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockRepo := new(mockRepository)
	mockCache := new(mockCache)
	searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
	exportSvc := NewExportService(searchSvc, logger)

	t.Run("limit within bounds", func(t *testing.T) {
		filters := domain.SearchFilters{
			Pagination: domain.Pagination{
				Limit: 5000,
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		require.NoError(t, err)
		assert.Equal(t, 5000, filters.Pagination.Limit)
	})

	t.Run("limit exceeds maximum", func(t *testing.T) {
		filters := domain.SearchFilters{
			Pagination: domain.Pagination{
				Limit: 50000, // Exceeds max of 10000
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "export limit cannot exceed 10000")
	})

	t.Run("zero limit defaults to max", func(t *testing.T) {
		filters := domain.SearchFilters{
			Pagination: domain.Pagination{
				Limit: 0,
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		require.NoError(t, err)
		assert.Equal(t, 10000, filters.Pagination.Limit)
	})
}
