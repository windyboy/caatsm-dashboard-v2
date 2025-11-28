package persistence

import "time"

// Telegram represents a telegram message as stored in the database and serialized for API responses.
// This is a persistence/transport concern, distinct from the domain entity.
// JSON tags are included for API serialization.
type Telegram struct {
	MessageID    string    `json:"message_id"`
	Type         string    `json:"type"`
	Time         time.Time `json:"time"`
	FlightNumber string    `json:"flight_number"`
	Source       string    `json:"source"`
	Destination  string    `json:"destination"`
	Priority     int       `json:"priority"`
	Content      string    `json:"content"`
	RawData      string    `json:"raw_data"`
}

// SearchFilter captures filters available to the search API.
// This is a transport/API concern for request/response handling.
type SearchFilter struct {
	Query       string
	Type        []string
	Source      []string
	Destination []string
	Priority    []int
	TimeRange   TimeWindow
	Page        Pagination
}

// Pagination parameters for list endpoints.
type Pagination struct {
	Limit  int
	Offset int
	SortBy string
	Order  string // asc/desc
}

// SearchResult wraps the outcome of a search query.
type SearchResult struct {
	Telegrams []Telegram
	Total     int64
	Page      Pagination
}

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

// RouteStat indicates the number of telegrams between a source/destination pair.
type RouteStat struct {
	Source      string
	Destination string
	Count       int64
}

// ExportFormat enumerates export formats.
type ExportFormat string

const (
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatExcel ExportFormat = "xlsx"
	ExportFormatPDF   ExportFormat = "pdf"
)
