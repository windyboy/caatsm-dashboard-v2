package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// ExportService handles data export operations
type ExportService struct {
	searchSvc *SearchService
	logger    *zap.Logger
}

// NewExportService creates a new export service
func NewExportService(searchSvc *SearchService, logger *zap.Logger) *ExportService {
	return &ExportService{
		searchSvc: searchSvc,
		logger:    logger,
	}
}

// Export exports search results in the specified format
func (e *ExportService) Export(ctx context.Context, filters domain.SearchFilters, format domain.ExportFormat) ([]byte, error) {
	// Validate format first (before expensive operations)
	switch format {
	case domain.ExportFormatCSV:
		// Valid format, continue
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	// Validate export limits
	if err := e.validateExportFilters(&filters); err != nil {
		return nil, fmt.Errorf("export validation failed: %w", err)
	}

	// Get data using search service
	result, err := e.searchSvc.Search(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("export search failed: %w", err)
	}

	// Export based on format
	switch format {
	case domain.ExportFormatCSV:
		return e.exportCSV(result)
	default:
		// Should never reach here due to validation above
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// validateExportFilters validates filters for export operations
func (e *ExportService) validateExportFilters(filters *domain.SearchFilters) error {
	// Check export limits
	const maxExportRecords = 10000
	if filters.Pagination.Limit > maxExportRecords {
		return fmt.Errorf("export limit cannot exceed %d records", maxExportRecords)
	}

	// Set default limit for export if not specified
	if filters.Pagination.Limit <= 0 {
		filters.Pagination.Limit = maxExportRecords
	}

	return nil
}

// exportCSV exports data as CSV
func (e *ExportService) exportCSV(result *domain.SearchResult) ([]byte, error) {
	if result == nil {
		result = &domain.SearchResult{}
	}
	e.logger.Info("exporting to CSV", zap.Int("records", len(result.Telegrams)))

	// Create buffer to hold CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV headers
	headers := []string{
		"message_id",
		"type",
		"time",
		"flight_number",
		"source",
		"destination",
		"priority",
		"content",
	}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, telegram := range result.Telegrams {
		row := []string{
			telegram.MessageID,
			telegram.Type,
			telegram.Time.Format("2006-01-02 15:04:05"),
			telegram.FlightNumber,
			telegram.Source,
			telegram.Destination,
			strconv.Itoa(telegram.Priority),
			telegram.Content,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Flush any buffered data
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	e.logger.Info("CSV export completed", zap.Int("records", len(result.Telegrams)))
	return buf.Bytes(), nil
}
