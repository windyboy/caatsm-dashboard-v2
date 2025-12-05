package repository

import "time"

// TimeWindow describes a time range used for analytics queries.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}
