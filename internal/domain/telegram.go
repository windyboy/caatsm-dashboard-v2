package domain

import (
	"strings"
	"time"
)

// Telegram represents a domain entity for aviation telegrams
type Telegram struct {
	MessageID    string
	Type         string
	Time         time.Time
	FlightNumber string
	Source       string
	Destination  string
	Priority     int
	Content      string
	RawData      string
}

// Validate validates the telegram according to domain rules.
func (t *Telegram) Validate() error {
	if t.MessageID == "" {
		return ErrInvalidTelegram{Field: "message_id", Reason: "cannot be empty"}
	}
	
	// Validate message ID format (should not contain control characters)
	if strings.ContainsAny(t.MessageID, "\n\r\t") {
		return ErrInvalidTelegram{Field: "message_id", Reason: "contains invalid characters"}
	}
	
	if t.Time.IsZero() {
		return ErrInvalidTelegram{Field: "time", Reason: "cannot be zero"}
	}
	
	// Validate time is not in the future
	now := time.Now()
	if t.Time.After(now.Add(1 * time.Minute)) {
		return ErrInvalidTelegram{Field: "time", Reason: "cannot be in the future"}
	}
	
	if t.Type == "" {
		return ErrInvalidTelegram{Field: "type", Reason: "cannot be empty"}
	}
	
	// Validate priority range
	if t.Priority < 1 || t.Priority > 3 {
		return ErrInvalidTelegram{Field: "priority", Reason: "must be between 1 and 3"}
	}
	
	// Validate route if both source and destination are provided
	if t.Source != "" && t.Destination != "" {
		if t.Source == t.Destination {
			return ErrInvalidTelegram{Field: "route", Reason: "source and destination cannot be the same"}
		}
	}
	
	return nil
}

// IsHighPriority returns true if the telegram has high priority (priority 1)
func (t *Telegram) IsHighPriority() bool {
	return t.Priority == 1
}

// ExtractRoute returns a formatted route string (source → destination)
func (t *Telegram) ExtractRoute() string {
	if t.Source != "" && t.Destination != "" {
		return t.Source + " → " + t.Destination
	}
	if t.Source != "" {
		return t.Source
	}
	if t.Destination != "" {
		return t.Destination
	}
	return ""
}

// Normalize normalizes the telegram data (e.g., uppercase flight numbers, trim whitespace)
func (t *Telegram) Normalize() {
	// Uppercase flight number
	if t.FlightNumber != "" {
		t.FlightNumber = strings.ToUpper(strings.TrimSpace(t.FlightNumber))
	}
	
	// Uppercase and trim ICAO codes (source, destination)
	if t.Source != "" {
		t.Source = strings.ToUpper(strings.TrimSpace(t.Source))
	}
	if t.Destination != "" {
		t.Destination = strings.ToUpper(strings.TrimSpace(t.Destination))
	}
	
	// Trim whitespace from content
	t.Content = strings.TrimSpace(t.Content)
	
	// Uppercase message type
	if t.Type != "" {
		t.Type = strings.ToUpper(strings.TrimSpace(t.Type))
	}
	
	// Trim message ID
	t.MessageID = strings.TrimSpace(t.MessageID)
}
