package garuda

import (
	"fmt"
	"strings"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/rs/zerolog/log"
)

func normalize(garudaFlights []GarudaFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(garudaFlights))
	skippedCount := 0

	for _, f := range garudaFlights {
		normalized, err := normalizeFlight(f)
		if err != nil {
			skippedCount++
			continue
		}

		if err := normalized.Validate(); err != nil {
			log.Warn().
				Str("provider", ProviderName).
				Str("flight_number", normalized.FlightNumber).
				Err(err).
				Msg("Flight validation failed")
			skippedCount++
			continue
		}

		result = append(result, normalized)
	}

	if skippedCount > 0 {
		log.Info().
			Str("provider", ProviderName).
			Int("skipped", skippedCount).
			Int("total", len(garudaFlights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

// normalizeFlight converts a single Garuda flight to a domain Flight entity.
func normalizeFlight(f GarudaFlight) (domain.Flight, error) {
	// Parse departure time
	departureTime, err := parseDateTime(f.Departure.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := parseDateTime(f.Arrival.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	// Calculate stops from segments if available, otherwise use stops field
	stops := f.Stops
	if len(f.Segments) > 1 {
		stops = len(f.Segments) - 1
	}

	return domain.Flight{
		ID:           f.FlightID,
		FlightNumber: f.FlightID, // Use flight_id as flight number since it contains the flight identifier
		Airline: domain.AirlineInfo{
			Code: f.AirlineCode,
			Name: f.Airline,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.Departure.Airport,
			AirportName: formatAirportName(f.Departure.Airport, f.Departure.City),
			Terminal:    f.Departure.Terminal,
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.Arrival.Airport,
			AirportName: formatAirportName(f.Arrival.Airport, f.Arrival.City),
			Terminal:    f.Arrival.Terminal,
			DateTime:    arrivalTime,
		},
		Duration: domain.NewDurationInfo(f.DurationMinutes),
		Price: domain.PriceInfo{
			Amount:    f.Price.Amount,
			Currency:  f.Price.Currency,
			Formatted: util.FormatIDR(f.Price.Amount),
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   f.Baggage.CarryOn * DefaultCabinBaggageKg,
			CheckedKg: f.Baggage.Checked * DefaultCheckedBaggageKg,
		},
		Class:    normalizeClass(f.FareClass),
		Stops:    stops,
		Provider: ProviderName,
	}, nil
}

// parseDateTime parses an ISO 8601 datetime string to time.Time.
// Supports formats: "2006-01-02T15:04:05Z07:00" and "2006-01-02T15:04:05"
func parseDateTime(dateTime string) (time.Time, error) {
	// Try RFC3339 format first (with timezone)
	t, err := time.Parse(time.RFC3339, dateTime)
	if err == nil {
		return t, nil
	}

	// Try without timezone
	t, err = time.Parse("2006-01-02T15:04:05", dateTime)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime %q", dateTime)
}

// formatAirportName creates a formatted airport name from code and city.
func formatAirportName(code, city string) string {
	if city == "" {
		return code
	}
	return fmt.Sprintf("%s (%s)", city, code)
}

// normalizeClass normalizes the class string to lowercase standard values.
func normalizeClass(class string) string {
	normalized := strings.ToLower(strings.TrimSpace(class))

	switch normalized {
	case "economy", "eco", "y":
		return "economy"
	case "business", "biz", "j", "c":
		return "business"
	case "first", "f":
		return "first"
	default:
		return "economy" // Default to economy if unknown
	}
}