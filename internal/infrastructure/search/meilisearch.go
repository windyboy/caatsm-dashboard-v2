package search

import (
	"context"
	"fmt"
	"net/http"
	"time"

	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// NewMeilisearchClient initialises a Meilisearch service manager with sensible defaults.
func NewMeilisearchClient(cfg config.SearchConfig) (meilisearchClient.ServiceManager, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("meilisearch host is empty")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	opts := []meilisearchClient.Option{
		meilisearchClient.WithCustomClient(httpClient),
	}
	if cfg.APIKey != "" {
		opts = append(opts, meilisearchClient.WithAPIKey(cfg.APIKey))
	}

	manager := meilisearchClient.New(cfg.Host, opts...)
	return manager, nil
}

// Index implements ports.SearchIndex using Meilisearch.
type Index struct {
	client meilisearchClient.ServiceManager
	index  string
}

// NewMeilisearchIndex creates an Index with the provided client and index name.
func NewMeilisearchIndex(client meilisearchClient.ServiceManager, index string) *Index {
	return &Index{
		client: client,
		index:  index,
	}
}

// Ensure Index implements ports.SearchIndex
var _ ports.SearchIndex = (*Index)(nil)

// EnsureIndex creates or updates the Meilisearch index with proper settings.
func (i *Index) EnsureIndex(ctx context.Context) error {
	idx := i.client.Index(i.index)

	searchableAttributes := []string{"content", "flight_number", "message_id", "source", "destination"}
	_, err := idx.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		return fmt.Errorf("update searchable attributes: %w", err)
	}

	filterableAttributes := []string{"type", "source", "destination", "priority", "time"}
	filterableInterface := make([]any, len(filterableAttributes))
	for j, v := range filterableAttributes {
		filterableInterface[j] = v
	}
	_, err = idx.UpdateFilterableAttributes(&filterableInterface)
	if err != nil {
		return fmt.Errorf("update filterable attributes: %w", err)
	}

	sortableAttributes := []string{"time", "priority"}
	_, err = idx.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		return fmt.Errorf("update sortable attributes: %w", err)
	}

	primaryKey := "message_id"
	_, err = idx.UpdateIndex(&meilisearchClient.UpdateIndexRequestParams{
		PrimaryKey: primaryKey,
	})
	if err != nil {
		return fmt.Errorf("update primary key: %w", err)
	}

	return nil
}

// Index adds/updates a telegram document in Meilisearch.
func (i *Index) Index(ctx context.Context, telegram *domain.Telegram) error {
	idx := i.client.Index(i.index)

	doc := map[string]any{
		"message_id":    telegram.MessageID,
		"type":          telegram.Type,
		"time":          telegram.Time.Unix(),
		"flight_number": telegram.FlightNumber,
		"source":        telegram.Source,
		"destination":   telegram.Destination,
		"content":       telegram.Content,
		"priority":      telegram.Priority,
	}

	primaryKey := "message_id"
	_, err := idx.AddDocuments([]map[string]any{doc}, &primaryKey)
	if err != nil {
		return fmt.Errorf("index telegram: %w", err)
	}

	return nil
}

// BulkIndex ingests multiple telegrams.
func (i *Index) BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error {
	if len(telegrams) == 0 {
		return nil
	}

	idx := i.client.Index(i.index)

	docs := make([]map[string]any, len(telegrams))
	for j, telegram := range telegrams {
		docs[j] = map[string]any{
			"message_id":    telegram.MessageID,
			"type":          telegram.Type,
			"time":          telegram.Time.Unix(),
			"flight_number": telegram.FlightNumber,
			"source":        telegram.Source,
			"destination":   telegram.Destination,
			"content":       telegram.Content,
			"priority":      telegram.Priority,
		}
	}

	primaryKey := "message_id"
	_, err := idx.AddDocuments(docs, &primaryKey)
	if err != nil {
		return fmt.Errorf("bulk index telegrams: %w", err)
	}

	return nil
}

// Delete removes a telegram from the index.
func (i *Index) Delete(ctx context.Context, messageID string) error {
	idx := i.client.Index(i.index)

	_, err := idx.DeleteDocument(messageID)
	if err != nil {
		return fmt.Errorf("delete telegram: %w", err)
	}

	return nil
}

// Search performs a search query on the Meilisearch index.
func (i *Index) Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error) {
	idx := i.client.Index(i.index)

	searchRequest := &meilisearchClient.SearchRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	if filter != "" {
		searchRequest.Filter = filter
	}

	if len(sort) > 0 {
		searchRequest.Sort = sort
	}

	result, err := idx.Search(query, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search: %w", err)
	}

	return result, nil
}

// SearchAutocomplete performs an autocomplete search on the Meilisearch index.
func (i *Index) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error) {
	idx := i.client.Index(i.index)

	searchRequest := &meilisearchClient.SearchRequest{
		Query:                query,
		Limit:                limit,
		AttributesToRetrieve: attributes,
	}

	result, err := idx.Search(query, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("meilisearch autocomplete: %w", err)
	}

	return result, nil
}
