package middleware

import (
	"database/sql"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

func MaintenanceMode(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			if path == "/api/v1/health" || path == "/api/v1/auth/login" || path == "/api/v1/auth/register" {
				return next(c)
			}

			role, _ := c.Get("role").(string)
			if role == "admin" {
				return next(c)
			}

			var value string
			err := db.QueryRow(`SELECT value FROM settings WHERE key = 'maintenance_mode'`).Scan(&value)
			if err == nil && value == "true" {
				return response.Error(c, 503, "سیستم در حال تعمیر و نگهداری است")
			}

			return next(c)
		}
	}
}
