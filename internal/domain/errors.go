package domain

// ErrInvalidTelegram represents a validation error for a telegram
type ErrInvalidTelegram struct {
	Field  string
	Reason string
}

func (e ErrInvalidTelegram) Error() string {
	return "invalid telegram: " + e.Field + " " + e.Reason
}

