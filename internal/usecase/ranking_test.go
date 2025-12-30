package usecase

import (
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRankingScores(t *testing.T) {
	tests := []struct {
		name          string
		flights       []domain.Flight
		checkScores   bool
		expectedCount int
	}{
		{
			name:          "empty flights returns empty",
			flights:       []domain.Flight{},
			expectedCount: 0,
		},
		{
			name: "single flight gets score 0",
			flights: []domain.Flight{
				{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
			},
			checkScores:   true,
			expectedCount: 1,
		},
		{
			name: "multiple flights with different values",
			flights: []domain.Flight{
				{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
				{ID: "f2", Price: domain.PriceInfo{Amount: 800000}, Duration: domain.DurationInfo{TotalMinutes: 180}, Stops: 1},
				{ID: "f3", Price: domain.PriceInfo{Amount: 1200000}, Duration: domain.DurationInfo{TotalMinutes: 240}, Stops: 2},
			},
			checkScores:   true,
			expectedCount: 3,
		},
		{
			name: "all equal values get score 0",
			flights: []domain.Flight{
				{ID: "f1", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
				{ID: "f2", Price: domain.PriceInfo{Amount: 500000}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0},
			},
			checkScores:   true,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRankingScores(tt.flights)
			assert.Equal(t, tt.expectedCount, len(result))

			if tt.checkScores && len(result) == 1 {
				assert.Equal(t, 0.0, result[0].RankingScore)
			}

			if tt.checkScores && tt.name == "multiple flights with different values" {
				// Best flight (lowest price, duration, stops) should have lowest score
				assert.True(t, result[0].RankingScore < result[1].RankingScore)
				assert.True(t, result[1].RankingScore < result[2].RankingScore)
			}

			if tt.checkScores && tt.name == "all equal values get score 0" {
				for _, f := range result {
					assert.Equal(t, 0.0, f.RankingScore)
				}
			}

			// Verify original slice not mutated (check that RankingScore wasn't modified in original)
			if len(tt.flights) > 0 {
				assert.Equal(t, 0.0, tt.flights[0].RankingScore, "original flight should not have ranking score modified")
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"min value returns 0", 100, 100, 200, 0.0},
		{"max value returns 1", 200, 100, 200, 1.0},
		{"mid value returns 0.5", 150, 100, 200, 0.5},
		{"equal min and max returns 0", 100, 100, 100, 0.0},
		{"quarter value", 125, 100, 200, 0.25},
		{"three quarters value", 175, 100, 200, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeValue(tt.value, tt.min, tt.max)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestFindPriceRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty flights returns 0,0",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "single flight",
			flights: []domain.Flight{
				{Price: domain.PriceInfo{Amount: 500000}},
			},
			expectedMin: 500000,
			expectedMax: 500000,
		},
		{
			name: "multiple flights",
			flights: []domain.Flight{
				{Price: domain.PriceInfo{Amount: 500000}},
				{Price: domain.PriceInfo{Amount: 1200000}},
				{Price: domain.PriceInfo{Amount: 800000}},
			},
			expectedMin: 500000,
			expectedMax: 1200000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findPriceRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

func TestFindDurationRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin int
		expectedMax int
	}{
		{
			name:        "empty flights returns 0,0",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "single flight",
			flights: []domain.Flight{
				{Duration: domain.DurationInfo{TotalMinutes: 120}},
			},
			expectedMin: 120,
			expectedMax: 120,
		},
		{
			name: "multiple flights",
			flights: []domain.Flight{
				{Duration: domain.DurationInfo{TotalMinutes: 120}},
				{Duration: domain.DurationInfo{TotalMinutes: 240}},
				{Duration: domain.DurationInfo{TotalMinutes: 180}},
			},
			expectedMin: 120,
			expectedMax: 240,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findDurationRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

func TestFindStopsRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin int
		expectedMax int
	}{
		{
			name:        "empty flights returns 0,0",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "single flight",
			flights: []domain.Flight{
				{Stops: 1},
			},
			expectedMin: 1,
			expectedMax: 1,
		},
		{
			name: "multiple flights",
			flights: []domain.Flight{
				{Stops: 0},
				{Stops: 2},
				{Stops: 1},
			},
			expectedMin: 0,
			expectedMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findStopsRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

func TestSortFlights(t *testing.T) {
	baseTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	flights := []domain.Flight{
		{
			ID:           "f1",
			Price:        domain.PriceInfo{Amount: 800000},
			Duration:     domain.DurationInfo{TotalMinutes: 180},
			Departure:    domain.FlightPoint{DateTime: baseTime.Add(6 * time.Hour)},
			RankingScore: 0.5,
		},
		{
			ID:           "f2",
			Price:        domain.PriceInfo{Amount: 500000},
			Duration:     domain.DurationInfo{TotalMinutes: 120},
			Departure:    domain.FlightPoint{DateTime: baseTime.Add(2 * time.Hour)},
			RankingScore: 0.2,
		},
		{
			ID:           "f3",
			Price:        domain.PriceInfo{Amount: 1200000},
			Duration:     domain.DurationInfo{TotalMinutes: 240},
			Departure:    domain.FlightPoint{DateTime: baseTime.Add(12 * time.Hour)},
			RankingScore: 0.8,
		},
	}

	tests := []struct {
		name         string
		sortBy       domain.SortOption
		expectedIDs  []string
		emptyFlights bool
	}{
		{
			name:        "empty flights returns empty",
			sortBy:      domain.SortByBestValue,
			expectedIDs: []string{},
			emptyFlights: true,
		},
		{
			name:        "sort by best value",
			sortBy:      domain.SortByBestValue,
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "sort by price",
			sortBy:      domain.SortByPrice,
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "sort by duration",
			sortBy:      domain.SortByDuration,
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "sort by departure",
			sortBy:      domain.SortByDeparture,
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "invalid sort defaults to best value",
			sortBy:      domain.SortOption("invalid"),
			expectedIDs: []string{"f2", "f1", "f3"},
		},
		{
			name:        "empty sort defaults to best value",
			sortBy:      "",
			expectedIDs: []string{"f2", "f1", "f3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFlights := flights
			if tt.emptyFlights {
				testFlights = []domain.Flight{}
			}

			result := SortFlights(testFlights, tt.sortBy)
			assert.Equal(t, len(tt.expectedIDs), len(result))

			ids := make([]string, len(result))
			for i, f := range result {
				ids[i] = f.ID
			}
			assert.Equal(t, tt.expectedIDs, ids)

			// Verify original slice not mutated
			if len(testFlights) > 0 && len(result) > 0 {
				assert.NotEqual(t, &testFlights[0], &result[0])
			}
		})
	}
}

func TestSortFlightsStability(t *testing.T) {
	flights := []domain.Flight{
		{ID: "f1", Price: domain.PriceInfo{Amount: 500000}},
		{ID: "f2", Price: domain.PriceInfo{Amount: 500000}},
		{ID: "f3", Price: domain.PriceInfo{Amount: 500000}},
	}

	result := SortFlights(flights, domain.SortByPrice)

	// With stable sort and equal values, order should be preserved
	assert.Equal(t, "f1", result[0].ID)
	assert.Equal(t, "f2", result[1].ID)
	assert.Equal(t, "f3", result[2].ID)
}
