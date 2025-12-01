package domain

import (
	"fmt"
	"net/http"
)

// APIError represents a standardized API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes and their corresponding HTTP status codes
const (
	// Client errors (4xx)
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeRateLimit        = "RATE_LIMIT_EXCEEDED"
	ErrCodeValidationFailed = "VALIDATION_FAILED"

	// Server errors (5xx)
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
)

// Error code to HTTP status mapping
var errorCodeToStatus = map[string]int{
	ErrCodeInvalidInput:       http.StatusBadRequest,
	ErrCodeUnauthorized:       http.StatusUnauthorized,
	ErrCodeForbidden:          http.StatusForbidden,
	ErrCodeNotFound:           http.StatusNotFound,
	ErrCodeRateLimit:          http.StatusTooManyRequests,
	ErrCodeValidationFailed:   http.StatusBadRequest,
	ErrCodeInternalError:      http.StatusInternalServerError,
	ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
	ErrCodeTimeout:            http.StatusRequestTimeout,
}

// GetHTTPStatus returns the HTTP status code for an error code
func GetHTTPStatus(code string) int {
	if status, exists := errorCodeToStatus[code]; exists {
		return status
	}
	return http.StatusInternalServerError
}

// NewAPIError creates a new APIError
func NewAPIError(code, message string) APIError {
	return APIError{
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithDetails creates a new APIError with details
func NewAPIErrorWithDetails(code, message, details string) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined errors
var (
	ErrInvalidInput       = NewAPIError(ErrCodeInvalidInput, "Invalid input parameters")
	ErrUnauthorized       = NewAPIError(ErrCodeUnauthorized, "Authentication required")
	ErrForbidden          = NewAPIError(ErrCodeForbidden, "Access forbidden")
	ErrNotFound           = NewAPIError(ErrCodeNotFound, "Resource not found")
	ErrRateLimit          = NewAPIError(ErrCodeRateLimit, "Rate limit exceeded")
	ErrValidationFailed   = NewAPIError(ErrCodeValidationFailed, "Validation failed")
	ErrInternalError      = NewAPIError(ErrCodeInternalError, "Internal server error")
	ErrServiceUnavailable = NewAPIError(ErrCodeServiceUnavailable, "Service temporarily unavailable")
	ErrTimeout            = NewAPIError(ErrCodeTimeout, "Request timeout")
)

// ErrInvalidTelegram represents a validation error for a telegram
type ErrInvalidTelegram struct {
	Field  string
	Reason string
}

func (e ErrInvalidTelegram) Error() string {
	return "invalid telegram: " + e.Field + " " + e.Reason
}

// ToAPIError converts ErrInvalidTelegram to APIError
func (e ErrInvalidTelegram) ToAPIError() APIError {
	return NewAPIErrorWithDetails(ErrCodeValidationFailed, "Invalid telegram data", e.Field+" "+e.Reason)
}

// ErrInvalidFilter represents a validation error in search filters
type ErrInvalidFilter struct {
	Field  string
	Reason string
}

func (e ErrInvalidFilter) Error() string {
	return "invalid filter: " + e.Field + " " + e.Reason
}

// ToAPIError converts ErrInvalidFilter to APIError
func (e ErrInvalidFilter) ToAPIError() APIError {
	return NewAPIErrorWithDetails(ErrCodeValidationFailed, "Invalid search filter", e.Field+" "+e.Reason)
}
