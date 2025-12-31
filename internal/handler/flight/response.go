package flight

import (
	"fmt"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
)

// SearchResponse is the main response structure for flight search API.
// Matches the expected_result.json format.
type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       Metadata       `json:"metadata"`
	Flights        []FlightDTO    `json:"flights"`
}

// SearchCriteria echoes back the search parameters.
type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

// Metadata contains search execution statistics and provider information.
type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

// FlightDTO extends domain.Flight with additional formatted fields.
type FlightDTO struct {
	ID             string      `json:"id"`
	Provider       string      `json:"provider"`
	Airline        AirlineDTO  `json:"airline"`
	FlightNumber   string      `json:"flight_number"`
	Departure      LocationDTO `json:"departure"`
	Arrival        LocationDTO `json:"arrival"`
	Duration       DurationDTO `json:"duration"`
	Stops          int         `json:"stops"`
	Price          PriceDTO    `json:"price"`
	AvailableSeats int         `json:"available_seats"`
	CabinClass     string      `json:"cabin_class"`
	Aircraft       *string     `json:"aircraft"`
	Amenities      []string    `json:"amenities"`
	Baggage        BaggageDTO  `json:"baggage"`
}

// AirlineDTO contains airline information.
type AirlineDTO struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// LocationDTO contains airport and time information.
type LocationDTO struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

// DurationDTO contains both numeric and formatted duration.
type DurationDTO struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

// PriceDTO contains price information.
type PriceDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// BaggageDTO contains baggage allowance information.
type BaggageDTO struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

// NewSearchResponse creates a SearchResponse from domain objects.
func NewSearchResponse(
	criteria domain.SearchCriteria,
	flights []domain.Flight,
	metadata Metadata,
) SearchResponse {
	if flights == nil {
		flights = []domain.Flight{}
	}

	flightDTOs := make([]FlightDTO, len(flights))
	for i, flight := range flights {
		flightDTOs[i] = ToFlightDTO(flight)
	}

	return SearchResponse{
		SearchCriteria: SearchCriteria{
			Origin:        criteria.Origin,
			Destination:   criteria.Destination,
			DepartureDate: criteria.DepartureDate,
			Passengers:    criteria.Passengers,
			CabinClass:    criteria.Class,
		},
		Metadata: metadata,
		Flights:  flightDTOs,
	}
}

// ToFlightDTO converts domain.Flight to FlightDTO with formatted fields.
func ToFlightDTO(flight domain.Flight) FlightDTO {
	// Extract city from airport name if available, otherwise use airport code
	departureCity := flight.Departure.AirportName
	if departureCity == "" {
		departureCity = flight.Departure.AirportCode
	}
	arrivalCity := flight.Arrival.AirportName
	if arrivalCity == "" {
		arrivalCity = flight.Arrival.AirportCode
	}

	// Format baggage information
	carryOn := fmt.Sprintf("%dkg cabin", flight.Baggage.CabinKg)
	checked := fmt.Sprintf("%dkg checked", flight.Baggage.CheckedKg)
	if flight.Baggage.CabinKg == 0 {
		carryOn = "Not included"
	}
	if flight.Baggage.CheckedKg == 0 {
		checked = "Not included"
	}

	return FlightDTO{
		ID:       flight.ID,
		Provider: flight.Provider,
		Airline: AirlineDTO{
			Name: flight.Airline.Name,
			Code: flight.Airline.Code,
		},
		FlightNumber: flight.FlightNumber,
		Departure: LocationDTO{
			Airport:   flight.Departure.AirportCode,
			City:      departureCity,
			Datetime:  flight.Departure.DateTime.Format(time.RFC3339),
			Timestamp: flight.Departure.DateTime.Unix(),
		},
		Arrival: LocationDTO{
			Airport:   flight.Arrival.AirportCode,
			City:      arrivalCity,
			Datetime:  flight.Arrival.DateTime.Format(time.RFC3339),
			Timestamp: flight.Arrival.DateTime.Unix(),
		},
		Duration: DurationDTO{
			TotalMinutes: flight.Duration.TotalMinutes,
			Formatted:    flight.Duration.Formatted,
		},
		Stops: flight.Stops,
		Price: PriceDTO{
			Amount:   flight.Price.Amount,
			Currency: flight.Price.Currency,
		},
		AvailableSeats: 0, // Not available in domain.Flight, set in handler
		CabinClass:     flight.Class,
		Aircraft:       nil, // Not available in domain.Flight, set in handler
		Amenities:      []string{}, // Not available in domain.Flight, set in handler
		Baggage: BaggageDTO{
			CarryOn: carryOn,
			Checked: checked,
		},
	}
}

// formatDuration converts minutes to "Xh Ym" format.
func formatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60

	if hours == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

