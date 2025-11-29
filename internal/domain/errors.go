package domain

// ErrInvalidTelegram represents a validation error for a telegram
type ErrInvalidTelegram struct {
	Field  string
	Reason string
}

func (e ErrInvalidTelegram) Error() string {
	return "invalid telegram: " + e.Field + " " + e.Reason
}

// ErrInvalidFilter represents a validation error in search filters
type ErrInvalidFilter struct {
	Field  string
	Reason string
}

func (e ErrInvalidFilter) Error() string {
	return "invalid filter: " + e.Field + " " + e.Reason
}
