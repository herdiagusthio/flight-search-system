package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/herdiagusthio/flight-search-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		expectedLevel zerolog.Level
	}{
		{
			name: "debug level",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "debug",
					Format: "json",
				},
			},
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name: "info level",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name: "warn level",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "warn",
					Format: "json",
				},
			},
			expectedLevel: zerolog.WarnLevel,
		},
		{
			name: "error level",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "error",
					Format: "json",
				},
			},
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name: "default level (unknown)",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "unknown",
					Format: "json",
				},
			},
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name: "console format",
			cfg: &config.Config{
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "console",
				},
			},
			expectedLevel: zerolog.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupLogger(tt.cfg)
			assert.Equal(t, tt.expectedLevel, zerolog.GlobalLevel())
		})
	}
}

func TestSetupMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "middleware configured correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			setupMiddleware(e)

			// Add a test route to verify middleware chain
			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			// Verify middleware was added by making a request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Should not panic (Recover middleware is working)
			assert.NotPanics(t, func() {
				e.ServeHTTP(rec, req)
			})

			// Verify request completed successfully
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify request ID header is set (RequestID middleware is working)
			assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))
		})
	}
}

func TestSetupRouter(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health check endpoint returns healthy",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"healthy"}`,
		},
		{
			name:           "health check endpoint POST method not allowed",
			method:         http.MethodPost,
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "unknown endpoint returns not found",
			method:         http.MethodGet,
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
			}

			setupRouter(e, cfg)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestSetupRouterWithConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "development config",
			cfg: &config.Config{
				App: config.AppConfig{
					Env: "development",
				},
				Server: config.ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
			},
		},
		{
			name: "production config",
			cfg: &config.Config{
				App: config.AppConfig{
					Env: "production",
				},
				Server: config.ServerConfig{
					Port:         80,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Should not panic
			assert.NotPanics(t, func() {
				setupRouter(e, tt.cfg)
			})

			// Verify health endpoint works
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
