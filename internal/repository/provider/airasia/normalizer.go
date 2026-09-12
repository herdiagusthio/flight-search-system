package airasia

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/rs/zerolog/log"
)

func normalize(flights []AirAsiaFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(flights))
	skippedCount := 0

	for _, f := range flights {
		normalized, ok := normalizeSingle(f)
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
			Int("total", len(flights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

func normalizeSingle(f AirAsiaFlight) (*domain.Flight, bool) {
	departureTime, err := parseDateTime(f.DepartTime, f.FromAirport)
	if err != nil {
		return nil, false
	}

	arrivalTime, err := parseDateTime(f.ArriveTime, f.ToAirport)
	if err != nil {
		return nil, false
	}

	stopsCount := directFlightToStops(f.DirectFlight, f.Stops)
	flightID := generateFlightID(f)
	cabinKg, checkedKg := parseBaggageNote(f.BaggageNote)
	carryOnDesc, checkedDesc := formatBaggageDescriptions(f.BaggageNote, cabinKg, checkedKg)

	// Value Object Instantiation
	depPoint, err := domain.NewFlightPoint(f.FromAirport, "Unknown", "", departureTime, "UTC")
	if err != nil {
		return nil, false
	}

	arrPoint, err := domain.NewFlightPoint(f.ToAirport, "Unknown", "", arrivalTime, "UTC")
	if err != nil {
		return nil, false
	}

	// Price transformation: float64 IDR -> int64 cents
	// Since IDR typically doesn't use cents, we represent 1 IDR as 1 unit 
	// but keep the type int64 for consistency across other currencies.
	price := domain.PriceInfo{
		AmountCents: int64(math.Round(f.PriceIDR * 100)),
		Currency:    "IDR", // Simplified; in a real system, this would come from the API
		Formatted:   util.FormatIDR(f.PriceIDR),
	}

	// Use the Factory to ensure domain invariants
	flight, err := domain.NewFlight(
		flightID,
		f.FlightCode,
		domain.AirlineInfo{
			Code: extractAirlineCode(f.FlightCode),
			Name: f.Airline,
		},
		depPoint,
		arrPoint,
		domain.NewDurationInfo(hoursToMinutes(f.DurationHours)),
		price,
		domain.BaggageInfo{
			CabinKg:     cabinKg,
			CheckedKg:   checkedKg,
			CarryOnDesc: carryOnDesc,
			CheckedDesc: checkedDesc,
		},
		strings.ToLower(f.CabinClass),
		stopsCount,
		ProviderName,
	)

	if err != nil {
		return nil, false
	}

	// Attach optional data
	flight.AvailableSeats = f.Seats
	flight.Aircraft = "" 
	flight.Amenities = []string{}

	return flight, true
}

func generateFlightID(f AirAsiaFlight) string {
	return fmt.Sprintf("%s-%s-%s-%s", ProviderName, f.FlightCode, f.FromAirport, f.ToAirport)
}

func extractAirlineCode(flightCode string) string {
	if len(flightCode) >= 2 {
		return strings.ToUpper(flightCode[:2])
	}
	return "QZ"
}

func hoursToMinutes(hours float64) int {
	return int(math.Round(hours * 60))
}

func directFlightToStops(isDirect bool, stops []AirAsiaStop) int {
	if isDirect {
		return 0
	}
	if len(stops) > 0 {
		return len(stops)
	}
	return 1
}

func parseDateTime(datetime, airportCode string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, datetime)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02T15:04:05-0700", datetime)
	if err == nil {
		return t, nil
	}
	timezone := util.GetTimezoneByAirport(airportCode)
	t, err = util.ParseInTimezone("2006-01-02T15:04:05", datetime, timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse datetime %q for airport %s: %w", datetime, airportCode, err)
	}
	return t, nil
}

func parseBaggageNote(note string) (cabinKg, checkedKg int) {
	noteLower := strings.ToLower(note)
	cabinKg = 7
	if strings.Contains(noteLower, "checked bag") && strings.Contains(noteLower, "additional fee") {
		checkedKg = 0
	} else if strings.Contains(noteLower, "20kg") {
		checkedKg = 20
	} else if strings.Contains(noteLower, "15kg") {
		checkedKg = 15
	} else {
		checkedKg = 0
	}
	return cabinKg, checkedKg
}

func formatBaggageDescriptions(note string, cabinKg, checkedKg int) (carryOnDesc, checkedDesc string) {
	noteLower := strings.ToLower(note)
	if strings.Contains(noteLower, "cabin baggage only") {
		carryOnDesc = "Cabin baggage only"
	} else if cabinKg > 0 {
		carryOnDesc = fmt.Sprintf("%dkg cabin", cabinKg)
	} else {
		carryOnDesc = "Not included"
	}
	if strings.Contains(noteLower, "additional fee") && strings.Contains(noteLower, "checked") {
		checkedDesc = "Additional fee"
	} else if checkedKg > 0 {
		checkedDesc = fmt.Sprintf("%dkg checked", checkedKg)
	} else {
		checkedDesc = "Not included"
	}
	return carryOnDesc, checkedDesc
}
