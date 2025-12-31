package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/herdiagusthio/flight-search-system/internal/api"
	"github.com/herdiagusthio/flight-search-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

const (
	gracefullShutdownTimeout = 10 * time.Second
)

func main() {
	// Load config
	cfg := config.MustLoadConfig()

	// Setup logger
	api.SetupLogger(cfg)

	log.Info().
		Str("env", cfg.App.Env).
		Int("port", cfg.Server.Port).
		Msg("Configuration loaded")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Setup timeout
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Setup middleware
	api.SetupMiddleware(e)

	// Setup router
	api.SetupRouter(e, cfg)

	// Start server with gracefull shutdown
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	go func() {
		log.Info().Str("address", addr).Msg("Starting server...")
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for shutdown signal
	gracefullShutdown(e)
}

// gracefullShutdown handles graceful shutdown of the server
func gracefullShutdown(e *echo.Echo) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), gracefullShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
	}
	log.Info().Msg("Server stopped")
}
