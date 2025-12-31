package util

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

// RetryConfig defines the configuration for retry behavior with exponential backoff.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts to make (including the initial attempt).
	// Must be at least 1.
	MaxAttempts int

	// InitialDelay is the initial delay before the first retry.
	// The delay increases exponentially for subsequent retries.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	// Prevents delays from growing unbounded with exponential backoff.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases for each retry.
	// Common values: 2.0 (double), 1.5 (50% increase).
	Multiplier float64
}

// DefaultRetryConfig returns a retry configuration with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryableFunc is a function that can be retried.
// It should return an error if the operation failed and nil on success.
type RetryableFunc func() error

// ExecuteWithRetry executes the given function with exponential backoff retry logic.
// It respects context cancellation and applies jitter to prevent thundering herd.
//
// The function will:
// - Execute fn up to cfg.MaxAttempts times
// - Use exponential backoff between retries: initialDelay * multiplier^(attempt-1)
// - Cap the delay at cfg.MaxDelay
// - Add random jitter (±20%) to each delay to prevent synchronized retries
// - Return immediately if the context is cancelled
// - Return the last error if all attempts fail
//
// Example backoff with default config (100ms initial, 2.0 multiplier):
// - Attempt 1: execute immediately
// - Attempt 2: wait ~100ms (80ms-120ms with jitter)
// - Attempt 3: wait ~200ms (160ms-240ms with jitter)
// - If MaxAttempts=3, total retries=2, max time ~300ms
func ExecuteWithRetry(ctx context.Context, cfg RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			// Success - return immediately
			return nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate exponential backoff delay
		// Formula: initialDelay * multiplier^(attempt-1)
		// Example: 100ms * 2^0 = 100ms, 100ms * 2^1 = 200ms, 100ms * 2^2 = 400ms
		delay := time.Duration(float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt-1)))

		// Cap the delay at MaxDelay
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		// Add jitter (±20%) to prevent thundering herd
		// This randomizes the delay to avoid all clients retrying at the same time
		jitter := time.Duration((rand.Float64()*0.4 - 0.2) * float64(delay))
		delay += jitter

		// Ensure delay is non-negative (in case jitter made it negative)
		if delay < 0 {
			delay = 0
		}

		// Sleep with context cancellation support
		// This allows the retry loop to be interrupted if the context is cancelled
		select {
		case <-time.After(delay):
			// Delay completed, continue to next retry
			continue
		case <-ctx.Done():
			// Context cancelled, return immediately
			return ctx.Err()
		}
	}

	// All attempts exhausted, return the last error
	return lastErr
}

// ExecuteWithRetryLogged is like ExecuteWithRetry but includes structured logging.
// Use this version when you want to track retry attempts for observability.
//
// Logs include:
// - Debug level: each retry attempt with delay and error
// - Warn level: final failure after all attempts exhausted
func ExecuteWithRetryLogged(ctx context.Context, cfg RetryConfig, fn RetryableFunc, operationName string) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			// Success
			if attempt > 1 {
				log.Debug().
					Str("operation", operationName).
					Int("attempt", attempt).
					Int("max_attempts", cfg.MaxAttempts).
					Msg("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err

		// Log retry attempt
		if attempt < cfg.MaxAttempts {
			log.Debug().
				Str("operation", operationName).
				Int("attempt", attempt).
				Int("max_attempts", cfg.MaxAttempts).
				Err(err).
				Msg("Operation failed, will retry")
		}

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate exponential backoff delay
		delay := time.Duration(float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt-1)))

		// Cap the delay at MaxDelay
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		// Add jitter (±20%)
		jitter := time.Duration((rand.Float64()*0.4 - 0.2) * float64(delay))
		delay += jitter

		// Ensure delay is non-negative
		if delay < 0 {
			delay = 0
		}

		log.Debug().
			Str("operation", operationName).
			Int("attempt", attempt).
			Dur("delay", delay).
			Msg("Sleeping before retry")

		// Sleep with context cancellation support
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			log.Debug().
				Str("operation", operationName).
				Int("attempt", attempt).
				Msg("Retry cancelled by context")
			return ctx.Err()
		}
	}

	// All attempts exhausted
	log.Warn().
		Str("operation", operationName).
		Int("attempts", cfg.MaxAttempts).
		Err(lastErr).
		Msg("Operation failed after all retry attempts")

	return lastErr
}
