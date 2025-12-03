package domain

// SearchFilter captures filters for searching telegrams.
// This is a domain-level query object used by the application layer.
type SearchFilter struct {
	Query       string
	Type        []string
	Source      []string
	Destination []string
	Priority    []int
	TimeRange   TimeWindow
	Page        Pagination
}

// SearchResult wraps the result of a search query.
type SearchResult struct {
	Telegrams []Telegram
	Total     int64
	Page      Pagination
}

// TrafficSummary contains aggregated traffic metrics.
type TrafficSummary struct {
	TotalMessages int64            `json:"total"`
	ByType        map[string]int64 `json:"byType"`
	ByPriority    map[int]int64     `json:"byPriority"`
}

// RouteStat represents traffic statistics for a route.
type RouteStat struct {
	Source      string
	Destination string
	Count       int64
}

// HistoricalDataPoint represents a single data point in historical statistics.
type HistoricalDataPoint struct {
	Time  string `json:"time"`  // ISO 8601 timestamp
	Count int64  `json:"count"` // Message count for this time period
}

// HistoricalStats contains time-series data for message counts.
type HistoricalStats struct {
	Interval string                `json:"interval"` // "hour", "day", etc.
	Data     []HistoricalDataPoint `json:"data"`
}

// ExportFormat enumerates supported export formats.
type ExportFormat string

const (
	ExportFormatCSV ExportFormat = "csv"
)
