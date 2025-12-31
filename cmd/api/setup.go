package main

import (
	"github.com/herdiagusthio/flight-search-system/internal/api"
	"github.com/herdiagusthio/flight-search-system/internal/config"
	"github.com/labstack/echo/v4"
)

// SetupLogger configures the global logger based on the provided configuration
func SetupLogger(cfg *config.Config) {
	api.SetupLogger(cfg)
}

// SetupMiddleware configures the middleware stack for the Echo instance
func SetupMiddleware(e *echo.Echo) {
	api.SetupMiddleware(e)
}

// SetupRouter configures all routes and route-specific middleware
func SetupRouter(e *echo.Echo, cfg *config.Config) {
	api.SetupRouter(e, cfg)
}
