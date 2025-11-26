package domain

import "time"

// SearchFilters represents domain-level search criteria for telegrams.
// This is the domain concept, separate from infrastructure concerns like JSON tags.
type SearchFilters struct {
	Query       string
	Types       []string
	Sources     []string
	Destinations []string
	Priorities  []int
	TimeRange   TimeWindow
	Pagination  Pagination
}

// TimeWindow defines a time range for filtering telegrams.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// Pagination controls result set size and ordering.
type Pagination struct {
	Limit  int
	Offset int
	SortBy string
	Order  string // "asc" or "desc"
}

// Validate ensures the search filters are valid.
func (f *SearchFilters) Validate() error {
	if f.Pagination.Limit < 0 {
		return ErrInvalidFilter{Field: "limit", Reason: "must be non-negative"}
	}
	if f.Pagination.Offset < 0 {
		return ErrInvalidFilter{Field: "offset", Reason: "must be non-negative"}
	}
	if !f.TimeRange.Start.IsZero() && !f.TimeRange.End.IsZero() {
		if f.TimeRange.Start.After(f.TimeRange.End) {
			return ErrInvalidFilter{Field: "time_range", Reason: "start must be before end"}
		}
	}
	if f.Pagination.SortBy != "" {
		allowedSortFields := map[string]bool{
			"time":          true,
			"priority":      true,
			"message_id":    true,
			"type":          true,
			"flight_number": true,
			"source":        true,
			"destination":   true,
		}
		if !allowedSortFields[f.Pagination.SortBy] {
			return ErrInvalidFilter{Field: "sort_by", Reason: "invalid sort field"}
		}
	}
	if f.Pagination.Order != "" && f.Pagination.Order != "asc" && f.Pagination.Order != "desc" {
		return ErrInvalidFilter{Field: "order", Reason: "must be 'asc' or 'desc'"}
	}
	return nil
}

// DefaultPagination returns sensible defaults for pagination.
func DefaultPagination() Pagination {
	return Pagination{
		Limit:  50,
		Offset: 0,
		SortBy: "time",
		Order:  "desc",
	}
}

// ErrInvalidFilter represents a validation error in search filters.
type ErrInvalidFilter struct {
	Field  string
	Reason string
}

func (e ErrInvalidFilter) Error() string {
	return "invalid filter: " + e.Field + " " + e.Reason
}

