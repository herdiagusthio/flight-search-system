package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		// Zero and edge cases
		{
			name:     "zero minutes",
			minutes:  0,
			expected: "0m",
		},
		{
			name:     "negative minutes treated as zero",
			minutes:  -10,
			expected: "0m",
		},
		{
			name:     "negative large value",
			minutes:  -150,
			expected: "0m",
		},

		// Only minutes (< 1 hour)
		{
			name:     "single minute",
			minutes:  1,
			expected: "1m",
		},
		{
			name:     "45 minutes",
			minutes:  45,
			expected: "45m",
		},
		{
			name:     "59 minutes",
			minutes:  59,
			expected: "59m",
		},

		// Exact hours
		{
			name:     "exactly 1 hour",
			minutes:  60,
			expected: "1h",
		},
		{
			name:     "exactly 2 hours",
			minutes:  120,
			expected: "2h",
		},
		{
			name:     "exactly 10 hours",
			minutes:  600,
			expected: "10h",
		},
		{
			name:     "exactly 24 hours",
			minutes:  1440,
			expected: "24h",
		},

		// Hours and minutes
		{
			name:     "1 hour 1 minute",
			minutes:  61,
			expected: "1h 1m",
		},
		{
			name:     "1 hour 30 minutes",
			minutes:  90,
			expected: "1h 30m",
		},
		{
			name:     "2 hours 30 minutes",
			minutes:  150,
			expected: "2h 30m",
		},
		{
			name:     "2 hours 45 minutes",
			minutes:  165,
			expected: "2h 45m",
		},
		{
			name:     "3 hours 5 minutes",
			minutes:  185,
			expected: "3h 5m",
		},
		{
			name:     "4 hours 20 minutes",
			minutes:  260,
			expected: "4h 20m",
		},

		// Typical flight durations
		{
			name:     "typical short flight (1h 50m)",
			minutes:  110,
			expected: "1h 50m",
		},
		{
			name:     "typical medium flight (3h 15m)",
			minutes:  195,
			expected: "3h 15m",
		},
		{
			name:     "typical long flight (8h 45m)",
			minutes:  525,
			expected: "8h 45m",
		},

		// > 24 hours (edge cases)
		{
			name:     "25 hours 15 minutes",
			minutes:  1515,
			expected: "25h 15m",
		},
		{
			name:     "48 hours exact",
			minutes:  2880,
			expected: "48h",
		},
		{
			name:     "100 hours",
			minutes:  6000,
			expected: "100h",
		},

		// Real-world examples from Indonesian routes
		{
			name:     "CGK to DPS (typical 2h)",
			minutes:  120,
			expected: "2h",
		},
		{
			name:     "CGK to DPS with layover (5h 30m)",
			minutes:  330,
			expected: "5h 30m",
		},
		{
			name:     "short hop (40m)",
			minutes:  40,
			expected: "40m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.minutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration_Consistency(t *testing.T) {
	t.Run("same input produces same output", func(t *testing.T) {
		minutes := 150
		result1 := FormatDuration(minutes)
		result2 := FormatDuration(minutes)
		assert.Equal(t, result1, result2)
		assert.Equal(t, "2h 30m", result1)
	})

	t.Run("format is consistent across ranges", func(t *testing.T) {
		// All exact hours should end with "h"
		assert.Equal(t, "1h", FormatDuration(60))
		assert.Equal(t, "5h", FormatDuration(300))
		assert.Equal(t, "10h", FormatDuration(600))

		// All sub-hour should end with "m"
		assert.Equal(t, "1m", FormatDuration(1))
		assert.Equal(t, "30m", FormatDuration(30))
		assert.Equal(t, "59m", FormatDuration(59))

		// All mixed should have "h Xm" format
		assert.Equal(t, "1h 1m", FormatDuration(61))
		assert.Equal(t, "5h 30m", FormatDuration(330))
		assert.Equal(t, "10h 45m", FormatDuration(645))
	})
}

func TestFormatDuration_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{
			name:     "boundary: 59 minutes (just before 1h)",
			minutes:  59,
			expected: "59m",
		},
		{
			name:     "boundary: 60 minutes (exactly 1h)",
			minutes:  60,
			expected: "1h",
		},
		{
			name:     "boundary: 61 minutes (just after 1h)",
			minutes:  61,
			expected: "1h 1m",
		},
		{
			name:     "boundary: 119 minutes (just before 2h)",
			minutes:  119,
			expected: "1h 59m",
		},
		{
			name:     "boundary: 120 minutes (exactly 2h)",
			minutes:  120,
			expected: "2h",
		},
		{
			name:     "boundary: 121 minutes (just after 2h)",
			minutes:  121,
			expected: "2h 1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.minutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
