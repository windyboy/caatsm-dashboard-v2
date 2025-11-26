package models

import (
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
)

// Re-export types from infrastructure/persistence for backward compatibility during migration.
// These will be removed once all code is migrated to use persistence package directly.
type (
	Telegram       = persistence.Telegram
	SearchFilter   = persistence.SearchFilter
	SearchResult   = persistence.SearchResult
	Pagination     = persistence.Pagination
	TimeWindow     = persistence.TimeWindow
	TrafficSummary = persistence.TrafficSummary
	RouteStat      = persistence.RouteStat
	ExportFormat   = persistence.ExportFormat
)

const (
	ExportFormatCSV   = persistence.ExportFormatCSV
	ExportFormatExcel = persistence.ExportFormatExcel
	ExportFormatPDF   = persistence.ExportFormatPDF
)
