package usecase

import (
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilters(t *testing.T) {
	baseTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	flights := []domain.Flight{
		{
			ID:    "f1",
			Price: domain.PriceInfo{Amount: 500000, Currency: "IDR"},
			Stops: 0,
			Airline: domain.AirlineInfo{Code: "GA", Name: "Garuda"},
			Departure: domain.FlightPoint{DateTime: baseTime.Add(2 * time.Hour)},
			Arrival:   domain.FlightPoint{DateTime: baseTime.Add(4 * time.Hour)},
			Duration:  domain.DurationInfo{TotalMinutes: 120},
		},
		{
			ID:    "f2",
			Price: domain.PriceInfo{Amount: 800000, Currency: "IDR"},
			Stops: 1,
			Airline: domain.AirlineInfo{Code: "JT", Name: "Lion Air"},
			Departure: domain.FlightPoint{DateTime: baseTime.Add(6 * time.Hour)},
			Arrival:   domain.FlightPoint{DateTime: baseTime.Add(9 * time.Hour)},
			Duration:  domain.DurationInfo{TotalMinutes: 180},
		},
		{
			ID:    "f3",
			Price: domain.PriceInfo{Amount: 1200000, Currency: "IDR"},
			Stops: 2,
			Airline: domain.AirlineInfo{Code: "QZ", Name: "AirAsia"},
			Departure: domain.FlightPoint{DateTime: baseTime.Add(12 * time.Hour)},
			Arrival:   domain.FlightPoint{DateTime: baseTime.Add(16 * time.Hour)},
			Duration:  domain.DurationInfo{TotalMinutes: 240},
		},
	}

	tests := []struct {
		name     string
		filters  *domain.FilterOptions
		expected int
		checkIDs []string
	}{
		{
			name:     "no filters returns all",
			filters:  nil,
			expected: 3,
			checkIDs: []string{"f1", "f2", "f3"},
		},
		{
			name: "filter by max price",
			filters: &domain.FilterOptions{
				MaxPrice: ptrFloat64(900000),
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by max stops",
			filters: &domain.FilterOptions{
				MaxStops: ptrInt(1),
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by airlines",
			filters: &domain.FilterOptions{
				Airlines: []string{"GA", "JT"},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by airlines case insensitive",
			filters: &domain.FilterOptions{
				Airlines: []string{"ga", "jt"},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by departure time range",
			filters: &domain.FilterOptions{
				DepartureTimeRange: &domain.TimeRange{
					Start: baseTime,
					End:   baseTime.Add(8 * time.Hour),
				},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by arrival time range",
			filters: &domain.FilterOptions{
				ArrivalTimeRange: &domain.TimeRange{
					Start: baseTime,
					End:   baseTime.Add(10 * time.Hour),
				},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filter by duration range",
			filters: &domain.FilterOptions{
				DurationRange: &domain.DurationRange{
					MinMinutes: ptrInt(100),
					MaxMinutes: ptrInt(200),
				},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "combined filters",
			filters: &domain.FilterOptions{
				MaxPrice: ptrFloat64(1000000),
				MaxStops: ptrInt(1),
				Airlines: []string{"GA", "JT"},
			},
			expected: 2,
			checkIDs: []string{"f1", "f2"},
		},
		{
			name: "filters returning empty",
			filters: &domain.FilterOptions{
				MaxPrice: ptrFloat64(100000),
			},
			expected: 0,
			checkIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(flights, tt.filters)
			assert.Equal(t, tt.expected, len(result))

			ids := make([]string, len(result))
			for i, f := range result {
				ids[i] = f.ID
			}
			assert.ElementsMatch(t, tt.checkIDs, ids)
		})
	}
}

func TestFilterByMaxPrice(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Price: domain.PriceInfo{Amount: 500000}},
		{ID: "f2", Price: domain.PriceInfo{Amount: 800000}},
		{ID: "f3", Price: domain.PriceInfo{Amount: 1200000}},
	}

	tests := []struct {
		name     string
		maxPrice *float64
		expected int
	}{
		{"nil returns all", nil, 3},
		{"filter 900000", ptrFloat64(900000), 2},
		{"filter 500000", ptrFloat64(500000), 1},
		{"filter 0", ptrFloat64(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByMaxPrice(flights, tt.maxPrice)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestFilterByMaxStops(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Stops: 0},
		{ID: "f2", Stops: 1},
		{ID: "f3", Stops: 2},
	}

	tests := []struct {
		name     string
		maxStops *int
		expected int
	}{
		{"nil returns all", nil, 3},
		{"direct only", ptrInt(0), 1},
		{"max 1 stop", ptrInt(1), 2},
		{"max 2 stops", ptrInt(2), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByMaxStops(flights, tt.maxStops)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestFilterByAirlines(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Airline: domain.AirlineInfo{Code: "GA"}},
		{ID: "f2", Airline: domain.AirlineInfo{Code: "JT"}},
		{ID: "f3", Airline: domain.AirlineInfo{Code: "QZ"}},
	}

	tests := []struct {
		name     string
		airlines []string
		expected int
	}{
		{"nil returns all", nil, 3},
		{"empty returns all", []string{}, 3},
		{"single airline", []string{"GA"}, 1},
		{"two airlines", []string{"GA", "JT"}, 2},
		{"case insensitive", []string{"ga", "jt"}, 2},
		{"mixed case", []string{"Ga", "jT"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByAirlines(flights, tt.airlines)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestFilterByDepartureTime(t *testing.T) {
	// TimeRange.Contains only compares time of day (hour:minute), not date
	baseTime := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)

	flights := []domain.Flight{
		{ID: "f1", Departure: domain.FlightPoint{DateTime: baseTime.Add(6 * time.Hour)}},  // 06:00
		{ID: "f2", Departure: domain.FlightPoint{DateTime: baseTime.Add(12 * time.Hour)}}, // 12:00
		{ID: "f3", Departure: domain.FlightPoint{DateTime: baseTime.Add(18 * time.Hour)}}, // 18:00
	}

	tests := []struct {
		name      string
		timeRange *domain.TimeRange
		expected  int
	}{
		{"nil returns all", nil, 3},
		{
			"morning range",
			&domain.TimeRange{
				Start: time.Date(0, 1, 1, 5, 0, 0, 0, time.UTC),  // 05:00
				End:   time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC), // 13:00
			},
			2, // f1 (06:00) and f2 (12:00)
		},
		{
			"afternoon range",
			&domain.TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC), // 15:00
				End:   time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC), // 23:00
			},
			1, // f3 (18:00)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByDepartureTime(flights, tt.timeRange)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestFilterByArrivalTime(t *testing.T) {
	baseTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	flights := []domain.Flight{
		{ID: "f1", Arrival: domain.FlightPoint{DateTime: baseTime.Add(4 * time.Hour)}},
		{ID: "f2", Arrival: domain.FlightPoint{DateTime: baseTime.Add(9 * time.Hour)}},
		{ID: "f3", Arrival: domain.FlightPoint{DateTime: baseTime.Add(16 * time.Hour)}},
	}

	tests := []struct {
		name      string
		timeRange *domain.TimeRange
		expected  int
	}{
		{"nil returns all", nil, 3},
		{
			"early range",
			&domain.TimeRange{Start: baseTime, End: baseTime.Add(10 * time.Hour)},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByArrivalTime(flights, tt.timeRange)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestFilterByDuration(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Duration: domain.DurationInfo{TotalMinutes: 120}},
		{ID: "f2", Duration: domain.DurationInfo{TotalMinutes: 180}},
		{ID: "f3", Duration: domain.DurationInfo{TotalMinutes: 240}},
	}

	tests := []struct {
		name     string
		duration *domain.DurationRange
		expected int
	}{
		{"nil returns all", nil, 3},
		{
			"short flights",
			&domain.DurationRange{MinMinutes: ptrInt(100), MaxMinutes: ptrInt(150)},
			1,
		},
		{
			"medium and long",
			&domain.DurationRange{MinMinutes: ptrInt(150), MaxMinutes: ptrInt(300)},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByDuration(flights, tt.duration)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestBuildAirlineSet(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		checkKey string
		exists   bool
	}{
		{"uppercase stored as uppercase", []string{"GA"}, "GA", true},
		{"lowercase converted to uppercase", []string{"ga"}, "GA", true},
		{"mixed case converted", []string{"Ga"}, "GA", true},
		{"multiple airlines", []string{"GA", "jt", "Qz"}, "JT", true},
		{"not in set", []string{"GA"}, "JT", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := buildAirlineSet(tt.input)
			_, exists := set[tt.checkKey]
			assert.Equal(t, tt.exists, exists)
		})
	}
}

func TestIsAirlineInSet(t *testing.T) {
	set := map[string]struct{}{
		"GA": {},
		"JT": {},
		"QZ": {},
	}

	tests := []struct {
		name   string
		code   string
		exists bool
	}{
		{"exact match uppercase", "GA", true},
		{"lowercase match", "ga", true},
		{"mixed case match", "Ga", true},
		{"not in set", "ID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAirlineInSet(tt.code, set)
			assert.Equal(t, tt.exists, result)
		})
	}
}

// Helper functions
func ptrFloat64(v float64) *float64 { return &v }
func ptrInt(v int) *int              { return &v }
