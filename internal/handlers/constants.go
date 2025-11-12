package handlers

import "time"

const (
	// DefaultPageLimit is the default number of results per page for search queries.
	DefaultPageLimit = 50

	// DefaultAutocompleteSize is the default number of autocomplete suggestions to return.
	DefaultAutocompleteSize = 5

	// DefaultTimeWindow is the default time window for statistics queries (24 hours).
	DefaultTimeWindow = 24 * time.Hour
)
