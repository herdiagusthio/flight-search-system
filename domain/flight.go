package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidAirportCode = errors.New("invalid airport code: must be 3 uppercase letters")
	ErrInvalidCurrency    = errors.New("invalid currency: must be 3 uppercase letters (ISO 4217)")
	ErrInvalidFlightNum   = errors.New("invalid flight number: cannot be empty")
)

// AirportCode represents an IATA airport code (e.g., CGK, SIN).
type AirportCode string

func NewAirportCode(code string) (AirportCode, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if matched, _ := regexp.MatchString(`^[A-Z]{3}$`, code); !matched {
		return "", ErrInvalidAirportCode
	}
	return AirportCode(code), nil
}

func (a AirportCode) String() string {
	return string(a)
}

// Currency represents an ISO 4217 currency code (e.g., USD, IDR).
type Currency string

func NewCurrency(c string) (Currency, error) {
	c = strings.ToUpper(strings.TrimSpace(c))
	if matched, _ := regexp.MatchString(`^[A-Z]{3}$`, c); !matched {
		return "", ErrInvalidCurrency
	}
	return Currency(c), nil
}

func (c Currency) String() string {
	return string(c)
}

// FlightNumber represents a provider's flight identifier.
type FlightNumber string

func NewFlightNumber(fn string) (FlightNumber, error) {
	fn = strings.TrimSpace(fn)
	if fn == "" {
		return "", ErrInvalidFlightNum
	}
	return FlightNumber(fn), nil
}

func (fn FlightNumber) String() string {
	return string(fn)
}

type Flight struct {
	ID             string
	FlightNumber   FlightNumber
	Airline        AirlineInfo
	Departure      FlightPoint
	Arrival        FlightPoint
	Duration       DurationInfo
	Price          PriceInfo
	Baggage        BaggageInfo
	Class          string
	Stops          int
	Provider       string
	RankingScore   float64
	AvailableSeats int
	Aircraft       string
	Amenities      []string
}

type AirlineInfo struct {
	Code string
	Name string
	Logo string
}

type FlightPoint struct {
	AirportCode AirportCode
	AirportName string
	Terminal    string
	DateTime    time.Time
	Timezone    string
}

type DurationInfo struct {
	TotalMinutes int
	Formatted    string
}

type PriceInfo struct {
	AmountCents int64    // Changed from float64 to prevent precision errors
	Currency    Currency
	Formatted   string
}

type BaggageInfo struct {
	CabinKg     int
	CheckedKg   int
	CarryOnDesc string
	CheckedDesc stringS
}

func NewDurationInfo(totalMinutes int) DurationInfo {
	hours := totalMinutes / 60
	mins := totalMinutes % 60

	var formatted string
	if hours > 0 && mins > 0 {
		formatted = fmt.Sprintf("%dh %dm", hours, mins)
	} else if hours > 0 {
		formatted = fmt.Sprintf("%dh", hours)
	} else {
		formatted = fmt.Sprintf("%dm", mins)
	}

	return DurationInfo{
		TotalMinutes: totalMinutes,
		Formatted:    formatted,
	}
}

func (f *Flight) Validate() error {
	if !f.Arrival.DateTime.After(f.Departure.DateTime) {
		return fmt.Errorf("%w: arrival time must be after departure", ErrInvalidFlightTimes)
	}
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
	return nil
}
