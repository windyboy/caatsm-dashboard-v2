package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/windy/caatsm-dashboard/internal/app"
)

// StreamCSV streams telegrams as CSV directly to the response.
// It reads from the telegram channel and writes CSV rows, flushing after each row.
// If an error occurs (from errCh or context cancellation), it returns immediately.
func StreamCSV(c echo.Context, stream <-chan *app.Telegram, errCh <-chan error) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=telegrams.csv")
	c.Response().WriteHeader(http.StatusOK)

	writer := csv.NewWriter(c.Response())
	defer writer.Flush()

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
		// Note: HTTP 200 already sent; client will see truncated response
		// Consider logging: log.Error("CSV header write failed", "error", err)
		return fmt.Errorf("write CSV headers: %w", err)
	}
	c.Response().Flush() // Ensure headers are sent immediately

	// Stream rows
	streamClosed := false
	errChClosed := false

	for {
		select {
		case <-c.Request().Context().Done():
			// Client disconnected, stop streaming
			return c.Request().Context().Err()

		case err, ok := <-errCh:
			if !ok {
				// Error channel closed
				errChClosed = true
				if streamClosed {
					return nil
				}
				continue
			}
			if err != nil {
				return fmt.Errorf("stream error: %w", err)
			}

		case telegram, ok := <-stream:
			if !ok {
				// Stream closed, all telegrams sent
				streamClosed = true
				if errChClosed {
					return nil
				}
				// Continue to check error channel
				continue
			}

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
				return fmt.Errorf("write CSV row: %w", err)
			}

			// Flush after each row to ensure frontend has progress feedback
			writer.Flush()
			if err := writer.Error(); err != nil {
				return fmt.Errorf("CSV writer error: %w", err)
			}
			c.Response().Flush()
		}

		// If both channels are closed, we're done
		if streamClosed && errChClosed {
			return nil
		}
	}
}
