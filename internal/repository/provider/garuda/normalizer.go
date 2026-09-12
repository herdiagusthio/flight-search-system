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
		normalized, ok := normalizeFlight(f)
		if !ok {
			skippedCount++
			continue
		}

		// Final domain validation check
		if err := normalized.Validate(); err != nil {
			log.Warn().
				Str("provider", ProviderName).
				Str("flight_number", normalized.FlightNumber.String()).
				Err(err).
				Msg("Flight validation failed")
			skippedCount++
			continue
		}

		result = append(result, *normalized)
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

func normalizeFlight(f GarudaFlight) (*domain.Flight, bool) {
	departureTime, err := parseDateTime(f.Departure.Time, f.Departure.Airport)
	if err != nil {
		return nil, false
	}

	arrivalTime, err := parseDateTime(f.Arrival.Time, f.Arrival.Airport)
	if err != nil {
		return nil, false
	}

	stops := f.Stops
	if len(f.Segments) > 1 {
		stops = len(f.Segments) - 1
	}

	// Value Object Instantiation
	depPoint, err := domain.NewFlightPoint(f.Departure.Airport, formatAirportName(f.Departure.Airport, f.Departure.City), f.Departure.Terminal, departureTime, "UTC")
	if err != nil {
		return nil, false
	}

	arrPoint, err := domain.NewFlightPoint(f.Arrival.Airport, formatAirportName(f.Arrival.Airport, f.Arrival.City), f.Arrival.Terminal, arrivalTime, "UTC")
	if err != nil {
		return nil, false
	}

	// Price transformation: float64 -> int64 cents
	price := domain.PriceInfo{
		AmountCents: int64(math.Round(f.Price.Amount * 100)),
		Currency:    "IDR", 
		Formatted:   util.FormatIDR(f.Price.Amount),
	}

	// Use the Factory to ensure domain invariants
	flight, err := domain.NewFlight(
		f.FlightID,
		f.FlightID,
		domain.AirlineInfo{
			Code: f.AirlineCode,
			Name: f.Airline,
		},
		depPoint,
		arrPoint,
		domain.NewDurationInfo(f.DurationMinutes),
		price,
		domain.BaggageInfo{
			CabinKg:   f.Baggage.CarryOn * DefaultCabinBaggageKg,
			CheckedKg: f.Baggage.Checked * DefaultCheckedBaggageKg,
		},
		normalizeClass(f.FareClass),
		stops,
		ProviderName,
	)

	if err != nil {
		return nil, false
	}

	// Attach optional data
	flight.AvailableSeats = f.AvailableSeats
	flight.Aircraft = f.Aircraft
	flight.Amenities = f.Amenities

	return flight, true
}

func parseDateTime(dateTime, airportCode string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, dateTime)
	if err == nil {
		return t, nil
	}
	timezone := util.GetTimezoneByAirport(airportCode)
	t, err = util.ParseInTimezone("2006-01-02T15:04:05", dateTime, timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse datetime %q for airport %s: %w", dateTime, airportCode, err)
	}
	return t, nil
}

func formatAirportName(code, city string) string {
	if city == "" {
		return code
	}
	return fmt.Sprintf("%s (%s)", city, code)
}

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
		return "economy"
	}
}
