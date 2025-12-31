package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDurationInfo(t *testing.T) {
	tests := []struct {
		name             string
		totalMinutes     int
		expectedFormatted string
	}{
		{
			name:             "hours and minutes",
			totalMinutes:     150,
			expectedFormatted: "2h 30m",
		},
		{
			name:             "hours only",
			totalMinutes:     120,
			expectedFormatted: "2h",
		},
		{
			name:             "minutes only",
			totalMinutes:     45,
			expectedFormatted: "45m",
		},
		{
			name:             "zero minutes",
			totalMinutes:     0,
			expectedFormatted: "0m",
		},
		{
			name:             "single hour",
			totalMinutes:     60,
			expectedFormatted: "1h",
		},
		{
			name:             "single minute",
			totalMinutes:     1,
			expectedFormatted: "1m",
		},
		{
			name:             "complex duration",
			totalMinutes:     185,
			expectedFormatted: "3h 5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDurationInfo(tt.totalMinutes)
			assert.Equal(t, tt.totalMinutes, result.TotalMinutes)
			assert.Equal(t, tt.expectedFormatted, result.Formatted)
		})
	}
}

func TestFlightValidate(t *testing.T) {
	now := time.Now()
	validFlight := Flight{
		ID:           "test-id",
		FlightNumber: "GA-123",
		Airline: AirlineInfo{
			Code: "GA",
			Name: "Garuda Indonesia",
		},
		Departure: FlightPoint{
			AirportCode: "CGK",
			DateTime:    now,
		},
		Arrival: FlightPoint{
			AirportCode: "DPS",
			DateTime:    now.Add(2 * time.Hour),
		},
		Duration: DurationInfo{
			TotalMinutes: 120,
			Formatted:    "2h",
		},
		Price: PriceInfo{
			Amount:   1500000,
			Currency: "IDR",
		},
		Class:    "economy",
		Stops:    0,
		Provider: "garuda",
	}

	tests := []struct {
		name        string
		modifyFlight func(*Flight)
		expectError bool
		errorType   error
	}{
		{
			name:         "valid flight",
			modifyFlight: nil,
			expectError:  false,
		},
		{
			name: "arrival before departure",
			modifyFlight: func(f *Flight) {
				f.Arrival.DateTime = f.Departure.DateTime.Add(-1 * time.Hour)
			},
			expectError: true,
			errorType:   ErrInvalidFlightTimes,
		},
		{
			name: "arrival same as departure",
			modifyFlight: func(f *Flight) {
				f.Arrival.DateTime = f.Departure.DateTime
			},
			expectError: true,
			errorType:   ErrInvalidFlightTimes,
		},
		{
			name: "missing flight number",
			modifyFlight: func(f *Flight) {
				f.FlightNumber = ""
			},
			expectError: true,
			errorType:   ErrMissingRequiredField,
		},
		{
			name: "missing airline code",
			modifyFlight: func(f *Flight) {
				f.Airline.Code = ""
			},
			expectError: true,
			errorType:   ErrMissingRequiredField,
		},
		{
			name: "missing departure airport",
			modifyFlight: func(f *Flight) {
				f.Departure.AirportCode = ""
			},
			expectError: true,
			errorType:   ErrMissingRequiredField,
		},
		{
			name: "missing arrival airport",
			modifyFlight: func(f *Flight) {
				f.Arrival.AirportCode = ""
			},
			expectError: true,
			errorType:   ErrMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flight := validFlight
			if tt.modifyFlight != nil {
				tt.modifyFlight(&flight)
			}

			err := flight.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}



func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "2h 30m", formatDuration(2, 30))
	assert.Equal(t, "1h 5m", formatDuration(1, 5))
}

func TestFormatHoursOnly(t *testing.T) {
	assert.Equal(t, "2h", formatHoursOnly(2))
	assert.Equal(t, "10h", formatHoursOnly(10))
}

func TestFormatMinutesOnly(t *testing.T) {
	assert.Equal(t, "30m", formatMinutesOnly(30))
	assert.Equal(t, "5m", formatMinutesOnly(5))
}
