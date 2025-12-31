package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status" example:"healthy"` // Service health status
}

// HealthCheck returns the health status of the service.
// @Summary		Health check
// @Description	Check if the API service is running and healthy
// @Tags		health
// @Accept		json
// @Produce		json
// @Success		200	{object}	HealthResponse	"Service is healthy"
// @Router		/health [get]
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, &HealthResponse{
		Status: "healthy",
	})
}

func SearchFlights(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusOK, result)
}
