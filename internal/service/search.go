package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/windy/caatsm-dashboard/internal/domain"
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
	
	var result *SearchResult
	var err error
	
	// If there's a query string, use Meilisearch for full-text search first
	// Then fetch full data from PostgreSQL with filters applied
	if filters.Query != "" && s.search != nil {
		result, err = s.searchWithMeilisearch(ctx, filters)
	} else {
		// No query string or Meilisearch unavailable - use PostgreSQL only
		result, err = s.repo.Search(ctx, filters)
	}
	
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

// searchWithMeilisearch performs a hybrid search: Meilisearch for full-text, PostgreSQL for filters
func (s *SearchService) searchWithMeilisearch(ctx context.Context, filters SearchFilters) (*SearchResult, error) {
	// Build Meilisearch request parameters
	meiliFilter := s.buildMeilisearchFilter(filters)
	
	sortArray := s.buildSortArray(filters.Pagination)
	limit := int64(filters.Pagination.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int64(filters.Pagination.Offset)
	if offset < 0 {
		offset = 0
	}
	
	// Search Meilisearch to get message IDs
	meiliResult, err := s.search.Search(ctx, filters.Query, meiliFilter, limit, offset, sortArray)
	if err != nil {
		s.logger.Warn("meilisearch failed, falling back to PostgreSQL",
			zap.Error(err),
			zap.String("query", filters.Query),
		)
		return s.repo.Search(ctx, filters)
	}
	
	// Extract message IDs from Meilisearch results
	messageIDs, total, err := s.extractMessageIDsFromMeilisearch(meiliResult)
	if err != nil {
		s.logger.Warn("failed to extract message IDs from meilisearch, falling back to PostgreSQL",
			zap.Error(err),
		)
		return s.repo.Search(ctx, filters)
	}
	
	if len(messageIDs) == 0 {
		return &SearchResult{
			Telegrams: []domain.Telegram{},
			Total:     0,
			Page:      filters.Pagination,
		}, nil
	}
	
	// Fetch full telegram data from PostgreSQL using message IDs
	pgFilters := filters
	pgFilters.Query = ""
	pgFilters.MessageIDs = messageIDs
	
	result, err := s.repo.Search(ctx, pgFilters)
	if err != nil {
		return nil, fmt.Errorf("fetch telegrams from PostgreSQL: %w", err)
	}
	
	result.Total = total
	return result, nil
}

// buildSortArray builds Meilisearch sort array from pagination
func (s *SearchService) buildSortArray(pagination domain.Pagination) []string {
	if pagination.SortBy == "" {
		return nil
	}
	order := "asc"
	if pagination.Order == "desc" {
		order = "desc"
	}
	return []string{fmt.Sprintf("%s:%s", pagination.SortBy, order)}
}

// buildMeilisearchFilter builds a Meilisearch filter string from SearchFilters
func (s *SearchService) buildMeilisearchFilter(filters SearchFilters) string {
	var conditions []string
	
	// Helper to build IN clause or equality
	buildFilter := func(field string, values []string) string {
		if len(values) == 0 {
			return ""
		}
		if len(values) == 1 {
			return fmt.Sprintf("%s = '%s'", field, values[0])
		}
		quoted := make([]string, len(values))
		for i, v := range values {
			quoted[i] = fmt.Sprintf("'%s'", v)
		}
		return fmt.Sprintf("%s IN [%s]", field, strings.Join(quoted, ", "))
	}
	
	buildIntFilter := func(field string, values []int) string {
		if len(values) == 0 {
			return ""
		}
		if len(values) == 1 {
			return fmt.Sprintf("%s = %d", field, values[0])
		}
		strs := make([]string, len(values))
		for i, v := range values {
			strs[i] = fmt.Sprintf("%d", v)
		}
		return fmt.Sprintf("%s IN [%s]", field, strings.Join(strs, ", "))
	}
	
	if filter := buildFilter("type", filters.Types); filter != "" {
		conditions = append(conditions, filter)
	}
	if filter := buildFilter("source", filters.Sources); filter != "" {
		conditions = append(conditions, filter)
	}
	if filter := buildFilter("destination", filters.Destinations); filter != "" {
		conditions = append(conditions, filter)
	}
	if filter := buildIntFilter("priority", filters.Priorities); filter != "" {
		conditions = append(conditions, filter)
	}
	
	if !filters.TimeRange.Start.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time >= %d", filters.TimeRange.Start.Unix()))
	}
	if !filters.TimeRange.End.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time <= %d", filters.TimeRange.End.Unix()))
	}
	
	if len(conditions) == 0 {
		return ""
	}
	return strings.Join(conditions, " AND ")
}

// extractMessageIDsFromMeilisearch extracts message IDs and total from Meilisearch result
func (s *SearchService) extractMessageIDsFromMeilisearch(meiliResult interface{}) ([]string, int64, error) {
	resultMap, ok := meiliResult.(map[string]interface{})
	if !ok {
		// Fallback to JSON marshaling if type assertion fails
		resultBytes, err := json.Marshal(meiliResult)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal meilisearch result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
			return nil, 0, fmt.Errorf("unmarshal meilisearch result: %w", err)
		}
	}
	
	// Extract total
	var total int64
	if estimatedTotal, ok := resultMap["estimatedTotalHits"].(float64); ok {
		total = int64(estimatedTotal)
	} else if totalHits, ok := resultMap["totalHits"].(float64); ok {
		total = int64(totalHits)
	}
	
	// Extract hits
	hits, ok := resultMap["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return []string{}, total, nil
	}
	
	messageIDs := make([]string, 0, len(hits))
	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		if messageID, ok := hitMap["message_id"].(string); ok && messageID != "" {
			messageIDs = append(messageIDs, messageID)
		}
	}
	
	return messageIDs, total, nil
}

// buildCacheKey creates a deterministic cache key for search filters
func (s *SearchService) buildCacheKey(filters SearchFilters) string {
	var sb strings.Builder
	sb.WriteString("search:")
	sb.WriteString(filters.Query)

	// Add filter components
	if len(filters.Types) > 0 {
		sb.WriteString(":types:")
		sb.WriteString(fmt.Sprintf("%v", filters.Types))
	}
	if len(filters.Sources) > 0 {
		sb.WriteString(":sources:")
		sb.WriteString(fmt.Sprintf("%v", filters.Sources))
	}
	if len(filters.Destinations) > 0 {
		sb.WriteString(":dests:")
		sb.WriteString(fmt.Sprintf("%v", filters.Destinations))
	}
	if len(filters.Priorities) > 0 {
		sb.WriteString(":priorities:")
		sb.WriteString(fmt.Sprintf("%v", filters.Priorities))
	}

	// Add time range
	if !filters.TimeRange.Start.IsZero() {
		sb.WriteString(":start:")
		sb.WriteString(fmt.Sprintf("%d", filters.TimeRange.Start.Unix()))
	}
	if !filters.TimeRange.End.IsZero() {
		sb.WriteString(":end:")
		sb.WriteString(fmt.Sprintf("%d", filters.TimeRange.End.Unix()))
	}

	// Add pagination
	sb.WriteString(":limit:")
	sb.WriteString(fmt.Sprintf("%d", filters.Pagination.Limit))
	sb.WriteString(":offset:")
	sb.WriteString(fmt.Sprintf("%d", filters.Pagination.Offset))
	sb.WriteString(":sort:")
	sb.WriteString(filters.Pagination.SortBy)
	sb.WriteString(":")
	sb.WriteString(filters.Pagination.Order)

	// Use FNV-1a for fast, non-cryptographic cache key hashing
	h := fnv.New64a()
	h.Write([]byte(sb.String()))
	return fmt.Sprintf("search:%016x", h.Sum64())
}
