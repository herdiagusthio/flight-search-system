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

func MustGetLocation(name string) *time.Location {
	loc, err := GetLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func InTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		return t, err
	}
	return t.In(loc), nil
}

func NowIn(timezone string) (time.Time, error) {
	return InTimezone(time.Now(), timezone)
}

func NowInJakarta() (time.Time, error) {
	loc := MustGetLocation(WIB)
	return time.Now().In(loc), nil
}

func NowInUTC() time.Time {
	return time.Now().UTC()
}

func ParseInTimezone(layout, value, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, loc)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func StartofDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndofDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func ClearLocationCache() {
	locationCache.Range(func(key, _ interface{}) bool {
		locationCache.Delete(key)
		return true
	})
}