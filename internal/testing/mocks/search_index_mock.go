package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/app/ports"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// SearchIndexMock is a mock implementation of ports.SearchIndex
type SearchIndexMock struct {
	mock.Mock
}

// Ensure SearchIndexMock implements ports.SearchIndex
var _ ports.SearchIndex = (*SearchIndexMock)(nil)

// Index mocks the Index method
func (m *SearchIndexMock) Index(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

// BulkIndex mocks the BulkIndex method
func (m *SearchIndexMock) BulkIndex(ctx context.Context, telegrams []*domain.Telegram) error {
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
func (m *SearchIndexMock) SearchAutocomplete(ctx context.Context, query string, limit int64, attributes []string) (any, error) {
	args := m.Called(ctx, query, limit, attributes)
	return args.Get(0), args.Error(1)
}
