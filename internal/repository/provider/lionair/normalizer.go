package lionair

import (
	"fmt"
	"math"
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
			Int("total", len(lionAirFlights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

func normalizeFlight(f LionAirFlight) (*domain.Flight, bool) {
	departureTime, err := parseDateTimeWithTimezone(f.Schedule.Departure, f.Schedule.DepartureTimezone, f.Route.From.Code)
	if err != nil {
		return nil, false
	}

	arrivalTime, err := parseDateTimeWithTimezone(f.Schedule.Arrival, f.Schedule.ArrivalTimezone, f.Route.To.Code)
	if err != nil {
		return nil, false
	}

	stops := 0
	if !f.IsDirect {
		stops = f.StopCount
		if stops == 0 && len(f.Layovers) > 0 {
			stops = len(f.Layovers)
		}
	}

	cabinKg := parseBaggageWeight(f.Services.BaggageAllowance.Cabin)
	checkedKg := parseBaggageWeight(f.Services.BaggageAllowance.Hold)

	var amenities []string
	if f.Services.WiFiAvailable {
		amenities = append(amenities, "wifi")
	}
	if f.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}

	// Value Object Instantiation
	depPoint, err := domain.NewFlightPoint(f.Route.From.Code, f.Route.From.Name, "", departureTime, f.Schedule.DepartureTimezone)
	if err != nil {
		return nil, false
	}

	arrPoint, err := domain.NewFlightPoint(f.Route.To.Code, f.Route.To.Name, "", arrivalTime, f.Schedule.ArrivalTimezone)
	if err != nil {
		return nil, false
	}

	// Price transformation: float64 -> int64 cents
	price := domain.PriceInfo{
		AmountCents: int64(math.Round(f.Pricing.Total * 100)),
		Currency:    "IDR",
		Formatted:   util.FormatIDR(f.Pricing.Total),
	}

	// Use the Factory to ensure domain invariants
	flight, err := domain.NewFlight(
		f.ID,
		f.ID,
		domain.AirlineInfo{
			Code: f.Carrier.IATA,
			Name: f.Carrier.Name,
		},
		depPoint,
		arrPoint,
		domain.NewDurationInfo(f.FlightTime),
		price,
		domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		normalizeClass(f.Pricing.FareType),
		stops,
		ProviderName,
	)

	if err != nil {
		return nil, false
	}

	// Attach optional data
	flight.AvailableSeats = f.SeatsLeft
	flight.Aircraft = f.PlaneType
	flight.Amenities = amenities

	return flight, true
}

func parseDateTimeWithTimezone(datetime, timezone, airportCode string) (time.Time, error) {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, datetime)
	if err != nil {
		layout = "2006-01-02 15:04:05"
		t, err = time.Parse(layout, datetime)
		if err != nil {
			return time.Time{}, fmt.Errorf("unable to parse datetime %q", datetime)
		}
	}

	loc, err := util.GetLocation(timezone)
	if err != nil {
		if airportCode != "" {
			inferredTz := util.GetTimezoneByAirport(airportCode)
			loc, err = util.GetLocation(inferredTz)
		}
		if err != nil {
			return t.UTC(), nil
		}
	}

	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		loc,
	), nil
}

func parseBaggageWeight(baggageStr string) int {
	cleaned := strings.TrimSpace(strings.ToLower(baggageStr))
	cleaned = strings.TrimSuffix(cleaned, "kg")
	cleaned = strings.TrimSpace(cleaned)
	weight, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}
	return weight
}

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
		return "economy"
	}
}
