package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

// queryService implements QueryService interface
type queryService struct {
	store repository.TelegramStore
	index repository.SearchIndex
}

// NewQueryService creates a new QueryService
func NewQueryService(
	store repository.TelegramStore,
	index repository.SearchIndex,
) QueryService {
	return &queryService{
		store: store,
		index: index,
	}
}

// Search performs a search query using the search index
func (s *queryService) Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error) {
	filterStr := buildFilterString(filter)

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

	rawResult, err := s.index.Search(ctx, filter.Query, filterStr, int64(filter.Page.Limit), int64(filter.Page.Offset), sort)
	if err != nil {
		return nil, fmt.Errorf("search index query: %w", err)
	}

	result := &persistence.SearchResult{
		Telegrams: []persistence.Telegram{},
		Total:     0,
		Page:      filter.Page,
	}

	_ = rawResult

	return result, nil
}

// Recent retrieves recent telegrams from the store
func (s *queryService) Recent(ctx context.Context, limit int) ([]*domain.Telegram, error) {
	filter := persistence.SearchFilter{
		Page: persistence.Pagination{
			Limit:  limit,
			Offset: 0,
			SortBy: "time",
			Order:  "desc",
		},
	}

	result, err := s.store.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search recent telegrams: %w", err)
	}

	telegrams := make([]*domain.Telegram, len(result.Telegrams))
	for i, tg := range result.Telegrams {
		telegrams[i] = domain.ToDomain(&tg)
	}

	return telegrams, nil
}

// buildFilterString builds a Meilisearch filter string from SearchFilter
func buildFilterString(filter persistence.SearchFilter) string {
	var parts []string

	if len(filter.Type) > 0 {
		typeParts := make([]string, len(filter.Type))
		for i, t := range filter.Type {
			typeParts[i] = fmt.Sprintf(`type = "%s"`, t)
		}
		parts = append(parts, "("+strings.Join(typeParts, " OR ")+")")
	}

	if len(filter.Source) > 0 {
		sourceParts := make([]string, len(filter.Source))
		for i, s := range filter.Source {
			sourceParts[i] = fmt.Sprintf(`source = "%s"`, s)
		}
		parts = append(parts, "("+strings.Join(sourceParts, " OR ")+")")
	}

	if len(filter.Destination) > 0 {
		destParts := make([]string, len(filter.Destination))
		for i, d := range filter.Destination {
			destParts[i] = fmt.Sprintf(`destination = "%s"`, d)
		}
		parts = append(parts, "("+strings.Join(destParts, " OR ")+")")
	}

	if len(filter.Priority) > 0 {
		priorityParts := make([]string, len(filter.Priority))
		for i, p := range filter.Priority {
			priorityParts[i] = fmt.Sprintf("priority = %d", p)
		}
		parts = append(parts, "("+strings.Join(priorityParts, " OR ")+")")
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " AND ")
}
