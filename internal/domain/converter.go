package domain

import (
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
)

// ToDomain converts a persistence.Telegram to a domain.Telegram
func ToDomain(m *persistence.Telegram) *Telegram {
	if m == nil {
		return nil
	}
	return &Telegram{
		MessageID:    m.MessageID,
		Type:         m.Type,
		Time:         m.Time,
		FlightNumber: m.FlightNumber,
		Source:       m.Source,
		Destination:  m.Destination,
		Priority:     m.Priority,
		Content:      m.Content,
		RawData:      m.RawData,
	}
}

// FromDomain converts a domain.Telegram to a persistence.Telegram
func FromDomain(d *Telegram) *persistence.Telegram {
	if d == nil {
		return nil
	}
	return &persistence.Telegram{
		MessageID:    d.MessageID,
		Type:         d.Type,
		Time:         d.Time,
		FlightNumber: d.FlightNumber,
		Source:       d.Source,
		Destination:  d.Destination,
		Priority:     d.Priority,
		Content:      d.Content,
		RawData:      d.RawData,
	}
}

// SearchFilterToDomain converts persistence.SearchFilter to domain.SearchFilters
func SearchFilterToDomain(p persistence.SearchFilter) SearchFilters {
	return SearchFilters{
		Query:        p.Query,
		Types:        p.Type,
		Sources:      p.Source,
		Destinations: p.Destination,
		Priorities:   p.Priority,
		TimeRange: TimeWindow{
			Start: p.TimeRange.Start,
			End:   p.TimeRange.End,
		},
		Pagination: Pagination{
			Limit:  p.Page.Limit,
			Offset: p.Page.Offset,
			SortBy: p.Page.SortBy,
			Order:  p.Page.Order,
		},
	}
}

// TimeWindowToDomain converts persistence.TimeWindow to domain.TimeWindow
func TimeWindowToDomain(p persistence.TimeWindow) TimeWindow {
	return TimeWindow{
		Start: p.Start,
		End:   p.End,
	}
}

// ExportFormatToDomain converts persistence.ExportFormat to domain.ExportFormat
func ExportFormatToDomain(p persistence.ExportFormat) ExportFormat {
	return ExportFormat(p)
}
