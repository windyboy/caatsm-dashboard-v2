package domain

import (
	"encoding/json"
	"fmt"
)

// Parser defines the interface for parsing raw telegram data into domain entities
// This allows for plugin-based parsing for different telegram formats
type Parser interface {
	// Parse parses raw data into a Telegram domain entity
	Parse(rawData string) (*Telegram, error)
	// Supports checks if this parser supports the given telegram type
	Supports(telegramType string) bool
}

// DefaultParser is the default JSON-based parser
type DefaultParser struct{}

// NewDefaultParser creates a new default parser
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse parses JSON raw data into a Telegram
func (p *DefaultParser) Parse(rawData string) (*Telegram, error) {
	var telegram Telegram
	if err := json.Unmarshal([]byte(rawData), &telegram); err != nil {
		return nil, fmt.Errorf("failed to parse telegram: %w", err)
	}
	return &telegram, nil
}

// Supports returns true for all types (default parser handles JSON)
func (p *DefaultParser) Supports(telegramType string) bool {
	return true
}

