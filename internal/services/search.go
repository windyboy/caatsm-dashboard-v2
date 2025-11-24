package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/repository"
	"github.com/windy/caatsm-dashboard/internal/repository/cache"
	"go.uber.org/zap"
)

// searchService implements SearchService using Meilisearch and caching.
type searchService struct {
	index  repository.SearchIndex
	store  repository.TelegramStore
	cache  *cache.Store
	logger *zap.Logger
}

// NewSearchService creates a new SearchService implementation.
func NewSearchService(
	index repository.SearchIndex,
	store repository.TelegramStore,
	cache *cache.Store,
	logger *zap.Logger,
) SearchService {
	return &searchService{
		index:  index,
		store:  store,
		cache:  cache,
		logger: logger,
	}
}

// Search performs a search query using Meilisearch.
func (s *searchService) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
	// Check cache first
	cacheKey := s.buildCacheKey(filter)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit - deserialize and return cached result
		var cachedResult models.SearchResult
		if err := json.Unmarshal(cached, &cachedResult); err == nil {
			s.logger.Debug("cache hit", zap.String("key", cacheKey))
			return &cachedResult, nil
		}
		// If deserialization fails, log warning and continue to fresh query
		s.logger.Warn("cache deserialization failed", zap.String("key", cacheKey), zap.Error(err))
	} else {
		s.logger.Debug("cache miss", zap.String("key", cacheKey))
	}

	// Build filters
	var filterParts []string

	if len(filter.Type) > 0 {
		typeFilters := make([]string, len(filter.Type))
		for i, t := range filter.Type {
			typeFilters[i] = fmt.Sprintf("type = %q", t)
		}
		filterParts = append(filterParts, "("+strings.Join(typeFilters, " OR ")+")")
	}

	if len(filter.Source) > 0 {
		sourceFilters := make([]string, len(filter.Source))
		for i, src := range filter.Source {
			sourceFilters[i] = fmt.Sprintf("source = %q", src)
		}
		filterParts = append(filterParts, "("+strings.Join(sourceFilters, " OR ")+")")
	}

	if len(filter.Destination) > 0 {
		destFilters := make([]string, len(filter.Destination))
		for i, dst := range filter.Destination {
			destFilters[i] = fmt.Sprintf("destination = %q", dst)
		}
		filterParts = append(filterParts, "("+strings.Join(destFilters, " OR ")+")")
	}

	if len(filter.Priority) > 0 {
		priorityFilters := make([]string, len(filter.Priority))
		for i, p := range filter.Priority {
			priorityFilters[i] = fmt.Sprintf("priority = %d", p)
		}
		filterParts = append(filterParts, "("+strings.Join(priorityFilters, " OR ")+")")
	}

	// Apply time range
	if !filter.TimeRange.Start.IsZero() || !filter.TimeRange.End.IsZero() {
		var timeFilters []string
		if !filter.TimeRange.Start.IsZero() {
			timeFilters = append(timeFilters, fmt.Sprintf("time >= %d", filter.TimeRange.Start.Unix()))
		}
		if !filter.TimeRange.End.IsZero() {
			timeFilters = append(timeFilters, fmt.Sprintf("time <= %d", filter.TimeRange.End.Unix()))
		}
		if len(timeFilters) > 0 {
			filterParts = append(filterParts, "("+strings.Join(timeFilters, " AND ")+")")
		}
	}

	// Combine all filters
	filterStr := ""
	if len(filterParts) > 0 {
		filterStr = strings.Join(filterParts, " AND ")
	}

	// Build sort
	var sort []string
	if filter.Page.SortBy != "" {
		sortBy := filter.Page.SortBy
		if filter.Page.Order == "ASC" {
			sortBy += ":asc"
		} else {
			sortBy += ":desc"
		}
		sort = []string{sortBy}
	}

	// Execute search via interface
	rawResult, err := s.index.Search(ctx, filter.Query, filterStr, int64(filter.Page.Limit), int64(filter.Page.Offset), sort)
	if err != nil {
		return nil, fmt.Errorf("search index query: %w", err)
	}

	// Type assert to Meilisearch result
	meiliResult, ok := rawResult.(*meilisearch.SearchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected search result type")
	}

	// Convert Meilisearch results to our model
	telegrams := make([]models.Telegram, 0, len(meiliResult.Hits))
	for _, hit := range meiliResult.Hits {
		var telegram models.Telegram
		// Parse hit into telegram struct using JSON marshaling
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			s.logger.Warn("failed to marshal hit", zap.Error(err))
			continue
		}

		var data map[string]any
		if err := json.Unmarshal(hitBytes, &data); err != nil {
			s.logger.Warn("failed to unmarshal hit", zap.Error(err))
			continue
		}

		if msgID, ok := data["message_id"].(string); ok {
			telegram.MessageID = msgID
		}
		if t, ok := data["type"].(string); ok {
			telegram.Type = t
		}
		if ts, ok := data["time"].(float64); ok {
			telegram.Time = time.Unix(int64(ts), 0)
		}
		if fn, ok := data["flight_number"].(string); ok {
			telegram.FlightNumber = fn
		}
		if src, ok := data["source"].(string); ok {
			telegram.Source = src
		}
		if dst, ok := data["destination"].(string); ok {
			telegram.Destination = dst
		}
		if content, ok := data["content"].(string); ok {
			telegram.Content = content
		}
		if priority, ok := data["priority"].(float64); ok {
			telegram.Priority = int(priority)
		}
		telegrams = append(telegrams, telegram)
	}

	// Get total from result
	total := meiliResult.EstimatedTotalHits
	if total <= 0 {
		// Fallback to totalHits if EstimatedTotalHits is not available
		total = meiliResult.TotalHits
	}
	if total <= 0 {
		// Fallback to hits length if both are not available
		total = int64(len(meiliResult.Hits))
	}
	if total == 0 {
		total = int64(len(telegrams))
	}

	searchResult := &models.SearchResult{
		Telegrams: telegrams,
		Total:     total,
		Page:      filter.Page,
	}

	// Cache result
	if resultBytes, err := json.Marshal(searchResult); err == nil {
		if err := s.cache.Set(ctx, cacheKey, resultBytes); err != nil {
			// Log error but don't fail the request
			s.logger.Warn("failed to cache result", zap.String("key", cacheKey), zap.Error(err))
		} else {
			s.logger.Debug("cached result", zap.String("key", cacheKey))
		}
	} else {
		s.logger.Warn("failed to marshal result for caching", zap.Error(err))
	}

	return searchResult, nil
}

// Autocomplete provides autocomplete suggestions using Meilisearch.
func (s *searchService) Autocomplete(ctx context.Context, term string, size int) ([]string, error) {
	if term == "" {
		return []string{}, nil
	}

	attributes := []string{"flight_number", "message_id", "source", "destination"}
	searchResult, err := s.index.SearchAutocomplete(ctx, term, int64(size), attributes)
	if err != nil {
		return nil, fmt.Errorf("autocomplete query: %w", err)
	}

	// Type assert to Meilisearch result
	result, ok := searchResult.(*meilisearch.SearchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected autocomplete result type")
	}

	suggestions := make([]string, 0, len(result.Hits))
	seen := make(map[string]bool)

	for _, hit := range result.Hits {
		// Parse hit using JSON marshaling
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			continue
		}

		var data map[string]any
		if err := json.Unmarshal(hitBytes, &data); err != nil {
			continue
		}

		// Extract flight_number, message_id, source, destination
		if fn, ok := data["flight_number"].(string); ok && fn != "" && !seen[fn] {
			suggestions = append(suggestions, fn)
			seen[fn] = true
		}
		if msgID, ok := data["message_id"].(string); ok && msgID != "" && !seen[msgID] {
			suggestions = append(suggestions, msgID)
			seen[msgID] = true
		}
		if src, ok := data["source"].(string); ok && src != "" && !seen[src] {
			suggestions = append(suggestions, src)
			seen[src] = true
		}
		if dst, ok := data["destination"].(string); ok && dst != "" && !seen[dst] {
			suggestions = append(suggestions, dst)
			seen[dst] = true
		}

		if len(suggestions) >= size {
			break
		}
	}

	return suggestions, nil
}

// buildCacheKey generates a deterministic cache key from search filter parameters.
//
// Cache key format: "search:{query}:{types}:{sources}:{destinations}:{priorities}:{start}:{end}:{limit}:{offset}:{sortBy}:{order}"
//
// This ensures that identical search parameters produce the same cache key,
// enabling effective cache hits for repeated queries.
func (s *searchService) buildCacheKey(filter models.SearchFilter) string {
	// Build deterministic cache key from all filter parameters
	key := fmt.Sprintf("search:%s", filter.Query)

	// Add filter arrays
	if len(filter.Type) > 0 {
		key += fmt.Sprintf(":types:%v", filter.Type)
	}
	if len(filter.Source) > 0 {
		key += fmt.Sprintf(":sources:%v", filter.Source)
	}
	if len(filter.Destination) > 0 {
		key += fmt.Sprintf(":dests:%v", filter.Destination)
	}
	if len(filter.Priority) > 0 {
		key += fmt.Sprintf(":priorities:%v", filter.Priority)
	}

	// Add time range
	if !filter.TimeRange.Start.IsZero() {
		key += fmt.Sprintf(":start:%d", filter.TimeRange.Start.Unix())
	}
	if !filter.TimeRange.End.IsZero() {
		key += fmt.Sprintf(":end:%d", filter.TimeRange.End.Unix())
	}

	// Add pagination
	key += fmt.Sprintf(":limit:%d:offset:%d", filter.Page.Limit, filter.Page.Offset)

	// Add sorting
	if filter.Page.SortBy != "" {
		key += fmt.Sprintf(":sort:%s:%s", filter.Page.SortBy, filter.Page.Order)
	}

	return key
}
