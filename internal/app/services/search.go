package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

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

	// Extract suggestions from Meilisearch response
	suggestions := make([]string, 0, limit)
	
	// Type assert to map structure that Meilisearch returns
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Extract hits from the response
		if hits, ok := resultMap["hits"].([]interface{}); ok {
			seen := make(map[string]bool) // Deduplicate suggestions
			
			for _, hit := range hits {
				if len(suggestions) >= limit {
					break
				}
				
				hitMap, ok := hit.(map[string]interface{})
				if !ok {
					continue
				}
				
				// Extract values from each attribute
				for _, attr := range attributes {
					if val, exists := hitMap[attr]; exists && val != nil {
						if strVal, ok := val.(string); ok && strVal != "" {
							// Add to suggestions if not already seen
							if !seen[strVal] {
								suggestions = append(suggestions, strVal)
								seen[strVal] = true
								
								if len(suggestions) >= limit {
									break
								}
							}
						}
					}
				}
			}
		} else {
			s.logger.Warn("autocomplete response missing hits field or wrong type",
				zap.String("query", query),
			)
		}
	} else {
		s.logger.Warn("autocomplete response type assertion failed",
			zap.String("query", query),
			zap.String("result_type", fmt.Sprintf("%T", result)),
		)
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
