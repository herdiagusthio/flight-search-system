package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, &HealthResponse{
		Status: "healthy",
	})
}

func SearchFlights(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusOK, result)
}
