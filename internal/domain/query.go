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
	TotalMessages int64
	ByType        map[string]int64
	ByPriority    map[int]int64
}

// RouteStat represents traffic statistics for a route.
type RouteStat struct {
	Source      string
	Destination string
	Count       int64
}

// ExportFormat enumerates supported export formats.
type ExportFormat string

const (
	ExportFormatCSV ExportFormat = "csv"
)
