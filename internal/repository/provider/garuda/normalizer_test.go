package garuda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	flights := []GarudaFlight{
		{
			FlightID:    "GA-123",
			Airline:     "Garuda Indonesia",
			AirlineCode: "GA",
			Departure: GarudaEndpoint{
				Airport:  "CGK",
				City:     "Jakarta",
				Terminal: "3",
				Time:     "2025-12-15T06:00:00+07:00",
			},
			Arrival: GarudaEndpoint{
				Airport: "DPS",
				City:    "Bali",
				Time:    "2025-12-15T08:00:00+08:00",
			},
			DurationMinutes: 120,
			Price: GarudaPrice{
				Amount:   1500000,
				Currency: "IDR",
			},
			FareClass: "Economy",
			Baggage: GarudaBaggage{
				CarryOn: 1,
				Checked: 1,
			},
			Stops: 0,
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "GA-123", result[0].FlightNumber)
	assert.Equal(t, "CGK", result[0].Departure.AirportCode)
	assert.Equal(t, "DPS", result[0].Arrival.AirportCode)
	assert.Equal(t, 120, result[0].Duration.TotalMinutes)
	assert.Equal(t, float64(1500000), result[0].Price.Amount)
	assert.Equal(t, "Rp 1.500.000", result[0].Price.Formatted)
	assert.Equal(t, "economy", result[0].Class)
}

func TestNormalizeFlight(t *testing.T) {
	tests := []struct {
		name        string
		flight      GarudaFlight
		expectError bool
		expectStops int
	}{
		{
			name: "valid direct flight",
			flight: GarudaFlight{
				FlightID:    "GA-100",
				Airline:     "Garuda Indonesia",
				AirlineCode: "GA",
				Departure: GarudaEndpoint{
					Airport: "CGK",
					Time:    "2025-12-15T10:00:00+07:00",
				},
				Arrival: GarudaEndpoint{
					Airport: "DPS",
					Time:    "2025-12-15T12:00:00+08:00",
				},
				DurationMinutes: 120,
				Price:           GarudaPrice{Amount: 1000000, Currency: "IDR"},
				FareClass:       "Y",
				Stops:           0,
			},
			expectError: false,
			expectStops: 0,
		},
		{
			name: "flight with segments",
			flight: GarudaFlight{
				FlightID:    "GA-200",
				Airline:     "Garuda Indonesia",
				AirlineCode: "GA",
				Departure: GarudaEndpoint{
					Airport: "CGK",
					Time:    "2025-12-15T08:00:00+07:00",
				},
				Arrival: GarudaEndpoint{
					Airport: "SIN",
					Time:    "2025-12-15T14:00:00+08:00",
				},
				DurationMinutes: 300,
				Price:           GarudaPrice{Amount: 2000000, Currency: "IDR"},
				FareClass:       "C",
				Stops:           1,
				Segments:        []GarudaSegment{{}, {}},
			},
			expectError: false,
			expectStops: 1,
		},
		{
			name: "invalid departure time",
			flight: GarudaFlight{
				FlightID: "GA-BAD",
				Departure: GarudaEndpoint{
					Airport: "CGK",
					Time:    "invalid",
				},
				Arrival: GarudaEndpoint{
					Airport: "DPS",
					Time:    "2025-12-15T12:00:00+07:00",
				},
			},
			expectError: true,
		},
		{
			name: "invalid arrival time",
			flight: GarudaFlight{
				FlightID: "GA-BAD2",
				Departure: GarudaEndpoint{
					Airport: "CGK",
					Time:    "2025-12-15T10:00:00+07:00",
				},
				Arrival: GarudaEndpoint{
					Airport: "DPS",
					Time:    "invalid",
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

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"RFC3339 format", "2025-12-15T10:30:00+07:00", false},
		{"without timezone", "2025-12-15T10:30:00", false},
		{"invalid format", "not-a-date", true},
		{"empty string", "", true},
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

func TestFormatAirportName(t *testing.T) {
	tests := []struct {
		code     string
		city     string
		expected string
	}{
		{"CGK", "Jakarta", "Jakarta (CGK)"},
		{"DPS", "", "DPS"},
		{"SUB", "Surabaya", "Surabaya (SUB)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatAirportName(tt.code, tt.city))
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
		{"business", "business"},
		{"Business", "business"},
		{"biz", "business"},
		{"J", "business"},
		{"C", "business"},
		{"first", "first"},
		{"First", "first"},
		{"F", "first"},
		{"unknown", "economy"},
		{"", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeClass(tt.input))
		})
	}
}
