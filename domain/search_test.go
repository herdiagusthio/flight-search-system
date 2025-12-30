package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchCriteriaValidate(t *testing.T) {
	validCriteria := SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		Class:         "economy",
	}

	tests := []struct {
		name           string
		modifyCriteria func(*SearchCriteria)
		expectError    bool
		errorContains  string
	}{
		{
			name:           "valid criteria",
			modifyCriteria: nil,
			expectError:    false,
		},
		{
			name: "empty origin",
			modifyCriteria: func(c *SearchCriteria) {
				c.Origin = ""
			},
			expectError:   true,
			errorContains: "origin is required",
		},
		{
			name: "invalid origin format - lowercase",
			modifyCriteria: func(c *SearchCriteria) {
				c.Origin = "cgk"
			},
			expectError:   true,
			errorContains: "3-letter IATA code",
		},
		{
			name: "invalid origin format - too short",
			modifyCriteria: func(c *SearchCriteria) {
				c.Origin = "CG"
			},
			expectError:   true,
			errorContains: "3-letter IATA code",
		},
		{
			name: "invalid origin format - too long",
			modifyCriteria: func(c *SearchCriteria) {
				c.Origin = "CGKK"
			},
			expectError:   true,
			errorContains: "3-letter IATA code",
		},
		{
			name: "empty destination",
			modifyCriteria: func(c *SearchCriteria) {
				c.Destination = ""
			},
			expectError:   true,
			errorContains: "destination is required",
		},
		{
			name: "invalid destination format",
			modifyCriteria: func(c *SearchCriteria) {
				c.Destination = "dps"
			},
			expectError:   true,
			errorContains: "3-letter IATA code",
		},
		{
			name: "same origin and destination",
			modifyCriteria: func(c *SearchCriteria) {
				c.Origin = "CGK"
				c.Destination = "CGK"
			},
			expectError:   true,
			errorContains: "must be different",
		},
		{
			name: "empty departure date",
			modifyCriteria: func(c *SearchCriteria) {
				c.DepartureDate = ""
			},
			expectError:   true,
			errorContains: "departureDate is required",
		},
		{
			name: "invalid date format - wrong separator",
			modifyCriteria: func(c *SearchCriteria) {
				c.DepartureDate = "2025/12/15"
			},
			expectError:   true,
			errorContains: "YYYY-MM-DD format",
		},
		{
			name: "invalid date format - invalid month",
			modifyCriteria: func(c *SearchCriteria) {
				c.DepartureDate = "2025-13-15"
			},
			expectError:   true,
			errorContains: "not a valid date",
		},
		{
			name: "invalid date format - invalid day",
			modifyCriteria: func(c *SearchCriteria) {
				c.DepartureDate = "2025-02-30"
			},
			expectError:   true,
			errorContains: "not a valid date",
		},
		{
			name: "zero passengers",
			modifyCriteria: func(c *SearchCriteria) {
				c.Passengers = 0
			},
			expectError:   true,
			errorContains: "at least 1",
		},
		{
			name: "negative passengers",
			modifyCriteria: func(c *SearchCriteria) {
				c.Passengers = -1
			},
			expectError:   true,
			errorContains: "at least 1",
		},
		{
			name: "invalid class",
			modifyCriteria: func(c *SearchCriteria) {
				c.Class = "premium"
			},
			expectError:   true,
			errorContains: "economy, business, first",
		},
		{
			name: "empty class is valid",
			modifyCriteria: func(c *SearchCriteria) {
				c.Class = ""
			},
			expectError: false,
		},
		{
			name: "business class is valid",
			modifyCriteria: func(c *SearchCriteria) {
				c.Class = "business"
			},
			expectError: false,
		},
		{
			name: "first class is valid",
			modifyCriteria: func(c *SearchCriteria) {
				c.Class = "first"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := validCriteria
			if tt.modifyCriteria != nil {
				tt.modifyCriteria(&criteria)
			}

			err := criteria.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidRequest)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchCriteriaSetDefaults(t *testing.T) {
	tests := []struct {
		name               string
		criteria           SearchCriteria
		expectedPassengers int
		expectedClass      string
	}{
		{
			name: "sets default passengers",
			criteria: SearchCriteria{
				Passengers: 0,
				Class:      "business",
			},
			expectedPassengers: 1,
			expectedClass:      "business",
		},
		{
			name: "sets default class",
			criteria: SearchCriteria{
				Passengers: 2,
				Class:      "",
			},
			expectedPassengers: 2,
			expectedClass:      "economy",
		},
		{
			name: "sets both defaults",
			criteria: SearchCriteria{
				Passengers: 0,
				Class:      "",
			},
			expectedPassengers: 1,
			expectedClass:      "economy",
		},
		{
			name: "preserves existing values",
			criteria: SearchCriteria{
				Passengers: 3,
				Class:      "first",
			},
			expectedPassengers: 3,
			expectedClass:      "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := tt.criteria
			criteria.SetDefaults()
			assert.Equal(t, tt.expectedPassengers, criteria.Passengers)
			assert.Equal(t, tt.expectedClass, criteria.Class)
		})
	}
}
