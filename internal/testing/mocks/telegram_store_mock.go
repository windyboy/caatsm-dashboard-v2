package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

// TelegramStoreMock is a mock implementation of repository.TelegramStore
type TelegramStoreMock struct {
	mock.Mock
}

// Ensure TelegramStoreMock implements repository.TelegramStore
var _ repository.TelegramStore = (*TelegramStoreMock)(nil)

// Save mocks the Save method
func (m *TelegramStoreMock) Save(ctx context.Context, telegram *persistence.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

// BulkSave mocks the BulkSave method
func (m *TelegramStoreMock) BulkSave(ctx context.Context, telegrams []*persistence.Telegram) error {
	args := m.Called(ctx, telegrams)
	return args.Error(0)
}

// Search mocks the Search method
func (m *TelegramStoreMock) Search(ctx context.Context, filter persistence.SearchFilter) (*persistence.SearchResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*persistence.SearchResult), args.Error(1)
}
