package domain

import (
	"errors"
	"fmt"
	"time"
)

// Internal validation errors
var (
	ErrInvalidFlightData = errors.New("invalid flight data")
)

// NewFlightPoint creates a validated FlightPoint.
func NewFlightPoint(code string, name, terminal string, dt time.Time, tz string) (FlightPoint, error) {
	airportCode, err := NewAirportCode(code)
	if err != nil {
		return FlightPoint{}, fmt.Errorf("departure/arrival code: %w", err)
	}

	return FlightPoint{
		AirportCode: airportCode,
		AirportName: name,
		Terminal:    terminal,
		DateTime:    dt,
		Timezone:    tz,
	}, nil
}

// NewFlight creates a validated Flight object.
// It ensures that the flight is logically sound (e.g., arrival after departure)
// and all required value objects are correctly instantiated.
func NewFlight(
	id string,
	flightNum string,
	airline AirlineInfo,
	dep FlightPoint,
	arr FlightPoint,
	dur DurationInfo,
	price PriceInfo,
	baggage BaggageInfo,
	class string,
	stops int,
	provider string,
) (*Flight, error) {
	// 1. Validate Flight Number
	fn, err := NewFlightNumber(flightNum)
	if err != nil {
		return nil, fmt.Errorf("flight number: %w", err)
	}

	// 2. Logical Invariant: Arrival must be after Departure
	if !arr.DateTime.After(dep.DateTime) {
		return nil, fmt.Errorf("%w: arrival time (%s) must be after departure time (%s)",
			ErrInvalidFlightTimes,
			arr.DateTime.Format(time.RFC3339),
			dep.DateTime.Format(time.RFC3339))
	}

	// 3. Required Field Checks
	if airline.Code == "" {
		return nil, fmt.Errorf("%w: Airline.Code", ErrMissingRequiredField)
	}

	return &Flight{
		ID:             id,
		FlightNumber:   fn,
		Airline:        airline,
		Departure:      dep,
		Arrival:        arr,
		Duration:       dur,
		Price:          price,
		Baggage:        baggage,
		Class:          class,
		Stops:          stops,
		Provider:       provider,
		AvailableSeats: 0, // Default
	}, nil
}
