package flight

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	airportCodeRegex = regexp.MustCompile(`^[A-Z]{3}$`)
	dateFormatRegex  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// SearchRequest represents the HTTP request for flight search.
type SearchRequest struct {
	Origin        string         `json:"origin" binding:"required" example:"CGK" validate:"required,len=3" format:"IATA code"`                                   // Origin airport IATA code (3 letters)
	Destination   string         `json:"destination" binding:"required" example:"DPS" validate:"required,len=3" format:"IATA code"`                           // Destination airport IATA code (3 letters)
	DepartureDate string         `json:"departureDate" binding:"required" example:"2025-01-15" format:"date" validate:"required"`                             // Departure date in YYYY-MM-DD format
	Passengers    int            `json:"passengers" binding:"required,min=1,max=9" example:"2" minimum:"1" maximum:"9" validate:"required,min=1,max=9"`      // Number of passengers (1-9)
	Class         string         `json:"class,omitempty" example:"economy" enums:"economy,business,first"`                                                    // Cabin class preference (optional)
	Filters       *FilterDTO     `json:"filters,omitempty"`                                                                                                   // Optional filters for search results
	SortBy        string         `json:"sortBy,omitempty" example:"price" enums:"best,price,duration,departure"`                                             // Sort order for results (optional)
}

// FilterDTO represents filter options in HTTP requests.
type FilterDTO struct {
	MaxPrice           *float64          `json:"maxPrice,omitempty" example:"5000000" minimum:"0"`                // Maximum price in IDR (optional)
	MaxStops           *int              `json:"maxStops,omitempty" example:"1" minimum:"0"`                      // Maximum number of stops (optional)
	Airlines           []string          `json:"airlines,omitempty" example:"GA,JT"`                              // Filter by airline codes (optional)
	DepartureTimeRange *TimeRangeDTO     `json:"departureTimeRange,omitempty"`                                    // Filter by departure time range (optional)
	ArrivalTimeRange   *TimeRangeDTO     `json:"arrivalTimeRange,omitempty"`                                      // Filter by arrival time range (optional)
	DurationRange      *DurationRangeDTO `json:"durationRange,omitempty"`                                         // Filter by flight duration range (optional)
}

// TimeRangeDTO represents a time range filter in HTTP requests.
// Time format should be "HH:MM" (24-hour format).
type TimeRangeDTO struct {
	Start string `json:"start" binding:"required" example:"06:00" pattern:"^([01]\\d|2[0-3]):([0-5]\\d)$"` // Start time in HH:MM format (24-hour)
	End   string `json:"end" binding:"required" example:"22:00" pattern:"^([01]\\d|2[0-3]):([0-5]\\d)$"`   // End time in HH:MM format (24-hour)
}

// DurationRangeDTO represents a duration range filter in HTTP requests.
type DurationRangeDTO struct {
	MinMinutes *int `json:"minMinutes,omitempty" example:"60" minimum:"0"`  // Minimum flight duration in minutes (optional)
	MaxMinutes *int `json:"maxMinutes,omitempty" example:"300" minimum:"0"` // Maximum flight duration in minutes (optional)
}

// Validate validates the search request.
func (r *SearchRequest) Validate() error {
	// Validate origin
	if r.Origin == "" {
		return fmt.Errorf("origin is required")
	}
	origin := strings.ToUpper(strings.TrimSpace(r.Origin))
	if !airportCodeRegex.MatchString(origin) {
		return fmt.Errorf("origin must be a valid 3-letter IATA code, got %q", r.Origin)
	}

	// Validate destination
	if r.Destination == "" {
		return fmt.Errorf("destination is required")
	}
	destination := strings.ToUpper(strings.TrimSpace(r.Destination))
	if !airportCodeRegex.MatchString(destination) {
		return fmt.Errorf("destination must be a valid 3-letter IATA code, got %q", r.Destination)
	}

	// Validate origin != destination
	if origin == destination {
		return fmt.Errorf("origin and destination must be different")
	}

	// Validate departure date
	if r.DepartureDate == "" {
		return fmt.Errorf("departureDate is required")
	}
	if !dateFormatRegex.MatchString(r.DepartureDate) {
		return fmt.Errorf("departureDate must be in YYYY-MM-DD format, got %q", r.DepartureDate)
	}
	if _, err := time.Parse("2006-01-02", r.DepartureDate); err != nil {
		return fmt.Errorf("departureDate is not a valid date: %s", r.DepartureDate)
	}

	// Validate passengers
	if r.Passengers < 1 {
		return fmt.Errorf("passengers must be at least 1")
	}
	if r.Passengers > 9 {
		return fmt.Errorf("passengers must be at most 9")
	}

	// Validate class (optional)
	if r.Class != "" {
		class := strings.ToLower(r.Class)
		if class != "economy" && class != "business" && class != "first" {
			return fmt.Errorf("class must be one of: economy, business, first; got %q", r.Class)
		}
	}

	// Validate sortBy (optional)
	if r.SortBy != "" {
		sortBy := strings.ToLower(r.SortBy)
		if sortBy != "best" && sortBy != "price" && sortBy != "duration" && sortBy != "departure" {
			return fmt.Errorf("sortBy must be one of: best, price, duration, departure; got %q", r.SortBy)
		}
	}

	// Validate filters
	if r.Filters != nil {
		if err := r.Filters.Validate(); err != nil {
			return fmt.Errorf("invalid filters: %w", err)
		}
	}

	return nil
}

// Normalize normalizes the request fields (uppercase airport codes, lowercase class and sortBy).
func (r *SearchRequest) Normalize() {
	r.Origin = strings.ToUpper(strings.TrimSpace(r.Origin))
	r.Destination = strings.ToUpper(strings.TrimSpace(r.Destination))
	if r.Class != "" {
		r.Class = strings.ToLower(r.Class)
	}
	if r.SortBy != "" {
		r.SortBy = strings.ToLower(r.SortBy)
	}
}

// Validate validates filter options.
func (f *FilterDTO) Validate() error {
	if f == nil {
		return nil
	}

	// Validate maxPrice
	if f.MaxPrice != nil && *f.MaxPrice < 0 {
		return fmt.Errorf("maxPrice must be non-negative")
	}

	// Validate maxStops
	if f.MaxStops != nil && *f.MaxStops < 0 {
		return fmt.Errorf("maxStops must be non-negative")
	}

	// Validate time ranges
	if f.DepartureTimeRange != nil {
		if err := f.DepartureTimeRange.Validate(); err != nil {
			return fmt.Errorf("departureTimeRange: %w", err)
		}
	}
	if f.ArrivalTimeRange != nil {
		if err := f.ArrivalTimeRange.Validate(); err != nil {
			return fmt.Errorf("arrivalTimeRange: %w", err)
		}
	}

	// Validate duration range
	if f.DurationRange != nil {
		if err := f.DurationRange.Validate(); err != nil {
			return fmt.Errorf("durationRange: %w", err)
		}
	}

	return nil
}

// Validate validates time range.
func (t *TimeRangeDTO) Validate() error {
	if t == nil {
		return nil
	}

	// Validate time format HH:MM
	timeRegex := regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)

	if !timeRegex.MatchString(t.Start) {
		return fmt.Errorf("start time must be in HH:MM format (24-hour), got %q", t.Start)
	}
	if !timeRegex.MatchString(t.End) {
		return fmt.Errorf("end time must be in HH:MM format (24-hour), got %q", t.End)
	}

	return nil
}

// Validate validates duration range.
func (d *DurationRangeDTO) Validate() error {
	if d == nil {
		return nil
	}

	if d.MinMinutes != nil && *d.MinMinutes < 0 {
		return fmt.Errorf("minMinutes must be non-negative")
	}
	if d.MaxMinutes != nil && *d.MaxMinutes < 0 {
		return fmt.Errorf("maxMinutes must be non-negative")
	}
	if d.MinMinutes != nil && d.MaxMinutes != nil && *d.MinMinutes > *d.MaxMinutes {
		return fmt.Errorf("minMinutes must be less than or equal to maxMinutes")
	}

	return nil
}
