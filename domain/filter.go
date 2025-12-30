package domain

import (
	"strings"
	"time"
)

// SortOption defines the available sorting options for flight results.
type SortOption string

const (
	SortByBestValue SortOption = "best"
	SortByPrice     SortOption = "price"
	SortByDuration  SortOption = "duration"
	SortByDeparture SortOption = "departure"
)

// IsValid checks if the sort option is a valid value.
func (s SortOption) IsValid() bool {
	switch s {
	case SortByBestValue, SortByPrice, SortByDuration, SortByDeparture:
		return true
	default:
		return false
	}
}

// FilterOptions defines optional filters to apply to flight results.
type FilterOptions struct {
	MaxPrice           *float64       `json:"maxPrice,omitempty"`
	MaxStops           *int           `json:"maxStops,omitempty"`
	Airlines           []string       `json:"airlines,omitempty"`
	DepartureTimeRange *TimeRange     `json:"departureTimeRange,omitempty"`
	ArrivalTimeRange   *TimeRange     `json:"arrivalTimeRange,omitempty"`
	DurationRange      *DurationRange `json:"durationRange,omitempty"`
}

// TimeRange represents a time window for filtering.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DurationRange represents a duration range filter for flights.
type DurationRange struct {
	MinMinutes *int `json:"minMinutes,omitempty"`
	MaxMinutes *int `json:"maxMinutes,omitempty"`
}

// IsValid checks if the duration range is valid.
func (dr *DurationRange) IsValid() bool {
	if dr == nil {
		return true
	}
	if dr.MinMinutes != nil && *dr.MinMinutes < 0 {
		return false
	}
	if dr.MaxMinutes != nil && *dr.MaxMinutes < 0 {
		return false
	}
	if dr.MinMinutes != nil && dr.MaxMinutes != nil && *dr.MinMinutes > *dr.MaxMinutes {
		return false
	}
	return true
}

// Contains checks if a given duration (in minutes) falls within the range.
func (dr *DurationRange) Contains(durationMinutes int) bool {
	if dr == nil {
		return true
	}
	if dr.MinMinutes != nil && durationMinutes < *dr.MinMinutes {
		return false
	}
	if dr.MaxMinutes != nil && durationMinutes > *dr.MaxMinutes {
		return false
	}
	return true
}

// Contains checks if a given time falls within the time range.
func (tr *TimeRange) Contains(t time.Time) bool {
	if tr == nil {
		return true
	}
	tMinutes := t.Hour()*60 + t.Minute()
	startMinutes := tr.Start.Hour()*60 + tr.Start.Minute()
	endMinutes := tr.End.Hour()*60 + tr.End.Minute()
	return tMinutes >= startMinutes && tMinutes <= endMinutes
}

// MatchesFlight checks if a flight matches all the filter criteria.
func (f *FilterOptions) MatchesFlight(flight Flight) bool {
	if f == nil {
		return true
	}
	if f.MaxPrice != nil && flight.Price.Amount > *f.MaxPrice {
		return false
	}
	if f.MaxStops != nil && flight.Stops > *f.MaxStops {
		return false
	}
	if len(f.Airlines) > 0 {
		found := false
		flightCode := strings.ToUpper(flight.Airline.Code)
		for _, code := range f.Airlines {
			if strings.ToUpper(code) == flightCode {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.DepartureTimeRange != nil && !f.DepartureTimeRange.Contains(flight.Departure.DateTime) {
		return false
	}
	if f.ArrivalTimeRange != nil && !f.ArrivalTimeRange.Contains(flight.Arrival.DateTime) {
		return false
	}
	if f.DurationRange != nil && !f.DurationRange.Contains(flight.Duration.TotalMinutes) {
		return false
	}
	return true
}

// ParseSortOption converts a string to a SortOption.
// Returns SortByBestValue if the string is empty or invalid.
func ParseSortOption(s string) SortOption {
	option := SortOption(s)
	if option.IsValid() {
		return option
	}
	return SortByBestValue
}
