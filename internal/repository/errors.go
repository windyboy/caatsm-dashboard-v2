package repository

import "errors"

// ErrNotImplemented is returned by methods that are placeholders.
var ErrNotImplemented = errors.New("not implemented")

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

// ErrInvalidInput is returned when input validation fails.
var ErrInvalidInput = errors.New("invalid input")

// ErrTransactionFailed is returned when a database transaction operation fails.
var ErrTransactionFailed = errors.New("transaction failed")

// ErrTooManyResults is returned when a query returns too many results.
var ErrTooManyResults = errors.New("too many results")
