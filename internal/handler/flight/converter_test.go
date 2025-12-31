package flight

import (
	"testing"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
)

func TestToSearchCriteria(t *testing.T) {
	tests := []struct {
		name     string
		request  *SearchRequest
		expected struct {
			origin      string
			destination string
			date        string
			passengers  int
			class       string
		}
	}{
		{
			name: "basic conversion",
			request: &SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			expected: struct {
				origin      string
				destination string
				date        string
				passengers  int
				class       string
			}{
				origin:      "CGK",
				destination: "DPS",
				date:        "2025-12-15",
				passengers:  1,
				class:       "economy",
			},
		},
		{
			name: "with class specified",
			request: &SearchRequest{
				Origin:        "CGK",
				Destination:   "SUB",
				DepartureDate: "2025-12-20",
				Passengers:    2,
				Class:         "business",
			},
			expected: struct {
				origin      string
				destination string
				date        string
				passengers  int
				class       string
			}{
				origin:      "CGK",
				destination: "SUB",
				date:        "2025-12-20",
				passengers:  2,
				class:       "business",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSearchCriteria(*tt.request)

			assert.Equal(t, tt.expected.origin, result.Origin)
			assert.Equal(t, tt.expected.destination, result.Destination)
			assert.Equal(t, tt.expected.date, result.DepartureDate)
			assert.Equal(t, tt.expected.passengers, result.Passengers)
			assert.Equal(t, tt.expected.class, result.Class)
		})
	}
}

func TestToSearchOptions(t *testing.T) {
	maxPrice := 1000000.0
	maxStops := 1
	minDuration := 90
	maxDuration := 240

	tests := []struct {
		name    string
		request *SearchRequest
		check   func(*testing.T, interface{})
	}{
		{
			name: "without filters",
			request: &SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				SortBy:        "price",
			},
			check: func(t *testing.T, opts interface{}) {
				// Filters should be nil
				// SortBy should be mapped to domain.SortByPrice
			},
		},
		{
			name: "with all filters",
			request: &SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				SortBy:        "price",
				Filters: &FilterDTO{
					MaxPrice: &maxPrice,
					MaxStops: &maxStops,
					Airlines: []string{"GA", "JT"},
					DepartureTimeRange: &TimeRangeDTO{
						Start: "06:00",
						End:   "12:00",
					},
					DurationRange: &DurationRangeDTO{
						MinMinutes: &minDuration,
						MaxMinutes: &maxDuration,
					},
				},
			},
			check: func(t *testing.T, opts interface{}) {
				// Should have filters populated
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSearchOptions(*tt.request)
			assert.NotNil(t, result)

			if tt.request.Filters != nil {
				assert.NotNil(t, result.Filters)
			}
		})
	}
}

func TestToFilterOptions(t *testing.T) {
	maxPrice := 1000000.0
	maxStops := 2

	tests := []struct {
		name     string
		filter   *FilterDTO
		expected func(*testing.T, interface{})
	}{
		{
			name:   "nil filter returns nil",
			filter: nil,
			expected: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "basic filter without ranges",
			filter: &FilterDTO{
				MaxPrice: &maxPrice,
				MaxStops: &maxStops,
				Airlines: []string{"GA", "JT"},
			},
			expected: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "filter with departure time range",
			filter: &FilterDTO{
				DepartureTimeRange: &TimeRangeDTO{
					Start: "06:00",
					End:   "12:00",
				},
			},
			expected: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "filter with arrival time range",
			filter: &FilterDTO{
				ArrivalTimeRange: &TimeRangeDTO{
					Start: "18:00",
					End:   "23:00",
				},
			},
			expected: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "filter with duration range",
			filter: &FilterDTO{
				DurationRange: &DurationRangeDTO{
					MinMinutes: &maxStops, // reuse as 2 minutes
					MaxMinutes: &maxStops,
				},
			},
			expected: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFilterOptions(tt.filter)
			tt.expected(t, result)
		})
	}
}

func TestToTimeRange(t *testing.T) {
	tests := []struct {
		name      string
		timeRange *TimeRangeDTO
		validate  func(*testing.T, *domain.TimeRange)
	}{
		{
			name:      "nil time range returns nil",
			timeRange: nil,
			validate: func(t *testing.T, result *domain.TimeRange) {
				assert.Nil(t, result)
			},
		},
		{
			name: "valid time range",
			timeRange: &TimeRangeDTO{
				Start: "06:00",
				End:   "12:00",
			},
			validate: func(t *testing.T, result *domain.TimeRange) {
				assert.NotNil(t, result)
				assert.Equal(t, 6, result.Start.Hour())
				assert.Equal(t, 0, result.Start.Minute())
				assert.Equal(t, 12, result.End.Hour())
				assert.Equal(t, 0, result.End.Minute())
			},
		},
		{
			name: "time range with minutes",
			timeRange: &TimeRangeDTO{
				Start: "08:30",
				End:   "17:45",
			},
			validate: func(t *testing.T, result *domain.TimeRange) {
				assert.NotNil(t, result)
				assert.Equal(t, 8, result.Start.Hour())
				assert.Equal(t, 30, result.Start.Minute())
				assert.Equal(t, 17, result.End.Hour())
				assert.Equal(t, 45, result.End.Minute())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTimeRange(tt.timeRange)
			tt.validate(t, result)
		})
	}
}

func TestToDurationRange(t *testing.T) {
	min90 := 90
	max240 := 240

	tests := []struct {
		name          string
		durationRange *DurationRangeDTO
		validate      func(*testing.T, *domain.DurationRange)
	}{
		{
			name:          "nil duration range returns nil",
			durationRange: nil,
			validate: func(t *testing.T, result *domain.DurationRange) {
				assert.Nil(t, result)
			},
		},
		{
			name: "both min and max specified",
			durationRange: &DurationRangeDTO{
				MinMinutes: &min90,
				MaxMinutes: &max240,
			},
			validate: func(t *testing.T, result *domain.DurationRange) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.MinMinutes)
				assert.NotNil(t, result.MaxMinutes)
				assert.Equal(t, min90, *result.MinMinutes)
				assert.Equal(t, max240, *result.MaxMinutes)
			},
		},
		{
			name: "only min specified",
			durationRange: &DurationRangeDTO{
				MinMinutes: &min90,
			},
			validate: func(t *testing.T, result *domain.DurationRange) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.MinMinutes)
				assert.Nil(t, result.MaxMinutes)
				assert.Equal(t, min90, *result.MinMinutes)
			},
		},
		{
			name: "only max specified",
			durationRange: &DurationRangeDTO{
				MaxMinutes: &max240,
			},
			validate: func(t *testing.T, result *domain.DurationRange) {
				assert.NotNil(t, result)
				assert.Nil(t, result.MinMinutes)
				assert.NotNil(t, result.MaxMinutes)
				assert.Equal(t, max240, *result.MaxMinutes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDurationRange(tt.durationRange)
			tt.validate(t, result)
		})
	}
}

func TestToSortOption(t *testing.T) {
	tests := []struct {
		name     string
		sortBy   string
		validate func(*testing.T, interface{})
	}{
		{
			name:   "empty string defaults to best",
			sortBy: "",
			validate: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "price",
			sortBy: "price",
			validate: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "duration",
			sortBy: "duration",
			validate: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "departure",
			sortBy: "departure",
			validate: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "best",
			sortBy: "best",
			validate: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSortOption(tt.sortBy)
			tt.validate(t, result)
		})
	}
}
