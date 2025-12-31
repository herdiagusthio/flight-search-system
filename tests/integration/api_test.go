package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/internal/api"
	"github.com/herdiagusthio/flight-search-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() *echo.Echo {
	e := echo.New()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		Timeouts: config.TimeoutConfig{
			GlobalSearch: 5 * time.Second,
			Provider:     2 * time.Second,
		},
	}

	api.SetupMiddleware(e)
	api.SetupRouter(e, cfg)

	return e
}

func TestFlightSearchEndpoint(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus []int // Allow multiple valid status codes
		validateBody   func(t *testing.T, status int, body map[string]interface{})
	}{
		{
			name: "valid search request",
			requestBody: `{
				"origin": "CGK",
				"destination": "DPS",
				"departureDate": "2025-12-20",
				"passengers": 1,
				"class": "economy"
			}`,
			expectedStatus: []int{http.StatusOK, http.StatusServiceUnavailable},
			validateBody: func(t *testing.T, status int, body map[string]interface{}) {
				if status == http.StatusOK {
					// Verify success response structure
					assert.Contains(t, body, "search_criteria")
					assert.Contains(t, body, "metadata")
					assert.Contains(t, body, "flights")

					metadata := body["metadata"].(map[string]interface{})
					assert.Contains(t, metadata, "total_results")
					assert.Contains(t, metadata, "providers_queried")
					assert.Contains(t, metadata, "providers_succeeded")
					assert.Contains(t, metadata, "search_time_ms")
				} else {
					// Verify error response structure
					assert.Equal(t, "service_unavailable", body["code"])
					assert.NotEmpty(t, body["message"])
				}
			},
		},
		{
			name: "invalid origin airport code",
			requestBody: `{
				"origin": "INVALID",
				"destination": "DPS",
				"departureDate": "2025-12-20",
				"passengers": 1,
				"class": "economy"
			}`,
			expectedStatus: []int{http.StatusBadRequest},
			validateBody: func(t *testing.T, status int, body map[string]interface{}) {
				assert.Equal(t, "validation_error", body["code"])
				assert.NotEmpty(t, body["message"])
			},
		},
		{
			name: "missing required field",
			requestBody: `{
				"origin": "CGK",
				"destination": "DPS"
			}`,
			expectedStatus: []int{http.StatusBadRequest},
			validateBody: func(t *testing.T, status int, body map[string]interface{}) {
				assert.Equal(t, "validation_error", body["code"])
				assert.NotEmpty(t, body["message"])
			},
		},
		{
			name:           "malformed JSON",
			requestBody:    `{"origin": "CGK"`,
			expectedStatus: []int{http.StatusBadRequest},
			validateBody: func(t *testing.T, status int, body map[string]interface{}) {
				assert.Contains(t, body, "code")
				assert.Contains(t, body, "message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/search", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Verify status code is one of the expected values
			assert.Contains(t, tt.expectedStatus, rec.Code,
				"Expected one of %v, got %d", tt.expectedStatus, rec.Code)

			// Parse and validate response body
			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.validateBody != nil {
				tt.validateBody(t, rec.Code, response)
			}
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name           string
		method         string
		path           string
		origin         string
		expectedHeader string
	}{
		{
			name:           "OPTIONS preflight request",
			method:         http.MethodOptions,
			path:           "/api/v1/flights/search",
			origin:         "http://localhost:3000",
			expectedHeader: echo.HeaderAccessControlAllowOrigin,
		},
		{
			name:           "POST request with origin",
			method:         http.MethodPost,
			path:           "/api/v1/flights/search",
			origin:         "http://localhost:3000",
			expectedHeader: echo.HeaderAccessControlAllowOrigin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				reqBody := `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-20","passengers":1}`
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(reqBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req.Header.Set(echo.HeaderOrigin, tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.NotEmpty(t, rec.Header().Get(tt.expectedHeader))
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "health check endpoint",
			method: http.MethodGet,
			path:   "/health",
			body:   "",
		},
		{
			name:   "flight search endpoint",
			method: http.MethodPost,
			path:   "/api/v1/flights/search",
			body:   `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-20","passengers":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			requestID := rec.Header().Get(echo.HeaderXRequestID)
			assert.NotEmpty(t, requestID)
			assert.Greater(t, len(requestID), 0)
		})
	}
}

func TestHealthCheckEndpoint(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "GET health returns healthy",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "healthy"},
		},
		{
			name:           "POST method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != nil {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}
