package lionair

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdapter(t *testing.T) {
	adapter := NewAdapter("test/path.json", true)
	assert.NotNil(t, adapter)
	assert.True(t, adapter.skipSimulation)
}

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter("", true)
	assert.Equal(t, ProviderName, adapter.Name())
	assert.Equal(t, "lion_air", adapter.Name())
}

func TestAdapterSearch(t *testing.T) {
	mockDataPath := filepath.Join("..", "..", "..", "..", "external", "response-mock", "lion_air_search_response.json")

	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		t.Skip("Mock data file not found")
	}

	adapter := NewAdapter(mockDataPath, true)
	ctx := context.Background()

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	assert.NoError(t, err)
	assert.NotNil(t, flights)
}

func TestAdapterSearchWithInvalidPath(t *testing.T) {
	adapter := NewAdapter("nonexistent.json", true)
	ctx := context.Background()

	_, err := adapter.Search(ctx, domain.SearchCriteria{})

	assert.Error(t, err)
	var providerErr *domain.ProviderError
	assert.ErrorAs(t, err, &providerErr)
	assert.True(t, providerErr.Retryable)
}

func TestAdapterSearchWithInvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid")
	require.NoError(t, err)
	tmpFile.Close()

	adapter := NewAdapter(tmpFile.Name(), true)
	_, err = adapter.Search(context.Background(), domain.SearchCriteria{})

	assert.Error(t, err)
}

func TestAdapterSearchWithEmptyFlights(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{"data": {"available_flights": []}}`)
	require.NoError(t, err)
	tmpFile.Close()

	adapter := NewAdapter(tmpFile.Name(), true)
	flights, err := adapter.Search(context.Background(), domain.SearchCriteria{})

	assert.NoError(t, err)
	assert.Empty(t, flights)
}

func TestFilterFlights(t *testing.T) {
	now := time.Now()
	flights := []domain.Flight{
		{ID: "1", Departure: domain.FlightPoint{AirportCode: "CGK", DateTime: now}, Arrival: domain.FlightPoint{AirportCode: "DPS"}, Class: "economy"},
		{ID: "2", Departure: domain.FlightPoint{AirportCode: "CGK", DateTime: now}, Arrival: domain.FlightPoint{AirportCode: "SUB"}, Class: "business"},
	}

	tests := []struct {
		name          string
		criteria      domain.SearchCriteria
		expectedCount int
	}{
		{"no filter", domain.SearchCriteria{}, 2},
		{"filter by origin", domain.SearchCriteria{Origin: "CGK"}, 2},
		{"filter by destination", domain.SearchCriteria{Destination: "DPS"}, 1},
		{"filter by class", domain.SearchCriteria{Class: "economy"}, 1},
		{"filter by date", domain.SearchCriteria{DepartureDate: now.Format("2006-01-02")}, 2},
		{"no matches", domain.SearchCriteria{Origin: "XXX"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFlights(flights, tt.criteria)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestAdapterSearchWithContextTimeout(t *testing.T) {
	adapter := NewAdapter("test.json", true)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	_, err := adapter.Search(ctx, domain.SearchCriteria{})
	assert.Error(t, err)
}

func TestAdapterSearchFiltering(t *testing.T) {
	mockDataPath := filepath.Join("..", "..", "..", "..", "external", "response-mock", "lion_air_search_response.json")

	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		t.Skip("Mock data file not found")
	}

	tests := []struct {
		name     string
		criteria domain.SearchCriteria
	}{
		{"filter by origin", domain.SearchCriteria{Origin: "CGK"}},
		{"filter by destination", domain.SearchCriteria{Destination: "DPS"}},
		{"filter by class", domain.SearchCriteria{Class: "economy"}},
		{"filter by date", domain.SearchCriteria{DepartureDate: "2025-12-15"}},
		{"combined filters", domain.SearchCriteria{Origin: "CGK", Destination: "DPS"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(mockDataPath, true)
			_, err := adapter.Search(context.Background(), tt.criteria)
			assert.NoError(t, err)
		})
	}
}
