// Package api Flight Search API
//
//	@title						Flight Search API
//	@version					1.0
//	@description				RESTful API for searching and aggregating flight information from multiple airline providers
//	@description				This API aggregates flight data from Indonesian airlines including Garuda Indonesia, Lion Air, Batik Air, and AirAsia
//
//	@contact.name				Flight Search API Support
//	@contact.email				support@flightsearch.example.com
//
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//
//	@host						localhost:8080
//
//	@schemes					http https
//
//	@tag.name					flights
//	@tag.description			Flight search and information operations
//	@tag.name					health
//	@tag.description			Service health check operations
//
//	@accept						json
//	@produce					json
//
//	@securitydefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				API Key for authentication (optional)
package api

import (
	"os"
	"time"

	_ "github.com/herdiagusthio/flight-search-system/docs" // Import generated Swagger docs
	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/internal/config"
	"github.com/herdiagusthio/flight-search-system/internal/handler/flight"
	"github.com/herdiagusthio/flight-search-system/internal/handler/httputil"
	"github.com/herdiagusthio/flight-search-system/internal/repository/provider/airasia"
	"github.com/herdiagusthio/flight-search-system/internal/repository/provider/batikair"
	"github.com/herdiagusthio/flight-search-system/internal/repository/provider/garuda"
	"github.com/herdiagusthio/flight-search-system/internal/repository/provider/lionair"
	"github.com/herdiagusthio/flight-search-system/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupLogger configures the global logger based on the provided configuration
func SetupLogger(cfg *config.Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if cfg.Logging.Format != "json" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	// Set log level from config
	switch cfg.Logging.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetupMiddleware configures the middleware stack for the Echo instance
func SetupMiddleware(e *echo.Echo) {
	// Recovery middleware to handle panics
	e.Use(middleware.Recover())

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Logger middleware with zerolog integration
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("request_id", v.RequestID).
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Msg("HTTP request")
			return nil
		},
	}))
}

// SetupDependencies initializes all dependencies (providers, usecase, handlers)
func SetupDependencies(cfg *config.Config) *flight.FlightHandler {
	// Initialize provider adapters
	garudaProvider := garuda.NewAdapter("external/response-mock/garuda_indonesia_search_response.json", false)
	lionProvider := lionair.NewAdapter("external/response-mock/lion_air_search_response.json", false)
	batikProvider := batikair.NewAdapter("external/response-mock/batik_air_search_response.json", false)
	airasiaProvider := airasia.NewAdapter("external/response-mock/airasia_search_response.json", false)

	providers := []domain.FlightProvider{
		garudaProvider,
		lionProvider,
		batikProvider,
		airasiaProvider,
	}

	// Initialize usecase with timeout configuration
	usecaseConfig := &usecase.Config{
		GlobalTimeout:   cfg.Timeouts.GlobalSearch,
		ProviderTimeout: cfg.Timeouts.Provider,
	}
	searchUseCase := usecase.NewFlightSearchUseCase(providers, usecaseConfig)

	// Initialize and return flight handler
	return flight.NewFlightHandler(searchUseCase, &log.Logger)
}

// SetupRouter configures all routes and route-specific middleware
func SetupRouter(e *echo.Echo, cfg *config.Config) {
	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	
	// Health check endpoint at root level
	e.GET("/health", func(c echo.Context) error {
		return httputil.HealthCheck(c)
	})

	// Initialize dependencies
	flightHandler := SetupDependencies(cfg)

	// API v1 group with middleware
	v1 := e.Group("/api/v1")
	
	// Configure CORS middleware for API routes
	v1.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"}, // Configure based on environment in production
		AllowMethods:     []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, echo.HeaderXRequestID},
		AllowCredentials: false,
	}))

	// Configure timeout middleware for API routes
	v1.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 5 * time.Second,
	}))

	// Register flight routes
	flights := v1.Group("/flights")
	flights.POST("/search", flightHandler.HandleSearch)
}
