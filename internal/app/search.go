package app

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SearchService handles search operations
type SearchService struct {
	repo   Repository
	cache  Cache
	search SearchIndex
	pub    EventPublisher
	logger *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	repo Repository,
	cache Cache,
	search SearchIndex,
	pub EventPublisher,
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
func (s *SearchService) Search(ctx context.Context, filters SearchFilters) (*SearchResult, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, fmt.Errorf("invalid search filters: %w", err)
	}

	// Try cache first

	cacheKey := s.buildCacheKey(filters)

	if s.cache != nil {

		cached, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			s.logger.Warn("cache get error", zap.String("key", cacheKey), zap.Error(err))
		} else if cached != nil {

			resultBytes, err := json.Marshal(cached)

			if err != nil {
				s.logger.Warn("failed to marshal cached search result", zap.String("key", cacheKey), zap.Error(err))
				_ = s.cache.Delete(ctx, cacheKey)
			} else {
				var result SearchResult

				if err := json.Unmarshal(resultBytes, &result); err != nil {

					s.logger.Warn("failed to unmarshal cached search result", zap.String("key", cacheKey), zap.Error(err))

					_ = s.cache.Delete(ctx, cacheKey)
				} else {
					s.logger.Debug("cache hit", zap.String("key", cacheKey))
					return &result, nil

				}

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
		event := SearchPerformed{
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

// AutocompleteSuggestion represents a single autocomplete suggestion with its type
type AutocompleteSuggestion struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// Autocomplete provides search suggestions with type labels for better readability
func (s *SearchService) Autocomplete(ctx context.Context, query string, limit int) ([]string, error) {
	suggestions, err := s.AutocompleteWithTypes(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Convert to simple string array for backward compatibility
	result := make([]string, len(suggestions))
	for i, sug := range suggestions {
		result[i] = sug.Value
	}
	return result, nil
}

// AutocompleteWithTypes provides search suggestions with type information.
// It queries the Meilisearch index for autocomplete results across multiple attributes
// (flight_number, message_id, source, destination) and returns structured suggestions
// with value, type, and human-readable label. The method ensures deduplication
// and prioritizes certain attribute types in the results.
func (s *SearchService) AutocompleteWithTypes(ctx context.Context, query string, limit int) ([]AutocompleteSuggestion, error) {
	if query == "" {
		return []AutocompleteSuggestion{}, nil
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

	if result == nil || len(result.Hits) == 0 {
		s.logger.Debug("autocomplete response has no hits",
			zap.String("query", query),
		)
		return []AutocompleteSuggestion{}, nil
	}

	// Map attribute names to display labels
	attrLabels := map[string]string{
		"flight_number": "Flight",
		"message_id":    "Message ID",
		"source":        "From",
		"destination":   "To",
	}

	// Extract suggestions from typed hits with type information
	seen := make(map[string]bool) // key: "type:value"
	suggestions := make([]AutocompleteSuggestion, 0, limit)

	for _, hit := range result.Hits {
		if len(suggestions) >= limit {
			break
		}

		// Extract values from each attribute with priority order
		// Use domain AutocompleteHit struct fields instead of map access
		attributes := []struct {
			attr  string
			value string
		}{
			{"flight_number", hit.FlightNumber},
			{"message_id", hit.MessageID},
			{"source", hit.Source},
			{"destination", hit.Destination},
		}

		for _, attrData := range attributes {
			if len(suggestions) >= limit {
				break
			}

			if attrData.value == "" {
				continue
			}

			// Create unique key for deduplication
			key := fmt.Sprintf("%s:%s", attrData.attr, attrData.value)
			if !seen[key] {
				label := attrLabels[attrData.attr]
				if label == "" {
					label = attrData.attr
				}

				suggestions = append(suggestions, AutocompleteSuggestion{
					Value: attrData.value,
					Type:  attrData.attr,
					Label: label,
				})
				seen[key] = true
			}
		}
	}

	return suggestions, nil
}

// buildCacheKey creates a deterministic cache key for search filters
func (s *SearchService) buildCacheKey(filters SearchFilters) string {
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
