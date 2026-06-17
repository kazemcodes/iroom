package middleware

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type maintenanceCache struct {
	mu        sync.RWMutex
	enabled   bool
	expiresAt time.Time
}

func MaintenanceMode(db *sql.DB, jwtSecret string) echo.MiddlewareFunc {
	cache := &maintenanceCache{}
	ttl := 30 * time.Second

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

			if role == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && parts[0] == "Bearer" {
					if claims, err := jwt.Validate(jwtSecret, parts[1]); err == nil && claims.Role == "admin" {
						return next(c)
					}
				}
			}

			cache.mu.RLock()
			if time.Now().Before(cache.expiresAt) {
				enabled := cache.enabled
				cache.mu.RUnlock()
				if enabled {
					return response.Error(c, 503, "سیستم در حال تعمیر و نگهداری است")
				}
				return next(c)
			}
			cache.mu.RUnlock()

			var value string
			err := db.QueryRow(`SELECT value FROM settings WHERE key = 'maintenance_mode'`).Scan(&value)
			isEnabled := err == nil && value == "true"

			cache.mu.Lock()
			cache.enabled = isEnabled
			cache.expiresAt = time.Now().Add(ttl)
			cache.mu.Unlock()

			if isEnabled {
				return response.Error(c, 503, "سیستم در حال تعمیر و نگهداری است")
			}

			return next(c)
		}
	}
}
