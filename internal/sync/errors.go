package sync

import "errors"

var (
	// ErrBadData indicates that the telegram data is invalid and should not be retried
	ErrBadData = errors.New("bad telegram data")

	// ErrAppFailure indicates an application layer failure that can be retried
	ErrAppFailure = errors.New("application failure")

	// ErrInfraFailure indicates an infrastructure failure that can be retried
	ErrInfraFailure = errors.New("infrastructure failure")
)

// IsRetryableError checks if an error should trigger a retry (NACK)
func IsRetryableError(err error) bool {
	return errors.Is(err, ErrAppFailure) || errors.Is(err, ErrInfraFailure)
}

// IsBadDataError checks if an error indicates bad data that should not be retried (ACK)
func IsBadDataError(err error) bool {
	return errors.Is(err, ErrBadData)
}
