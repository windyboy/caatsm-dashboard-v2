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
	index      repository.SearchIndex
	store      repository.TelegramStore
	cache      *cache.Store
	logger     *zap.Logger
	meili      meilisearch.ServiceManager
	meiliIndex string
}

// NewSearchService creates a new SearchService implementation.
func NewSearchService(
	index repository.SearchIndex,
	store repository.TelegramStore,
	cache *cache.Store,
	logger *zap.Logger,
	meili meilisearch.ServiceManager,
	meiliIndex string,
) SearchService {
	return &searchService{
		index:      index,
		store:      store,
		cache:      cache,
		logger:     logger,
		meili:      meili,
		meiliIndex: meiliIndex,
	}
}

// Search performs a search query using Meilisearch.
func (s *searchService) Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error) {
	// Check cache first
	cacheKey := s.buildCacheKey(filter)
	if _, err := s.cache.Get(ctx, cacheKey); err == nil {
		// Return cached result if available
		// For simplicity, we'll skip cache deserialization here
		// In production, you'd deserialize the cached JSON
	}

	// Build Meilisearch query
	idx := s.meili.Index(s.meiliIndex)

	searchRequest := &meilisearch.SearchRequest{
		Query:  filter.Query,
		Limit:  int64(filter.Page.Limit),
		Offset: int64(filter.Page.Offset),
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
	if len(filterParts) > 0 {
		filterStr := strings.Join(filterParts, " AND ")
		searchRequest.Filter = filterStr
	}

	// Apply sorting
	if filter.Page.SortBy != "" {
		sortBy := filter.Page.SortBy
		if filter.Page.Order == "ASC" {
			sortBy += ":asc"
		} else {
			sortBy += ":desc"
		}
		searchRequest.Sort = []string{sortBy}
	}

	// Execute search
	result, err := idx.Search(filter.Query, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("meilisearch query: %w", err)
	}

	// Convert Meilisearch results to our model
	telegrams := make([]models.Telegram, 0, len(result.Hits))
	for _, hit := range result.Hits {
		var telegram models.Telegram
		// Parse hit into telegram struct using JSON marshaling
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			s.logger.Warn("failed to marshal hit", zap.Error(err))
			continue
		}

		var data map[string]interface{}
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

	// Cache result
	// In production, you'd serialize the result to JSON and cache it
	// For now, we'll skip caching

	// Get total from result
	total := result.EstimatedTotalHits
	if total <= 0 {
		// Fallback to totalHits if EstimatedTotalHits is not available
		total = result.TotalHits
	}
	if total <= 0 {
		// Fallback to hits length if both are not available
		total = int64(len(result.Hits))
	}
	if total == 0 {
		total = int64(len(telegrams))
	}

	return &models.SearchResult{
		Telegrams: telegrams,
		Total:     total,
		Page:      filter.Page,
	}, nil
}

// Autocomplete provides autocomplete suggestions using Meilisearch.
func (s *searchService) Autocomplete(ctx context.Context, term string, size int) ([]string, error) {
	if term == "" {
		return []string{}, nil
	}

	idx := s.meili.Index(s.meiliIndex)

	searchRequest := &meilisearch.SearchRequest{
		Query:                term,
		Limit:                int64(size),
		AttributesToRetrieve: []string{"flight_number", "message_id", "source", "destination"},
	}

	result, err := idx.Search(term, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("autocomplete query: %w", err)
	}

	suggestions := make([]string, 0, len(result.Hits))
	seen := make(map[string]bool)

	for _, hit := range result.Hits {
		// Parse hit using JSON marshaling
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			continue
		}

		var data map[string]interface{}
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

func (s *searchService) buildCacheKey(filter models.SearchFilter) string {
	// Build a cache key from filter parameters
	// This is a simplified version; in production, you'd use a proper hash
	return fmt.Sprintf("search:%s:%v:%v:%v:%v", filter.Query, filter.Type, filter.Source, filter.Destination, filter.Priority)
}
