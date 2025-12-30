package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSortOptionIsValid(t *testing.T) {
	tests := []struct {
		option   SortOption
		expected bool
	}{
		{SortByBestValue, true},
		{SortByPrice, true},
		{SortByDuration, true},
		{SortByDeparture, true},
		{SortOption("invalid"), false},
		{SortOption(""), false},
		{SortOption("BEST"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.option), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.option.IsValid())
		})
	}
}

func TestParseSortOption(t *testing.T) {
	tests := []struct {
		input    string
		expected SortOption
	}{
		{"best", SortByBestValue},
		{"price", SortByPrice},
		{"duration", SortByDuration},
		{"departure", SortByDeparture},
		{"invalid", SortByBestValue},
		{"", SortByBestValue},
		{"PRICE", SortByBestValue},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseSortOption(tt.input))
		})
	}
}

func TestDurationRangeIsValid(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name     string
		dr       *DurationRange
		expected bool
	}{
		{
			name:     "nil range is valid",
			dr:       nil,
			expected: true,
		},
		{
			name:     "empty range is valid",
			dr:       &DurationRange{},
			expected: true,
		},
		{
			name:     "only min is valid",
			dr:       &DurationRange{MinMinutes: intPtr(60)},
			expected: true,
		},
		{
			name:     "only max is valid",
			dr:       &DurationRange{MaxMinutes: intPtr(180)},
			expected: true,
		},
		{
			name:     "min less than max is valid",
			dr:       &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			expected: true,
		},
		{
			name:     "min equals max is valid",
			dr:       &DurationRange{MinMinutes: intPtr(120), MaxMinutes: intPtr(120)},
			expected: true,
		},
		{
			name:     "min greater than max is invalid",
			dr:       &DurationRange{MinMinutes: intPtr(180), MaxMinutes: intPtr(60)},
			expected: false,
		},
		{
			name:     "negative min is invalid",
			dr:       &DurationRange{MinMinutes: intPtr(-10)},
			expected: false,
		},
		{
			name:     "negative max is invalid",
			dr:       &DurationRange{MaxMinutes: intPtr(-10)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dr.IsValid())
		})
	}
}

func TestDurationRangeContains(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name     string
		dr       *DurationRange
		duration int
		expected bool
	}{
		{
			name:     "nil range contains any",
			dr:       nil,
			duration: 100,
			expected: true,
		},
		{
			name:     "within range",
			dr:       &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			duration: 120,
			expected: true,
		},
		{
			name:     "at min boundary",
			dr:       &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			duration: 60,
			expected: true,
		},
		{
			name:     "at max boundary",
			dr:       &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			duration: 180,
			expected: true,
		},
		{
			name:     "below min",
			dr:       &DurationRange{MinMinutes: intPtr(60)},
			duration: 30,
			expected: false,
		},
		{
			name:     "above max",
			dr:       &DurationRange{MaxMinutes: intPtr(180)},
			duration: 200,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dr.Contains(tt.duration))
		})
	}
}

func TestTimeRangeContains(t *testing.T) {
	makeTime := func(hour, min int) time.Time {
		return time.Date(2025, 1, 1, hour, min, 0, 0, time.UTC)
	}

	tests := []struct {
		name     string
		tr       *TimeRange
		t        time.Time
		expected bool
	}{
		{
			name:     "nil range contains any",
			tr:       nil,
			t:        makeTime(12, 0),
			expected: true,
		},
		{
			name:     "within range",
			tr:       &TimeRange{Start: makeTime(8, 0), End: makeTime(12, 0)},
			t:        makeTime(10, 0),
			expected: true,
		},
		{
			name:     "at start boundary",
			tr:       &TimeRange{Start: makeTime(8, 0), End: makeTime(12, 0)},
			t:        makeTime(8, 0),
			expected: true,
		},
		{
			name:     "at end boundary",
			tr:       &TimeRange{Start: makeTime(8, 0), End: makeTime(12, 0)},
			t:        makeTime(12, 0),
			expected: true,
		},
		{
			name:     "before range",
			tr:       &TimeRange{Start: makeTime(8, 0), End: makeTime(12, 0)},
			t:        makeTime(7, 0),
			expected: false,
		},
		{
			name:     "after range",
			tr:       &TimeRange{Start: makeTime(8, 0), End: makeTime(12, 0)},
			t:        makeTime(13, 0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tr.Contains(tt.t))
		})
	}
}

func TestFilterOptionsMatchesFlight(t *testing.T) {
	floatPtr := func(v float64) *float64 { return &v }
	intPtr := func(v int) *int { return &v }

	baseFlight := Flight{
		Airline: AirlineInfo{Code: "GA"},
		Price:   PriceInfo{Amount: 1500000},
		Stops:   1,
		Departure: FlightPoint{
			DateTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		Arrival: FlightPoint{
			DateTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		Duration: DurationInfo{TotalMinutes: 120},
	}

	tests := []struct {
		name     string
		filter   *FilterOptions
		flight   Flight
		expected bool
	}{
		{
			name:     "nil filter matches all",
			filter:   nil,
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "empty filter matches all",
			filter:   &FilterOptions{},
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "price within max",
			filter:   &FilterOptions{MaxPrice: floatPtr(2000000)},
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "price exceeds max",
			filter:   &FilterOptions{MaxPrice: floatPtr(1000000)},
			flight:   baseFlight,
			expected: false,
		},
		{
			name:     "stops within max",
			filter:   &FilterOptions{MaxStops: intPtr(2)},
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "stops exceeds max",
			filter:   &FilterOptions{MaxStops: intPtr(0)},
			flight:   baseFlight,
			expected: false,
		},
		{
			name:     "airline matches",
			filter:   &FilterOptions{Airlines: []string{"GA", "QZ"}},
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "airline matches case insensitive",
			filter:   &FilterOptions{Airlines: []string{"ga"}},
			flight:   baseFlight,
			expected: true,
		},
		{
			name:     "airline does not match",
			filter:   &FilterOptions{Airlines: []string{"QZ", "JT"}},
			flight:   baseFlight,
			expected: false,
		},
		{
			name: "departure within time range",
			filter: &FilterOptions{
				DepartureTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			flight:   baseFlight,
			expected: true,
		},
		{
			name: "departure outside time range",
			filter: &FilterOptions{
				DepartureTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 18, 0, 0, 0, time.UTC),
				},
			},
			flight:   baseFlight,
			expected: false,
		},
		{
			name: "arrival within time range",
			filter: &FilterOptions{
				ArrivalTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
				},
			},
			flight:   baseFlight,
			expected: true,
		},
		{
			name: "arrival outside time range",
			filter: &FilterOptions{
				ArrivalTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 18, 0, 0, 0, time.UTC),
				},
			},
			flight:   baseFlight,
			expected: false,
		},
		{
			name: "duration within range",
			filter: &FilterOptions{
				DurationRange: &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			},
			flight:   baseFlight,
			expected: true,
		},
		{
			name: "duration outside range",
			filter: &FilterOptions{
				DurationRange: &DurationRange{MaxMinutes: intPtr(60)},
			},
			flight:   baseFlight,
			expected: false,
		},
		{
			name: "multiple filters all match",
			filter: &FilterOptions{
				MaxPrice: floatPtr(2000000),
				MaxStops: intPtr(2),
				Airlines: []string{"GA"},
			},
			flight:   baseFlight,
			expected: true,
		},
		{
			name: "multiple filters one fails",
			filter: &FilterOptions{
				MaxPrice: floatPtr(2000000),
				MaxStops: intPtr(0),
				Airlines: []string{"GA"},
			},
			flight:   baseFlight,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.filter.MatchesFlight(tt.flight))
		})
	}
}
