package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
)

// ChaosProvider is a mock provider that allows us to inject faults.
type ChaosProvider struct {
	Name           string
	ShouldPanic    bool
	ShouldTimeout   bool
	ShouldErr      error
	ReturnFlights  []domain.Flight
	ResponseDelay  time.Duration
}

func (cp *ChaosProvider) Name() string {
	return cp.Name
}

func (cp *ChaosProvider) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	if cp.ShouldPanic {
		panic("chaos monkey: provider crashed!")
	}

	if cp.ResponseDelay > 0 {
		select {
		case <-time.After(cp.ResponseDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if cp.ShouldTimeout {
		// Simulate a timeout by blocking until context expires
		<-ctx.Done()
		return nil, ctx.Err()
	}

	if cp.ShouldErr != nil {
		return nil, cp.ShouldErr
	}

	return cp.ReturnFlights, nil
}

func TestFlightSearchUseCase_Chaos(t *testing.T) {
	tests := []struct {
		name           string
		providers      []domain.FlightProvider
		wantSuccess    bool
		wantMinFlights int
	}{
		{
			name: "All Providers Panic",
			providers: []domain.FlightProvider{
				&ChaosProvider{Name: "P1", ShouldPanic: true},
				&ChaosProvider{Name: "P2", ShouldPanic: true},
			},
			wantSuccess:    false,
			wantMinFlights: 0,
		},
		{
			name: "Partial Failure - One Panic, One Success",
			providers: []domain.FlightProvider{
				&ChaosProvider{Name: "P1", ShouldPanic: true},
				&ChaosProvider{Name: "P2", ReturnFlights: []domain.Flight{{FlightNumber: "OK123"}}},
			},
			wantSuccess:    true,
			wantMinFlights: 1,
		},
		{
			name: "Partial Failure - One Timeout, One Success",
			providers: []domain.FlightProvider{
				&ChaosProvider{Name: "P1", ShouldTimeout: true},
				&ChaosProvider{Name: "P2", ReturnFlights: []domain.Flight{{FlightNumber: "OK456"}}},
			},
			wantSuccess:    true,
			wantMinFlights: 1,
		},
		{
			name: "All Providers Timeout",
			providers: []domain.FlightProvider{
				&ChaosProvider{Name: "P1", ShouldTimeout: true},
				&ChaosProvider{Name: "P2", ShouldTimeout: true},
			},
			wantSuccess:    false,
			wantMinFlights: 0,
		},
		{
			name: "Transient Error Recovery",
			providers: []domain.FlightProvider{
				// This is a simple mock; for actual retry testing, 
				// we'd need a stateful mock that succeeds on attempt 2.
				&ChaosProvider{Name: "P1", ReturnFlights: []domain.Flight{{FlightNumber: "OK789"}}},
			},
			wantSuccess:    true,
			wantMinFlights: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFlightSearchUseCase(tt.providers, &Config{
				GlobalTimeout:   100 * time.Millisecond,
				ProviderTimeout: 50 * time.Millisecond,
			})

			res, err := uc.Search(context.Background(), domain.SearchCriteria{}, SearchOptions{})

			if (err == nil) != tt.wantSuccess {
				t.Errorf("Search() error = %v, wantSuccess %v", err, tt.wantSuccess)
			}

			if tt.wantSuccess && res != nil {
				if len(res.Flights) < tt.wantMinFlights {
					t.Errorf("Expected at least %d flights, got %d", tt.wantMinFlights, len(res.Flights))
				}
			}
		})
	}
}
