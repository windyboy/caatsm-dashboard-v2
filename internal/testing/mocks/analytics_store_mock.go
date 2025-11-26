package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/repository"
)

// AnalyticsStoreMock is a mock implementation of repository.AnalyticsStore
type AnalyticsStoreMock struct {
	mock.Mock
}

// Ensure AnalyticsStoreMock implements repository.AnalyticsStore
var _ repository.AnalyticsStore = (*AnalyticsStoreMock)(nil)

// TrafficSummary mocks the TrafficSummary method
func (m *AnalyticsStoreMock) TrafficSummary(ctx context.Context, window domain.TimeWindow) (*domain.TrafficSummary, error) {
	args := m.Called(ctx, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TrafficSummary), args.Error(1)
}

// RouteStats mocks the RouteStats method
func (m *AnalyticsStoreMock) RouteStats(ctx context.Context, limit int) ([]domain.RouteStat, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RouteStat), args.Error(1)
}

