package flight

import (
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSearchResponse(t *testing.T) {
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		Class:         "economy",
	}

	flights := []domain.Flight{
		{
			ID:           "QZ520_AirAsia",
			FlightNumber: "QZ520",
			Provider:     "AirAsia",
			Airline: domain.AirlineInfo{
				Code: "QZ",
				Name: "AirAsia",
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				AirportName: "Jakarta",
				DateTime:    time.Date(2025, 12, 15, 4, 45, 0, 0, time.UTC),
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				AirportName: "Denpasar",
				DateTime:    time.Date(2025, 12, 15, 7, 25, 0, 0, time.UTC),
			},
			Duration: domain.DurationInfo{
				TotalMinutes: 100,
				Formatted:    "1h 40m",
			},
			Stops: 0,
			Price: domain.PriceInfo{
				Amount:   650000,
				Currency: "IDR",
			},
			Baggage: domain.BaggageInfo{
				CabinKg:   7,
				CheckedKg: 20,
			},
			Class: "economy",
		},
	}

	metadata := Metadata{
		TotalResults:       1,
		ProvidersQueried:   4,
		ProvidersSucceeded: 4,
		ProvidersFailed:    0,
		SearchTimeMs:       285,
		CacheHit:           false,
	}

	response := NewSearchResponse(criteria, flights, metadata)

	// Check search criteria
	assert.Equal(t, "CGK", response.SearchCriteria.Origin)
	assert.Equal(t, "DPS", response.SearchCriteria.Destination)
	assert.Equal(t, "2025-12-15", response.SearchCriteria.DepartureDate)
	assert.Equal(t, 1, response.SearchCriteria.Passengers)
	assert.Equal(t, "economy", response.SearchCriteria.CabinClass)

	// Check metadata
	assert.Equal(t, 1, response.Metadata.TotalResults)
	assert.Equal(t, 4, response.Metadata.ProvidersQueried)
	assert.Equal(t, 4, response.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, response.Metadata.ProvidersFailed)
	assert.Equal(t, int64(285), response.Metadata.SearchTimeMs)
	assert.False(t, response.Metadata.CacheHit)

	// Check flights
	assert.Len(t, response.Flights, 1)
	flight := response.Flights[0]
	assert.Equal(t, "QZ520_AirAsia", flight.ID)
	assert.Equal(t, "AirAsia", flight.Provider)
	assert.Equal(t, "AirAsia", flight.Airline.Name)
	assert.Equal(t, "QZ", flight.Airline.Code)
	assert.Equal(t, "QZ520", flight.FlightNumber)
}

func TestNewSearchResponse_EmptyFlights(t *testing.T) {
	criteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		Class:         "economy",
	}

	metadata := Metadata{
		TotalResults:       0,
		ProvidersQueried:   4,
		ProvidersSucceeded: 4,
		ProvidersFailed:    0,
		SearchTimeMs:       150,
		CacheHit:           false,
	}

	response := NewSearchResponse(criteria, nil, metadata)

	assert.Equal(t, 0, len(response.Flights))
	assert.NotNil(t, response.Flights)
}

func TestToFlightDTO(t *testing.T) {
	departureTime := time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC)
	arrivalTime := time.Date(2025, 12, 15, 8, 50, 0, 0, time.UTC)

	flight := domain.Flight{
		ID:           "GA400",
		FlightNumber: "GA400",
		Provider:     "Garuda Indonesia",
		Airline: domain.AirlineInfo{
			Code: "GA",
			Name: "Garuda Indonesia",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			AirportName: "Jakarta",
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			AirportName: "Denpasar",
			DateTime:    arrivalTime,
		},
		Duration: domain.DurationInfo{
			TotalMinutes: 110,
			Formatted:    "1h 50m",
		},
		Stops: 0,
		Price: domain.PriceInfo{
			Amount:   1250000,
			Currency: "IDR",
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   7,
			CheckedKg: 20,
		},
		Class: "economy",
	}

	dto := ToFlightDTO(flight)

	assert.Equal(t, "GA400", dto.ID)
	assert.Equal(t, "Garuda Indonesia", dto.Provider)
	assert.Equal(t, "Garuda Indonesia", dto.Airline.Name)
	assert.Equal(t, "GA", dto.Airline.Code)
	assert.Equal(t, "GA400", dto.FlightNumber)

	// Check departure
	assert.Equal(t, "CGK", dto.Departure.Airport)
	assert.Equal(t, "Jakarta", dto.Departure.City)
	assert.Equal(t, departureTime.Format(time.RFC3339), dto.Departure.Datetime)
	assert.Equal(t, departureTime.Unix(), dto.Departure.Timestamp)

	// Check arrival
	assert.Equal(t, "DPS", dto.Arrival.Airport)
	assert.Equal(t, "Denpasar", dto.Arrival.City)
	assert.Equal(t, arrivalTime.Format(time.RFC3339), dto.Arrival.Datetime)
	assert.Equal(t, arrivalTime.Unix(), dto.Arrival.Timestamp)

	// Check duration
	assert.Equal(t, 110, dto.Duration.TotalMinutes)
	assert.Equal(t, "1h 50m", dto.Duration.Formatted)

	// Check other fields
	assert.Equal(t, 0, dto.Stops)
	assert.Equal(t, 1250000.0, dto.Price.Amount)
	assert.Equal(t, "IDR", dto.Price.Currency)
	assert.Equal(t, "economy", dto.CabinClass)
	assert.Equal(t, "7kg cabin", dto.Baggage.CarryOn)
	assert.Equal(t, "20kg checked", dto.Baggage.Checked)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{
			name:     "only minutes",
			minutes:  45,
			expected: "45m",
		},
		{
			name:     "only hours",
			minutes:  120,
			expected: "2h",
		},
		{
			name:     "hours and minutes",
			minutes:  110,
			expected: "1h 50m",
		},
		{
			name:     "hours and minutes - multiple hours",
			minutes:  260,
			expected: "4h 20m",
		},
		{
			name:     "zero minutes",
			minutes:  0,
			expected: "0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.minutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBaggage(t *testing.T) {
	// Test baggage formatting in ToFlightDTO
	tests := []struct {
		name        string
		cabinKg     int
		checkedKg   int
		expectedCO  string
		expectedCh  string
	}{
		{
			name:        "zero baggage",
			cabinKg:     0,
			checkedKg:   0,
			expectedCO:  "Not included",
			expectedCh:  "Not included",
		},
		{
			name:        "valid baggage",
			cabinKg:     7,
			checkedKg:   20,
			expectedCO:  "7kg cabin",
			expectedCh:  "20kg checked",
		},
		{
			name:        "cabin only",
			cabinKg:     10,
			checkedKg:   0,
			expectedCO:  "10kg cabin",
			expectedCh:  "Not included",
		},
		{
			name:        "checked only",
			cabinKg:     0,
			checkedKg:   25,
			expectedCO:  "Not included",
			expectedCh:  "25kg checked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flight := domain.Flight{
				ID:           "TEST123",
				FlightNumber: "TEST123",
				Provider:     "Test Airline",
				Airline: domain.AirlineInfo{
					Code: "TA",
					Name: "Test Airline",
				},
				Departure: domain.FlightPoint{
					AirportCode: "CGK",
					DateTime:    time.Now(),
				},
				Arrival: domain.FlightPoint{
					AirportCode: "DPS",
					DateTime:    time.Now(),
				},
				Duration: domain.DurationInfo{
					TotalMinutes: 100,
					Formatted:    "1h 40m",
				},
				Price: domain.PriceInfo{
					Amount:   100000,
					Currency: "IDR",
				},
				Baggage: domain.BaggageInfo{
					CabinKg:   tt.cabinKg,
					CheckedKg: tt.checkedKg,
				},
				Class: "economy",
			}

			dto := ToFlightDTO(flight)
			assert.Equal(t, tt.expectedCO, dto.Baggage.CarryOn)
			assert.Equal(t, tt.expectedCh, dto.Baggage.Checked)
		})
	}
}

func TestToFlightDTO_NilFields(t *testing.T) {
	departureTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	arrivalTime := time.Date(2025, 12, 15, 12, 45, 0, 0, time.UTC)

	flight := domain.Flight{
		ID:           "QZ524",
		FlightNumber: "QZ524",
		Provider:     "AirAsia",
		Airline: domain.AirlineInfo{
			Code: "QZ",
			Name: "AirAsia",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			DateTime:    arrivalTime,
		},
		Duration: domain.DurationInfo{
			TotalMinutes: 105,
			Formatted:    "1h 45m",
		},
		Stops: 0,
		Price: domain.PriceInfo{
			Amount:   720000,
			Currency: "IDR",
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   0,
			CheckedKg: 0,
		},
		Class: "economy",
	}

	dto := ToFlightDTO(flight)

	assert.Nil(t, dto.Aircraft)
	assert.Equal(t, "Not included", dto.Baggage.CarryOn)
	assert.Equal(t, "Not included", dto.Baggage.Checked)
	assert.Empty(t, dto.Amenities)
	assert.Equal(t, "CGK", dto.Departure.City) // Falls back to airport code when name is empty
	assert.Equal(t, "DPS", dto.Arrival.City)
}
