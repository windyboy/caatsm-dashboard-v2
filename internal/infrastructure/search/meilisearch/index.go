package meilisearch

import (
	"context"

	meilisearchClient "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/internal/domain"
	oldmeili "github.com/windy/caatsm-dashboard/internal/repository/meili"
)

// Index implements search indexing using Meilisearch.
// This is the new infrastructure layer implementation.
type Index struct {
	oldIndex *oldmeili.Index
}

// New creates a new Meilisearch index adapter.
func New(client meilisearchClient.ServiceManager, indexName string) *Index {
	return &Index{
		oldIndex: oldmeili.New(client, indexName),
	}
}

// EnsureIndex creates or updates the Meilisearch index with proper settings.
func (i *Index) EnsureIndex(ctx context.Context) error {
	return i.oldIndex.EnsureIndex(ctx)
}

// Index adds/updates a telegram document in Meilisearch.
func (i *Index) Index(ctx context.Context, telegram *domain.Telegram) error {
	return i.oldIndex.Index(ctx, telegram)
}

// BulkIndex indexes multiple telegrams.
func (i *Index) BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error {
	return i.oldIndex.BulkIndex(ctx, telegrams)
}

// Delete removes a telegram from the index.
func (i *Index) Delete(ctx context.Context, messageID string) error {
	return i.oldIndex.Delete(ctx, messageID)
}

// Search performs a search query.
func (i *Index) Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error) {
	return i.oldIndex.Search(ctx, query, filter, limit, offset, sort)
}

// SearchAutocomplete provides autocomplete suggestions.
func (i *Index) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error) {
	return i.oldIndex.SearchAutocomplete(ctx, query, limit, attributes)
}

