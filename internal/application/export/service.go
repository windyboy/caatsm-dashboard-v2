package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

const (
	MaxExportLimit = 10000
)

// Service implements export use cases.
type Service struct {
	searchService SearchService
	logger        *zap.Logger
}

// SearchService defines the interface for search operations needed by export.
type SearchService interface {
	Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error)
}

// NewService creates a new export service.
func NewService(searchService SearchService, logger *zap.Logger) *Service {
	return &Service{
		searchService: searchService,
		logger:        logger,
	}
}

// Export exports search results in the specified format.
func (e *Service) Export(ctx context.Context, filter persistence.SearchFilter, format persistence.ExportFormat) ([]byte, error) {
	// Set a higher limit for export
	filter.Page.Limit = MaxExportLimit
	filter.Page.Offset = 0

	result, err := e.searchService.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search for export: %w", err)
	}

	switch format {
	case persistence.ExportFormatCSV:
		return e.exportCSV(result)
	case persistence.ExportFormatExcel:
		return e.exportExcel(result)
	case persistence.ExportFormatPDF:
		return e.exportPDF(result)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (e *Service) exportCSV(result *persistence.SearchResult) ([]byte, error) {
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

func (e *Service) exportExcel(result *persistence.SearchResult) ([]byte, error) {
	// For Excel export, we'll use JSON for now
	// In production, you'd use a library like excelize
	return e.exportJSON(result)
}

func (e *Service) exportPDF(result *persistence.SearchResult) ([]byte, error) {
	// For PDF export, we'll use JSON for now
	// In production, you'd use a library like go-pdf
	return e.exportJSON(result)
}

func (e *Service) exportJSON(result *persistence.SearchResult) ([]byte, error) {
	return json.Marshal(result)
}

