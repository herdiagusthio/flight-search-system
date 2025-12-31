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
			Int("total", len(flights)).
			Msg("Skipped invalid flights during normalization")
	}

	return result
}

// normalizeSingle converts a single AirAsiaFlight to a domain.Flight.
// Returns false if the flight cannot be normalized (e.g., invalid datetime).
func normalizeSingle(f AirAsiaFlight) (domain.Flight, bool) {
	departureTime, err := parseDateTime(f.DepartTime)
	if err != nil {
		return domain.Flight{}, false
	}

	arrivalTime, err := parseDateTime(f.ArriveTime)
	if err != nil {
		return domain.Flight{}, false
	}

	stopsCount := directFlightToStops(f.DirectFlight, f.Stops)
	flightID := generateFlightID(f)
	cabinKg, checkedKg := parseBaggageNote(f.BaggageNote)

	// Parse baggage descriptions from the note
	carryOnDesc, checkedDesc := formatBaggageDescriptions(f.BaggageNote, cabinKg, checkedKg)

	return domain.Flight{
		ID:           flightID,
		FlightNumber: f.FlightCode,
		Airline: domain.AirlineInfo{
			Code: extractAirlineCode(f.FlightCode),
			Name: f.Airline,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.FromAirport,
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.ToAirport,
			DateTime:    arrivalTime,
		},
		Duration: domain.DurationInfo{
			TotalMinutes: hoursToMinutes(f.DurationHours),
			Formatted:    util.FormatDuration(hoursToMinutes(f.DurationHours)),
		},
		Price: domain.PriceInfo{
			Amount:   f.PriceIDR,
			Currency: "IDR",
			Formatted: util.FormatIDR(f.PriceIDR),
		},
		Baggage: domain.BaggageInfo{
			CabinKg:     cabinKg,
			CheckedKg:   checkedKg,
			CarryOnDesc: carryOnDesc,
			CheckedDesc: checkedDesc,
		},
		Class:          strings.ToLower(f.CabinClass),
		Stops:          stopsCount,
		Provider:       ProviderName,
		AvailableSeats: f.Seats,
		Aircraft:       "", // AirAsia mock data doesn't include aircraft info
		Amenities:      []string{}, // AirAsia mock data doesn't include amenities
	}, true
}

// generateFlightID creates a unique identifier for a flight.
func generateFlightID(f AirAsiaFlight) string {
	return fmt.Sprintf("%s-%s-%s-%s", ProviderName, f.FlightCode, f.FromAirport, f.ToAirport)
}

// extractAirlineCode extracts the 2-character IATA airline code from a flight code.
// Flight codes follow the format {airline_code}{flight_number}, e.g., "QZ520" → "QZ"
// Falls back to the default airlineCode constant if the flight code is too short.
func extractAirlineCode(flightCode string) string {
	if len(flightCode) >= 2 {
		return strings.ToUpper(flightCode[:2])
	}
	return airlineCode
}

// hoursToMinutes converts float hours to integer minutes with proper rounding.
// Examples: 1.75 → 105, 2.5 → 150, 0.5 → 30
func hoursToMinutes(hours float64) int {
	return int(math.Round(hours * 60))
}

// directFlightToStops converts the direct_flight boolean to stops count.
// If direct_flight is true, returns 0.
// If direct_flight is false, returns the actual number of stops or 1 if unknown.
func directFlightToStops(isDirect bool, stops []AirAsiaStop) int {
	if isDirect {
		return 0
	}
	// If stops array is provided, use its length
	if len(stops) > 0 {
		return len(stops)
	}
	// Default to 1 stop if not direct but no stops array
	return 1
}

// parseDateTime parses an ISO 8601 datetime string to time.Time.
// Supports formats with timezone offset (e.g., "2025-12-15T06:00:00+07:00").
func parseDateTime(datetime string) (time.Time, error) {
	// Try standard RFC3339 format first
	t, err := time.Parse(time.RFC3339, datetime)
	if err == nil {
		return t, nil
	}

	// Try without colon in timezone (e.g., +0700)
	t, err = time.Parse("2006-01-02T15:04:05-0700", datetime)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", datetime)
}

// parseBaggageNote extracts baggage weights from a baggage note string.
// AirAsia typically provides a note like "Cabin baggage only, checked bags additional fee"
// which means only cabin baggage is included (default 7kg), no checked baggage.
func parseBaggageNote(note string) (cabinKg, checkedKg int) {
	noteLower := strings.ToLower(note)

	// Default cabin baggage for AirAsia is 7kg
	cabinKg = 7

	// Check if checked baggage is included
	if strings.Contains(noteLower, "checked bag") && strings.Contains(noteLower, "additional fee") {
		// No checked baggage included
		checkedKg = 0
	} else if strings.Contains(noteLower, "20kg") {
		checkedKg = 20
	} else if strings.Contains(noteLower, "15kg") {
		checkedKg = 15
	} else {
		// Default to no checked baggage for low-cost carrier
		checkedKg = 0
	}

	return cabinKg, checkedKg
}

// formatBaggageDescriptions generates descriptive baggage strings from the baggage note.
// For AirAsia, the baggage note contains descriptive text like "Cabin baggage only, checked bags additional fee"
func formatBaggageDescriptions(note string, cabinKg, checkedKg int) (carryOnDesc, checkedDesc string) {
	noteLower := strings.ToLower(note)

	// Determine carry-on description
	if strings.Contains(noteLower, "cabin baggage only") {
		carryOnDesc = "Cabin baggage only"
	} else if cabinKg > 0 {
		carryOnDesc = fmt.Sprintf("%dkg cabin", cabinKg)
	} else {
		carryOnDesc = "Not included"
	}

	// Determine checked baggage description
	if strings.Contains(noteLower, "additional fee") && strings.Contains(noteLower, "checked") {
		checkedDesc = "Additional fee"
	} else if checkedKg > 0 {
		checkedDesc = fmt.Sprintf("%dkg checked", checkedKg)
	} else {
		checkedDesc = "Not included"
	}

	return carryOnDesc, checkedDesc
}
