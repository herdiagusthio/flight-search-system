package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements domain.FlightProvider for testing
type mockProvider struct {
	name    string
	flights []domain.Flight
	err     error
	delay   time.Duration
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if m.err != nil {
		return nil, m.err
	}
	return m.flights, nil
}

func TestNewFlightSearchUseCase(t *testing.T) {
	tests := []struct {
		name            string
		providers       []domain.FlightProvider
		config          *Config
		expectGlobal    time.Duration
		expectProvider  time.Duration
	}{
		{
			name:           "nil config uses defaults",
			providers:      []domain.FlightProvider{},
			config:         nil,
			expectGlobal:   DefaultGlobalTimeout,
			expectProvider: DefaultProviderTimeout,
		},
		{
			name:      "custom config",
			providers: []domain.FlightProvider{},
			config: &Config{
				GlobalTimeout:   10 * time.Second,
				ProviderTimeout: 3 * time.Second,
			},
			expectGlobal:   10 * time.Second,
			expectProvider: 3 * time.Second,
		},
		{
			name:      "partial config uses defaults for unset values",
			providers: []domain.FlightProvider{},
			config: &Config{
				GlobalTimeout: 10 * time.Second,
			},
			expectGlobal:   10 * time.Second,
			expectProvider: DefaultProviderTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFlightSearchUseCase(tt.providers, tt.config).(*flightSearchUseCase)
			assert.Equal(t, tt.expectGlobal, uc.globalTimeout)
			assert.Equal(t, tt.expectProvider, uc.providerTimeout)
		})
	}
}

func TestSearch_NoProviders(t *testing.T) {
	uc := NewFlightSearchUseCase([]domain.FlightProvider{}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAllProvidersFailed)
}

func TestSearch_SingleProvider(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
	}

	provider := &mockProvider{
		name:    "test_provider",
		flights: flights,
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Flights))
	assert.Equal(t, 1, result.Metadata.ProvidersQueried)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)
}

func TestSearch_MultipleProviders(t *testing.T) {
	provider1 := &mockProvider{
		name: "provider1",
		flights: []domain.Flight{
			{ID: "p1_f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
		},
	}

	provider2 := &mockProvider{
		name: "provider2",
		flights: []domain.Flight{
			{ID: "p2_f1", Price: domain.PriceInfo{Amount: 600000}, Duration: domain.DurationInfo{TotalMinutes: 140}, Stops: 1},
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider1, provider2}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Flights))
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 2, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)
}

func TestSearch_ProviderError(t *testing.T) {
	provider1 := &mockProvider{
		name: "provider1",
		err:  errors.New("provider error"),
	}

	provider2 := &mockProvider{
		name: "provider2",
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider1, provider2}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Flights))
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)
}

func TestSearch_AllProvidersFail(t *testing.T) {
	provider1 := &mockProvider{
		name: "provider1",
		err:  errors.New("error 1"),
	}

	provider2 := &mockProvider{
		name: "provider2",
		err:  errors.New("error 2"),
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider1, provider2}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAllProvidersFailed)
}

func TestSearch_WithFilters(t *testing.T) {
	provider := &mockProvider{
		name: "test_provider",
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
			{ID: "f2", Price: domain.PriceInfo{Amount: 800000}, Duration: domain.DurationInfo{TotalMinutes: 180}, Stops: 1},
			{ID: "f3", Price: domain.PriceInfo{Amount: 1200000}, Duration: domain.DurationInfo{TotalMinutes: 240}, Stops: 2},
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	maxPrice := float64(900000)
	opts := SearchOptions{
		Filters: &domain.FilterOptions{
			MaxPrice: &maxPrice,
		},
		SortBy: domain.SortByPrice,
	}

	result, err := uc.Search(ctx, criteria, opts)

	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Flights))
	// Should be sorted by price
	assert.Equal(t, "f1", result.Flights[0].ID)
	assert.Equal(t, "f2", result.Flights[1].ID)
}

func TestSearch_WithSorting(t *testing.T) {
	provider := &mockProvider{
		name: "test_provider",
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 800000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
			{ID: "f2", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 180}, Stops: 1},
			{ID: "f3", Price: domain.PriceInfo{Amount: 1200000}, Duration: domain.DurationInfo{TotalMinutes: 240}, Stops: 2},
		},
	}

	tests := []struct {
		name        string
		sortBy      domain.SortOption
		expectedIDs []string
	}{
		{
			name:        "sort by price",
			sortBy:      domain.SortByPrice,
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "sort by duration",
			sortBy:      domain.SortByDuration,
			expectedIDs: []string{"f1", "f2", "f3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
			ctx := context.Background()
			criteria := domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2024-12-25",
				Passengers:    1,
			}

			opts := SearchOptions{
				SortBy: tt.sortBy,
			}

			result, err := uc.Search(ctx, criteria, opts)

			require.NoError(t, err)
			assert.Equal(t, 3, len(result.Flights))

			ids := make([]string, len(result.Flights))
			for i, f := range result.Flights {
				ids[i] = f.ID
			}
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}

func TestSearch_Timeout(t *testing.T) {
	// Provider that takes too long
	provider := &mockProvider{
		name:  "slow_provider",
		delay: 3 * time.Second,
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 500000}},
		},
	}

	config := &Config{
		GlobalTimeout:   100 * time.Millisecond,
		ProviderTimeout: 50 * time.Millisecond,
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, config)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAllProvidersFailed)
}

func TestSearch_ContextCancellation(t *testing.T) {
	provider := &mockProvider{
		name:  "provider",
		delay: 1 * time.Second,
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 500000}},
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	// Cancel immediately
	cancel()

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAllProvidersFailed)
}

func TestSearch_EmptyResults(t *testing.T) {
	provider := &mockProvider{
		name:    "provider",
		flights: []domain.Flight{}, // No flights
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	require.NoError(t, err)
	assert.Equal(t, 0, len(result.Flights))
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, DefaultGlobalTimeout, config.GlobalTimeout)
	assert.Equal(t, DefaultProviderTimeout, config.ProviderTimeout)
}

func TestDefaultSearchOptions(t *testing.T) {
	opts := DefaultSearchOptions()
	assert.Nil(t, opts.Filters)
	assert.Equal(t, domain.SortByBestValue, opts.SortBy)
}

// panicProvider is a mock provider that panics during search
type panicProvider struct {
	name string
}

func (p *panicProvider) Name() string {
	return p.name
}

func (p *panicProvider) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	panic("provider panic!")
}

func TestSearch_ProviderPanic(t *testing.T) {
	provider1 := &panicProvider{name: "panic_provider"}
	provider2 := &mockProvider{
		name: "good_provider",
		flights: []domain.Flight{
			{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
		},
	}

	uc := NewFlightSearchUseCase([]domain.FlightProvider{provider1, provider2}, nil)
	ctx := context.Background()
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2024-12-25",
		Passengers:    1,
	}

	result, err := uc.Search(ctx, criteria, DefaultSearchOptions())

	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Flights))
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)
}
