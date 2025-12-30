package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLocation(t *testing.T) {
	// Clear cache before tests
	ClearLocationCache()

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

func TestMustGetLocation(t *testing.T) {
	ClearLocationCache()

	tests := []struct {
		name        string
		timezone    string
		shouldPanic bool
	}{
		{
			name:        "valid timezone should not panic",
			timezone:    WIB,
			shouldPanic: false,
		},
		{
			name:        "invalid timezone should panic",
			timezone:    "Invalid/Timezone",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					MustGetLocation(tt.timezone)
				})
			} else {
				assert.NotPanics(t, func() {
					loc := MustGetLocation(tt.timezone)
					assert.NotNil(t, loc)
					assert.Equal(t, tt.timezone, loc.String())
				})
			}
		})
	}
}

func TestInTimezone(t *testing.T) {
	ClearLocationCache()

	// Create a fixed time for testing
	baseTime := time.Date(2024, 12, 25, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		input    time.Time
		timezone string
		wantErr  bool
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:     "convert UTC to Jakarta",
			input:    baseTime,
			timezone: WIB,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, WIB, result.Location().String())
				// UTC+7 means Jakarta time is 7 hours ahead
				assert.Equal(t, 19, result.Hour()) // 12 + 7
			},
		},
		{
			name:     "convert UTC to Singapore",
			input:    baseTime,
			timezone: SGT,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, SGT, result.Location().String())
				// UTC+8
				assert.Equal(t, 20, result.Hour()) // 12 + 8
			},
		},
		{
			name:     "convert UTC to Tokyo",
			input:    baseTime,
			timezone: JST,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, JST, result.Location().String())
				// UTC+9
				assert.Equal(t, 21, result.Hour()) // 12 + 9
			},
		},
		{
			name:     "invalid timezone returns error",
			input:    baseTime,
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
		{
			name:     "same instant in different timezone",
			input:    baseTime,
			timezone: WIB,
			wantErr:  false,
			validate: func(t *testing.T, result time.Time) {
				// Should be the same instant in time
				assert.True(t, baseTime.Equal(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InTimezone(tt.input, tt.timezone)

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

func TestNowIn(t *testing.T) {
	ClearLocationCache()

	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "get current time in Jakarta",
			timezone: WIB,
			wantErr:  false,
		},
		{
			name:     "get current time in Singapore",
			timezone: SGT,
			wantErr:  false,
		},
		{
			name:     "get current time in UTC",
			timezone: UTC,
			wantErr:  false,
		},
		{
			name:     "invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now, err := NowIn(tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.timezone, now.Location().String())
				// Verify it's recent (within last second)
				assert.WithinDuration(t, time.Now(), now, time.Second)
			}
		})
	}
}

func TestNowInJakarta(t *testing.T) {
	ClearLocationCache()

	t.Run("returns current time in Jakarta timezone", func(t *testing.T) {
		jakartaTime, err := NowInJakarta()

		assert.NoError(t, err)
		assert.Equal(t, WIB, jakartaTime.Location().String())
		assert.WithinDuration(t, time.Now(), jakartaTime, time.Second)
	})
}

func TestNowInUTC(t *testing.T) {
	t.Run("returns current time in UTC", func(t *testing.T) {
		utcTime := NowInUTC()

		assert.Equal(t, "UTC", utcTime.Location().String())
		assert.WithinDuration(t, time.Now().UTC(), utcTime, time.Second)
	})
}

func TestParseInTimezone(t *testing.T) {
	ClearLocationCache()

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

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "format date in UTC",
			input:    time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC),
			expected: "2024-12-25",
		},
		{
			name:     "format date with single digit month and day",
			input:    time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-05",
		},
		{
			name:     "format date in different timezone",
			input:    time.Date(2024, 12, 31, 23, 59, 59, 0, MustGetLocation(WIB)),
			expected: "2024-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "format time in afternoon",
			input:    time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC),
			expected: "14:30",
		},
		{
			name:     "format time in morning",
			input:    time.Date(2024, 12, 25, 9, 5, 0, 0, time.UTC),
			expected: "09:05",
		},
		{
			name:     "format midnight",
			input:    time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: "00:00",
		},
		{
			name:     "format time just before midnight",
			input:    time.Date(2024, 12, 25, 23, 59, 0, 0, time.UTC),
			expected: "23:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "format full datetime",
			input:    time.Date(2024, 12, 25, 14, 30, 45, 0, time.UTC),
			expected: "2024-12-25 14:30:45",
		},
		{
			name:     "format datetime with zeros",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-01 00:00:00",
		},
		{
			name:     "format datetime end of year",
			input:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2024-12-31 23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartofDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:  "start of day from afternoon time",
			input: time.Date(2024, 12, 25, 14, 30, 45, 123456789, time.UTC),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
				assert.Equal(t, 0, result.Nanosecond())
				assert.Equal(t, "UTC", result.Location().String())
			},
		},
		{
			name:  "start of day preserves timezone",
			input: time.Date(2024, 12, 25, 14, 30, 45, 0, MustGetLocation(WIB)),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, WIB, result.Location().String())
			},
		},
		{
			name:  "start of day from midnight",
			input: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
			},
		},
		{
			name:  "start of day from end of day",
			input: time.Date(2024, 12, 25, 23, 59, 59, 999999999, time.UTC),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Nanosecond())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartofDay(tt.input)
			tt.validate(t, result)
		})
	}
}

func TestEndofDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:  "end of day from morning time",
			input: time.Date(2024, 12, 25, 9, 30, 45, 0, time.UTC),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, 23, result.Hour())
				assert.Equal(t, 59, result.Minute())
				assert.Equal(t, 59, result.Second())
				assert.Equal(t, int(time.Second-time.Nanosecond), result.Nanosecond())
				assert.Equal(t, "UTC", result.Location().String())
			},
		},
		{
			name:  "end of day preserves timezone",
			input: time.Date(2024, 12, 25, 14, 30, 45, 0, MustGetLocation(WIB)),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 23, result.Hour())
				assert.Equal(t, 59, result.Minute())
				assert.Equal(t, 59, result.Second())
				assert.Equal(t, WIB, result.Location().String())
			},
		},
		{
			name:  "end of day from midnight",
			input: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 25, result.Day())
				assert.Equal(t, 23, result.Hour())
				assert.Equal(t, 59, result.Minute())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndofDay(tt.input)
			tt.validate(t, result)
		})
	}
}

func TestClearLocationCache(t *testing.T) {
	t.Run("clears all cached locations", func(t *testing.T) {
		// Populate cache
		_, err := GetLocation(WIB)
		require.NoError(t, err)
		_, err = GetLocation(SGT)
		require.NoError(t, err)
		_, err = GetLocation(JST)
		require.NoError(t, err)

		// Verify cache has items
		count := 0
		locationCache.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 3, count, "cache should have 3 items")

		// Clear cache
		ClearLocationCache()

		// Verify cache is empty
		count = 0
		locationCache.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count, "cache should be empty after clearing")
	})

	t.Run("can reload locations after clearing cache", func(t *testing.T) {
		ClearLocationCache()

		// Load location
		loc1, err := GetLocation(WIB)
		require.NoError(t, err)
		require.NotNil(t, loc1)

		// Clear and reload
		ClearLocationCache()
		loc2, err := GetLocation(WIB)
		require.NoError(t, err)
		require.NotNil(t, loc2)

		// Should still work correctly
		assert.Equal(t, WIB, loc2.String())
	})
}

// TestTimezoneConstants verifies that all predefined timezone constants are valid
func TestTimezoneConstants(t *testing.T) {
	ClearLocationCache()

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
	ClearLocationCache()

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

// TestStartEndOfDayRelationship verifies the relationship between start and end of day
func TestStartEndOfDayRelationship(t *testing.T) {
	now := time.Now()

	start := StartofDay(now)
	end := EndofDay(now)

	// End should be after start
	assert.True(t, end.After(start), "end of day should be after start of day")

	// They should be on the same day
	assert.Equal(t, start.Year(), end.Year())
	assert.Equal(t, start.Month(), end.Month())
	assert.Equal(t, start.Day(), end.Day())

	// Duration between them should be just under 24 hours
	duration := end.Sub(start)
	assert.True(t, duration < 24*time.Hour)
	assert.True(t, duration >= 23*time.Hour+59*time.Minute)
}
