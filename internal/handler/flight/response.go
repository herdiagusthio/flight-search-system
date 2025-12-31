package flight

import (
	"fmt"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
)

// SearchResponse is the main response structure for flight search API.
// Matches the expected_result.json format.
type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"` // Echo of the search criteria submitted
	Metadata       Metadata       `json:"metadata"`        // Search execution metadata and statistics
	Flights        []FlightDTO    `json:"flights"`         // List of available flights matching criteria
}

// SearchCriteria echoes back the search parameters.
type SearchCriteria struct {
	Origin        string `json:"origin" example:"CGK"`         // Origin airport IATA code
	Destination   string `json:"destination" example:"DPS"`   // Destination airport IATA code
	DepartureDate string `json:"departure_date" example:"2025-01-15"` // Departure date
	Passengers    int    `json:"passengers" example:"2"`      // Number of passengers
	CabinClass    string `json:"cabin_class" example:"economy"` // Cabin class
}

// Metadata contains search execution statistics and provider information.
type Metadata struct {
	TotalResults       int   `json:"total_results" example:"15"`       // Total number of flights found
	ProvidersQueried   int   `json:"providers_queried" example:"4"`    // Number of providers queried
	ProvidersSucceeded int   `json:"providers_succeeded" example:"4"`  // Number of providers that responded successfully
	ProvidersFailed    int   `json:"providers_failed" example:"0"`     // Number of providers that failed
	SearchTimeMs       int64 `json:"search_time_ms" example:"1234"`    // Total search execution time in milliseconds
	CacheHit           bool  `json:"cache_hit" example:"false"`        // Whether result was served from cache
}

// FlightDTO extends domain.Flight with additional formatted fields.
type FlightDTO struct {
	ID             string      `json:"id" example:"GA-12345"`                // Unique flight identifier
	Provider       string      `json:"provider" example:"Garuda Indonesia"`  // Provider/airline name
	Airline        AirlineDTO  `json:"airline"`                              // Airline information
	FlightNumber   string      `json:"flight_number" example:"GA-123"`       // Flight number
	Departure      LocationDTO `json:"departure"`                            // Departure information
	Arrival        LocationDTO `json:"arrival"`                              // Arrival information
	Duration       DurationDTO `json:"duration"`                             // Flight duration
	Stops          int         `json:"stops" example:"0"`                    // Number of stops (0 for direct)
	Price          PriceDTO    `json:"price"`                                // Price information
	AvailableSeats int         `json:"available_seats" example:"0"`          // Available seats (0 if not available)
	CabinClass     string      `json:"cabin_class" example:"economy"`        // Cabin class
	Aircraft       *string     `json:"aircraft" example:"Boeing 737"`        // Aircraft type (nullable)
	Amenities      []string    `json:"amenities" example:"WiFi,Meals"`       // Available amenities
	Baggage        BaggageDTO  `json:"baggage"`                              // Baggage allowance
}

// AirlineDTO contains airline information.
type AirlineDTO struct {
	Name string `json:"name" example:"Garuda Indonesia"` // Airline full name
	Code string `json:"code" example:"GA"`              // IATA airline code
}

// LocationDTO contains airport and time information.
type LocationDTO struct {
	Airport   string `json:"airport" example:"CGK"`                           // Airport IATA code
	City      string `json:"city" example:"Jakarta"`                          // City name
	Datetime  string `json:"datetime" example:"2025-01-15T08:00:00Z"`        // ISO 8601 datetime
	Timestamp int64  `json:"timestamp" example:"1736928000"`                  // Unix timestamp
}

// DurationDTO contains both numeric and formatted duration.
type DurationDTO struct {
	TotalMinutes int    `json:"total_minutes" example:"90"`  // Total duration in minutes
	Formatted    string `json:"formatted" example:"1h 30m"` // Human-readable formatted duration
}

// PriceDTO contains price information.
type PriceDTO struct {
	Amount   float64 `json:"amount" example:"1500000"`   // Price amount
	Currency string  `json:"currency" example:"IDR"`    // Currency code (ISO 4217)
}

// BaggageDTO contains baggage allowance information.
type BaggageDTO struct {
	CarryOn string `json:"carry_on" example:"7kg cabin"`    // Cabin baggage allowance
	Checked string `json:"checked" example:"20kg checked"`  // Checked baggage allowance
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

