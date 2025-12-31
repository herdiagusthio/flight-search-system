package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeInvalidRequest,
		Message: message,
	})
}

func InvalidRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeInvalidRequest,
		Message: MsgInvalidRequestBody,
	})
}

func ValidationError(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeValidationError,
		Message: MsgValidationFailed,
	})
}

func ValidationErrorWithMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeValidationError,
		Message: message,
	})
}

func ServiceUnavailable(c echo.Context) error {
	return c.JSON(http.StatusServiceUnavailable, &ErrorDetail{
		Code:    CodeServiceUnavailable,
		Message: MsgServiceUnavailable,
	})
}

func GatewayTimeout(c echo.Context) error {
	return c.JSON(http.StatusGatewayTimeout, &ErrorDetail{
		Code:    CodeTimeout,
		Message: MsgTimeout,
	})
}

func RequestCancelled(c echo.Context) error {
	return c.JSON(http.StatusGatewayTimeout, &ErrorDetail{
		Code:    CodeTimeout,
		Message: MsgRequestCancelled,
	})
}

func InternalError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, &ErrorDetail{
		Code:    CodeInternalError,
		Message: MsgInternalError,
	})
}

func InternalServerErrorWithMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, &ErrorDetail{
		Code:    CodeInternalError,
		Message: message,
	})
}
