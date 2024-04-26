package helper

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func SuccessHandler(c echo.Context, result interface{}, statusCode ...int) error {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return c.JSON(code, result)
}

func FailedHandler(c echo.Context, errorMsg string, statusCode ...int) error {
	code := http.StatusInternalServerError
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return c.JSON(code, ErrorMessage{Message: errorMsg})
}
