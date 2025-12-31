// Package util provides utility functions for working with timezones and time operations.
// It includes caching mechanisms for improved performance when working with multiple timezone conversions.
package util

import (
	"fmt"
	"sync"
	"time"
)

// locationCache is a thread-safe cache for storing loaded time.Location objects.
// This prevents repeated calls to time.LoadLocation which can be expensive.
var locationCache sync.Map

// Common timezone constants used in the flight search system.
// These represent the IANA Time Zone Database names for various regions.
const (
	// UTC represents Coordinated Universal Time
	UTC = "UTC"

	// Indonesian time zones
	WIB  = "Asia/Jakarta"   // Western Indonesia Time (UTC+7)
	WITA = "Asia/Makassar"  // Central Indonesia Time (UTC+8)
	WIT  = "Asia/Jayapura"  // Eastern Indonesia Time (UTC+9)

	// Other Asian time zones
	SGT = "Asia/Singapore" // Singapore Time (UTC+8)
	JST = "Asia/Tokyo"     // Japan Standard Time (UTC+9)
)

func GetLocation(name string) (*time.Location, error) {
	if loc, ok := locationCache.Load(name); ok {
		return loc.(*time.Location), nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %q: %w", name, err)
	}

	locationCache.Store(name, loc)
	return loc, nil
}



func ParseInTimezone(layout, value, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, loc)
}



// GetTimezoneByAirport returns the IANA timezone for an Indonesian airport code.
// Indonesian airports are divided into three time zones:
//   - WIB (Western): Jakarta, Surabaya, Bandung, Medan, Semarang, Yogyakarta, etc.
//   - WITA (Central): Bali, Makassar, Balikpapan, Manado, etc.
//   - WIT (Eastern): Jayapura, Ambon, Timika, etc.
//
// For non-Indonesian or unknown airports, defaults to WIB.
func GetTimezoneByAirport(airportCode string) string {
	switch airportCode {
	// Central Indonesia Time (WITA) - UTC+8
	case "DPS", // Denpasar, Bali
		"UPG", // Makassar
		"BPN", // Balikpapan
		"MDC", // Manado
		"PLW", // Palu
		"KDI", // Kendari
		"LOP", // Lombok
		"BDJ": // Banjarmasin
		return WITA

	// Eastern Indonesia Time (WIT) - UTC+9
	case "DJJ", // Jayapura
		"AMQ", // Ambon
		"TIM", // Timika
		"MKQ", // Merauke
		"SOQ", // Sorong
		"BIK": // Biak
		return WIT

	// Western Indonesia Time (WIB) - UTC+7 (default)
	default:
		// Includes: CGK, SUB, BDO, KNO, SRG, JOG, PLM, PKU, BTH, PNK, etc.
		return WIB
	}
}