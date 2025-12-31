package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLocation(t *testing.T) {
	// Clear cache before tests
	clearLocationCache()

	tests := []struct {
		name      string
		timezone  string
		wantErr   bool
		errMsg    string
		cacheable bool // Whether this location should be cached
	}{
		{
			name:      "valid UTC timezone",
			timezone:  "UTC",
			wantErr:   false,
			cacheable: true,
		},
		{
			name:      "valid Asia/Jakarta timezone",
			timezone:  WIB,
			wantErr:   false,
			cacheable: true,
		},
		{
			name:      "valid Asia/Singapore timezone",
			timezone:  SGT,
			wantErr:   false,
			cacheable: true,
		},
		{
			name:      "valid Asia/Tokyo timezone",
			timezone:  JST,
			wantErr:   false,
			cacheable: true,
		},
		{
			name:     "invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
			errMsg:   "failed to load timezone",
		},
		{
			name:      "empty timezone defaults to UTC",
			timezone:  "",
			wantErr:   false,
			cacheable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := GetLocation(tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, loc)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loc)
				// Empty string defaults to UTC, so check for UTC instead
				expectedName := tt.timezone
				if expectedName == "" {
					expectedName = "UTC"
				}
				assert.Equal(t, expectedName, loc.String())

				// Test caching - second call should return cached value
				if tt.cacheable {
					loc2, err2 := GetLocation(tt.timezone)
					assert.NoError(t, err2)
					assert.Equal(t, loc, loc2, "should return same cached instance")
				}
			}
		})
	}
}



func TestParseInTimezone(t *testing.T) {
	clearLocationCache()

	tests := []struct {
		name     string
		layout   string
		value    string
		timezone string
		wantErr  bool
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:     "parse date in Jakarta timezone",
			layout:   "2006-01-02 15:04:05",
			value:    "2024-12-25 14:30:45",
			timezone: WIB,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
				assert.Equal(t, 45, result.Second())
				assert.Equal(t, WIB, result.Location().String())
			},
		},
		{
			name:     "parse date only",
			layout:   "2006-01-02",
			value:    "2024-12-25",
			timezone: SGT,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, SGT, result.Location().String())
			},
		},
		{
			name:     "parse time only",
			layout:   "15:04",
			value:    "14:30",
			timezone: JST,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
				assert.Equal(t, JST, result.Location().String())
			},
		},
		{
			name:     "invalid date format",
			layout:   "2006-01-02",
			value:    "invalid-date",
			timezone: WIB,
			wantErr:  true,
		},
		{
			name:     "invalid timezone",
			layout:   "2006-01-02",
			value:    "2024-12-25",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
		{
			name:     "mismatched layout and value",
			layout:   "2006-01-02",
			value:    "2024-12-25 14:30:45",
			timezone: WIB,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInTimezone(tt.layout, tt.value, tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}



// TestTimezoneConstants verifies that all predefined timezone constants are valid
func TestTimezoneConstants(t *testing.T) {
	clearLocationCache()

	tests := []struct {
		name     string
		timezone string
	}{
		{"UTC constant", UTC},
		{"WIB constant", WIB},
		{"WITA constant", WITA},
		{"WIT constant", WIT},
		{"SGT constant", SGT},
		{"JST constant", JST},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := GetLocation(tt.timezone)
			assert.NoError(t, err, "timezone constant %s should be valid", tt.timezone)
			assert.NotNil(t, loc)
		})
	}
}

// TestConcurrentAccess verifies thread-safety of the location cache
func TestConcurrentAccess(t *testing.T) {
	clearLocationCache()

	const goroutines = 100
	const iterations = 10

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			timezones := []string{WIB, SGT, JST, UTC, WITA, WIT}
			for j := 0; j < iterations; j++ {
				tz := timezones[j%len(timezones)]
				loc, err := GetLocation(tz)
				assert.NoError(t, err)
				assert.NotNil(t, loc)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// TestGetTimezoneByAirport verifies airport code to timezone mapping
func TestGetTimezoneByAirport(t *testing.T) {
	tests := []struct {
		name         string
		airportCode  string
		expectedZone string
	}{
		// Western Indonesia (WIB) - UTC+7
		{"Jakarta CGK", "CGK", WIB},
		{"Surabaya SUB", "SUB", WIB},
		{"Bandung BDO", "BDO", WIB},
		{"Medan KNO", "KNO", WIB},
		{"Semarang SRG", "SRG", WIB},
		{"Yogyakarta JOG", "JOG", WIB},
		{"Palembang PLM", "PLM", WIB},
		{"Batam BTH", "BTH", WIB},

		// Central Indonesia (WITA) - UTC+8
		{"Bali DPS", "DPS", WITA},
		{"Makassar UPG", "UPG", WITA},
		{"Balikpapan BPN", "BPN", WITA},
		{"Manado MDC", "MDC", WITA},
		{"Palu PLW", "PLW", WITA},
		{"Kendari KDI", "KDI", WITA},
		{"Lombok LOP", "LOP", WITA},
		{"Banjarmasin BDJ", "BDJ", WITA},

		// Eastern Indonesia (WIT) - UTC+9
		{"Jayapura DJJ", "DJJ", WIT},
		{"Ambon AMQ", "AMQ", WIT},
		{"Timika TIM", "TIM", WIT},
		{"Merauke MKQ", "MKQ", WIT},
		{"Sorong SOQ", "SOQ", WIT},
		{"Biak BIK", "BIK", WIT},

		// Unknown/International defaults to WIB
		{"Unknown XYZ", "XYZ", WIB},
		{"International", "SIN", WIB},
		{"Empty code", "", WIB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimezoneByAirport(tt.airportCode)
			assert.Equal(t, tt.expectedZone, result)
		})
	}
}

// clearLocationCache clears the location cache for testing purposes.
func clearLocationCache() {
	locationCache.Range(func(key, _ interface{}) bool {
		locationCache.Delete(key)
		return true
	})
}

