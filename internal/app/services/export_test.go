package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"fmt"

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
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

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

		mockRepo.On("Search", mock.Anything, mock.Anything).Return(searchResult, nil)
		mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, nil)           // Cache miss
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil) // Cache set

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
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

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
		mockRepo.On("Search", mock.Anything, mock.Anything).Return(searchResult, nil)
		mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, nil)           // Cache miss
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil) // Cache set

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
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

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

		mockRepo.On("Search", mock.Anything, mock.Anything).Return(searchResult, nil)
		mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, nil)           // Cache miss
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil) // Cache set

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

	t.Run("unsupported format returns error", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

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
	exportSvc := NewExportService(searchSvc, mockRepo, logger)

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
				Limit: 50000, // Exceeds max
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("export limit cannot exceed %d", domain.MaxExportRecords))
	})

	t.Run("zero limit allowed", func(t *testing.T) {
		filters := domain.SearchFilters{
			Pagination: domain.Pagination{
				Limit: 0,
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		require.NoError(t, err)
		assert.Equal(t, 0, filters.Pagination.Limit)
	})

	t.Run("time range within 90 days", func(t *testing.T) {
		filters := domain.SearchFilters{
			TimeRange: domain.TimeWindow{
				Start: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC), // 29 days
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		require.NoError(t, err)
	})

	t.Run("time range exceeds 90 days", func(t *testing.T) {
		filters := domain.SearchFilters{
			TimeRange: domain.TimeWindow{
				Start: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC), // 121 days
			},
		}

		err := exportSvc.validateExportFilters(&filters)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "range cannot exceed 90 days")
	})
}

func TestExportService_ExportStream(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("successful streaming export", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

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

		mockRepo.On("StreamSearch", ctx, filters).Return(searchResult, nil)

		telegramCh, errCh, err := exportSvc.ExportStream(ctx, filters)

		require.NoError(t, err)
		assert.NotNil(t, telegramCh)
		assert.NotNil(t, errCh)

		// Collect results
		var telegrams []*domain.Telegram
		var errs []error
		done := make(chan bool, 2)

		go func() {
			for telegram := range telegramCh {
				telegrams = append(telegrams, telegram)
			}
			done <- true
		}()

		go func() {
			for err := range errCh {
				errs = append(errs, err)
			}
			done <- true
		}()

		<-done
		<-done

		assert.Len(t, telegrams, 2)
		assert.Len(t, errs, 0)
		assert.Equal(t, "MSG-001", telegrams[0].MessageID)
		assert.Equal(t, "MSG-002", telegrams[1].MessageID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("streaming export with error", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

		filters := domain.SearchFilters{
			Query: "test",
		}

		mockRepo.On("StreamSearch", ctx, filters).Return(nil, errors.New("stream search failed"))

		telegramCh, errCh, err := exportSvc.ExportStream(ctx, filters)

		require.NoError(t, err)
		assert.NotNil(t, telegramCh)
		assert.NotNil(t, errCh)

		// Collect results
		var telegrams []*domain.Telegram
		var errs []error
		done := make(chan bool, 2)

		go func() {
			for telegram := range telegramCh {
				telegrams = append(telegrams, telegram)
			}
			done <- true
		}()

		go func() {
			for err := range errCh {
				errs = append(errs, err)
			}
			done <- true
		}()

		<-done
		<-done

		assert.Len(t, telegrams, 0)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "stream search failed")

		mockRepo.AssertExpectations(t)
	})

	t.Run("streaming export validation failure", func(t *testing.T) {
		ctx := context.Background()

		mockRepo := new(mockRepository)
		mockCache := new(mockCache)

		searchSvc := NewSearchService(mockRepo, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, mockRepo, logger)

		filters := domain.SearchFilters{
			TimeRange: domain.TimeWindow{
				Start: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC), // 121 days
			},
		}

		telegramCh, errCh, err := exportSvc.ExportStream(ctx, filters)

		assert.Error(t, err)
		assert.Nil(t, telegramCh)
		assert.Nil(t, errCh)
		assert.Contains(t, err.Error(), "range cannot exceed 90 days")
	})

	t.Run("streaming export without repository", func(t *testing.T) {
		ctx := context.Background()

		mockCache := new(mockCache)

		searchSvc := NewSearchService(nil, mockCache, nil, nil, logger)
		exportSvc := NewExportService(searchSvc, nil, logger) // No repository

		filters := domain.SearchFilters{
			Query: "test",
		}

		telegramCh, errCh, err := exportSvc.ExportStream(ctx, filters)

		assert.Error(t, err)
		assert.Nil(t, telegramCh)
		assert.Nil(t, errCh)
		assert.Contains(t, err.Error(), "streaming export not available")
	})
}
