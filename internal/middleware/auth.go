package middleware

import (
	"strings"

	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

func Auth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Unauthorized(c, "توکن ارائه نشده است")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Unauthorized(c, "فرمت توکن نامعتبر است")
			}

			claims, err := jwt.Validate(secret, parts[1])
			if err != nil {
				return response.Unauthorized(c, "توکن نامعتبر یا منقضی شده است")
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			return next(c)
		}
	}
}

func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if role != "admin" {
				return response.Forbidden(c, "دسترسی غیرمجاز")
			}
			return next(c)
		}
	}
}

func TeacherOrAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if role != "admin" && role != "teacher" {
				return response.Forbidden(c, "فقط مدیر و مدرس اجازه دسترسی دارند")
			}
			return next(c)
		}
	}
}

func APIKeyAuth(validKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get("X-API-Key")
			if key == "" {
				key = c.QueryParam("api_key")
			}

			if validKey == "" {
				return response.Unauthorized(c, "API key not configured on server")
			}

			if key != validKey {
				return response.Unauthorized(c, "کلید API نامعتبر است")
			}

			c.Set("role", "admin")
			return next(c)
		}
	}
}
