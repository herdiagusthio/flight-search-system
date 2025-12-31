package lionair

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/rs/zerolog/log"
)

func normalize(lionAirFlights []LionAirFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(lionAirFlights))
	skippedCount := 0

	for _, f := range lionAirFlights {
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
			Int("total", len(lionAirFlights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

// normalizeFlight converts a single Lion Air flight to a domain Flight entity.
func normalizeFlight(f LionAirFlight) (domain.Flight, error) {
	// Parse departure time with timezone
	departureTime, err := parseDateTimeWithTimezone(f.Schedule.Departure, f.Schedule.DepartureTimezone, f.Route.From.Code)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}

	// Parse arrival time with timezone
	arrivalTime, err := parseDateTimeWithTimezone(f.Schedule.Arrival, f.Schedule.ArrivalTimezone, f.Route.To.Code)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	// Parse baggage allowances
	cabinKg := parseBaggageWeight(f.Services.BaggageAllowance.Cabin)
	checkedKg := parseBaggageWeight(f.Services.BaggageAllowance.Hold)

	// Calculate stops
	stops := 0
	if !f.IsDirect {
		stops = f.StopCount
		if stops == 0 && len(f.Layovers) > 0 {
			stops = len(f.Layovers)
		}
	}

	// Build amenities from services fields
	var amenities []string
	if f.Services.WiFiAvailable {
		amenities = append(amenities, "wifi")
	}
	if f.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}
	if amenities == nil {
		amenities = []string{}
	}

	return domain.Flight{
		ID:           f.ID,
		FlightNumber: f.ID,
		Airline: domain.AirlineInfo{
			Code: f.Carrier.IATA,
			Name: f.Carrier.Name,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.Route.From.Code,
			AirportName: f.Route.From.Name,
			DateTime:    departureTime,
			Timezone:    f.Schedule.DepartureTimezone,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.Route.To.Code,
			AirportName: f.Route.To.Name,
			DateTime:    arrivalTime,
			Timezone:    f.Schedule.ArrivalTimezone,
		},
		Duration: domain.NewDurationInfo(f.FlightTime),
		Price: domain.PriceInfo{
			Amount:    f.Pricing.Total,
			Currency:  f.Pricing.Currency,
			Formatted: util.FormatIDR(f.Pricing.Total),
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		Class:          normalizeClass(f.Pricing.FareType),
		Stops:          stops,
		Provider:       ProviderName,
		AvailableSeats: f.SeatsLeft,
		Aircraft:       f.PlaneType,
		Amenities:      amenities,
	}, nil
}

// parseDateTimeWithTimezone parses a datetime string with a separate timezone.
// The datetime format is "2006-01-02T15:04:05" (ISO 8601 without offset).
// If timezone is invalid/empty, tries to look up timezone by airport code.
func parseDateTimeWithTimezone(datetime, timezone, airportCode string) (time.Time, error) {
	// Try parsing with T separator (ISO 8601 format)
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, datetime)
	if err != nil {
		// Try with space separator as fallback
		layout = "2006-01-02 15:04:05"
		t, err = time.Parse(layout, datetime)
		if err != nil {
			return time.Time{}, fmt.Errorf("unable to parse datetime %q", datetime)
		}
	}

	// Use caching for location loading
	loc, err := util.GetLocation(timezone)
	if err != nil {
		// If explicit timezone fails or is empty, try looking up by airport code
		if airportCode != "" {
			inferredTz := util.GetTimezoneByAirport(airportCode)
			loc, err = util.GetLocation(inferredTz)
		}
		
		// If still failed or no airport code, fallback to UTC
		if err != nil {
			return t.UTC(), nil
		}
	}

	// Create time in the specified/inferred timezone
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		loc,
	), nil
}

// parseBaggageWeight extracts the weight in kg from a baggage string like "7 kg".
func parseBaggageWeight(baggageStr string) int {
	// Remove "kg" suffix and trim spaces
	cleaned := strings.TrimSpace(strings.ToLower(baggageStr))
	cleaned = strings.TrimSuffix(cleaned, "kg")
	cleaned = strings.TrimSpace(cleaned)

	weight, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}
	return weight
}

// normalizeClass normalizes the class string to lowercase standard values.
func normalizeClass(class string) string {
	normalized := strings.ToLower(strings.TrimSpace(class))

	switch normalized {
	case "economy", "eco", "y", "economy_class":
		return "economy"
	case "business", "biz", "j", "c", "business_class":
		return "business"
	case "first", "f", "first_class":
		return "first"
	default:
		return "economy" // Default to economy if unknown
	}
}
