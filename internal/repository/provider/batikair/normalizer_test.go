package batikair

import (
	"testing"

	"github.com/herdiagusthio/flight-search-system/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	flights := []entity.BatikAirFlight{
		{
			FlightNumber:      "ID-123",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "DPS",
			DepartureDateTime: "2025-12-15T06:00:00+07:00",
			ArrivalDateTime:   "2025-12-15T08:00:00+08:00",
			TravelTime:        "2h 0m",
			NumberOfStops:     0,
			Fare: entity.BatikAirFare{
				BasePrice:    1000000,
				Taxes:        200000,
				TotalPrice:   1200000,
				CurrencyCode: "IDR",
				Class:        "Y",
			},
			BaggageInfo: "7kg cabin, 20kg checked",
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "ID-123", result[0].FlightNumber)
	assert.Equal(t, "CGK", result[0].Departure.AirportCode)
	assert.Equal(t, "DPS", result[0].Arrival.AirportCode)
	assert.Equal(t, 120, result[0].Duration.TotalMinutes)
	assert.Equal(t, float64(1200000), result[0].Price.Amount)
}

func TestNormalizeFlight(t *testing.T) {
	tests := []struct {
		name        string
		flight      entity.BatikAirFlight
		expectError bool
	}{
		{
			name: "valid flight",
			flight: entity.BatikAirFlight{
				FlightNumber:      "ID-100",
				AirlineName:       "Batik Air",
				AirlineIATA:       "ID",
				Origin:            "CGK",
				Destination:       "DPS",
				DepartureDateTime: "2025-12-15T10:00:00+07:00",
				ArrivalDateTime:   "2025-12-15T12:00:00+08:00",
				TravelTime:        "2h 0m",
				Fare:              entity.BatikAirFare{TotalPrice: 1000000, CurrencyCode: "IDR", Class: "Y"},
			},
			expectError: false,
		},
		{
			name: "invalid departure",
			flight: entity.BatikAirFlight{
				DepartureDateTime: "invalid",
				ArrivalDateTime:   "2025-12-15T12:00:00+07:00",
				TravelTime:        "2h",
			},
			expectError: true,
		},
		{
			name: "invalid arrival",
			flight: entity.BatikAirFlight{
				DepartureDateTime: "2025-12-15T10:00:00+07:00",
				ArrivalDateTime:   "invalid",
				TravelTime:        "2h",
			},
			expectError: true,
		},
		{
			name: "invalid travel time",
			flight: entity.BatikAirFlight{
				DepartureDateTime: "2025-12-15T10:00:00+07:00",
				ArrivalDateTime:   "2025-12-15T12:00:00+07:00",
				TravelTime:        "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeFlight(tt.flight)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"RFC3339", "2025-12-15T10:30:00+07:00", false},
		{"without colon", "2025-12-15T10:30:00+0700", false},
		{"no timezone", "2025-12-15T10:30:00", false},
		{"invalid", "bad", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTime(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseDurationString(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		expectError bool
	}{
		{"2h 30m", 150, false},
		{"1h", 60, false},
		{"45m", 45, false},
		{"0h 30m", 30, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDurationString(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseBaggageInfo(t *testing.T) {
	tests := []struct {
		input         string
		expectCabin   int
		expectChecked int
	}{
		{"7kg cabin, 20kg checked", 7, 20},
		{"10kg cabin, 25kg checked", 10, 25},
		{"", 7, 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cabin, checked := parseBaggageInfo(tt.input)
			assert.Equal(t, tt.expectCabin, cabin)
			assert.Equal(t, tt.expectChecked, checked)
		})
	}
}

func TestMapCabinClass(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Y", "economy"},
		{"W", "premium_economy"},
		{"C", "business"},
		{"J", "business"},
		{"F", "first"},
		{"X", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapCabinClass(tt.input))
		})
	}
}

func TestNormalizeWithValidationFailure(t *testing.T) {
	// Flight with arrival before departure
	flights := []entity.BatikAirFlight{
		{
			FlightNumber:      "ID-BAD",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "DPS",
			DepartureDateTime: "2025-12-15T12:00:00+07:00",
			ArrivalDateTime:   "2025-12-15T10:00:00+07:00", // Before departure
			TravelTime:        "2h 0m",
			NumberOfStops:     0,
			Fare: entity.BatikAirFare{
				TotalPrice:   1000000,
				CurrencyCode: "IDR",
				Class:        "Y",
			},
		},
	}

	result := normalize(flights)
	assert.Empty(t, result)
}

func TestNormalizeFlightPriceFallback(t *testing.T) {
	// Flight with zero TotalPrice, should use BasePrice + Taxes
	flight := entity.BatikAirFlight{
		FlightNumber:      "ID-100",
		AirlineName:       "Batik Air",
		AirlineIATA:       "ID",
		Origin:            "CGK",
		Destination:       "DPS",
		DepartureDateTime: "2025-12-15T10:00:00+07:00",
		ArrivalDateTime:   "2025-12-15T12:00:00+08:00",
		TravelTime:        "2h 0m",
		Fare: entity.BatikAirFare{
			BasePrice:    800000,
			Taxes:        100000,
			TotalPrice:   0,
			CurrencyCode: "IDR",
			Class:        "Y",
		},
	}

	result, err := normalizeFlight(flight)
	assert.NoError(t, err)
	assert.Equal(t, float64(900000), result.Price.Amount) // BasePrice + Taxes
}

func TestNormalizeWithMultipleFlights(t *testing.T) {
	flights := []entity.BatikAirFlight{
		{
			FlightNumber:      "ID-1",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "DPS",
			DepartureDateTime: "2025-12-15T06:00:00+07:00",
			ArrivalDateTime:   "2025-12-15T08:00:00+08:00",
			TravelTime:        "2h 0m",
			Fare:              entity.BatikAirFare{TotalPrice: 1000000, CurrencyCode: "IDR", Class: "Y"},
		},
		{
			FlightNumber:      "ID-2",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "SUB",
			DepartureDateTime: "2025-12-15T09:00:00+07:00",
			ArrivalDateTime:   "2025-12-15T10:30:00+07:00",
			TravelTime:        "1h 30m",
			Fare:              entity.BatikAirFare{TotalPrice: 800000, CurrencyCode: "IDR", Class: "C"},
		},
	}

	result := normalize(flights)
	assert.Len(t, result, 2)
}
