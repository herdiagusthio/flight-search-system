package usecase

import (
	"github.com/herdiagusthio/flight-search-system/domain"
)

// SearchOptions contains optional search parameters.
type SearchOptions struct {
	Filters *domain.FilterOptions
	SortBy  domain.SortOption
}

// DefaultSearchOptions returns SearchOptions with sensible defaults.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Filters: nil,
		SortBy:  domain.SortByBestValue,
	}
}
