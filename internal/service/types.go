package service

import (
	"github.com/windy/caatsm-dashboard/internal/app"
	"github.com/windy/caatsm-dashboard/internal/domain"
)

// Type aliases for domain types used in services
type (
	Telegram       = domain.Telegram
	SearchFilters  = domain.SearchFilters
	SearchResult   = domain.SearchResult
	TimeWindow     = domain.TimeWindow
	TrafficSummary = domain.TrafficSummary
	RouteStat      = domain.RouteStat
	ExportFormat   = domain.ExportFormat
)

// Interface definitions that reference app ports
type (
	Repository     = app.Repository
	Cache          = app.Cache
	SearchIndex    = app.SearchIndex
	EventPublisher = app.EventPublisher
)

// WSMessage re-exported from app package for use in services
type WSMessage = app.WSMessage

// AutocompleteSuggestion represents a single autocomplete suggestion with its type
type AutocompleteSuggestion struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// Domain event types re-exported
type (
	TelegramReceived = domain.TelegramReceived
	SearchPerformed  = domain.SearchPerformed
)

// Re-export domain constants
const (
	MaxExportRecords = domain.MaxExportRecords
	ExportFormatCSV  = domain.ExportFormatCSV
)

