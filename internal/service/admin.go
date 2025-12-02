package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// AdminService provides administrative operations.
type AdminService struct {
	repo           app.Repository
	streamPublisher app.StreamPublisher
	logger         *zap.Logger
	reindexMutex   sync.Mutex
	reindexRunning bool
}

// NewAdminService creates a new admin service.
func NewAdminService(
	repo app.Repository,
	streamPublisher app.StreamPublisher,
	logger *zap.Logger,
) *AdminService {
	return &AdminService{
		repo:            repo,
		streamPublisher: streamPublisher,
		logger:          logger,
	}
}

// ReindexJobID represents a reindex job identifier.
type ReindexJobID struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

// Reindex triggers a reindex operation for the specified time range.
func (s *AdminService) Reindex(ctx context.Context, from, to time.Time) (interface{}, error) {
	// Check if reindex is already running
	s.reindexMutex.Lock()
	if s.reindexRunning {
		s.reindexMutex.Unlock()
		return nil, fmt.Errorf("reindex operation already in progress")
	}
	s.reindexRunning = true
	s.reindexMutex.Unlock()

	// Ensure we unlock on exit
	defer func() {
		s.reindexMutex.Lock()
		s.reindexRunning = false
		s.reindexMutex.Unlock()
	}()

	jobID := &ReindexJobID{
		ID:        fmt.Sprintf("reindex-%d", time.Now().Unix()),
		StartedAt: time.Now(),
		From:      from,
		To:        to,
	}

	s.logger.Info("starting reindex operation",
		zap.String("job_id", jobID.ID),
		zap.Time("from", from),
		zap.Time("to", to),
	)

	// Query PostgreSQL for messages in the specified time range
	filters := domain.SearchFilters{
		TimeRange: domain.TimeWindow{
			Start: from,
			End:   to,
		},
		Pagination: domain.Pagination{
			Limit:  1000, // Process in batches
			Offset: 0,
		},
	}

	totalProcessed := 0
	for {
		// Use StreamSearch to prevent OOM
		telegramChan, errChan := s.repo.StreamSearch(ctx, filters)

		batchCount := 0
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case telegram, ok := <-telegramChan:
				if !ok {
					// Channel closed, check for errors
					select {
					case err := <-errChan:
						if err != nil {
							return nil, fmt.Errorf("stream search error: %w", err)
						}
					default:
					}
					goto nextBatch
				}

				// Push message to Redis Stream job:index
				if err := s.streamPublisher.Publish(ctx, "job:index", telegram); err != nil {
					s.logger.Warn("failed to publish message to stream",
						zap.String("message_id", telegram.MessageID),
						zap.Error(err),
					)
					// Continue processing other messages
					continue
				}

				batchCount++
				totalProcessed++

			case err := <-errChan:
				if err != nil {
					return nil, fmt.Errorf("stream search error: %w", err)
				}
				goto nextBatch
			}
		}

	nextBatch:
		if batchCount == 0 {
			// No more messages
			break
		}

		s.logger.Info("reindex batch processed",
			zap.String("job_id", jobID.ID),
			zap.Int("batch_count", batchCount),
			zap.Int("total_processed", totalProcessed),
		)

		// Move to next batch
		filters.Pagination.Offset += filters.Pagination.Limit
	}

	s.logger.Info("reindex operation completed",
		zap.String("job_id", jobID.ID),
		zap.Int("total_processed", totalProcessed),
	)

	return jobID, nil
}

