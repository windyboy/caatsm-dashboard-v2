package events

import "github.com/windy/caatsm-dashboard/internal/repository"

// Re-export repository errors for backward compatibility.
var (
	ErrNotImplemented  = repository.ErrNotImplemented
	ErrNotFound        = repository.ErrNotFound
	ErrInvalidInput    = repository.ErrInvalidInput
	ErrTransactionFailed = repository.ErrTransactionFailed
	ErrTooManyResults  = repository.ErrTooManyResults
)

