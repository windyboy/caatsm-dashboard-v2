package application

import (
	"context"

	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/models"
)

// TelegramService defines the interface for telegram command operations
type TelegramService interface {
	// SaveTelegram saves a telegram to storage and indexes it
	SaveTelegram(ctx context.Context, telegram *domain.Telegram) error
}

// QueryService defines the interface for query operations
type QueryService interface {
	// Search performs a search query
	Search(ctx context.Context, filter models.SearchFilter) (*models.SearchResult, error)
	// Recent retrieves recent telegrams
	Recent(ctx context.Context, limit int) ([]*domain.Telegram, error)
}

