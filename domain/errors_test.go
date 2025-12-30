package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		err           error
		retryable     bool
		expectedMsg   string
	}{
		{
			name:        "basic provider error",
			provider:    "airasia",
			err:         errors.New("connection failed"),
			retryable:   false,
			expectedMsg: "provider airasia: connection failed",
		},
		{
			name:        "retryable provider error",
			provider:    "garuda",
			err:         errors.New("rate limited"),
			retryable:   true,
			expectedMsg: "provider garuda: rate limited",
		},
		{
			name:        "timeout error",
			provider:    "lion_air",
			err:         ErrProviderTimeout,
			retryable:   false,
			expectedMsg: "provider lion_air: provider timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var providerErr *ProviderError
			if tt.retryable {
				providerErr = NewRetryableProviderError(tt.provider, tt.err)
			} else {
				providerErr = NewProviderError(tt.provider, tt.err)
			}

			assert.Equal(t, tt.expectedMsg, providerErr.Error())
			assert.Equal(t, tt.provider, providerErr.Provider)
			assert.Equal(t, tt.retryable, providerErr.Retryable)
			assert.True(t, errors.Is(providerErr, tt.err))
		})
	}
}

func TestNewProviderTimeoutError(t *testing.T) {
	err := NewProviderTimeoutError("test_provider")
	assert.Equal(t, "test_provider", err.Provider)
	assert.True(t, errors.Is(err, ErrProviderTimeout))
	assert.False(t, err.Retryable)
}

func TestNewProviderUnavailableError(t *testing.T) {
	err := NewProviderUnavailableError("test_provider")
	assert.Equal(t, "test_provider", err.Provider)
	assert.True(t, errors.Is(err, ErrProviderUnavailable))
	assert.False(t, err.Retryable)
}

func TestProviderErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	providerErr := NewProviderError("test", originalErr)
	
	assert.Equal(t, originalErr, providerErr.Unwrap())
	assert.True(t, errors.Is(providerErr, originalErr))
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		message     string
		expectedErr string
	}{
		{
			name:        "origin field error",
			field:       "origin",
			message:     "is required",
			expectedErr: "origin: is required",
		},
		{
			name:        "date field error",
			field:       "departureDate",
			message:     "must be in YYYY-MM-DD format",
			expectedErr: "departureDate: must be in YYYY-MM-DD format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message)
			assert.Equal(t, tt.expectedErr, err.Error())
			assert.Equal(t, tt.field, err.Field)
			assert.Equal(t, tt.message, err.Message)
		})
	}
}

func TestWrapInvalidRequest(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple message",
			format:   "origin is required",
			args:     nil,
			expected: "invalid request: origin is required",
		},
		{
			name:     "formatted message",
			format:   "invalid value: %s",
			args:     []interface{}{"abc"},
			expected: "invalid request: invalid value: abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapInvalidRequest(tt.format, tt.args...)
			assert.Equal(t, tt.expected, err.Error())
			assert.True(t, errors.Is(err, ErrInvalidRequest))
		})
	}
}

func TestIsInvalidRequest(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is invalid request",
			err:      WrapInvalidRequest("test"),
			expected: true,
		},
		{
			name:     "is not invalid request",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInvalidRequest(tt.err))
		})
	}
}

func TestIsAllProvidersFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is all providers failed",
			err:      ErrAllProvidersFailed,
			expected: true,
		},
		{
			name:     "is not all providers failed",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAllProvidersFailed(tt.err))
		})
	}
}

func TestIsProviderTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is provider timeout",
			err:      NewProviderTimeoutError("test"),
			expected: true,
		},
		{
			name:     "is not provider timeout",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsProviderTimeout(tt.err))
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are distinct
	sentinelErrors := []error{
		ErrInvalidRequest,
		ErrAllProvidersFailed,
		ErrProviderTimeout,
		ErrProviderUnavailable,
		ErrNoFlightsFound,
		ErrInvalidFlightTimes,
		ErrMissingRequiredField,
	}

	for i, err1 := range sentinelErrors {
		for j, err2 := range sentinelErrors {
			if i == j {
				assert.True(t, errors.Is(err1, err2))
			} else {
				assert.False(t, errors.Is(err1, err2))
			}
		}
	}
}
