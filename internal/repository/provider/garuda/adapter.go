package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
)

const (
	ProviderName = "garuda_indonesia"
	DefaultCabinBaggageKg = 7
	DefaultCheckedBaggageKg = 20
)

// Adapter implements the domain.FlightProvider interface for Garuda Indonesia.
// It reads from mock JSON data and normalizes it to the unified Flight domain model.
type Adapter struct {
	// mockDataPath is the path to the mock JSON data file.
	mockDataPath string
	// skipSimulation disables delay simulation for deterministic testing.
	skipSimulation bool
}

// NewAdapter creates a new Garuda Indonesia adapter.
// The mockDataPath parameter specifies the path to the mock JSON data file.
func NewAdapter(mockDataPath string, skipSimulation bool) *Adapter {
	return &Adapter{
		mockDataPath:   mockDataPath,
		skipSimulation: skipSimulation,
	}
}

// Name returns the unique identifier for this provider.
// Implements domain.FlightProvider.
func (a *Adapter) Name() string {
	return ProviderName
}

// Search queries the provider for available flights matching the criteria.
// It reads from mock JSON data and returns normalized flight entities.
// Simulates real-world conditions: Fast response (50-100ms delay).
// Implements domain.FlightProvider.
func (a *Adapter) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	if !a.skipSimulation {
		delay := time.Duration(50+rand.Intn(51)) * time.Millisecond
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, &domain.ProviderError{
				Provider:  ProviderName,
				Err:       ctx.Err(),
				Retryable: false,
			}
		}
	}

	select {
	case <-ctx.Done():
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       ctx.Err(),
			Retryable: false,
		}
	default:
	}

	data, err := os.ReadFile(a.mockDataPath)
	if err != nil {
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       fmt.Errorf("failed to read mock data: %w", err),
			Retryable: true,
		}
	}

	var response GarudaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       fmt.Errorf("failed to parse JSON: %w", err),
			Retryable: false,
		}
	}

	if len(response.Flights) == 0 {
		return []domain.Flight{}, nil
	}

	flights := normalize(response.Flights)
	return filterFlights(flights, criteria), nil
}

func filterFlights(flights []domain.Flight, criteria domain.SearchCriteria) []domain.Flight {
	result := make([]domain.Flight, 0, len(flights))

	for _, f := range flights {
		if criteria.Origin != "" && f.Departure.AirportCode != criteria.Origin {
			continue
		}
		if criteria.Destination != "" && f.Arrival.AirportCode != criteria.Destination {
			continue
		}
		if criteria.DepartureDate != "" {
			flightDate := f.Departure.DateTime.Format("2006-01-02")
			if flightDate != criteria.DepartureDate {
				continue
			}
		}
		if criteria.Class != "" && f.Class != criteria.Class {
			continue
		}
		result = append(result, f)
	}

	return result
}