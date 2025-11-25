package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/models"
	"github.com/windy/caatsm-dashboard/internal/services"
)

var _ services.StatsService = (*StatsServiceMock)(nil)

// StatsServiceMock is a mock implementation of services.StatsService
type StatsServiceMock struct {
	mock.Mock
}

// TrafficSummary mocks the TrafficSummary method
func (m *StatsServiceMock) TrafficSummary(ctx context.Context, window models.TimeWindow) (*models.TrafficSummary, error) {
	args := m.Called(ctx, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrafficSummary), args.Error(1)
}

// TopRoutes mocks the TopRoutes method
func (m *StatsServiceMock) TopRoutes(ctx context.Context, limit int) ([]models.RouteStat, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RouteStat), args.Error(1)
}

