package meili

import (
	"context"
	"fmt"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/internal/models"
)

// Index implements SearchIndex using Meilisearch.
type Index struct {
	client meilisearch.ServiceManager
	index  string
}

// New creates an Index with the provided client and index name.
func New(client meilisearch.ServiceManager, index string) *Index {
	return &Index{
		client: client,
		index:  index,
	}
}

// EnsureIndex creates or updates the Meilisearch index with proper settings.
func (i *Index) EnsureIndex(ctx context.Context) error {
	idx := i.client.Index(i.index)

	// Configure searchable attributes
	searchableAttributes := []string{"content", "flight_number", "message_id", "source", "destination"}
	_, err := idx.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		return fmt.Errorf("update searchable attributes: %w", err)
	}

	// Configure filterable attributes - convert []string to []interface{}
	filterableAttributes := []string{"type", "source", "destination", "priority", "time"}
	filterableInterface := make([]interface{}, len(filterableAttributes))
	for j, v := range filterableAttributes {
		filterableInterface[j] = v
	}
	_, err = idx.UpdateFilterableAttributes(&filterableInterface)
	if err != nil {
		return fmt.Errorf("update filterable attributes: %w", err)
	}

	// Configure sortable attributes
	sortableAttributes := []string{"time", "priority"}
	_, err = idx.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		return fmt.Errorf("update sortable attributes: %w", err)
	}

	// Set primary key using UpdateIndex
	primaryKey := "message_id"
	_, err = idx.UpdateIndex(&meilisearch.UpdateIndexRequestParams{
		PrimaryKey: primaryKey,
	})
	if err != nil {
		// Ignore error if primary key is already set
		// In production, you might want to check the error type
		return fmt.Errorf("update primary key: %w", err)
	}

	return nil
}

// Index adds/updates a telegram document in Meilisearch.
func (i *Index) Index(ctx context.Context, telegram *models.Telegram) error {
	idx := i.client.Index(i.index)

	doc := map[string]interface{}{
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
	_, err := idx.AddDocuments([]map[string]interface{}{doc}, &primaryKey)
	if err != nil {
		return fmt.Errorf("index telegram: %w", err)
	}

	return nil
}

// BulkIndex ingests multiple telegrams.
func (i *Index) BulkIndex(ctx context.Context, telegrams []*models.Telegram) error {
	if len(telegrams) == 0 {
		return nil
	}

	idx := i.client.Index(i.index)

	docs := make([]map[string]interface{}, len(telegrams))
	for i, telegram := range telegrams {
		docs[i] = map[string]interface{}{
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
func (i *Index) Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (interface{}, error) {
	idx := i.client.Index(i.index)

	searchRequest := &meilisearch.SearchRequest{
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
func (i *Index) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (interface{}, error) {
	idx := i.client.Index(i.index)

	searchRequest := &meilisearch.SearchRequest{
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
