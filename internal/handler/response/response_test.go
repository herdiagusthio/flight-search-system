package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(c echo.Context) error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "BadRequest with custom message",
			handler: func(c echo.Context) error {
				return BadRequest(c, "custom bad request message")
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeInvalidRequest,
			expectedMsg:    "custom bad request message",
		},
		{
			name: "InvalidRequest",
			handler: func(c echo.Context) error {
				return InvalidRequest(c)
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeInvalidRequest,
			expectedMsg:    MsgInvalidRequestBody,
		},
		{
			name: "ValidationError",
			handler: func(c echo.Context) error {
				return ValidationError(c)
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeValidationError,
			expectedMsg:    MsgValidationFailed,
		},
		{
			name: "ValidationErrorWithMessage",
			handler: func(c echo.Context) error {
				return ValidationErrorWithMessage(c, "field 'name' is required")
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeValidationError,
			expectedMsg:    "field 'name' is required",
		},
		{
			name: "ServiceUnavailable",
			handler: func(c echo.Context) error {
				return ServiceUnavailable(c)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   CodeServiceUnavailable,
			expectedMsg:    MsgServiceUnavailable,
		},
		{
			name: "GatewayTimeout",
			handler: func(c echo.Context) error {
				return GatewayTimeout(c)
			},
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   CodeTimeout,
			expectedMsg:    MsgTimeout,
		},
		{
			name: "RequestCancelled",
			handler: func(c echo.Context) error {
				return RequestCancelled(c)
			},
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   CodeTimeout,
			expectedMsg:    MsgRequestCancelled,
		},
		{
			name: "InternalError",
			handler: func(c echo.Context) error {
				return InternalError(c)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   CodeInternalError,
			expectedMsg:    MsgInternalError,
		},
		{
			name: "InternalServerErrorWithMessage",
			handler: func(c echo.Context) error {
				return InternalServerErrorWithMessage(c, "database connection failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   CodeInternalError,
			expectedMsg:    "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := tt.handler(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var errDetail ErrorDetail
			err = json.Unmarshal(rec.Body.Bytes(), &errDetail)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, errDetail.Code)
			assert.Equal(t, tt.expectedMsg, errDetail.Message)
		})
	}
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns healthy status",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"healthy"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := HealthCheck(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestSearchFlights(t *testing.T) {
	type FlightResult struct {
		Flights []struct {
			ID    string `json:"id"`
			Price int    `json:"price"`
		} `json:"flights"`
	}

	tests := []struct {
		name           string
		result         interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "returns flight results",
			result: map[string]interface{}{
				"flights": []map[string]interface{}{
					{"id": "FL001", "price": 250},
					{"id": "FL002", "price": 300},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"flights":[{"id":"FL001","price":250},{"id":"FL002","price":300}]}`,
		},
		{
			name: "returns empty flight results",
			result: map[string]interface{}{
				"flights": []map[string]interface{}{},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"flights":[]}`,
		},
		{
			name:           "returns nil result",
			result:         nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/flights", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := SearchFlights(c, tt.result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestErrorDetailSerialization(t *testing.T) {
	tests := []struct {
		name         string
		errorDetail  ErrorDetail
		expectedJSON string
	}{
		{
			name: "basic error detail",
			errorDetail: ErrorDetail{
				Code:    CodeInvalidRequest,
				Message: "test message",
			},
			expectedJSON: `{"code":"invalid_request","message":"test message"}`,
		},
		{
			name: "error detail with details",
			errorDetail: ErrorDetail{
				Code:    CodeValidationError,
				Message: "validation failed",
				Details: map[string]string{
					"field": "email",
					"error": "invalid format",
				},
			},
			expectedJSON: `{"code":"validation_error","message":"validation failed","details":{"field":"email","error":"invalid format"}}`,
		},
		{
			name: "error detail with empty details omits field",
			errorDetail: ErrorDetail{
				Code:    CodeInternalError,
				Message: "internal error",
				Details: nil,
			},
			expectedJSON: `{"code":"internal_error","message":"internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.errorDetail)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(data))
		})
	}
}

func TestConstants(t *testing.T) {
	// Test error codes
	codeTests := []struct {
		name     string
		constant string
		expected string
	}{
		{"CodeInvalidRequest", CodeInvalidRequest, "invalid_request"},
		{"CodeValidationError", CodeValidationError, "validation_error"},
		{"CodeServiceUnavailable", CodeServiceUnavailable, "service_unavailable"},
		{"CodeTimeout", CodeTimeout, "timeout"},
		{"CodeInternalError", CodeInternalError, "internal_error"},
	}

	for _, tt := range codeTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}

	// Test error messages
	msgTests := []struct {
		name     string
		constant string
		expected string
	}{
		{"MsgInvalidRequestBody", MsgInvalidRequestBody, "Failed to parse request body"},
		{"MsgValidationFailed", MsgValidationFailed, "Request validation failed"},
		{"MsgServiceUnavailable", MsgServiceUnavailable, "All flight providers are currently unavailable"},
		{"MsgTimeout", MsgTimeout, "Request timed out"},
		{"MsgRequestCancelled", MsgRequestCancelled, "Request was cancelled"},
		{"MsgInternalError", MsgInternalError, "An unexpected error occurred"},
	}

	for _, tt := range msgTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}
