package domain

import (
	"fmt"
	"regexp"
	"time"
)

// SearchCriteria defines the parameters for a flight search request.
type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departureDate"`
	Passengers    int    `json:"passengers"`
	Class         string `json:"class,omitempty"`
}

var airportCodeRegex = regexp.MustCompile(`^[A-Z]{3}$`)
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var validClasses = map[string]bool{
	"economy":  true,
	"business": true,
	"first":    true,
}

// Validate checks if the search criteria is valid.
func (s *SearchCriteria) Validate() error {
	if s.Origin == "" {
		return fmt.Errorf("%w: origin is required", ErrInvalidRequest)
	}
	if !airportCodeRegex.MatchString(s.Origin) {
		return fmt.Errorf("%w: origin must be a valid 3-letter IATA code, got %q", ErrInvalidRequest, s.Origin)
	}
	if s.Destination == "" {
		return fmt.Errorf("%w: destination is required", ErrInvalidRequest)
	}
	if !airportCodeRegex.MatchString(s.Destination) {
		return fmt.Errorf("%w: destination must be a valid 3-letter IATA code, got %q", ErrInvalidRequest, s.Destination)
	}
	if s.Origin == s.Destination {
		return fmt.Errorf("%w: origin and destination must be different", ErrInvalidRequest)
	}
	if s.DepartureDate == "" {
		return fmt.Errorf("%w: departureDate is required", ErrInvalidRequest)
	}
	if !dateRegex.MatchString(s.DepartureDate) {
		return fmt.Errorf("%w: departureDate must be in YYYY-MM-DD format, got %q", ErrInvalidRequest, s.DepartureDate)
	}
	if _, err := time.Parse("2006-01-02", s.DepartureDate); err != nil {
		return fmt.Errorf("%w: departureDate is not a valid date: %s", ErrInvalidRequest, s.DepartureDate)
	}
	if s.Passengers < 1 {
		return fmt.Errorf("%w: passengers must be at least 1", ErrInvalidRequest)
	}
	if s.Class != "" && !validClasses[s.Class] {
		return fmt.Errorf("%w: class must be one of: economy, business, first; got %q", ErrInvalidRequest, s.Class)
	}
	return nil
}

// SetDefaults applies default values to empty optional fields.
func (s *SearchCriteria) SetDefaults() {
	if s.Passengers == 0 {
		s.Passengers = 1
	}
	if s.Class == "" {
		s.Class = "economy"
	}
}