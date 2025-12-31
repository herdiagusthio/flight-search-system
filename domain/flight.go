package domain

import (
	"fmt"
	"strconv"
	"time"
)

// Flight represents a single flight offering from a provider.
type Flight struct {
	ID             string       `json:"id"`
	FlightNumber   string       `json:"flightNumber"`
	Airline        AirlineInfo  `json:"airline"`
	Departure      FlightPoint  `json:"departure"`
	Arrival        FlightPoint  `json:"arrival"`
	Duration       DurationInfo `json:"duration"`
	Price          PriceInfo    `json:"price"`
	Baggage        BaggageInfo  `json:"baggage"`
	Class          string       `json:"class"`
	Stops          int          `json:"stops"`
	Provider       string       `json:"provider"`
	RankingScore   float64      `json:"rankingScore,omitempty"`
	AvailableSeats int          `json:"availableSeats"`
	Aircraft       string       `json:"aircraft,omitempty"`
	Amenities      []string     `json:"amenities,omitempty"`
}

// AirlineInfo contains information about an airline.
type AirlineInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Logo string `json:"logo,omitempty"`
}

// FlightPoint represents a departure or arrival point.
type FlightPoint struct {
	AirportCode string    `json:"airportCode"`
	AirportName string    `json:"airportName,omitempty"`
	Terminal    string    `json:"terminal,omitempty"`
	DateTime    time.Time `json:"dateTime"`
	Timezone    string    `json:"timezone,omitempty"`
}

// DurationInfo contains flight duration information.
type DurationInfo struct {
	TotalMinutes int    `json:"totalMinutes"`
	Formatted    string `json:"formatted"`
}

// PriceInfo contains pricing information for a flight.
type PriceInfo struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Formatted string  `json:"formatted,omitempty"`
}

// BaggageInfo contains baggage allowance information.
type BaggageInfo struct {
	CabinKg     int    `json:"cabinKg"`
	CheckedKg   int    `json:"checkedKg"`
	CarryOnDesc string `json:"carryOnDesc,omitempty"`
	CheckedDesc string `json:"checkedDesc,omitempty"`
}

// NewDurationInfo creates a DurationInfo from total minutes and formats it.
func NewDurationInfo(totalMinutes int) DurationInfo {
	hours := totalMinutes / 60
	mins := totalMinutes % 60

	var formatted string
	if hours > 0 && mins > 0 {
		formatted = formatDuration(hours, mins)
	} else if hours > 0 {
		formatted = formatHoursOnly(hours)
	} else {
		formatted = formatMinutesOnly(mins)
	}

	return DurationInfo{
		TotalMinutes: totalMinutes,
		Formatted:    formatted,
	}
}

// formatDuration formats hours and minutes as "Xh Ym".
func formatDuration(hours, mins int) string {
	return strconv.Itoa(hours) + "h " + strconv.Itoa(mins) + "m"
}

// formatHoursOnly formats hours as "Xh".
func formatHoursOnly(hours int) string {
	return strconv.Itoa(hours) + "h"
}

// formatMinutesOnly formats minutes as "Xm".
func formatMinutesOnly(mins int) string {
	return strconv.Itoa(mins) + "m"
}



// Validate checks if the flight data is valid and consistent.
// It returns an error if:
//   - Arrival time is not after departure time
//   - Required fields are missing (FlightNumber, Airline.Code, Origin, Destination)
//
// It logs a warning (but doesn't fail) if:
//   - Duration doesn't match the calculated time difference
//
// This method is used by provider adapters to ensure data integrity.
func (f *Flight) Validate() error {
	// Check that arrival is after departure
	if !f.Arrival.DateTime.After(f.Departure.DateTime) {
		return fmt.Errorf("%w: arrival time (%s) must be after departure time (%s)",
			ErrInvalidFlightTimes,
			f.Arrival.DateTime.Format(time.RFC3339),
			f.Departure.DateTime.Format(time.RFC3339))
	}

	// Check required fields
	if f.FlightNumber == "" {
		return fmt.Errorf("%w: FlightNumber", ErrMissingRequiredField)
	}

	if f.Airline.Code == "" {
		return fmt.Errorf("%w: Airline.Code", ErrMissingRequiredField)
	}

	if f.Departure.AirportCode == "" {
		return fmt.Errorf("%w: Departure.AirportCode", ErrMissingRequiredField)
	}

	if f.Arrival.AirportCode == "" {
		return fmt.Errorf("%w: Arrival.AirportCode", ErrMissingRequiredField)
	}

	// Note: Duration mismatch is logged as a warning in the provider adapters
	// but doesn't fail validation, as providers may calculate it differently

	return nil
}
