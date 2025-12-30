package garuda

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
	assert.Equal(t, "test/path.json", adapter.mockDataPath)
	assert.True(t, adapter.skipSimulation)
}

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter("", true)
	assert.Equal(t, ProviderName, adapter.Name())
	assert.Equal(t, "garuda_indonesia", adapter.Name())
}

func TestAdapterSearch(t *testing.T) {
	mockDataPath := filepath.Join("..", "..", "..", "..", "external", "response-mock", "garuda_indonesia_search_response.json")

	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		t.Skip("Mock data file not found, skipping integration test")
	}

	tests := []struct {
		name        string
		criteria    domain.SearchCriteria
		expectError bool
	}{
		{
			name:        "search all flights",
			criteria:    domain.SearchCriteria{},
			expectError: false,
		},
		{
			name:        "search with origin filter",
			criteria:    domain.SearchCriteria{Origin: "CGK"},
			expectError: false,
		},
		{
			name:        "search with destination filter",
			criteria:    domain.SearchCriteria{Destination: "DPS"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(mockDataPath, true)
			ctx := context.Background()

			flights, err := adapter.Search(ctx, tt.criteria)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, flights)
			}
		})
	}
}

func TestAdapterSearchWithInvalidPath(t *testing.T) {
	adapter := NewAdapter("nonexistent/path.json", true)
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

	_, err = tmpFile.WriteString("invalid json")
	require.NoError(t, err)
	tmpFile.Close()

	adapter := NewAdapter(tmpFile.Name(), true)
	ctx := context.Background()

	_, err = adapter.Search(ctx, domain.SearchCriteria{})

	assert.Error(t, err)
	var providerErr *domain.ProviderError
	assert.ErrorAs(t, err, &providerErr)
	assert.False(t, providerErr.Retryable)
}

func TestAdapterSearchWithContextCancellation(t *testing.T) {
	adapter := NewAdapter("test.json", true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.Search(ctx, domain.SearchCriteria{})

	assert.Error(t, err)
}

func TestAdapterSearchWithEmptyFlights(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{"flights": []}`)
	require.NoError(t, err)
	tmpFile.Close()

	adapter := NewAdapter(tmpFile.Name(), true)
	ctx := context.Background()

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	assert.NoError(t, err)
	assert.Empty(t, flights)
}

func TestFilterFlights(t *testing.T) {
	now := time.Now()
	flights := []domain.Flight{
		{
			ID: "1",
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				DateTime:    now,
			},
			Arrival: domain.FlightPoint{AirportCode: "DPS"},
			Class:   "economy",
		},
		{
			ID: "2",
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				DateTime:    now,
			},
			Arrival: domain.FlightPoint{AirportCode: "SUB"},
			Class:   "business",
		},
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
