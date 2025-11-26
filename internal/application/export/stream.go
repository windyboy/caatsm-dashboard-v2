package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

const (
	// ChunkSize defines how many records to fetch and write at a time
	ChunkSize = 1000
)

// ExportStream exports search results in streaming fashion to avoid loading all data into memory.
// This is the preferred method for large exports.
func (s *Service) ExportStream(ctx context.Context, filter persistence.SearchFilter, format persistence.ExportFormat, w io.Writer) error {
	switch format {
	case persistence.ExportFormatCSV:
		return s.exportCSVStream(ctx, filter, w)
	case persistence.ExportFormatExcel:
		// Excel requires full buffering, so fall back to buffered export
		data, err := s.Export(ctx, filter, format)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	case persistence.ExportFormatPDF:
		// PDF requires full buffering, so fall back to buffered export
		data, err := s.Export(ctx, filter, format)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportCSVStream streams CSV export in chunks to avoid memory issues
func (s *Service) exportCSVStream(ctx context.Context, filter persistence.SearchFilter, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"Message ID", "Type", "Time", "Flight Number", "Source", "Destination", "Priority", "Content"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	// Stream in chunks
	totalExported := 0
	for offset := 0; offset < MaxExportLimit; offset += ChunkSize {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Prepare filter for this chunk
		filter.Page.Limit = ChunkSize
		filter.Page.Offset = offset

		// Fetch chunk
		result, err := s.searchService.Search(ctx, filter)
		if err != nil {
			return fmt.Errorf("search chunk at offset %d: %w", offset, err)
		}

		// If no results, we're done
		if len(result.Telegrams) == 0 {
			break
		}

		// Write rows from this chunk
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
			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("write csv row: %w", err)
			}
			totalExported++
		}

		// Flush after each chunk to stream data progressively
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return fmt.Errorf("flush csv chunk: %w", err)
		}

		s.logger.Debug("exported chunk",
			zap.Int("offset", offset),
			zap.Int("chunk_size", len(result.Telegrams)),
			zap.Int("total_exported", totalExported),
		)

		// If we got fewer results than the chunk size, we've reached the end
		if len(result.Telegrams) < ChunkSize {
			break
		}

		// Stop if we've reached the maximum export limit
		if totalExported >= MaxExportLimit {
			s.logger.Warn("reached maximum export limit",
				zap.Int("max_limit", MaxExportLimit),
				zap.Int("total_exported", totalExported),
			)
			break
		}
	}

	s.logger.Info("export completed",
		zap.Int("total_exported", totalExported),
	)

	return nil
}
