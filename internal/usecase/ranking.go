package usecase

import (
	"math"
	"sort"

	"github.com/herdiagusthio/flight-search-system/domain"
)

// Ranking weights (total = 1.0).
const (
	weightPrice    = 0.5 // Price impact: 50%
	weightDuration = 0.3 // Duration impact: 30%
	weightStops    = 0.2 // Stops impact: 20%
)

// CalculateRankingScores calculates ranking score for each flight.
// Score = 0.5×price + 0.3×duration + 0.2×stops (normalized to [0,1]).
// Lower score = better value.
func CalculateRankingScores(flights []domain.Flight) []domain.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Find min/max for normalization
	minPrice, maxPrice := findPriceRange(flights)
	minDuration, maxDuration := findDurationRange(flights)
	minStops, maxStops := findStopsRange(flights)

	// Calculate scores - create a copy to avoid mutating input
	result := make([]domain.Flight, len(flights))
	for i, f := range flights {
		result[i] = f

		normPrice := normalizeValue(f.Price.Amount, minPrice, maxPrice)
		normDuration := normalizeValue(float64(f.Duration.TotalMinutes), float64(minDuration), float64(maxDuration))
		normStops := normalizeValue(float64(f.Stops), float64(minStops), float64(maxStops))

		result[i].RankingScore = (weightPrice * normPrice) +
			(weightDuration * normDuration) +
			(weightStops * normStops)
	}

	return result
}

// normalizeValue normalizes value to [0,1] range.
// Returns 0 when min == max to avoid division by zero.
func normalizeValue(value, min, max float64) float64 {
	if max == min {
		return 0 // All values equal = all optimal
	}
	return (value - min) / (max - min)
}

// findPriceRange finds the minimum and maximum price across all flights.
func findPriceRange(flights []domain.Flight) (min, max float64) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxFloat64
	max = 0

	for _, f := range flights {
		if f.Price.Amount < min {
			min = f.Price.Amount
		}
		if f.Price.Amount > max {
			max = f.Price.Amount
		}
	}
	return min, max
}

// findDurationRange finds the minimum and maximum duration (in minutes) across all flights.
func findDurationRange(flights []domain.Flight) (min, max int) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxInt
	max = 0

	for _, f := range flights {
		if f.Duration.TotalMinutes < min {
			min = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > max {
			max = f.Duration.TotalMinutes
		}
	}
	return min, max
}

// findStopsRange finds the minimum and maximum number of stops across all flights.
func findStopsRange(flights []domain.Flight) (min, max int) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxInt
	max = 0

	for _, f := range flights {
		if f.Stops < min {
			min = f.Stops
		}
		if f.Stops > max {
			max = f.Stops
		}
	}
	return min, max
}

// SortFlights sorts flights by the specified option using stable sorting.
// Defaults to SortByBestValue if sortBy is invalid.
func SortFlights(flights []domain.Flight, sortBy domain.SortOption) []domain.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Copy to avoid mutating input
	result := make([]domain.Flight, len(flights))
	copy(result, flights)

	// Single flight doesn't need sorting
	if len(result) == 1 {
		return result
	}

	// Default to best value if sortBy is empty or invalid
	if sortBy == "" || !sortBy.IsValid() {
		sortBy = domain.SortByBestValue
	}

	switch sortBy {
	case domain.SortByBestValue:
		// Lower score = better value
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].RankingScore < result[j].RankingScore
		})
	case domain.SortByPrice:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Price.Amount < result[j].Price.Amount
		})
	case domain.SortByDuration:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Duration.TotalMinutes < result[j].Duration.TotalMinutes
		})
	case domain.SortByDeparture:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Departure.DateTime.Before(result[j].Departure.DateTime)
		})
	}

	return result
}
