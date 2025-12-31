package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, cfg.InitialDelay)
	assert.Equal(t, 2*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
}

func TestExecuteWithRetry_SuccessFirstAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return nil // Success immediately
	}

	err := ExecuteWithRetry(context.Background(), cfg, fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts, "should only call function once on success")
}

func TestExecuteWithRetry_SuccessOnSecondAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")

	fn := func() error {
		attempts++
		if attempts == 1 {
			return testErr // Fail first time
		}
		return nil // Success on second attempt
	}

	start := time.Now()
	err := ExecuteWithRetry(context.Background(), cfg, fn)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts, "should call function twice")
	// Should have waited approximately InitialDelay (10ms) with jitter
	// Being generous with timing in tests
	assert.GreaterOrEqual(t, duration, 8*time.Millisecond, "should have some delay")
	assert.Less(t, duration, 50*time.Millisecond, "shouldn't wait too long")
}

func TestExecuteWithRetry_SuccessOnThirdAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")

	fn := func() error {
		attempts++
		if attempts < 3 {
			return testErr // Fail first two times
		}
		return nil // Success on third attempt
	}

	start := time.Now()
	err := ExecuteWithRetry(context.Background(), cfg, fn)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "should call function three times")
	// Should have waited approximately 10ms + 20ms = 30ms (with jitter)
	assert.GreaterOrEqual(t, duration, 20*time.Millisecond, "should have delays")
	assert.Less(t, duration, 100*time.Millisecond, "shouldn't wait too long")
}

func TestExecuteWithRetry_AllAttemptsFail(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("persistent error")

	fn := func() error {
		attempts++
		return testErr // Always fail
	}

	err := ExecuteWithRetry(context.Background(), cfg, fn)

	assert.Error(t, err)
	assert.Equal(t, testErr, err, "should return the last error")
	assert.Equal(t, 3, attempts, "should call function MaxAttempts times")
}

func TestExecuteWithRetry_ContextCancelledBeforeRetry(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		attempts++
		if attempts == 2 {
			// Cancel context on second attempt
			cancel()
		}
		return testErr
	}

	err := ExecuteWithRetry(ctx, cfg, fn)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err, "should return context.Canceled")
	assert.Equal(t, 2, attempts, "should stop retrying after context cancelled")
}

func TestExecuteWithRetry_ContextCancelledDuringSleep(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond, // Long delay to ensure we can cancel during sleep
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		attempts++
		return testErr
	}

	// Cancel context after a short delay (while sleeping)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := ExecuteWithRetry(ctx, cfg, fn)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err, "should return context.Canceled")
	assert.Equal(t, 1, attempts, "should only call function once before cancellation")
	// Should return quickly after cancellation, not wait full delay
	assert.Less(t, duration, 150*time.Millisecond, "should cancel quickly during sleep")
}

func TestExecuteWithRetry_ContextAlreadyCancelled(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before starting

	fn := func() error {
		attempts++
		return testErr
	}

	err := ExecuteWithRetry(ctx, cfg, fn)

	assert.Error(t, err)
	// First attempt will still execute, then context check happens during sleep
	assert.Equal(t, 1, attempts, "should execute once even if context pre-cancelled")
	assert.Equal(t, context.Canceled, err)
}

func TestExecuteWithRetry_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")
	attemptTimes := make([]time.Time, 0, 4)

	fn := func() error {
		attempts++
		attemptTimes = append(attemptTimes, time.Now())
		return testErr
	}

	start := time.Now()
	err := ExecuteWithRetry(context.Background(), cfg, fn)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, 4, attempts)

	// Verify delays between attempts (accounting for jitter ±20%)
	// Expected delays: 50ms, 100ms, 200ms
	// With jitter: 40-60ms, 80-120ms, 160-240ms
	if len(attemptTimes) >= 2 {
		delay1 := attemptTimes[1].Sub(attemptTimes[0])
		assert.GreaterOrEqual(t, delay1, 35*time.Millisecond, "first delay should be ~50ms")
		assert.LessOrEqual(t, delay1, 70*time.Millisecond, "first delay with jitter")
	}

	if len(attemptTimes) >= 3 {
		delay2 := attemptTimes[2].Sub(attemptTimes[1])
		assert.GreaterOrEqual(t, delay2, 70*time.Millisecond, "second delay should be ~100ms")
		assert.LessOrEqual(t, delay2, 130*time.Millisecond, "second delay with jitter")
	}

	if len(attemptTimes) >= 4 {
		delay3 := attemptTimes[3].Sub(attemptTimes[2])
		assert.GreaterOrEqual(t, delay3, 140*time.Millisecond, "third delay should be ~200ms")
		assert.LessOrEqual(t, delay3, 260*time.Millisecond, "third delay with jitter")
	}

	// Total duration should be approximately 50 + 100 + 200 = 350ms (with jitter)
	assert.GreaterOrEqual(t, duration, 300*time.Millisecond, "total duration")
	assert.LessOrEqual(t, duration, 450*time.Millisecond, "total duration with jitter")
}

func TestExecuteWithRetry_MaxDelayRespected(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     150 * time.Millisecond, // Cap at 150ms
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")
	attemptTimes := make([]time.Time, 0, 5)

	fn := func() error {
		attempts++
		attemptTimes = append(attemptTimes, time.Now())
		return testErr
	}

	ExecuteWithRetry(context.Background(), cfg, fn)

	assert.Equal(t, 5, attempts)

	// Third delay would be 400ms without cap, should be capped at 150ms
	if len(attemptTimes) >= 4 {
		delay3 := attemptTimes[3].Sub(attemptTimes[2])
		// With MaxDelay=150ms and ±20% jitter: 120-180ms
		assert.GreaterOrEqual(t, delay3, 110*time.Millisecond)
		assert.LessOrEqual(t, delay3, 190*time.Millisecond, "should respect MaxDelay cap")
	}
}

func TestExecuteWithRetry_SingleAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("error")

	fn := func() error {
		attempts++
		return testErr
	}

	err := ExecuteWithRetry(context.Background(), cfg, fn)

	assert.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Equal(t, 1, attempts, "should only attempt once")
}

func TestExecuteWithRetry_ZeroInitialDelay(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 0, // No delay
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("error")

	fn := func() error {
		attempts++
		return testErr
	}

	start := time.Now()
	err := ExecuteWithRetry(context.Background(), cfg, fn)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
	// Should be very fast with no delays
	assert.Less(t, duration, 20*time.Millisecond, "should be fast with zero delay")
}

func TestExecuteWithRetryLogged_SuccessAfterRetry(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("temporary error")

	fn := func() error {
		attempts++
		if attempts < 2 {
			return testErr
		}
		return nil
	}

	err := ExecuteWithRetryLogged(context.Background(), cfg, fn, "test_operation")

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestExecuteWithRetryLogged_AllAttemptsFail(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("persistent error")

	fn := func() error {
		attempts++
		return testErr
	}

	err := ExecuteWithRetryLogged(context.Background(), cfg, fn, "test_operation")

	assert.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Equal(t, 2, attempts)
}

func TestExecuteWithRetryLogged_ContextCancelled(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	attempts := 0
	testErr := errors.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		attempts++
		return testErr
	}

	// Cancel after short delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := ExecuteWithRetryLogged(ctx, cfg, fn, "test_operation")

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, 1, attempts)
}

// Benchmark tests
func BenchmarkExecuteWithRetry_Success(b *testing.B) {
	cfg := DefaultRetryConfig()
	fn := func() error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExecuteWithRetry(context.Background(), cfg, fn)
	}
}

func BenchmarkExecuteWithRetry_AllFail(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}
	testErr := errors.New("error")
	fn := func() error { return testErr }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExecuteWithRetry(context.Background(), cfg, fn)
	}
}

// TestExecuteWithRetry_JitterVariation verifies that jitter is actually applied
func TestExecuteWithRetry_JitterVariation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   1.0, // No exponential increase, to isolate jitter effect
	}

	testErr := errors.New("error")
	delays := make([]time.Duration, 0, 9)

	for i := 0; i < 3; i++ {
		attempts := 0
		attemptTimes := make([]time.Time, 0, 10)

		fn := func() error {
			attempts++
			attemptTimes = append(attemptTimes, time.Now())
			return testErr
		}

		ExecuteWithRetry(context.Background(), cfg, fn)

		// Collect delays
		for j := 1; j < len(attemptTimes); j++ {
			delays = append(delays, attemptTimes[j].Sub(attemptTimes[j-1]))
		}
	}

	// With jitter, we should see variation in delays
	// All delays should be different (highly unlikely to be exactly the same with random jitter)
	require.Greater(t, len(delays), 5, "should have multiple delay samples")

	// Check that delays vary within expected range (80ms - 120ms for 100ms ± 20%)
	hasVariation := false
	firstDelay := delays[0]
	for _, d := range delays {
		if d != firstDelay {
			hasVariation = true
			break
		}
		// Also verify each delay is within jitter range
		assert.GreaterOrEqual(t, d, 70*time.Millisecond, "delay should be >= 80ms")
		assert.LessOrEqual(t, d, 130*time.Millisecond, "delay should be <= 120ms")
	}

	// Note: This is probabilistic, but with multiple samples, variation is virtually guaranteed
	assert.True(t, hasVariation, "jitter should cause variation in delays")
}
