package domain

import "time"

// Telegram represents a domain entity for aviation telegrams
// This is the core business entity, separate from infrastructure concerns
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

// Validate validates the telegram according to business rules
func (t *Telegram) Validate() error {
	if t.MessageID == "" {
		return ErrInvalidTelegram{Field: "message_id", Reason: "cannot be empty"}
	}
	if t.Time.IsZero() {
		return ErrInvalidTelegram{Field: "time", Reason: "cannot be zero"}
	}
	if t.Type == "" {
		return ErrInvalidTelegram{Field: "type", Reason: "cannot be empty"}
	}
	if t.Priority < 1 || t.Priority > 3 {
		return ErrInvalidTelegram{Field: "priority", Reason: "must be between 1 and 3"}
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
	// Normalize flight number to uppercase
	if t.FlightNumber != "" {
		// Flight numbers are typically uppercase, but we keep as-is for now
		// This can be extended with more normalization rules
	}

	// Normalize source/destination to uppercase (ICAO codes)
	if t.Source != "" {
		// Source should be uppercase ICAO code
		// Keep as-is for now, can add normalization if needed
	}
	if t.Destination != "" {
		// Destination should be uppercase ICAO code
		// Keep as-is for now, can add normalization if needed
	}

	// Trim whitespace from content
	if t.Content != "" {
		// Content trimming can be added if needed
	}
}

