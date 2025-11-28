package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// CacheMock is a mock implementation of cache.Store
type CacheMock struct {
	mock.Mock
}

// Set mocks the Set method
func (m *CacheMock) Set(ctx context.Context, key string, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

// Get mocks the Get method
func (m *CacheMock) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
