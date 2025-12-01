package repository

import "time"

// TimeWindow describes a time range used for analytics queries.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// TrafficSummary contains aggregated metrics for dashboard visualisations.
type TrafficSummary struct {
	TotalMessages int64
	ByType        map[string]int64
	ByPriority    map[int]int64
}
