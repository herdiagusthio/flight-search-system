package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProviderWithRetryable is a mock provider that can return retryable or non-retryable errors
type mockProviderWithRetryable struct {
	name          string
	callCount     int
	successAfter  int // Succeed after this many calls (0 = always fail)
	retryable     bool
	returnFlights []domain.Flight
	returnError   error
}

func (m *mockProviderWithRetryable) Name() string {
	return m.name
}

func (m *mockProviderWithRetryable) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	m.callCount++

	// Check if we should succeed this time
	if m.successAfter > 0 && m.callCount >= m.successAfter {
		return m.returnFlights, nil
	}

	// Return error (retryable or not)
	if m.returnError != nil {
		if m.retryable {
			return nil, domain.NewRetryableProviderError(m.name, m.returnError)
		}
		return nil, domain.NewProviderError(m.name, m.returnError)
	}

	return m.returnFlights, nil
}

func TestSearch_RetryableError_SuccessOnRetry(t *testing.T) {
	// Provider fails first time, succeeds on second attempt
	provider := &mockProviderWithRetryable{
		name:          "retryable_provider",
		successAfter:  2, // Fail first, succeed on second
		retryable:     true,
		returnError:   errors.New("temporary network error"),
		returnFlights: []domain.Flight{{FlightNumber: "TEST-123"}},
	}

	cfg := &Config{
		GlobalTimeout:   5 * time.Second,
		ProviderTimeout: 2 * time.Second,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond, // Short delay for faster tests
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, cfg)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Flights, 1)
	assert.Equal(t, "TEST-123", resp.Flights[0].FlightNumber)
	assert.Equal(t, 2, provider.callCount, "should have called provider twice (1 failure + 1 success)")
	assert.Equal(t, 1, resp.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, resp.Metadata.ProvidersFailed)
}

func TestSearch_RetryableError_AllAttemptsFail(t *testing.T) {
	// Provider always fails with retryable error
	provider := &mockProviderWithRetryable{
		name:        "always_fail_retryable",
		retryable:   true,
		returnError: errors.New("persistent network issue"),
	}

	cfg := &Config{
		GlobalTimeout:   5 * time.Second,
		ProviderTimeout: 2 * time.Second,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, cfg)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	// Should return ErrAllProvidersFailed since the only provider failed
	require.Error(t, err)
	assert.Equal(t, domain.ErrAllProvidersFailed, err)
	assert.Nil(t, resp)
	assert.Equal(t, 3, provider.callCount, "should have retried 3 times")
}

func TestSearch_NonRetryableError_NoRetry(t *testing.T) {
	// Provider fails with non-retryable error
	provider := &mockProviderWithRetryable{
		name:        "non_retryable_provider",
		retryable:   false,
		returnError: errors.New("invalid API key"),
	}

	cfg := &Config{
		GlobalTimeout:   5 * time.Second,
		ProviderTimeout: 2 * time.Second,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, cfg)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	// Should return ErrAllProvidersFailed
	require.Error(t, err)
	assert.Equal(t, domain.ErrAllProvidersFailed, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, provider.callCount, "should NOT have retried (non-retryable error)")
}

func TestSearch_MixedProviders_RetryableAndNonRetryable(t *testing.T) {
	// One provider with retryable error (succeeds on retry)
	provider1 := &mockProviderWithRetryable{
		name:          "retryable",
		successAfter:  2,
		retryable:     true,
		returnError:   errors.New("temporary error"),
		returnFlights: []domain.Flight{{FlightNumber: "RETRY-1"}},
	}

	// One provider with non-retryable error
	provider2 := &mockProviderWithRetryable{
		name:        "non_retryable",
		retryable:   false,
		returnError: errors.New("auth error"),
	}

	// One provider that succeeds immediately
	provider3 := &mockProviderWithRetryable{
		name:          "success",
		returnFlights: []domain.Flight{{FlightNumber: "SUCCESS-1"}},
	}

	cfg := &Config{
		GlobalTimeout:   5 * time.Second,
		ProviderTimeout: 2 * time.Second,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider1, provider2, provider3}, cfg)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Flights, 2, "should have flights from provider1 and provider3")
	assert.Equal(t, 2, provider1.callCount, "retryable provider should have been retried")
	assert.Equal(t, 1, provider2.callCount, "non-retryable provider should NOT have been retried")
	assert.Equal(t, 1, provider3.callCount, "successful provider should only be called once")
	assert.Equal(t, 2, resp.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, resp.Metadata.ProvidersFailed)
}

func TestSearch_RetryRespectsContext(t *testing.T) {
	// Provider always fails with retryable error
	provider := &mockProviderWithRetryable{
		name:        "slow_retryable",
		retryable:   true,
		returnError: errors.New("network timeout"),
	}

	cfg := &Config{
		GlobalTimeout:   200 * time.Millisecond, // Very short timeout
		ProviderTimeout: 100 * time.Millisecond,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  10, // Many attempts, but should be cancelled by context
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     200 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, cfg)

	start := time.Now()
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})
	duration := time.Since(start)

	// Should fail due to context timeout
	require.Error(t, err)
	assert.Nil(t, resp)

	// Should not have exhausted all retry attempts (context should cancel retries)
	assert.Less(t, provider.callCount, 10, "retries should have been cancelled by context")

	// Should have timed out around the global timeout
	assert.Less(t, duration, 500*time.Millisecond, "should timeout quickly")
}

func TestSearch_CustomRetryConfig(t *testing.T) {
	// Provider fails first 4 times, succeeds on 5th
	provider := &mockProviderWithRetryable{
		name:          "needs_many_retries",
		successAfter:  5,
		retryable:     true,
		returnError:   errors.New("intermittent error"),
		returnFlights: []domain.Flight{{FlightNumber: "RETRY-5"}},
	}

	// Custom retry config with more attempts
	cfg := &Config{
		GlobalTimeout:   10 * time.Second,
		ProviderTimeout: 2 * time.Second,
		RetryConfig: util.RetryConfig{
			MaxAttempts:  5, // Allow 5 attempts
			InitialDelay: 5 * time.Millisecond,
			MaxDelay:     30 * time.Millisecond,
			Multiplier:   1.5,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, cfg)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Flights, 1)
	assert.Equal(t, 5, provider.callCount, "should have called provider 5 times")
	assert.Equal(t, 1, resp.Metadata.ProvidersSucceeded)
}

func TestSearch_DefaultRetryConfig(t *testing.T) {
	// Provider succeeds on 2nd attempt
	provider := &mockProviderWithRetryable{
		name:          "default_retry_test",
		successAfter:  2,
		retryable:     true,
		returnError:   errors.New("temporary error"),
		returnFlights: []domain.Flight{{FlightNumber: "DEFAULT-1"}},
	}

	// Use nil config to get defaults
	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, provider.callCount, "should use default retry config")
}

// TestSearch_RetryWithProviderTimeout verifies that provider timeouts are handled correctly with retries
func TestSearch_RetryWithProviderTimeout(t *testing.T) {
	// Provider that times out
	slowProvider := &mockProviderSlow{
		name:  "timeout_provider",
		delay: 500 * time.Millisecond, // Will timeout
	}

	cfg := &Config{
		GlobalTimeout:   5 * time.Second,
		ProviderTimeout: 100 * time.Millisecond, // Short provider timeout
		RetryConfig: util.RetryConfig{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{slowProvider}, cfg)

	start := time.Now()
	resp, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})
	duration := time.Since(start)

	// Should fail (all providers failed)
	require.Error(t, err)
	assert.Nil(t, resp)

	// Should have attempted retries but timed out each time
	// Total time should be: ~100ms (attempt 1) + ~10ms (retry delay) + cancellation
	// Being more lenient with timing in tests
	assert.GreaterOrEqual(t, duration, 80*time.Millisecond, "should have taken time for at least one attempt")
	assert.Less(t, duration, 500*time.Millisecond, "but not too long")
}

// mockProviderSlow simulates a slow provider
type mockProviderSlow struct {
	name  string
	delay time.Duration
}

func (m *mockProviderSlow) Name() string {
	return m.name
}

func (m *mockProviderSlow) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	select {
	case <-time.After(m.delay):
		return []domain.Flight{{FlightNumber: "SLOW-1"}}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
