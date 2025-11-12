package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/models"
	"go.uber.org/zap"
)

// exportService implements ExportService.
type exportService struct {
	searchService SearchService
	logger        *zap.Logger
}

// NewExportService creates a new ExportService implementation.
func NewExportService(searchService SearchService, logger *zap.Logger) ExportService {
	return &exportService{
		searchService: searchService,
		logger:        logger,
	}
}

// Export exports search results in the specified format.
func (e *exportService) Export(ctx context.Context, filter models.SearchFilter, format models.ExportFormat) ([]byte, error) {
	// Set a higher limit for export
	filter.Page.Limit = 10000
	filter.Page.Offset = 0

	result, err := e.searchService.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search for export: %w", err)
	}

	switch format {
	case models.ExportFormatCSV:
		return e.exportCSV(result)
	case models.ExportFormatExcel:
		return e.exportExcel(result)
	case models.ExportFormatPDF:
		return e.exportPDF(result)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (e *exportService) exportCSV(result *models.SearchResult) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Message ID", "Type", "Time", "Flight Number", "Source", "Destination", "Priority", "Content"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	// Write rows
	for _, telegram := range result.Telegrams {
		row := []string{
			telegram.MessageID,
			telegram.Type,
			telegram.Time.Format(time.RFC3339),
			telegram.FlightNumber,
			telegram.Source,
			telegram.Destination,
			fmt.Sprintf("%d", telegram.Priority),
			telegram.Content,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}

func (e *exportService) exportExcel(result *models.SearchResult) ([]byte, error) {
	// For Excel export, we'll use JSON for now
	// In production, you'd use a library like excelize
	return e.exportJSON(result)
}

func (e *exportService) exportPDF(result *models.SearchResult) ([]byte, error) {
	// For PDF export, we'll use JSON for now
	// In production, you'd use a library like go-pdf
	return e.exportJSON(result)
}

func (e *exportService) exportJSON(result *models.SearchResult) ([]byte, error) {
	return json.Marshal(result)
}
