package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
)

// QueryServiceMock implements QueryService interface
// (interface check removed to avoid import cycle with internal/application)

// QueryServiceMock is a mock implementation of application.QueryService
type QueryServiceMock struct {
	mock.Mock
}

// Search mocks the Search method
func (m *QueryServiceMock) Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*persistence.SearchResult), args.Error(1)
}

// Recent mocks the Recent method
func (m *QueryServiceMock) Recent(ctx context.Context, limit int) ([]*domain.Telegram, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Telegram), args.Error(1)
}
