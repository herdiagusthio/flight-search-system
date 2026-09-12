package domain

import (
	"errors"
	"fmt"
)

// ErrorCategory defines the nature of the failure to determine the recovery strategy.
type ErrorCategory int

const (
	CategoryUnknown ErrorCategory = iota
	CategoryTransient // Retryable: network glitch, 503, rate limit
	CategoryPermanent // Fatal: 400 Bad Request, 401 Unauthorized, 404 Not Found
	CategoryCritical  // System failure: panic, database down
)

// DomainError is the base for all domain-level errors.
type DomainError struct {
	Category ErrorCategory
	Message  string
	Err      error
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.CategoryName(), e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func (e *DomainError) Category() ErrorCategory {
	return e.Category
}

func (e *DomainError) CategoryName() string {
	switch e.Category {
	case CategoryTransient:
		return "TRANSIENT"
	case CategoryPermanent:
		return "PERMANENT"
	case CategoryCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Sentinel errors for common scenarios
var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrAllProvidersFailed = errors.New("all providers failed")
	ErrProviderTimeout = errors.New("provider timeout")
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrNoFlightsFound = errors.New("no flights found")
	ErrInvalidFlightTimes = errors.New("invalid flight times")
	ErrMissingRequiredField = errors.New("missing required field")
)

// NewDomainError creates a structured domain error.
func NewDomainError(cat ErrorCategory, msg string, err error) error {
	return &DomainError{
		Category: cat,
		Message:  msg,
		Err:      err,
	}
}

// IsTransient checks if an error is retryable.
func IsTransient(err error) bool {
	var dErr *DomainError
	if errors.As(err, &dErr) {
		return dErr.Category == CategoryTransient
	}
	// Default to non-retryable if unknown
	return false
}

// IsPermanent checks if an error is a fatal request error.
func IsPermanent(err error) bool {
	var dErr *DomainError
	if errors.As(err, &dErr) {
		return dErr.Category == CategoryPermanent
	}
	return false
}
