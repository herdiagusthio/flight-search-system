package util

import (
	"fmt"
	"math"
	"strings"
)

// FormatIDR formats a float64 amount to Indonesian Rupiah (IDR) format.
//
// The function follows Indonesian currency conventions:
//   - Prefix: "Rp " (with space)
//   - Thousand separator: dot (.)
//   - No decimal places
//
// Examples:
//   - 0 → "Rp 0"
//   - 500 → "Rp 500"
//   - 1500 → "Rp 1.500"
//   - 1500000 → "Rp 1.500.000"
//   - 1500000000 → "Rp 1.500.000.000"
//
// Note:
//   - Decimal values are rounded using standard rounding rules
//   - Very large numbers are supported up to int64 limits
func FormatIDR(amount float64) string {
	// Round and convert to integer for formatting
	intAmount := int64(math.Round(amount))

	// Handle zero case
	if intAmount == 0 {
		return "Rp 0"
	}

	// Convert to string
	str := fmt.Sprintf("%d", intAmount)

	// Add thousand separators (dots)
	var result strings.Builder
	length := len(str)

	for i, char := range str {
		// Add dot before every 3rd digit from the right
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(char)
	}

	// Build final string with prefix
	return "Rp " + result.String()
}
