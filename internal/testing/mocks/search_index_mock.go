package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/app"
)

// SearchIndexMock is a mock implementation of app.SearchIndex
type SearchIndexMock struct {
	mock.Mock
}

// Ensure SearchIndexMock implements app.SearchIndex
var _ app.SearchIndex = (*SearchIndexMock)(nil)

// EnsureIndex mocks the EnsureIndex method
func (m *SearchIndexMock) EnsureIndex(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Index mocks the Index method
func (m *SearchIndexMock) Index(ctx context.Context, telegram *app.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

// BulkIndex mocks the BulkIndex method
func (m *SearchIndexMock) BulkIndex(ctx context.Context, telegrams []*app.Telegram) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *SearchIndexMock) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

// Search mocks the Search method
func (m *SearchIndexMock) Search(ctx context.Context, query string, filter string, limit, offset int64, sort []string) (any, error) {
	args := m.Called(ctx, query, filter, limit, offset, sort)
	return args.Get(0), args.Error(1)
}

// SearchAutocomplete mocks the SearchAutocomplete method
func (m *SearchIndexMock) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (*app.AutocompleteResponse, error) {
	args := m.Called(ctx, query, limit, attributes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*app.AutocompleteResponse), args.Error(1)
}
