package lionair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	flights := []LionAirFlight{
		{
			ID: "JT-123",
			Carrier: LionAirCarrier{
				IATA: "JT",
				Name: "Lion Air",
			},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK", Name: "Soekarno-Hatta"},
				To:   LionAirAirport{Code: "DPS", Name: "Ngurah Rai"},
			},
			Schedule: LionAirSchedule{
				Departure:         "2025-12-15T06:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T08:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			FlightTime: 120,
			IsDirect:   true,
			Pricing: LionAirPricing{
				Total:    900000,
				Currency: "IDR",
				FareType: "economy",
			},
			Services: LionAirServices{
				BaggageAllowance: LionAirBaggageAllowance{
					Cabin: "7 kg",
					Hold:  "20 kg",
				},
			},
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "JT-123", result[0].FlightNumber)
	assert.Equal(t, "CGK", result[0].Departure.AirportCode)
	assert.Equal(t, "DPS", result[0].Arrival.AirportCode)
	assert.Equal(t, 120, result[0].Duration.TotalMinutes)
	assert.Equal(t, "2h", result[0].Duration.Formatted)
	assert.Equal(t, float64(900000), result[0].Price.Amount)
	assert.Equal(t, "Rp 900.000", result[0].Price.Formatted)
	assert.Equal(t, 0, result[0].Stops)
}

func TestNormalizeFlight(t *testing.T) {
	tests := []struct {
		name        string
		flight      LionAirFlight
		expectError bool
		expectStops int
	}{
		{
			name: "valid direct flight",
			flight: LionAirFlight{
				ID:      "JT-100",
				Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
				Route: LionAirRoute{
					From: LionAirAirport{Code: "CGK"},
					To:   LionAirAirport{Code: "DPS"},
				},
				Schedule: LionAirSchedule{
					Departure:         "2025-12-15T10:00:00",
					DepartureTimezone: "Asia/Jakarta",
					Arrival:           "2025-12-15T12:00:00",
					ArrivalTimezone:   "Asia/Makassar",
				},
				FlightTime: 120,
				IsDirect:   true,
				Pricing:    LionAirPricing{Total: 800000, Currency: "IDR", FareType: "Y"},
				Services: LionAirServices{
					BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "20 kg"},
				},
			},
			expectError: false,
			expectStops: 0,
		},
		{
			name: "connecting flight",
			flight: LionAirFlight{
				ID:      "JT-200",
				Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
				Route: LionAirRoute{
					From: LionAirAirport{Code: "CGK"},
					To:   LionAirAirport{Code: "SIN"},
				},
				Schedule: LionAirSchedule{
					Departure:         "2025-12-15T08:00:00",
					DepartureTimezone: "Asia/Jakarta",
					Arrival:           "2025-12-15T14:00:00",
					ArrivalTimezone:   "Asia/Singapore",
				},
				FlightTime: 300,
				IsDirect:   false,
				StopCount:  1,
				Pricing:    LionAirPricing{Total: 1500000, Currency: "IDR", FareType: "C"},
				Services: LionAirServices{
					BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "30 kg"},
				},
			},
			expectError: false,
			expectStops: 1,
		},
		{
			name: "invalid departure time",
			flight: LionAirFlight{
				ID: "JT-BAD",
				Schedule: LionAirSchedule{
					Departure:         "invalid",
					DepartureTimezone: "Asia/Jakarta",
					Arrival:           "2025-12-15T12:00:00",
					ArrivalTimezone:   "Asia/Jakarta",
				},
			},
			expectError: true,
		},
		{
			name: "invalid arrival time",
			flight: LionAirFlight{
				ID: "JT-BAD2",
				Schedule: LionAirSchedule{
					Departure:         "2025-12-15T10:00:00",
					DepartureTimezone: "Asia/Jakarta",
					Arrival:           "invalid",
					ArrivalTimezone:   "Asia/Jakarta",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeFlight(tt.flight)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectStops, result.Stops)
			}
		})
	}
}

func TestParseDateTimeWithTimezone(t *testing.T) {
	tests := []struct {
		name        string
		datetime    string
		timezone    string
		expectError bool
	}{
		{"valid ISO format", "2025-12-15T10:30:00", "Asia/Jakarta", false},
		{"with space separator", "2025-12-15 10:30:00", "Asia/Jakarta", false},
		{"invalid timezone", "2025-12-15T10:30:00", "Invalid/Zone", false}, // Falls back to UTC
		{"invalid datetime", "invalid", "Asia/Jakarta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTimeWithTimezone(tt.datetime, tt.timezone)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseBaggageWeight(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7 kg", 7},
		{"20kg", 20},
		{"15 KG", 15},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseBaggageWeight(tt.input))
		})
	}
}

func TestNormalizeClass(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"economy", "economy"},
		{"Economy", "economy"},
		{"eco", "economy"},
		{"Y", "economy"},
		{"economy_class", "economy"},
		{"business", "business"},
		{"Business", "business"},
		{"biz", "business"},
		{"J", "business"},
		{"C", "business"},
		{"business_class", "business"},
		{"first", "first"},
		{"First", "first"},
		{"F", "first"},
		{"first_class", "first"},
		{"unknown", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeClass(tt.input))
		})
	}
}

func TestNormalizeWithValidationFailure(t *testing.T) {
	// Flight with arrival before departure
	flights := []LionAirFlight{
		{
			ID:      "JT-BAD",
			Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "DPS"},
			},
			Schedule: LionAirSchedule{
				Departure:         "2025-12-15T12:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T10:00:00", // Before departure
				ArrivalTimezone:   "Asia/Jakarta",
			},
			FlightTime: 120,
			IsDirect:   true,
			Pricing:    LionAirPricing{Total: 800000, Currency: "IDR", FareType: "Y"},
			Services:   LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "20 kg"}},
		},
	}

	result := normalize(flights)
	assert.Empty(t, result)
}

func TestNormalizeFlightWithLayovers(t *testing.T) {
	flight := LionAirFlight{
		ID:      "JT-300",
		Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
		Route: LionAirRoute{
			From: LionAirAirport{Code: "CGK"},
			To:   LionAirAirport{Code: "SIN"},
		},
		Schedule: LionAirSchedule{
			Departure:         "2025-12-15T08:00:00",
			DepartureTimezone: "Asia/Jakarta",
			Arrival:           "2025-12-15T16:00:00",
			ArrivalTimezone:   "Asia/Singapore",
		},
		FlightTime: 360,
		IsDirect:   false,
		StopCount:  0, // No StopCount set
		Layovers: []LionAirLayover{
			{Airport: "KUL", DurationMinutes: 60},
			{Airport: "BTH", DurationMinutes: 45},
		},
		Pricing:  LionAirPricing{Total: 2000000, Currency: "IDR", FareType: "Y"},
		Services: LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "30 kg"}},
	}

	result, err := normalizeFlight(flight)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.Stops) // Should use length of Layovers
}

func TestNormalizeWithMultipleFlights(t *testing.T) {
	flights := []LionAirFlight{
		{
			ID:      "JT-1",
			Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "DPS"},
			},
			Schedule: LionAirSchedule{
				Departure: "2025-12-15T06:00:00", DepartureTimezone: "Asia/Jakarta",
				Arrival: "2025-12-15T08:00:00", ArrivalTimezone: "Asia/Makassar",
			},
			FlightTime: 120, IsDirect: true,
			Pricing:  LionAirPricing{Total: 800000, Currency: "IDR", FareType: "Y"},
			Services: LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "20 kg"}},
		},
		{
			ID:      "JT-2",
			Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "SUB"},
			},
			Schedule: LionAirSchedule{
				Departure: "2025-12-15T09:00:00", DepartureTimezone: "Asia/Jakarta",
				Arrival: "2025-12-15T10:30:00", ArrivalTimezone: "Asia/Jakarta",
			},
			FlightTime: 90, IsDirect: true,
			Pricing:  LionAirPricing{Total: 600000, Currency: "IDR", FareType: "C"},
			Services: LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "25 kg"}},
		},
	}

	result := normalize(flights)
	assert.Len(t, result, 2)
}

func TestNormalizeWithMixedValidAndInvalidFlights(t *testing.T) {
	flights := []LionAirFlight{
		{
			ID:      "JT-VALID",
			Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "DPS"},
			},
			Schedule: LionAirSchedule{
				Departure: "2025-12-15T06:00:00", DepartureTimezone: "Asia/Jakarta",
				Arrival: "2025-12-15T08:00:00", ArrivalTimezone: "Asia/Makassar",
			},
			FlightTime: 120, IsDirect: true,
			Pricing:  LionAirPricing{Total: 800000, Currency: "IDR", FareType: "Y"},
			Services: LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "20 kg"}},
		},
		{
			ID: "JT-INVALID-TIME",
			Schedule: LionAirSchedule{
				Departure: "invalid", DepartureTimezone: "Asia/Jakarta",
				Arrival: "2025-12-15T08:00:00", ArrivalTimezone: "Asia/Jakarta",
			},
		},
		{
			ID:      "JT-VALID-2",
			Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "SUB"},
			},
			Schedule: LionAirSchedule{
				Departure: "2025-12-15T09:00:00", DepartureTimezone: "Asia/Jakarta",
				Arrival: "2025-12-15T10:30:00", ArrivalTimezone: "Asia/Jakarta",
			},
			FlightTime: 90, IsDirect: true,
			Pricing:  LionAirPricing{Total: 600000, Currency: "IDR", FareType: "Y"},
			Services: LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "25 kg"}},
		},
	}

	result := normalize(flights)
	assert.Len(t, result, 2) // Should skip the invalid one
}

func TestNormalizeFlightWithStopCountOnly(t *testing.T) {
	flight := LionAirFlight{
		ID:      "JT-400",
		Carrier: LionAirCarrier{IATA: "JT", Name: "Lion Air"},
		Route: LionAirRoute{
			From: LionAirAirport{Code: "CGK"},
			To:   LionAirAirport{Code: "SIN"},
		},
		Schedule: LionAirSchedule{
			Departure:         "2025-12-15T08:00:00",
			DepartureTimezone: "Asia/Jakarta",
			Arrival:           "2025-12-15T14:00:00",
			ArrivalTimezone:   "Asia/Singapore",
		},
		FlightTime: 300,
		IsDirect:   false,
		StopCount:  2, // StopCount set but no layovers array
		Layovers:   nil,
		Pricing:    LionAirPricing{Total: 1800000, Currency: "IDR", FareType: "Y"},
		Services:   LionAirServices{BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "30 kg"}},
	}

	result, err := normalizeFlight(flight)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.Stops) // Should use StopCount
}
