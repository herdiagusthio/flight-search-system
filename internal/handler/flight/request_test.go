package flight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request SearchRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with all fields",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				Class:         "economy",
				SortBy:        "price",
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal fields",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    2,
			},
			wantErr: false,
		},
		{
			name: "missing origin",
			request: SearchRequest{
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "origin is required",
		},
		{
			name: "missing destination",
			request: SearchRequest{
				Origin:        "CGK",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "destination is required",
		},
		{
			name: "missing departure date",
			request: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
				Passengers:  1,
			},
			wantErr: true,
			errMsg:  "departureDate is required",
		},
		{
			name: "invalid origin - too short",
			request: SearchRequest{
				Origin:        "CG",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "origin must be a valid 3-letter IATA code",
		},
		{
			name: "invalid origin - too long",
			request: SearchRequest{
				Origin:        "CGKX",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "origin must be a valid 3-letter IATA code",
		},
		{
			name: "invalid origin - contains numbers",
			request: SearchRequest{
				Origin:        "CG1",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "origin must be a valid 3-letter IATA code",
		},
		{
			name: "invalid destination - too short",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DP",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "destination must be a valid 3-letter IATA code",
		},
		{
			name: "same origin and destination",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "CGK",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "origin and destination must be different",
		},
		{
			name: "invalid date format - wrong separator",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025/12/15",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "departureDate must be in YYYY-MM-DD format",
		},
		{
			name: "invalid date format - wrong order",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "15-12-2025",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "departureDate must be in YYYY-MM-DD format",
		},
		{
			name: "invalid date - not a real date",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-13-45",
				Passengers:    1,
			},
			wantErr: true,
			errMsg:  "departureDate is not a valid date",
		},
		{
			name: "passengers less than 1",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    0,
			},
			wantErr: true,
			errMsg:  "passengers must be at least 1",
		},
		{
			name: "passengers more than 9",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    10,
			},
			wantErr: true,
			errMsg:  "passengers must be at most 9",
		},
		{
			name: "invalid class",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				Class:         "premium",
			},
			wantErr: true,
			errMsg:  "class must be one of: economy, business, first",
		},
		{
			name: "valid class - business",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				Class:         "business",
			},
			wantErr: false,
		},
		{
			name: "invalid sortBy",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				SortBy:        "cheapest",
			},
			wantErr: true,
			errMsg:  "sortBy must be one of: best, price, duration, departure",
		},
		{
			name: "valid sortBy - duration",
			request: SearchRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				SortBy:        "duration",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchRequest_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		request  SearchRequest
		expected SearchRequest
	}{
		{
			name: "normalize lowercase airport codes",
			request: SearchRequest{
				Origin:      "cgk",
				Destination: "dps",
			},
			expected: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
			},
		},
		{
			name: "normalize uppercase class",
			request: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
				Class:       "ECONOMY",
			},
			expected: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
				Class:       "economy",
			},
		},
		{
			name: "normalize uppercase sortBy",
			request: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
				SortBy:      "PRICE",
			},
			expected: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
				SortBy:      "price",
			},
		},
		{
			name: "trim whitespace from airport codes",
			request: SearchRequest{
				Origin:      " CGK ",
				Destination: " DPS ",
			},
			expected: SearchRequest{
				Origin:      "CGK",
				Destination: "DPS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.Normalize()
			assert.Equal(t, tt.expected.Origin, tt.request.Origin)
			assert.Equal(t, tt.expected.Destination, tt.request.Destination)
			assert.Equal(t, tt.expected.Class, tt.request.Class)
			assert.Equal(t, tt.expected.SortBy, tt.request.SortBy)
		})
	}
}

func TestFilterDTO_Validate(t *testing.T) {
	maxPrice := 1000000.0
	negativePrice := -100.0
	maxStops := 1
	negativeStops := -1

	tests := []struct {
		name    string
		filter  *FilterDTO
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil filter is valid",
			filter:  nil,
			wantErr: false,
		},
		{
			name: "valid filter with all fields",
			filter: &FilterDTO{
				MaxPrice: &maxPrice,
				MaxStops: &maxStops,
				Airlines: []string{"GA", "JT"},
			},
			wantErr: false,
		},
		{
			name: "negative maxPrice",
			filter: &FilterDTO{
				MaxPrice: &negativePrice,
			},
			wantErr: true,
			errMsg:  "maxPrice must be non-negative",
		},
		{
			name: "negative maxStops",
			filter: &FilterDTO{
				MaxStops: &negativeStops,
			},
			wantErr: true,
			errMsg:  "maxStops must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeRangeDTO_Validate(t *testing.T) {
	tests := []struct {
		name      string
		timeRange *TimeRangeDTO
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "nil time range is valid",
			timeRange: nil,
			wantErr:   false,
		},
		{
			name: "valid time range",
			timeRange: &TimeRangeDTO{
				Start: "06:00",
				End:   "12:00",
			},
			wantErr: false,
		},
		{
			name: "valid time range - 24 hour format",
			timeRange: &TimeRangeDTO{
				Start: "18:30",
				End:   "23:59",
			},
			wantErr: false,
		},
		{
			name: "invalid start time format",
			timeRange: &TimeRangeDTO{
				Start: "6:00",
				End:   "12:00",
			},
			wantErr: true,
			errMsg:  "start time must be in HH:MM format",
		},
		{
			name: "invalid end time format",
			timeRange: &TimeRangeDTO{
				Start: "06:00",
				End:   "12:0",
			},
			wantErr: true,
			errMsg:  "end time must be in HH:MM format",
		},
		{
			name: "invalid hour - too high",
			timeRange: &TimeRangeDTO{
				Start: "25:00",
				End:   "12:00",
			},
			wantErr: true,
			errMsg:  "start time must be in HH:MM format",
		},
		{
			name: "invalid minute - too high",
			timeRange: &TimeRangeDTO{
				Start: "12:60",
				End:   "13:00",
			},
			wantErr: true,
			errMsg:  "start time must be in HH:MM format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.timeRange.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDurationRangeDTO_Validate(t *testing.T) {
	min90 := 90
	max240 := 240
	negative := -10

	tests := []struct {
		name          string
		durationRange *DurationRangeDTO
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "nil duration range is valid",
			durationRange: nil,
			wantErr:       false,
		},
		{
			name: "valid duration range - both min and max",
			durationRange: &DurationRangeDTO{
				MinMinutes: &min90,
				MaxMinutes: &max240,
			},
			wantErr: false,
		},
		{
			name: "valid duration range - only min",
			durationRange: &DurationRangeDTO{
				MinMinutes: &min90,
			},
			wantErr: false,
		},
		{
			name: "valid duration range - only max",
			durationRange: &DurationRangeDTO{
				MaxMinutes: &max240,
			},
			wantErr: false,
		},
		{
			name: "negative minMinutes",
			durationRange: &DurationRangeDTO{
				MinMinutes: &negative,
			},
			wantErr: true,
			errMsg:  "minMinutes must be non-negative",
		},
		{
			name: "negative maxMinutes",
			durationRange: &DurationRangeDTO{
				MaxMinutes: &negative,
			},
			wantErr: true,
			errMsg:  "maxMinutes must be non-negative",
		},
		{
			name: "minMinutes greater than maxMinutes",
			durationRange: &DurationRangeDTO{
				MinMinutes: &max240,
				MaxMinutes: &min90,
			},
			wantErr: true,
			errMsg:  "minMinutes must be less than or equal to maxMinutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.durationRange.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchRequest_WithFilters(t *testing.T) {
	maxPrice := 1000000.0
	maxStops := 1

	request := SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		Filters: &FilterDTO{
			MaxPrice: &maxPrice,
			MaxStops: &maxStops,
			DepartureTimeRange: &TimeRangeDTO{
				Start: "06:00",
				End:   "12:00",
			},
			DurationRange: &DurationRangeDTO{
				MinMinutes: &maxStops, // reuse as 1 minute for test
				MaxMinutes: &maxStops,
			},
		},
	}

	err := request.Validate()
	assert.NoError(t, err)
}
