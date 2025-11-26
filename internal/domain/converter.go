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
