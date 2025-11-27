package errors

import "errors"

// Common application errors
var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternal           = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// Domain-specific errors
var (
	ErrInvalidTelegram     = errors.New("invalid telegram")
	ErrInvalidFilter       = errors.New("invalid filter")
	ErrTimeRangeTooLarge   = errors.New("time range exceeds maximum limit")
	ErrExportLimitExceeded = errors.New("export limit exceeded")
)

// IsNotFound checks if an error indicates a not found condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput checks if an error indicates invalid input
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if an error indicates unauthorized access
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}
