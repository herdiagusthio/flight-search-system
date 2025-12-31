package util

import "fmt"

// FormatDuration formats total minutes to human-readable duration format.
//
// The function converts minutes to a user-friendly format:
//   - Only minutes: "Xm" (e.g., "45m")
//   - Only hours: "Xh" (e.g., "2h")
//   - Hours and minutes: "Xh Ym" (e.g., "2h 30m")
//
// Examples:
//   - 0 → "0m"
//   - 45 → "45m"
//   - 60 → "1h"
//   - 90 → "1h 30m"
//   - 150 → "2h 30m"
//   - 1440 → "24h" (24 hours)
//
// Edge cases:
//   - Negative values are treated as 0
//   - Very large numbers are supported (e.g., 48h, 100h)
func FormatDuration(totalMinutes int) string {
	// Handle negative values
	if totalMinutes < 0 {
		totalMinutes = 0
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	// Only minutes (< 1 hour)
	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	// Only hours (exact hour)
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	// Hours and minutes
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
