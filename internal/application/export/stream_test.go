package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap/zaptest"
)

// mockSearchService is a mock implementation of SearchService for testing
type mockSearchService struct {
	searchFunc func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error)
	callCount  int
}

func (m *mockSearchService) Search(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
	m.callCount++
	if m.searchFunc != nil {
		return m.searchFunc(ctx, filter)
	}
	return &domain.SearchResult{}, nil
}

func TestExportStream_CSV_SmallDataset(t *testing.T) {
	// Create mock service that returns a small dataset
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			// Return 10 telegrams
			telegrams := make([]domain.Telegram, 10)
			for i := 0; i < 10; i++ {
				telegrams[i] = domain.Telegram{
					MessageID:    "TEST-" + string(rune('0'+i)),
					Type:         "AFTN",
					Time:         time.Now(),
					FlightNumber: "CA123",
					Source:       "ZBAA",
					Destination:  "ZSPD",
					Priority:     2,
					Content:      "Test content",
				}
			}
			return &domain.SearchResult{
				Telegrams: telegrams,
				Total:     10,
			}, nil
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormatCSV, &buf)
	require.NoError(t, err)

	// Parse CSV and verify
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + 10 rows
	assert.Equal(t, 11, len(records))
	assert.Equal(t, []string{"Message ID", "Type", "Time", "Flight Number", "Source", "Destination", "Priority", "Content"}, records[0])

	// Verify search was called once
	assert.Equal(t, 1, mockService.callCount)
}

func TestExportStream_CSV_LargeDataset(t *testing.T) {
	// Create mock service that returns data in chunks
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			// Simulate chunked responses
			chunkSize := filter.Page.Limit
			if chunkSize == 0 {
				chunkSize = ChunkSize
			}

			// Stop after 2 chunks (2000 records)
			if filter.Page.Offset >= 2000 {
				return &domain.SearchResult{
					Telegrams: []domain.Telegram{},
					Total:     2000,
				}, nil
			}

			// Return telegrams based on offset
			telegrams := make([]domain.Telegram, chunkSize)
			for i := 0; i < chunkSize; i++ {
				idx := filter.Page.Offset + i
				telegrams[i] = domain.Telegram{
					MessageID:    "TEST-" + string(rune('0'+idx%10)),
					Type:         "AFTN",
					Time:         time.Now(),
					FlightNumber: "CA123",
					Source:       "ZBAA",
					Destination:  "ZSPD",
					Priority:     2,
					Content:      "Test content",
				}
			}

			return &domain.SearchResult{
				Telegrams: telegrams,
				Total:     2000,
			}, nil
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormatCSV, &buf)
	require.NoError(t, err)

	// Parse CSV and verify
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + 2000 rows
	assert.Equal(t, 2001, len(records))

	// Verify search was called 3 times (2 data chunks + 1 empty check)
	assert.Equal(t, 3, mockService.callCount)
}

func TestExportStream_CSV_Cancellation(t *testing.T) {
	// Create mock service that simulates slow responses
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			// Simulate some work
			select {
			case <-time.After(10 * time.Millisecond):
				telegrams := make([]domain.Telegram, ChunkSize)
				for i := 0; i < ChunkSize; i++ {
					telegrams[i] = domain.Telegram{
						MessageID:    "TEST-" + string(rune('0'+i)),
						Type:         "AFTN",
						Time:         time.Now(),
						FlightNumber: "CA123",
						Source:       "ZBAA",
						Destination:  "ZSPD",
						Priority:     2,
						Content:      "Test content",
					}
				}
				return &domain.SearchResult{
					Telegrams: telegrams,
					Total:     10000,
				}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	err := service.ExportStream(ctx, domain.SearchFilter{}, domain.ExportFormatCSV, &buf)

	// Should return context cancellation error
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

func TestExportStream_CSV_SearchError(t *testing.T) {
	// Create mock service that returns error
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			return nil, errors.New("search failed")
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormatCSV, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "search chunk")
}

func TestExportStream_CSV_MaxExportLimit(t *testing.T) {
	// Create mock service that always returns full chunks
	callCount := 0
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			callCount++
			// Always return full chunks to simulate more data available
			telegrams := make([]domain.Telegram, ChunkSize)
			for i := 0; i < ChunkSize; i++ {
				telegrams[i] = domain.Telegram{
					MessageID:    "TEST-" + string(rune('0'+i)),
					Type:         "AFTN",
					Time:         time.Now(),
					FlightNumber: "CA123",
					Source:       "ZBAA",
					Destination:  "ZSPD",
					Priority:     2,
					Content:      "Test content",
				}
			}
			return &domain.SearchResult{
				Telegrams: telegrams,
				Total:     50000, // More than MaxExportLimit
			}, nil
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormatCSV, &buf)
	require.NoError(t, err)

	// Parse CSV and verify
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + MaxExportLimit rows
	assert.Equal(t, MaxExportLimit+1, len(records))

	// Verify search was called MaxExportLimit/ChunkSize times
	expectedCalls := MaxExportLimit / ChunkSize
	assert.Equal(t, expectedCalls, callCount)
}

func TestExportStream_UnsupportedFormat(t *testing.T) {
	mockService := &mockSearchService{}
	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormat("unknown"), &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestExportStream_CSV_EmptyResult(t *testing.T) {
	// Create mock service that returns no data
	mockService := &mockSearchService{
		searchFunc: func(ctx context.Context, filter domain.SearchFilter) (*domain.SearchResult, error) {
			return &domain.SearchResult{
				Telegrams: []domain.Telegram{},
				Total:     0,
			}, nil
		},
	}

	service := NewService(mockService, zaptest.NewLogger(t))

	var buf bytes.Buffer
	err := service.ExportStream(context.Background(), domain.SearchFilter{}, domain.ExportFormatCSV, &buf)
	require.NoError(t, err)

	// Should only have header
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines))
	assert.Contains(t, lines[0], "Message ID")
}
