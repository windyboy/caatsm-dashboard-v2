package services

import "time"

const (
	// MaxExportLimit is the maximum number of records to export in a single request.
	MaxExportLimit = 10000

	// DefaultStatsLimit is the default number of top routes to return in statistics.
	DefaultStatsLimit = 10

	// DefaultTimeWindow is the default time window for statistics queries (24 hours).
	DefaultTimeWindow = 24 * time.Hour
)
