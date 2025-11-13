package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// TelegramServiceMock is a mock implementation of application.TelegramService
// Note: We don't import application package to avoid import cycles
type TelegramServiceMock struct {
	mock.Mock
}

// SaveTelegram mocks the SaveTelegram method
func (m *TelegramServiceMock) SaveTelegram(ctx context.Context, telegram *domain.Telegram) error {
	args := m.Called(ctx, telegram)
	return args.Error(0)
}

