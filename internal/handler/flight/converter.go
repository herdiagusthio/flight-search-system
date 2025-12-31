package flight

import (
	"fmt"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/internal/usecase"
)

// ToSearchCriteria converts SearchRequest to domain.SearchCriteria.
func ToSearchCriteria(req SearchRequest) domain.SearchCriteria {
	criteria := domain.SearchCriteria{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: req.DepartureDate,
		Passengers:    req.Passengers,
		Class:         req.Class,
	}

	// Apply defaults
	criteria.SetDefaults()

	return criteria
}

// ToSearchOptions converts DTO fields to usecase.SearchOptions.
func ToSearchOptions(req SearchRequest) usecase.SearchOptions {
	options := usecase.SearchOptions{
		Filters: ToFilterOptions(req.Filters),
		SortBy:  ToSortOption(req.SortBy),
	}

	return options
}

// ToFilterOptions converts FilterDTO to domain.FilterOptions.
func ToFilterOptions(dto *FilterDTO) *domain.FilterOptions {
	if dto == nil {
		return nil
	}

	filters := &domain.FilterOptions{
		MaxPrice: dto.MaxPrice,
		MaxStops: dto.MaxStops,
		Airlines: dto.Airlines,
	}

	// Convert time ranges
	if dto.DepartureTimeRange != nil {
		filters.DepartureTimeRange = ToTimeRange(dto.DepartureTimeRange)
	}
	if dto.ArrivalTimeRange != nil {
		filters.ArrivalTimeRange = ToTimeRange(dto.ArrivalTimeRange)
	}

	// Convert duration range
	if dto.DurationRange != nil {
		filters.DurationRange = ToDurationRange(dto.DurationRange)
	}

	return filters
}

// ToTimeRange converts TimeRangeDTO to domain.TimeRange.
// The DTO contains HH:MM strings which are converted to time.Time.
func ToTimeRange(dto *TimeRangeDTO) *domain.TimeRange {
	if dto == nil {
		return nil
	}

	// Parse start and end times (HH:MM format)
	// Use a reference date (doesn't matter which, we only care about time of day)
	referenceDate := "2006-01-02"
	startTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", referenceDate, dto.Start))
	if err != nil {
		// This should not happen if Validate() was called
		return nil
	}

	endTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", referenceDate, dto.End))
	if err != nil {
		// This should not happen if Validate() was called
		return nil
	}

	return &domain.TimeRange{
		Start: startTime,
		End:   endTime,
	}
}

// ToDurationRange converts DurationRangeDTO to domain.DurationRange.
func ToDurationRange(dto *DurationRangeDTO) *domain.DurationRange {
	if dto == nil {
		return nil
	}

	return &domain.DurationRange{
		MinMinutes: dto.MinMinutes,
		MaxMinutes: dto.MaxMinutes,
	}
}

// ToSortOption converts string sortBy to domain.SortOption.
func ToSortOption(sortBy string) domain.SortOption {
	switch sortBy {
	case "price":
		return domain.SortByPrice
	case "duration":
		return domain.SortByDuration
	case "departure":
		return domain.SortByDeparture
	case "best":
		return domain.SortByBestValue
	default:
		// Default to best value if empty or invalid
		return domain.SortByBestValue
	}
}
