package domain

// SearchResponse represents the aggregated response from a flight search.
type SearchResponse struct {
	SearchCriteria SearchCriteriaResponse `json:"search_criteria"`
	Metadata       SearchMetadata         `json:"metadata"`
	Flights        []Flight               `json:"flights"`
}

// SearchCriteriaResponse represents the search criteria in the response.
type SearchCriteriaResponse struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

// SearchMetadata contains metadata about the search execution.
type SearchMetadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

// NewSearchResponse creates a new SearchResponse.
func NewSearchResponse(criteria *SearchCriteria, flights []Flight, metadata SearchMetadata) SearchResponse {
	if flights == nil {
		flights = []Flight{}
	}
	metadata.TotalResults = len(flights)

	criteriaResp := SearchCriteriaResponse{
		Origin:        criteria.Origin,
		Destination:   criteria.Destination,
		DepartureDate: criteria.DepartureDate,
		Passengers:    criteria.Passengers,
		CabinClass:    criteria.Class,
	}

	return SearchResponse{
		SearchCriteria: criteriaResp,
		Metadata:       metadata,
		Flights:        flights,
	}
}

// ProviderResult represents the result from a single provider query.
type ProviderResult struct {
	Provider   string
	Flights    []Flight
	Error      error
	DurationMs int64
}

// IsSuccess returns true if the provider query succeeded.
func (pr *ProviderResult) IsSuccess() bool {
	return pr.Error == nil
}