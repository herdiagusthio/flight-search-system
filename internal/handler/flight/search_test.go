package flight

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewFlightHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()

	handler := NewFlightHandler(mockUseCase, &logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUseCase, handler.searchUseCase)
	assert.Equal(t, &logger, handler.logger)
}

func TestHandleSearch_Success_WithResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	// Prepare request
	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1,
		"class": "economy"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Expected domain response
	expectedCriteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		Class:         "economy",
	}

	flights := []domain.Flight{
		{
			ID:           "QZ520",
			FlightNumber: "QZ520",
			Provider:     "AirAsia",
			Airline: domain.AirlineInfo{
				Code: "QZ",
				Name: "AirAsia",
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				AirportName: "Jakarta",
				DateTime:    time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC),
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				AirportName: "Denpasar",
				DateTime:    time.Date(2025, 12, 15, 8, 30, 0, 0, time.UTC),
			},
			Duration: domain.DurationInfo{
				TotalMinutes: 150,
				Formatted:    "2h 30m",
			},
			Price: domain.PriceInfo{
				Amount:   650000,
				Currency: "IDR",
			},
			Baggage: domain.BaggageInfo{
				CabinKg:   7,
				CheckedKg: 20,
			},
			Class: "economy",
			Stops: 0,
		},
	}

	domainResponse := &domain.SearchResponse{
		Flights: flights,
		Metadata: domain.SearchMetadata{
			TotalResults:       1,
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
			ProvidersFailed:    0,
			CacheHit:           false,
		},
	}

	mockUseCase.EXPECT().
		Search(gomock.Any(), expectedCriteria, gomock.Any()).
		Return(domainResponse, nil)

	// Execute
	err := handler.HandleSearch(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response SearchResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "CGK", response.SearchCriteria.Origin)
	assert.Equal(t, "DPS", response.SearchCriteria.Destination)
	assert.Equal(t, 1, response.Metadata.TotalResults)
	assert.Equal(t, 4, response.Metadata.ProvidersQueried)
	assert.Len(t, response.Flights, 1)
}

func TestHandleSearch_Success_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	domainResponse := &domain.SearchResponse{
		Flights: []domain.Flight{},
		Metadata: domain.SearchMetadata{
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
			ProvidersFailed:    0,
			CacheHit:           false,
		},
	}

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domainResponse, nil)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response SearchResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Metadata.TotalResults)
	assert.Empty(t, response.Flights)
}

func TestHandleSearch_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{invalid json`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "invalid_request", response["code"])
}

func TestHandleSearch_ValidationError_MissingOrigin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "validation_error", response["code"])
	assert.Contains(t, response["message"], "origin is required")
}

func TestHandleSearch_ValidationError_InvalidAirportCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CG",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "validation_error", response["code"])
	assert.Contains(t, response["message"], "origin must be a valid 3-letter IATA code")
}

func TestHandleSearch_ValidationError_InvalidDateFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "15-12-2025",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "validation_error", response["code"])
	assert.Contains(t, response["message"], "departureDate must be in YYYY-MM-DD format")
}

func TestHandleSearch_DomainInvalidRequestError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, domain.ErrInvalidRequest)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "invalid_request", response["code"])
}

func TestHandleSearch_TimeoutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, context.DeadlineExceeded)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "timeout", response["code"])
	assert.Contains(t, response["message"], "timed out")
}

func TestHandleSearch_ProviderUnavailableError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, domain.ErrProviderUnavailable)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "service_unavailable", response["code"])
}

func TestHandleSearch_GenericError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("unexpected database error"))

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "internal_error", response["code"])
}

func TestHandleSearch_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	maxPrice := 1000000.0
	maxStops := 1
	reqBody := `{
		"origin": "CGK",
		"destination": "DPS",
		"departureDate": "2025-12-15",
		"passengers": 1,
		"filters": {
			"maxPrice": 1000000,
			"maxStops": 1,
			"airlines": ["GA", "QZ"],
			"departureTimeRange": {
				"start": "06:00",
				"end": "12:00"
			}
		},
		"sortBy": "price"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	domainResponse := &domain.SearchResponse{
		Flights: []domain.Flight{},
		Metadata: domain.SearchMetadata{
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
			ProvidersFailed:    0,
			CacheHit:           false,
		},
	}

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) {
			assert.NotNil(t, opts.Filters)
			assert.Equal(t, &maxPrice, opts.Filters.MaxPrice)
			assert.Equal(t, &maxStops, opts.Filters.MaxStops)
			assert.Equal(t, []string{"GA", "QZ"}, opts.Filters.Airlines)
			assert.NotNil(t, opts.Filters.DepartureTimeRange)
			assert.Equal(t, domain.SortByPrice, opts.SortBy)
		}).
		Return(domainResponse, nil)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleSearch_NormalizeRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := usecase.NewMockFlightSearchUseCase(ctrl)
	logger := zerolog.Nop()
	handler := NewFlightHandler(mockUseCase, &logger)

	// Request with lowercase airport codes
	reqBody := `{
		"origin": "cgk",
		"destination": "dps",
		"departureDate": "2025-12-15",
		"passengers": 1,
		"class": "ECONOMY"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/flights/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	domainResponse := &domain.SearchResponse{
		Flights: []domain.Flight{},
		Metadata: domain.SearchMetadata{
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
			ProvidersFailed:    0,
			CacheHit:           false,
		},
	}

	mockUseCase.EXPECT().
		Search(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) {
			// Should be normalized to uppercase
			assert.Equal(t, "CGK", criteria.Origin)
			assert.Equal(t, "DPS", criteria.Destination)
			assert.Equal(t, "economy", criteria.Class)
		}).
		Return(domainResponse, nil)

	err := handler.HandleSearch(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
