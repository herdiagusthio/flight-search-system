package batikair

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/pkg/util"
	"github.com/rs/zerolog/log"
)

var durationRegex = regexp.MustCompile(`(?:(\\d+)h)?\\s*(?:(\\d+)m)?`)

func normalize(batikAirFlights []BatikAirFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(batikAirFlights))
	skippedCount := 0

	for _, f := range batikAirFlights {
		normalized, ok := normalizeFlight(f)
		if !ok {
			skippedCount++
			continue
		}

		// Final domain validation check
		if err := normalized.Validate(); err != nil {
			log, la := log.Warn().
				Str("provider", ProviderName).
				Str("flight_number", normalized.FlightNumber.String()).
				Err(err).
				Msg("Flight validation failed")
			_ = la
			skippedCount++
			continue
		}

		result = append(result, *normalized)
	}

	if skippedCount > 0 {
		log.Info().
			Str("provider", ProviderName).
			Int("skipped", skippedCount).
			Int("total", len(batikAirFlights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

func normalizeFlight(f BatikAirFlight) (*domain.Flight, bool) {
	departureTime, err := parseDateTime(f.DepartureDateTime, f.Origin)
	if err != nil {
		return nil, false
	}

	arrivalTime, err := parseDateTime(f.ArrivalDateTime, f.Destination)
	if err != nil {
		return nil, false
	}

	durationMinutes, err := parseDurationString(f.TravelTime)
	if err != nil {
		return nil, false
	}

	cabinKg, checkedKg := parseBaggageInfo(f.BaggageInfo)

	totalPrice := f.Fare.TotalPrice
	if totalPrice == 0 {
		totalPrice = f.Fare.BasePrice + f.Fare.Taxes
	}

	// Value Object Instantiation
	depPoint, err := domain.NewFlightPoint(f.Origin, "Unknown", "", departureTime, "UTC")
	if err != nil {
		return nil, false
	}

	arrPoint, err := domain.NewFlightPoint(f.Destination, "Unknown", "", arrivalTime, "UTC")
	if err != nil {
		return nil, false
	}

	// Price transformation: float64 -> int64 cents
	price := domain.PriceInfo{
		AmountCents: int64(math.Round(totalPrice * 100)),
		Currency:    "IDR",
		Formatted:   util.FormatIDR(totalPrice),
	}

	// Use the Factory to ensure domain invariants
	flight, err := domain.NewFlight(
		f.FlightNumber,
		f.FlightNumber,
		domain.AirlineInfo{
			Code: f.AirlineIATA,
			Name: f.AirlineName,
		},
		depPoint,
		arrPoint,
		domain.NewDurationInfo(durationMinutes),
		price,
		domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		mapCabinClass(f.Fare.Class),
		f.NumberOfStops,
		ProviderName,
	)

	if err != nil {
		return nil, false
	}

	// Attach optional data
	flight.AvailableSeats = f.SeatsAvailable
	flight.Aircraft = f.AircraftModel
	flight.Amenities = f.OnboardServices

	return flight, true
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

func parseDurationString(duration string) (int, error) {
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	matches := durationRegex.FindStringSubmatch(duration)
	if matches == nil || (matches[1] == "" && matches[2] == "") {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	var hours, minutes int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "", _ {
		minutes, _ = strconv.Atoi(matches[2])
	}

	return hours*60 + minutes, nil
}

func parseBaggageInfo(baggageInfo string) (cabinKg, checkedKg int) {
	cabinKg = 7
	checkedKg = 20
	if baggageInfo == "" {
		return
	}
	info := strings.ToLower(baggageInfo)
	cabinRegex := regexp.MustCompile(`(\\d+)\\s*kg\\s*cabin`)
	if matches := cabinRegex.FindStringSubmatch(info); len(matches) > 1 {
		cabinKg, _ = strconv.Atoi(matches[1])
	}
	checkedRegex := regexp.MustCompile(`(\\d+)\\s*kg\\s*checked`)
	if matches := checkedRegex.FindStringSubmatch(info); len(matches) > 1 {
		checkedKg, _ = strconv.Atoi(matches[1])
	}
	return
}

func mapCabinClass(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	classMap := map[string]string{
		"Y": "economy",
		"W": "premium_economy",
		"C": "business",
		"J": "business",
		"F": "first",
	}
	if class, ok := classMap[code]; ok {
		return class
	}
	return "economy"
}
