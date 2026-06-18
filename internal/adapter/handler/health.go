package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
	dbPath    string
}

func NewHealthHandler(db *sql.DB, dbPath string) *HealthHandler {
	return &HealthHandler{db: db, startTime: time.Now(), dbPath: dbPath}
}

func (h *HealthHandler) Health(c echo.Context) error {
	uptime := time.Since(h.startTime)
	dbSize := "unknown"
	if info, err := os.Stat(h.dbPath); err == nil {
		dbSize = formatBytes(info.Size())
	}

	var activeRooms int64
	h.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status = 'live'`).Scan(&activeRooms)
	var totalUsers int64
	h.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"uptime":      formatUptime(uptime),
		"db_size":     dbSize,
		"webrtc_status": "pion_builtin",
		"active_rooms": activeRooms,
		"total_users": totalUsers,
	})
}

func formatUptime(d time.Duration) string {
	totalMinutes := int64(d.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatBytes(bytes int64) string {
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.0f MB", float64(bytes)/(1024*1024))
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%.0f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%d B", bytes)
}
