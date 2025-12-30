package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SimpleFlightProviderMock is a simple test implementation of FlightProvider
type SimpleFlightProviderMock struct {
	ProviderName string
	Flights      []Flight
	Err          error
}

func NewSimpleFlightProviderMock(name string, flights []Flight, err error) *SimpleFlightProviderMock {
	return &SimpleFlightProviderMock{
		ProviderName: name,
		Flights:      flights,
		Err:          err,
	}
}

func (m *SimpleFlightProviderMock) Name() string {
	return m.ProviderName
}

func (m *SimpleFlightProviderMock) Search(ctx context.Context, criteria SearchCriteria) ([]Flight, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Flights, nil
}

func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()
	assert.NotNil(t, registry)
	assert.Empty(t, registry.GetAll())
	assert.Empty(t, registry.Names())
}

func TestProviderRegistryRegister(t *testing.T) {
	tests := []struct {
		name          string
		providers     []*SimpleFlightProviderMock
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "register single provider",
			providers:     []*SimpleFlightProviderMock{NewSimpleFlightProviderMock("garuda", nil, nil)},
			expectedCount: 1,
			expectedNames: []string{"garuda"},
		},
		{
			name: "register multiple providers",
			providers: []*SimpleFlightProviderMock{
				NewSimpleFlightProviderMock("garuda", nil, nil),
				NewSimpleFlightProviderMock("airasia", nil, nil),
				NewSimpleFlightProviderMock("lion_air", nil, nil),
			},
			expectedCount: 3,
			expectedNames: []string{"garuda", "airasia", "lion_air"},
		},
		{
			name: "register duplicate replaces",
			providers: []*SimpleFlightProviderMock{
				NewSimpleFlightProviderMock("garuda", []Flight{{ID: "1"}}, nil),
				NewSimpleFlightProviderMock("garuda", []Flight{{ID: "2"}}, nil),
			},
			expectedCount: 1,
			expectedNames: []string{"garuda"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewProviderRegistry()

			for _, p := range tt.providers {
				registry.Register(p)
			}

			assert.Len(t, registry.GetAll(), tt.expectedCount)

			names := registry.Names()
			for _, expectedName := range tt.expectedNames {
				assert.Contains(t, names, expectedName)
			}
		})
	}
}

func TestProviderRegistryRegisterNil(t *testing.T) {
	registry := NewProviderRegistry()
	registry.Register(nil)
	assert.Empty(t, registry.GetAll())
}

func TestProviderRegistryGet(t *testing.T) {
	registry := NewProviderRegistry()
	garuda := NewSimpleFlightProviderMock("garuda", nil, nil)
	airasia := NewSimpleFlightProviderMock("airasia", nil, nil)

	registry.Register(garuda)
	registry.Register(airasia)

	tests := []struct {
		name     string
		lookup   string
		expected FlightProvider
	}{
		{
			name:     "get existing provider",
			lookup:   "garuda",
			expected: garuda,
		},
		{
			name:     "get another existing provider",
			lookup:   "airasia",
			expected: airasia,
		},
		{
			name:     "get non-existing provider",
			lookup:   "lion_air",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.Get(tt.lookup)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Name(), result.Name())
			}
		})
	}
}

func TestProviderRegistryGetAll(t *testing.T) {
	registry := NewProviderRegistry()

	assert.Empty(t, registry.GetAll())

	registry.Register(NewSimpleFlightProviderMock("garuda", nil, nil))
	registry.Register(NewSimpleFlightProviderMock("airasia", nil, nil))

	providers := registry.GetAll()
	assert.Len(t, providers, 2)

	names := make(map[string]bool)
	for _, p := range providers {
		names[p.Name()] = true
	}
	assert.True(t, names["garuda"])
	assert.True(t, names["airasia"])
}

func TestProviderRegistryNames(t *testing.T) {
	registry := NewProviderRegistry()

	assert.Empty(t, registry.Names())

	registry.Register(NewSimpleFlightProviderMock("garuda", nil, nil))
	registry.Register(NewSimpleFlightProviderMock("airasia", nil, nil))
	registry.Register(NewSimpleFlightProviderMock("lion_air", nil, nil))

	names := registry.Names()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "garuda")
	assert.Contains(t, names, "airasia")
	assert.Contains(t, names, "lion_air")
}

func TestSimpleFlightProviderMockSearch(t *testing.T) {
	flights := []Flight{
		{ID: "1", FlightNumber: "GA-123"},
		{ID: "2", FlightNumber: "GA-456"},
	}

	tests := []struct {
		name            string
		provider        *SimpleFlightProviderMock
		expectedFlights int
		expectedError   error
	}{
		{
			name:            "successful search",
			provider:        NewSimpleFlightProviderMock("garuda", flights, nil),
			expectedFlights: 2,
			expectedError:   nil,
		},
		{
			name:            "search with error",
			provider:        NewSimpleFlightProviderMock("airasia", nil, ErrProviderTimeout),
			expectedFlights: 0,
			expectedError:   ErrProviderTimeout,
		},
		{
			name:            "empty results",
			provider:        NewSimpleFlightProviderMock("lion", []Flight{}, nil),
			expectedFlights: 0,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.provider.Search(context.Background(), SearchCriteria{})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedFlights)
			}
		})
	}
}
