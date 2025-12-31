package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatIDR(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		// Zero and small amounts
		{
			name:     "zero amount",
			amount:   0,
			expected: "Rp 0",
		},
		{
			name:     "small amount less than 1000",
			amount:   500,
			expected: "Rp 500",
		},
		{
			name:     "small amount 750",
			amount:   750,
			expected: "Rp 750",
		},
		{
			name:     "small amount 499",
			amount:   499,
			expected: "Rp 499",
		},

		// Thousands
		{
			name:     "exact thousand",
			amount:   1000,
			expected: "Rp 1.000",
		},
		{
			name:     "one and a half thousand",
			amount:   1500,
			expected: "Rp 1.500",
		},
		{
			name:     "multiple thousands",
			amount:   5000,
			expected: "Rp 5.000",
		},
		{
			name:     "tens of thousands",
			amount:   50000,
			expected: "Rp 50.000",
		},
		{
			name:     "hundreds of thousands",
			amount:   500000,
			expected: "Rp 500.000",
		},

		// Millions
		{
			name:     "one million",
			amount:   1000000,
			expected: "Rp 1.000.000",
		},
		{
			name:     "one and a half million",
			amount:   1500000,
			expected: "Rp 1.500.000",
		},
		{
			name:     "complex millions",
			amount:   2345000,
			expected: "Rp 2.345.000",
		},
		{
			name:     "tens of millions",
			amount:   50000000,
			expected: "Rp 50.000.000",
		},

		// Billions
		{
			name:     "one billion",
			amount:   1000000000,
			expected: "Rp 1.000.000.000",
		},
		{
			name:     "multiple billions",
			amount:   5500000000,
			expected: "Rp 5.500.000.000",
		},

		// Decimal rounding
		{
			name:     "decimal rounds down (1500400.4)",
			amount:   1500400.4,
			expected: "Rp 1.500.400",
		},
		{
			name:     "decimal rounds up (1500500.5)",
			amount:   1500500.5,
			expected: "Rp 1.500.501",
		},
		{
			name:     "decimal rounds up (1500600.6)",
			amount:   1500600.6,
			expected: "Rp 1.500.601",
		},
		{
			name:     "decimal with cents (1500000.50)",
			amount:   1500000.50,
			expected: "Rp 1.500.001",
		},

		// Very large numbers
		{
			name:     "trillion",
			amount:   1000000000000,
			expected: "Rp 1.000.000.000.000",
		},
		{
			name:     "complex large number",
			amount:   123456789000,
			expected: "Rp 123.456.789.000",
		},

		// Real-world flight prices
		{
			name:     "typical economy flight",
			amount:   850000,
			expected: "Rp 850.000",
		},
		{
			name:     "typical business flight",
			amount:   3250000,
			expected: "Rp 3.250.000",
		},
		{
			name:     "budget airline price",
			amount:   299000,
			expected: "Rp 299.000",
		},
		{
			name:     "premium route price",
			amount:   12500000,
			expected: "Rp 12.500.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIDR(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatIDR_Consistency(t *testing.T) {
	t.Run("same input produces same output", func(t *testing.T) {
		amount := 1500000.0
		result1 := FormatIDR(amount)
		result2 := FormatIDR(amount)
		assert.Equal(t, result1, result2)
	})

	t.Run("integer amounts produce consistent results", func(t *testing.T) {
		result1 := FormatIDR(1500400)
		result2 := FormatIDR(1500400.0)
		assert.Equal(t, result1, result2)
		assert.Equal(t, "Rp 1.500.400", result1)
	})
}

func TestFormatIDR_RoundingBehavior(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{
			name:     "0.5 rounds to 1",
			amount:   0.5,
			expected: "Rp 1",
		},
		{
			name:     "0.4 rounds to 0",
			amount:   0.4,
			expected: "Rp 0",
		},
		{
			name:     "1500.5 rounds to 1501",
			amount:   1500.5,
			expected: "Rp 1.501",
		},
		{
			name:     "1500.4 rounds to 1500",
			amount:   1500.4,
			expected: "Rp 1.500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIDR(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
