package usecase

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -destination=flight_search_mock.go -package=usecase github.com/flight-search/flight-search-and-aggregation-system/internal/usecase FlightSearchUseCase

// Default timeout values.
const (
	DefaultGlobalTimeout   = 5 * time.Second
	DefaultProviderTimeout = 2 * time.Second
)

// FlightSearchUseCase defines flight search operations.
type FlightSearchUseCase interface {
	// Search queries all providers and returns aggregated results.
	Search(ctx context.Context, criteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error)
}

// flightSearchUseCase implements FlightSearchUseCase.
type flightSearchUseCase struct {
	providers       []domain.FlightProvider
	globalTimeout   time.Duration
	providerTimeout time.Duration
	retryConfig     util.RetryConfig
}

// Config contains configuration options for the use case.
type Config struct {
	GlobalTimeout   time.Duration
	ProviderTimeout time.Duration
	RetryConfig     util.RetryConfig
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		GlobalTimeout:   DefaultGlobalTimeout,
		ProviderTimeout: DefaultProviderTimeout,
		RetryConfig:     util.DefaultRetryConfig(),
	}
}

// NewFlightSearchUseCase creates a FlightSearchUseCase.
// Uses default timeouts if config is nil.
func NewFlightSearchUseCase(providers []domain.FlightProvider, config *Config) FlightSearchUseCase {
	cfg := DefaultConfig()
	if config != nil {
		if config.GlobalTimeout > 0 {
			cfg.GlobalTimeout = config.GlobalTimeout
		}
		if config.ProviderTimeout > 0 {
			cfg.ProviderTimeout = config.ProviderTimeout
		}
		if config.RetryConfig.MaxAttempts > 0 {
			cfg.RetryConfig = config.RetryConfig
		}
	}

	return &flightSearchUseCase{
		providers:       providers,
		globalTimeout:   cfg.GlobalTimeout,
		providerTimeout: cfg.ProviderTimeout,
		retryConfig:     cfg.RetryConfig,
	}
}

// providerResult holds the result from a single provider query.
type providerResult struct {
	Provider string
	Flights  []domain.Flight
	Error    error
	Duration time.Duration
}

// Search implements FlightSearchUseCase.Search using Scatter-Gather pattern.
func (uc *flightSearchUseCase) Search(ctx context.Context, criteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error) {
	startTime := time.Now()

	// Handle case with no providers
	if len(uc.providers) == 0 {
		return nil, domain.ErrAllProvidersFailed
	}

	// Create context with global timeout
	ctx, cancel := context.WithTimeout(ctx, uc.globalTimeout)
	defer cancel()

	// Buffered channel to prevent goroutine blocking
	resultsChan := make(chan providerResult, len(uc.providers))

	// WaitGroup to track goroutine completion
	var wg sync.WaitGroup

	// Scatter: launch goroutines for each provider
	for _, provider := range uc.providers {
		wg.Add(1)
		go func(p domain.FlightProvider) {
			defer wg.Done()
			uc.queryProvider(ctx, p, criteria, resultsChan)
		}(provider)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Gather: collect results
	var allFlights []domain.Flight
	var failedProviders []string
	queriedProviders := make([]string, 0, len(uc.providers))

	for result := range resultsChan {
		queriedProviders = append(queriedProviders, result.Provider)
		if result.Error != nil {
			failedProviders = append(failedProviders, result.Provider)
			continue
		}
		allFlights = append(allFlights, result.Flights...)
	}

	// Check if context was cancelled before we got all results
	if ctx.Err() != nil && len(queriedProviders) < len(uc.providers) {
		// Record remaining providers as failed
		for _, p := range uc.providers {
			found := false
			for _, q := range queriedProviders {
				if q == p.Name() {
					found = true
					break
				}
			}
			if !found {
				queriedProviders = append(queriedProviders, p.Name())
				failedProviders = append(failedProviders, p.Name())
			}
		}
	}

	// Check if all providers failed
	if len(failedProviders) == len(uc.providers) {
		return nil, domain.ErrAllProvidersFailed
	}

	// Apply filtering using the dedicated filter module
	filtered := ApplyFilters(allFlights, opts.Filters)

	// Calculate ranking scores using the dedicated ranking module
	ranked := CalculateRankingScores(filtered)

	// Sort results using the dedicated sorting module
	sorted := SortFlights(ranked, opts.SortBy)

	// Build response with new format
	successfulProviders := len(uc.providers) - len(failedProviders)
	response := domain.NewSearchResponse(
		&criteria,
		sorted,
		domain.SearchMetadata{
			TotalResults:       len(sorted),
			ProvidersQueried:   len(uc.providers),
			ProvidersSucceeded: successfulProviders,
			ProvidersFailed:    len(failedProviders),
			SearchTimeMs:       time.Since(startTime).Milliseconds(),
			CacheHit:           false, // Not implemented yet
		},
	)

	return &response, nil
}

// queryProvider queries a single provider with timeout and panic recovery.
func (uc *flightSearchUseCase) queryProvider(ctx context.Context, provider domain.FlightProvider, criteria domain.SearchCriteria, results chan<- providerResult) {
	// Per-provider timeout
	ctx, cancel := context.WithTimeout(ctx, uc.providerTimeout)
	defer cancel()

	start := time.Now()
	providerName := provider.Name()

	// Panic recovery to prevent one provider from crashing the whole search
	defer func() {
		if r := recover(); r != nil {
			results <- providerResult{
				Provider: providerName,
				Error:    fmt.Errorf("provider panic: %v", r),
				Duration: time.Since(start),
			}
		}
	}()

	var flights []domain.Flight
	var lastErr error

	// Execute provider search with retry logic for retryable errors
retryLoop:
	for attempt := 1; attempt <= uc.retryConfig.MaxAttempts; attempt++ {
		flights, lastErr = provider.Search(ctx, criteria)

		// Check if context was cancelled during the provider call
		if ctx.Err() != nil {
			log.Debug().
				Str("provider", providerName).
				Int("attempt", attempt).
				Msg("Context cancelled during provider call")
			lastErr = ctx.Err()
			break retryLoop
		}

		// Success - return immediately
		if lastErr == nil {
			if attempt > 1 {
				log.Debug().
					Str("provider", providerName).
					Int("attempt", attempt).
					Msg("Provider succeeded after retry")
			}
			break retryLoop
		}

		// Check if error is retryable
		shouldRetry := true
		if providerErr, ok := lastErr.(*domain.ProviderError); ok {
			if !providerErr.Retryable {
				// Non-retryable error, don't retry
				log.Debug().
					Str("provider", providerName).
					Err(lastErr).
					Bool("retryable", false).
					Msg("Provider returned non-retryable error, will not retry")
				shouldRetry = false
			} else {
				log.Debug().
					Str("provider", providerName).
					Err(lastErr).
					Bool("retryable", true).
					Int("attempt", attempt).
					Int("max_attempts", uc.retryConfig.MaxAttempts).
					Msg("Provider returned retryable error")
			}
		} else {
			// Unknown error type, treat as retryable for backwards compatibility
			log.Debug().
				Str("provider", providerName).
				Err(lastErr).
				Int("attempt", attempt).
				Int("max_attempts", uc.retryConfig.MaxAttempts).
				Msg("Provider returned error (treating as retryable)")
		}

		// Don't retry if error is non-retryable or we've exhausted attempts
		if !shouldRetry || attempt >= uc.retryConfig.MaxAttempts {
			if shouldRetry && attempt >= uc.retryConfig.MaxAttempts {
				log.Warn().
					Str("provider", providerName).
					Int("attempts", uc.retryConfig.MaxAttempts).
					Err(lastErr).
					Msg("Provider failed after all retry attempts")
			}
			break retryLoop
		}

		// Calculate exponential backoff delay
		// Formula: initialDelay * multiplier^(attempt-1)
		delay := time.Duration(float64(uc.retryConfig.InitialDelay) * 
			math.Pow(uc.retryConfig.Multiplier, float64(attempt-1)))
		if delay > uc.retryConfig.MaxDelay {
			delay = uc.retryConfig.MaxDelay
		}

		// Add jitter (±20%)
		jitter := time.Duration((rand.Float64()*0.4 - 0.2) * float64(delay))
		delay += jitter
		if delay < 0 {
			delay = 0
		}

		log.Debug().
			Str("provider", providerName).
			Int("attempt", attempt).
			Dur("delay", delay).
			Msg("Sleeping before retry")

		// Sleep with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			log.Debug().
				Str("provider", providerName).
				Int("attempt", attempt).
				Msg("Retry cancelled by context")
			lastErr = ctx.Err()
			break retryLoop
		}
	}

	results <- providerResult{
		Provider: providerName,
		Flights:  flights,
		Error:    lastErr,
		Duration: time.Since(start),
	}
}