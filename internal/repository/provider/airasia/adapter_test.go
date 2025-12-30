package airasia

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
	adapter := NewAdapter("test/path.json", false)
	assert.NotNil(t, adapter)
	assert.Equal(t, "test/path.json", adapter.mockDataPath)
	assert.False(t, adapter.skipSimulation)
}

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter("", true)
	assert.Equal(t, ProviderName, adapter.Name())
	assert.Equal(t, "airasia", adapter.Name())
}

func TestAdapterSearch(t *testing.T) {
	mockDataPath := filepath.Join("..", "..", "..", "..", "external", "response-mock", "airasia_search_response.json")

	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		t.Skip("Mock data file not found, skipping integration test")
	}

	tests := []struct {
		name          string
		criteria      domain.SearchCriteria
		expectError   bool
		minFlights    int
	}{
		{
			name: "search all flights",
			criteria: domain.SearchCriteria{
				Origin:      "",
				Destination: "",
			},
			expectError: false,
			minFlights:  0,
		},
		{
			name: "search with origin filter",
			criteria: domain.SearchCriteria{
				Origin: "CGK",
			},
			expectError: false,
			minFlights:  0,
		},
		{
			name: "search with destination filter",
			criteria: domain.SearchCriteria{
				Destination: "DPS",
			},
			expectError: false,
			minFlights:  0,
		},
		{
			name: "search with class filter",
			criteria: domain.SearchCriteria{
				Class: "economy",
			},
			expectError: false,
			minFlights:  0,
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
				assert.GreaterOrEqual(t, len(flights), tt.minFlights)
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
	assert.Equal(t, ProviderName, providerErr.Provider)
	assert.True(t, providerErr.Retryable)
}

func TestAdapterSearchWithInvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid json content")
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
	mockDataPath := filepath.Join("..", "..", "..", "..", "external", "response-mock", "airasia_search_response.json")

	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		t.Skip("Mock data file not found, skipping test")
	}

	adapter := NewAdapter(mockDataPath, true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.Search(ctx, domain.SearchCriteria{})

	assert.Error(t, err)
	var providerErr *domain.ProviderError
	assert.ErrorAs(t, err, &providerErr)
	assert.False(t, providerErr.Retryable)
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
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
			},
			Class: "economy",
		},
		{
			ID: "2",
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				DateTime:    now,
			},
			Arrival: domain.FlightPoint{
				AirportCode: "SUB",
			},
			Class: "business",
		},
		{
			ID: "3",
			Departure: domain.FlightPoint{
				AirportCode: "SUB",
				DateTime:    now.Add(24 * time.Hour),
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
			},
			Class: "economy",
		},
	}

	tests := []struct {
		name           string
		criteria       domain.SearchCriteria
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:          "no filter",
			criteria:      domain.SearchCriteria{},
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name:          "filter by origin",
			criteria:      domain.SearchCriteria{Origin: "CGK"},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name:          "filter by destination",
			criteria:      domain.SearchCriteria{Destination: "DPS"},
			expectedCount: 2,
			expectedIDs:   []string{"1", "3"},
		},
		{
			name:          "filter by class",
			criteria:      domain.SearchCriteria{Class: "economy"},
			expectedCount: 2,
			expectedIDs:   []string{"1", "3"},
		},
		{
			name:          "filter by date",
			criteria:      domain.SearchCriteria{DepartureDate: now.Format("2006-01-02")},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name:          "combined filters",
			criteria:      domain.SearchCriteria{Origin: "CGK", Destination: "DPS", Class: "economy"},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "no matches",
			criteria:      domain.SearchCriteria{Origin: "XXX"},
			expectedCount: 0,
			expectedIDs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFlights(flights, tt.criteria)
			assert.Len(t, result, tt.expectedCount)

			resultIDs := make([]string, len(result))
			for i, f := range result {
				resultIDs[i] = f.ID
			}
			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}
