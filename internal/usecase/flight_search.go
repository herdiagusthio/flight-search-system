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

const (
	DefaultGlobalTimeout   = 5 * time.Second
	DefaultProviderTimeout = 2 * time.Second
)

type FlightSearchUseCase interface {
	Search(ctx context.Context, criteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error)
}

type flightSearchUseCase struct {
	providers       []domain.FlightProvider
	globalTimeout   time.Duration
	providerTimeout time.Duration
	retryConfig     util.RetryConfig
}

type Config struct {
	GlobalTimeout   time.Duration
	ProviderTimeout time.Duration
	RetryConfig     util.RetryConfig
}

func DefaultConfig() Config {
	return Config{
		GlobalTimeout:   DefaultGlobalTimeout,
		ProviderTimeout: DefaultProviderTimeout,
		RetryConfig:     util.DefaultRetryConfig(),
	}
}

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

type providerResult struct {
	Provider string
	Flights  []domain.Flight
	Error    error
	Duration time.Duration
}

// Search implements FlightSearchUseCase.Search using a high-resilience Scatter-Gather pattern.
func (uc *flightSearchUseCase) Search(ctx context.Context, javaCriteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error) {
	startTime := time.Now()

	if len(uc.providers) == 0 {
		return nil, domain.ErrAllProvidersFailed
	}

	// Global context for the entire search operation
	ctx, cancel := context.WithTimeout(ctx, uc.globalTimeout)
	defer cancel()

	resultsChan := make(chan providerResult, len(uc.providers))
	var wg sync.WaitGroup

	// SCATTER: Launch provider queries in parallel
	for _, provider := range uc.providers {
		wg.Add(1)
		go func(p domain.FlightProvider) {
			defer wg.Done()
			uc.queryProvider(ctx, p, javaCriteria, resultsChan)
		}(provider)
	}

	// Gather results as they arrive
	var allFlights []domain.Flight
	var failedProviders []string
	var successfulProvidersCount int

	// Wait for all providers or the global timeout
	// We use a separate goroutine to close the channel when all work is done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Error != nil {
			failedProviders = append(failedProviders, result.Provider)
			continue
		}
		allFlights = append(allFlights, result.Flights...)
		successfulProvidersCount++
	}

	// PARTIAL SUCCESS LOGIC: 
	// We return results if AT LEAST ONE provider succeeded, even if others failed.
	if successfulProvidersCount == 0 {
		return nil, domain.ErrAllProvidersFailed
	}

	// Apply filtering and ranking
	filtered := ApplyFilters(allFlights, opts.Filters)
	ranked := CalculateRankingScores(filtered)
	sorted := SortFlights(ranked, opts.SortBy)

	return domain.NewSearchResponse(
		&javaCriteria,
		sorted,
		domain.SearchMetadata{
			TotalResults:       len(sorted),
			ProvidersQueried:   len(uc.providers),
			ProvidersSucceeded: successfulProvidersCount,
			ProvidersFailed:    len(failedProviders),
			SearchTimeMs:       time.Since(startTime).Milliseconds(),
			CacheHit:           false,
		},
	), nil
}

func (uc *flightSearchUseCase) queryProvider(ctx context.Context, provider domain.FlightProvider, criteria domain.SearchCriteria, results chan<- providerResult) {
	// Per-provider timeout
	ctx, cancel := context.WithTimeout(ctx, uc.providerTimeout)
	defer cancel()

	start := time.Now()
	providerName := provider.Name()

	// Panic recovery to prevent one provider from crashing the whole system
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

	// Retry loop using the new Error Taxonomy
	for attempt := 1; attempt <= uc.retryConfig.MaxAttempts; attempt++ {
		flights, lastErr = provider.Search(ctx, criteria)

		if lastErr == nil {
			results <- providerResult{
				Provider: providerName,
				Flights:  flights,
				Error:    nil,
				Duration: time.Since(start),
			}
			return
		}

		// Check if error is transient (retryable) using the new taxonomy
		if !domain.IsTransient(lastErr) {
			log.Debug().Str("provider", providerName).Err(lastErr).Msg("Non-retryable provider error")
			break
		}

		if attempt >= uc.retryConfig.MaxAttempts {
			break
		}

		// Exponential backoff with jitter
		delay := time.Duration(float64(uc.retryConfig.InitialDelay) * math.Pow(uc.retryConfig.Multiplier, float64(attempt-1)))
		jitter := time.Duration((rand.Float64()*0.4 - 0.2) * float64(delay))
		delay += jitter

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			lastErr = ctx.Err()
			break
		}
	}

	// Final result after all retries fail
	results <- providerResult{
		Provider: providerName,
		Error:    lastErr,
		Duration: time.Since(start),
	}
}
