package repository

import "time"

// TimeWindow describes a time range used for analytics queries.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// TrafficSummary contains aggregated metrics for dashboard visualisations.
type TrafficSummary struct {
	TotalMessages int64            `json:"total"`
	ByType        map[string]int64 `json:"byType"`
	ByPriority    map[int]int64     `json:"byPriority"`
}
