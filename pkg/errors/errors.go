package errors

import "errors"

// Common application errors
// These are generic HTTP/application-level errors used across the delivery layer.
// Domain-specific validation errors should be defined in their respective domain packages.
var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternal           = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
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
