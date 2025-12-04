package repository

import "time"

// TimeWindow describes a time range used for analytics queries.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// TrafficSummary contains aggregated metrics for dashboard visualisations.
// NOTE: This type is deprecated. Use domain.TrafficSummary instead.
type TrafficSummary struct {
	TotalMessages  int64            `json:"total"`
	ByType         map[string]int64 `json:"byType"`
	ActiveRoutes   int64            `json:"activeRoutes"`
	MessagesPerSec float64          `json:"messagesPerSec"`
	TimeWindow     string           `json:"timeWindow"`
}
