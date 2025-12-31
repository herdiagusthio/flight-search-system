package flight

import (
	"context"
	"errors"
	"time"

	"github.com/herdiagusthio/flight-search-system/domain"
	"github.com/herdiagusthio/flight-search-system/internal/handler/httputil"
	"github.com/herdiagusthio/flight-search-system/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// FlightHandler handles HTTP requests for flight search operations.
type FlightHandler struct {
	searchUseCase usecase.FlightSearchUseCase
	logger        *zerolog.Logger
}

// NewFlightHandler creates a new FlightHandler instance.
func NewFlightHandler(searchUseCase usecase.FlightSearchUseCase, logger *zerolog.Logger) *FlightHandler {
	return &FlightHandler{
		searchUseCase: searchUseCase,
		logger:        logger,
	}
}

// HandleSearch processes flight search requests.
// @Summary		Search for flights
// @Description	Search for available flights from multiple airline providers based on search criteria
// @Description	This endpoint aggregates flight data from Garuda Indonesia, Lion Air, Batik Air, and AirAsia
// @Tags		flights
// @Accept		json
// @Produce		json
// @Param		request	body		SearchRequest	true	"Flight search parameters"
// @Success		200		{object}	SearchResponse	"Successful flight search with results"
// @Failure		400		{object}	httputil.ErrorDetail	"Invalid request body or validation error"
// @Failure		504		{object}	httputil.ErrorDetail	"Gateway timeout - search took too long"
// @Failure		503		{object}	httputil.ErrorDetail	"Service unavailable - all providers failed"
// @Failure		500		{object}	httputil.ErrorDetail	"Internal server error"
// @Router		/api/v1/flights/search [post]
// It parses the request, validates input, calls the use case, and returns the response.
func (h *FlightHandler) HandleSearch(c echo.Context) error {
	start := time.Now()
	ctx := c.Request().Context()

	// Parse request body
	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Str("method", "HandleSearch").
			Msg("Failed to parse request body")
		return httputil.InvalidRequest(c)
	}

	// Normalize request (uppercase airport codes, lowercase options)
	req.Normalize()

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn().
			Err(err).
			Str("method", "HandleSearch").
			Interface("request", req).
			Msg("Request validation failed")
		return httputil.ValidationErrorWithMessage(c, err.Error())
	}

	// Convert DTO to domain models
	criteria := ToSearchCriteria(req)
	options := ToSearchOptions(req)

	// Log search request
	h.logger.Info().
		Str("method", "HandleSearch").
		Str("origin", criteria.Origin).
		Str("destination", criteria.Destination).
		Str("date", criteria.DepartureDate).
		Int("passengers", criteria.Passengers).
		Msg("Processing flight search request")

	// Execute search
	result, err := h.searchUseCase.Search(ctx, criteria, options)
	if err != nil {
		return h.handleError(c, err, start)
	}

	// Calculate total processing time
	processingTime := time.Since(start).Milliseconds()

	// Update metadata with processing time
	metadata := Metadata{
		TotalResults:       result.Metadata.TotalResults,
		ProvidersQueried:   result.Metadata.ProvidersQueried,
		ProvidersSucceeded: result.Metadata.ProvidersSucceeded,
		ProvidersFailed:    result.Metadata.ProvidersFailed,
		SearchTimeMs:       processingTime,
		CacheHit:           result.Metadata.CacheHit,
	}

	// Build response
	respDTO := NewSearchResponse(criteria, result.Flights, metadata)

	h.logger.Info().
		Str("method", "HandleSearch").
		Int("total_results", metadata.TotalResults).
		Int("providers_succeeded", metadata.ProvidersSucceeded).
		Int64("processing_time_ms", processingTime).
		Msg("Flight search completed successfully")

	return httputil.SearchFlights(c, respDTO)
}

// handleError processes errors from the use case and returns appropriate HTTP responses.
func (h *FlightHandler) handleError(c echo.Context, err error, start time.Time) error {
	processingTime := time.Since(start).Milliseconds()

	// Check for specific domain errors
	if errors.Is(err, domain.ErrInvalidRequest) {
		h.logger.Warn().
			Err(err).
			Str("method", "HandleSearch").
			Int64("processing_time_ms", processingTime).
			Msg("Invalid request from domain layer")
		return httputil.BadRequest(c, err.Error())
	}

	if errors.Is(err, context.DeadlineExceeded) {
		h.logger.Error().
			Err(err).
			Str("method", "HandleSearch").
			Int64("processing_time_ms", processingTime).
			Msg("Search timeout")
		return httputil.GatewayTimeout(c)
	}

	if errors.Is(err, domain.ErrProviderUnavailable) || errors.Is(err, domain.ErrAllProvidersFailed) {
		h.logger.Error().
			Err(err).
			Str("method", "HandleSearch").
			Int64("processing_time_ms", processingTime).
			Msg("All providers unavailable")
		return httputil.ServiceUnavailable(c)
	}

	// Generic error
	h.logger.Error().
		Err(err).
		Str("method", "HandleSearch").
		Int64("processing_time_ms", processingTime).
		Msg("Unexpected error during search")
	return httputil.InternalError(c)
}
