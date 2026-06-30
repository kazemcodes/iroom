package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c echo.Context, status int, msg string) error {
	return c.JSON(status, APIResponse{
		Success: false,
		Error:   msg,
	})
}

func BadRequest(c echo.Context, msg string) error {
	return Error(c, http.StatusBadRequest, msg)
}

func Unauthorized(c echo.Context, msg string) error {
	return Error(c, http.StatusUnauthorized, msg)
}

func Forbidden(c echo.Context, msg string) error {
	return Error(c, http.StatusForbidden, msg)
}

func NotFound(c echo.Context, msg string) error {
	return Error(c, http.StatusNotFound, msg)
}

func InternalError(c echo.Context, msg string) error {
	return Error(c, http.StatusInternalServerError, msg)
}

func Conflict(c echo.Context, msg string) error {
	return Error(c, http.StatusConflict, msg)
}
