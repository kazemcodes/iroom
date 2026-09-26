package handler

import (
	"fmt"
	"os"
	"time"

	"github.com/iroom/iroom/internal/database"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	db        database.DB
	startTime time.Time
	dbPath    string
}

func NewHealthHandler(db database.DB, dbPath string) *HealthHandler {
	return &HealthHandler{db: db, startTime: time.Now(), dbPath: dbPath}
}

func (h *HealthHandler) Health(c echo.Context) error {
	uptime := time.Since(h.startTime)

	// Only SQLite has a file we can measure; external databases report "n/a".
	dbSize := "n/a"
	if h.db.Driver() == database.DriverSQLite && h.dbPath != "" {
		if info, err := statFile(h.dbPath); err == nil {
			dbSize = formatBytes(info)
		}
	}

	var activeRooms int64
	h.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status = 'live'`).Scan(&activeRooms)
	var totalUsers int64
	h.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)

	return response.Success(c, map[string]interface{}{
		"status":        "ok",
		"uptime":        formatUptime(uptime),
		"db_driver":     string(h.db.Driver()),
		"db_size":       dbSize,
		"webrtc_status": "pion_builtin",
		"active_rooms":  activeRooms,
		"total_users":   totalUsers,
	})
}

// statFile returns the size in bytes of the file at path.
func statFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
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
