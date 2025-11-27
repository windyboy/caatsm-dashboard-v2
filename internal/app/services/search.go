package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"go.uber.org/zap"
)

// SearchService handles search operations
type SearchService struct {
	repo   ports.Repository
	cache  ports.Cache
	search ports.SearchIndex
	pub    ports.EventPublisher
	logger *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	repo ports.Repository,
	cache ports.Cache,
	search ports.SearchIndex,
	pub ports.EventPublisher,
	logger *zap.Logger,
) *SearchService {
	return &SearchService{
		repo:   repo,
		cache:  cache,
		search: search,
		pub:    pub,
		logger: logger,
	}
}

// Search performs a search operation with caching
func (s *SearchService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, fmt.Errorf("invalid search filters: %w", err)
	}

	// Try cache first
	cacheKey := s.buildCacheKey(filters)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			if result, ok := cached.(*domain.SearchResult); ok {
				s.logger.Debug("cache hit", zap.String("key", cacheKey))
				return result, nil
			}
		}
	}

	// Perform search
	start := time.Now()
	result, err := s.repo.Search(ctx, filters)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("search failed",
			zap.Error(err),
			zap.String("query", filters.Query),
			zap.Duration("duration", duration),
		)
		return nil, fmt.Errorf("search operation failed: %w", err)
	}

	// Cache result
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, result); err != nil {
			s.logger.Warn("failed to cache search result", zap.Error(err))
		}
	}

	// Publish search event
	if s.pub != nil {
		event := domain.SearchPerformed{
			Query:       filters.Query,
			ResultCount: result.Total,
			Duration:    duration,
			Timestamp:   time.Now(),
		}
		if err := s.pub.Publish(ctx, event); err != nil {
			s.logger.Warn("failed to publish search event", zap.Error(err))
		}
	}

	s.logger.Info("search completed",
		zap.String("query", filters.Query),
		zap.Int64("total", result.Total),
		zap.Int("returned", len(result.Telegrams)),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// Autocomplete provides search suggestions
func (s *SearchService) Autocomplete(ctx context.Context, query string, limit int) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	attributes := []string{"flight_number", "message_id", "source", "destination"}
	result, err := s.search.SearchAutocomplete(ctx, query, int64(limit), attributes)
	if err != nil {
		return nil, fmt.Errorf("autocomplete search failed: %w", err)
	}

	// Type assert to Meilisearch SDK's SearchResponse
	searchResp, ok := result.(*meilisearch.SearchResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse autocomplete response: expected *meilisearch.SearchResponse, got %T", result)
	}

	if searchResp.Hits == nil {
		s.logger.Warn("autocomplete response has nil hits",
			zap.String("query", query),
		)
		return []string{}, nil
	}

	// Extract suggestions from typed hits
	seen := make(map[string]bool)
	suggestions := make([]string, 0, limit)

	for _, hit := range searchResp.Hits {
		if len(suggestions) >= limit {
			break
		}

		// Decode hit into a map to access dynamic fields
		var hitData map[string]interface{}
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			s.logger.Error("failed to marshal hit",
				zap.Error(err),
				zap.String("query", query),
			)
			return nil, fmt.Errorf("failed to marshal hit: %w", err)
		}

		if err := json.Unmarshal(hitBytes, &hitData); err != nil {
			s.logger.Error("failed to unmarshal hit",
				zap.Error(err),
				zap.String("query", query),
			)
			return nil, fmt.Errorf("failed to unmarshal hit: %w", err)
		}

		// Extract values from each attribute
		for _, attr := range attributes {
			if len(suggestions) >= limit {
				break
			}

			val, exists := hitData[attr]
			if !exists || val == nil {
				continue
			}

			strVal, ok := val.(string)
			if !ok || strVal == "" {
				continue
			}

			// Add to suggestions if not already seen
			if !seen[strVal] {
				suggestions = append(suggestions, strVal)
				seen[strVal] = true
			}
		}
	}

	return suggestions, nil
}

// buildCacheKey creates a deterministic cache key for search filters
func (s *SearchService) buildCacheKey(filters domain.SearchFilters) string {
	key := fmt.Sprintf("search:%s", filters.Query)

	// Add filter components
	if len(filters.Types) > 0 {
		key += fmt.Sprintf(":types:%v", filters.Types)
	}
	if len(filters.Sources) > 0 {
		key += fmt.Sprintf(":sources:%v", filters.Sources)
	}
	if len(filters.Destinations) > 0 {
		key += fmt.Sprintf(":dests:%v", filters.Destinations)
	}
	if len(filters.Priorities) > 0 {
		key += fmt.Sprintf(":priorities:%v", filters.Priorities)
	}

	// Add time range
	if !filters.TimeRange.Start.IsZero() {
		key += fmt.Sprintf(":start:%d", filters.TimeRange.Start.Unix())
	}
	if !filters.TimeRange.End.IsZero() {
		key += fmt.Sprintf(":end:%d", filters.TimeRange.End.Unix())
	}

	// Add pagination
	key += fmt.Sprintf(":limit:%d:offset:%d:sort:%s:%s",
		filters.Pagination.Limit,
		filters.Pagination.Offset,
		filters.Pagination.SortBy,
		filters.Pagination.Order,
	)

	// Create hash for consistent key length
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("search:%x", hash)
}
