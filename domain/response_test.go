package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSearchResponse(t *testing.T) {
	tests := []struct {
		name           string
		criteria       *SearchCriteria
		flights        []Flight
		metadata       SearchMetadata
		expectedTotal  int
	}{
		{
			name: "with flights",
			criteria: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    2,
				Class:         "economy",
			},
			flights: []Flight{
				{ID: "1", FlightNumber: "GA-123"},
				{ID: "2", FlightNumber: "GA-456"},
			},
			metadata: SearchMetadata{
				ProvidersQueried:   3,
				ProvidersSucceeded: 2,
				ProvidersFailed:    1,
				SearchTimeMs:       150,
			},
			expectedTotal: 2,
		},
		{
			name: "with nil flights",
			criteria: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				Class:         "business",
			},
			flights: nil,
			metadata: SearchMetadata{
				ProvidersQueried: 2,
			},
			expectedTotal: 0,
		},
		{
			name: "with empty flights",
			criteria: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			flights:       []Flight{},
			metadata:      SearchMetadata{},
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewSearchResponse(tt.criteria, tt.flights, tt.metadata)

			assert.Equal(t, tt.criteria.Origin, response.SearchCriteria.Origin)
			assert.Equal(t, tt.criteria.Destination, response.SearchCriteria.Destination)
			assert.Equal(t, tt.criteria.DepartureDate, response.SearchCriteria.DepartureDate)
			assert.Equal(t, tt.criteria.Passengers, response.SearchCriteria.Passengers)
			assert.Equal(t, tt.criteria.Class, response.SearchCriteria.CabinClass)
			assert.Equal(t, tt.expectedTotal, response.Metadata.TotalResults)
			assert.NotNil(t, response.Flights)
			assert.Len(t, response.Flights, tt.expectedTotal)
		})
	}
}

func TestProviderResultIsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   ProviderResult
		expected bool
	}{
		{
			name: "success with flights",
			result: ProviderResult{
				Provider: "garuda",
				Flights:  []Flight{{ID: "1"}},
				Error:    nil,
			},
			expected: true,
		},
		{
			name: "success with no flights",
			result: ProviderResult{
				Provider: "garuda",
				Flights:  []Flight{},
				Error:    nil,
			},
			expected: true,
		},
		{
			name: "failure with error",
			result: ProviderResult{
				Provider: "airasia",
				Flights:  nil,
				Error:    ErrProviderTimeout,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsSuccess())
		})
	}
}

func TestSearchMetadataFields(t *testing.T) {
	metadata := SearchMetadata{
		TotalResults:       10,
		ProvidersQueried:   4,
		ProvidersSucceeded: 3,
		ProvidersFailed:    1,
		SearchTimeMs:       250,
		CacheHit:           true,
	}

	assert.Equal(t, 10, metadata.TotalResults)
	assert.Equal(t, 4, metadata.ProvidersQueried)
	assert.Equal(t, 3, metadata.ProvidersSucceeded)
	assert.Equal(t, 1, metadata.ProvidersFailed)
	assert.Equal(t, int64(250), metadata.SearchTimeMs)
	assert.True(t, metadata.CacheHit)
}

func TestSearchCriteriaResponseFields(t *testing.T) {
	response := SearchCriteriaResponse{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		CabinClass:    "business",
	}

	assert.Equal(t, "CGK", response.Origin)
	assert.Equal(t, "DPS", response.Destination)
	assert.Equal(t, "2025-12-15", response.DepartureDate)
	assert.Equal(t, 2, response.Passengers)
	assert.Equal(t, "business", response.CabinClass)
}
