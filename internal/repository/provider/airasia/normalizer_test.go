package airasia

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ-123",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T06:00:00+07:00",
			ArriveTime:    "2025-12-15T08:30:00+08:00",
			DurationHours: 2.5,
			PriceIDR:      750000,
			CabinClass:    "Economy",
			DirectFlight:  true,
			BaggageNote:   "7kg cabin",
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "QZ-123", result[0].FlightNumber)
	assert.Equal(t, "CGK", result[0].Departure.AirportCode)
	assert.Equal(t, "DPS", result[0].Arrival.AirportCode)
	assert.Equal(t, 150, result[0].Duration.TotalMinutes)
	assert.Equal(t, "2h 30m", result[0].Duration.Formatted)
	assert.Equal(t, float64(750000), result[0].Price.Amount)
	assert.Equal(t, "IDR", result[0].Price.Currency)
	assert.Equal(t, "Rp 750.000", result[0].Price.Formatted)
	assert.Equal(t, "economy", result[0].Class)
	assert.Equal(t, 0, result[0].Stops)
	assert.Equal(t, ProviderName, result[0].Provider)
}

func TestNormalizeSingle(t *testing.T) {
	tests := []struct {
		name        string
		flight      AirAsiaFlight
		expectOK    bool
		expectStops int
	}{
		{
			name: "valid direct flight",
			flight: AirAsiaFlight{
				FlightCode:    "QZ-100",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "DPS",
				DepartTime:    "2025-12-15T10:00:00+07:00",
				ArriveTime:    "2025-12-15T12:30:00+08:00",
				DurationHours: 2.5,
				PriceIDR:      500000,
				CabinClass:    "Economy",
				DirectFlight:  true,
			},
			expectOK:    true,
			expectStops: 0,
		},
		{
			name: "valid connecting flight with stops array",
			flight: AirAsiaFlight{
				FlightCode:    "QZ-200",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "SIN",
				DepartTime:    "2025-12-15T08:00:00+07:00",
				ArriveTime:    "2025-12-15T14:00:00+08:00",
				DurationHours: 5.0,
				PriceIDR:      1200000,
				CabinClass:    "Business",
				DirectFlight:  false,
				Stops:         []AirAsiaStop{{Airport: "KUL"}},
			},
			expectOK:    true,
			expectStops: 1,
		},
		{
			name: "connecting flight without stops array",
			flight: AirAsiaFlight{
				FlightCode:    "QZ-300",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "BKK",
				DepartTime:    "2025-12-15T06:00:00+07:00",
				ArriveTime:    "2025-12-15T12:00:00+07:00",
				DurationHours: 6.0,
				PriceIDR:      1500000,
				DirectFlight:  false,
			},
			expectOK:    true,
			expectStops: 1,
		},
		{
			name: "invalid departure time",
			flight: AirAsiaFlight{
				FlightCode: "QZ-400",
				DepartTime: "invalid-time",
				ArriveTime: "2025-12-15T10:00:00+07:00",
			},
			expectOK: false,
		},
		{
			name: "invalid arrival time",
			flight: AirAsiaFlight{
				FlightCode: "QZ-500",
				DepartTime: "2025-12-15T10:00:00+07:00",
				ArriveTime: "invalid-time",
			},
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := normalizeSingle(tt.flight)

			assert.Equal(t, tt.expectOK, ok)
			if tt.expectOK {
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
		expectHour  int
	}{
		{
			name:        "RFC3339 with colon timezone",
			input:       "2025-12-15T10:30:00+07:00",
			expectError: false,
			expectHour:  10,
		},
		{
			name:        "without colon in timezone",
			input:       "2025-12-15T10:30:00+0700",
			expectError: false,
			expectHour:  10,
		},
		{
			name:        "invalid format",
			input:       "not-a-date",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.input, "CGK")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectHour, result.Hour())
			}
		})
	}
}

func TestHoursToMinutes(t *testing.T) {
	tests := []struct {
		hours    float64
		expected int
	}{
		{1.0, 60},
		{1.5, 90},
		{2.5, 150},
		{0.5, 30},
		{0.0, 0},
		{1.75, 105},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, hoursToMinutes(tt.hours))
		})
	}
}

func TestDirectFlightToStops(t *testing.T) {
	tests := []struct {
		name     string
		isDirect bool
		stops    []AirAsiaStop
		expected int
	}{
		{
			name:     "direct flight",
			isDirect: true,
			stops:    nil,
			expected: 0,
		},
		{
			name:     "connecting with stops array",
			isDirect: false,
			stops:    []AirAsiaStop{{}, {}},
			expected: 2,
		},
		{
			name:     "connecting without stops array",
			isDirect: false,
			stops:    nil,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, directFlightToStops(tt.isDirect, tt.stops))
		})
	}
}

func TestParseBaggageNote(t *testing.T) {
	tests := []struct {
		note          string
		expectCabin   int
		expectChecked int
	}{
		{"Cabin baggage only, checked bags additional fee", 7, 0},
		{"7kg cabin, 20kg checked", 7, 20},
		{"15kg checked baggage included", 7, 15},
		{"", 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.note, func(t *testing.T) {
			cabin, checked := parseBaggageNote(tt.note)
			assert.Equal(t, tt.expectCabin, cabin)
			assert.Equal(t, tt.expectChecked, checked)
		})
	}
}

func TestGenerateFlightID(t *testing.T) {
	flight := AirAsiaFlight{
		FlightCode:  "QZ-123",
		FromAirport: "CGK",
		ToAirport:   "DPS",
	}

	id := generateFlightID(flight)
	assert.Equal(t, "airasia-QZ-123-CGK-DPS", id)
}

func TestExtractAirlineCode(t *testing.T) {
	tests := []struct {
		name       string
		flightCode string
		expected   string
	}{
		{
			name:       "standard QZ flight",
			flightCode: "QZ520",
			expected:   "QZ",
		},
		{
			name:       "flight with hyphen",
			flightCode: "QZ-123",
			expected:   "QZ",
		},
		{
			name:       "lowercase converts to uppercase",
			flightCode: "qz999",
			expected:   "QZ",
		},
		{
			name:       "mixed case",
			flightCode: "Qz100",
			expected:   "QZ",
		},
		{
			name:       "different airline code",
			flightCode: "GA123",
			expected:   "GA",
		},
		{
			name:       "empty string fallback",
			flightCode: "",
			expected:   "QZ", // Falls back to airlineCode constant
		},
		{
			name:       "single char fallback",
			flightCode: "Q",
			expected:   "QZ", // Falls back to airlineCode constant
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAirlineCode(tt.flightCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeWithInvalidFlights(t *testing.T) {
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ-VALID",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T06:00:00+07:00",
			ArriveTime:    "2025-12-15T08:00:00+08:00",
			DurationHours: 2.0,
			PriceIDR:      500000,
			DirectFlight:  true,
		},
		{
			FlightCode: "QZ-INVALID",
			DepartTime: "invalid",
			ArriveTime: "2025-12-15T08:00:00+08:00",
		},
	}

	result := normalize(flights)

	// Should only include valid flight
	assert.Len(t, result, 1)
	assert.Equal(t, "QZ-VALID", result[0].FlightNumber)
}

func TestNormalizeWithValidationFailure(t *testing.T) {
	// Flight where arrival is before departure (validation will fail)
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ-BAD",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T10:00:00+07:00",
			ArriveTime:    "2025-12-15T08:00:00+07:00", // Before departure
			DurationHours: 2.0,
			PriceIDR:      500000,
			DirectFlight:  true,
		},
	}

	result := normalize(flights)

	// Should be empty because validation fails
	assert.Empty(t, result)
}

func BenchmarkNormalize(b *testing.B) {
	flights := make([]AirAsiaFlight, 100)
	baseTime := time.Now()

	for i := range flights {
		flights[i] = AirAsiaFlight{
			FlightCode:    "QZ-" + string(rune('0'+i%10)),
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    baseTime.Format(time.RFC3339),
			ArriveTime:    baseTime.Add(2 * time.Hour).Format(time.RFC3339),
			DurationHours: 2.0,
			PriceIDR:      500000,
			DirectFlight:  true,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalize(flights)
	}
}

func TestFormatBaggageDescriptions(t *testing.T) {
	tests := []struct {
		name          string
		note          string
		cabinKg       int
		checkedKg     int
		expectCarryOn string
		expectChecked string
	}{
		{
			name:          "cabin baggage only with additional fee",
			note:          "Cabin baggage only, checked bags additional fee",
			cabinKg:       7,
			checkedKg:     0,
			expectCarryOn: "Cabin baggage only",
			expectChecked: "Additional fee",
		},
		{
			name:          "numeric baggage with no special note",
			note:          "7kg cabin, 20kg checked",
			cabinKg:       7,
			checkedKg:     20,
			expectCarryOn: "7kg cabin",
			expectChecked: "20kg checked",
		},
		{
			name:          "no baggage included",
			note:          "",
			cabinKg:       0,
			checkedKg:     0,
			expectCarryOn: "Not included",
			expectChecked: "Not included",
		},
		{
			name:          "cabin only no checked",
			note:          "7kg cabin",
			cabinKg:       7,
			checkedKg:     0,
			expectCarryOn: "7kg cabin",
			expectChecked: "Not included",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			carryOn, checked := formatBaggageDescriptions(tt.note, tt.cabinKg, tt.checkedKg)
			assert.Equal(t, tt.expectCarryOn, carryOn)
			assert.Equal(t, tt.expectChecked, checked)
		})
	}
}

func TestNormalizeSingleWithNewFields(t *testing.T) {
	tests := []struct {
		name                 string
		flight               AirAsiaFlight
		expectSeats          int
		expectAircraft       string
		expectAmenities      []string
		expectCarryOnDesc    string
		expectCheckedDesc    string
	}{
		{
			name: "flight with seats and baggage description",
			flight: AirAsiaFlight{
				FlightCode:    "QZ-100",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "DPS",
				DepartTime:    "2025-12-15T10:00:00+07:00",
				ArriveTime:    "2025-12-15T12:30:00+08:00",
				DurationHours: 2.5,
				PriceIDR:      500000,
				CabinClass:    "Economy",
				DirectFlight:  true,
				Seats:         88,
				BaggageNote:   "Cabin baggage only, checked bags additional fee",
			},
			expectSeats:       88,
			expectAircraft:    "",
			expectAmenities:   []string{},
			expectCarryOnDesc: "Cabin baggage only",
			expectCheckedDesc: "Additional fee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := normalizeSingle(tt.flight)
			assert.True(t, ok)
			assert.Equal(t, tt.expectSeats, result.AvailableSeats)
			assert.Equal(t, tt.expectAircraft, result.Aircraft)
			assert.Equal(t, tt.expectAmenities, result.Amenities)
			assert.Equal(t, tt.expectCarryOnDesc, result.Baggage.CarryOnDesc)
			assert.Equal(t, tt.expectCheckedDesc, result.Baggage.CheckedDesc)
		})
	}
}

