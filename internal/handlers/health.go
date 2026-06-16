package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

// HealthHandler provides an enhanced health endpoint with system metrics.
type HealthHandler struct {
	db         *sql.DB
	startTime  time.Time
	dbPath     string
	livekitSvc *services.LiveKitService
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB, dbPath string, livekitSvc *services.LiveKitService) *HealthHandler {
	return &HealthHandler{
		db:         db,
		startTime:  time.Now(),
		dbPath:     dbPath,
		livekitSvc: livekitSvc,
	}
}

// HealthResponse represents the JSON response from the health endpoint.
type HealthResponse struct {
	Status        string `json:"status"`
	Uptime        string `json:"uptime"`
	DBSize        string `json:"db_size"`
	LiveKitStatus string `json:"livekit_status"`
	ActiveRooms   int64  `json:"active_rooms"`
	TotalUsers    int64  `json:"total_users"`
	TotalSessions int64  `json:"total_sessions"`
	TotalClasses  int64  `json:"total_classes"`
}

// Health returns detailed health and metrics information about the server.
func (h *HealthHandler) Health(c echo.Context) error {
	uptime := time.Since(h.startTime)

	dbSize, err := h.getDBSize()
	if err != nil {
		dbSize = "unknown"
	}

	lkStatus := h.checkLiveKit()

	activeRooms, _ := h.countActiveRooms()
	totalUsers, _ := h.countUsers()
	totalSessions, _ := h.countSessions()
	totalClasses, _ := h.countClasses()

	return c.JSON(http.StatusOK, HealthResponse{
		Status:        "ok",
		Uptime:        formatUptime(uptime),
		DBSize:        dbSize,
		LiveKitStatus: lkStatus,
		ActiveRooms:   activeRooms,
		TotalUsers:    totalUsers,
		TotalSessions: totalSessions,
		TotalClasses:  totalClasses,
	})
}

func (h *HealthHandler) getDBSize() (string, error) {
	info, err := os.Stat(h.dbPath)
	if err != nil {
		return "", err
	}
	return formatBytes(info.Size()), nil
}

func (h *HealthHandler) checkLiveKit() string {
	url := h.livekitSvc.GetURL()
	if url == "" {
		return "not_configured"
	}

	// Attempt a simple HTTP GET to the LiveKit server.
	// LiveKit serves a /health endpoint or at least responds on HTTP.
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://" + url + "/")
	if err != nil {
		// Try with https
		resp, err = client.Get("https://" + url + "/")
		if err != nil {
			return "disconnected"
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 500 {
		return "connected"
	}
	return "degraded"
}

func (h *HealthHandler) countActiveRooms() (int64, error) {
	var count int64
	err := h.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status = 'live'`).Scan(&count)
	return count, err
}

func (h *HealthHandler) countUsers() (int64, error) {
	var count int64
	err := h.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (h *HealthHandler) countSessions() (int64, error) {
	var count int64
	err := h.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count)
	return count, err
}

func (h *HealthHandler) countClasses() (int64, error) {
	var count int64
	err := h.db.QueryRow(`SELECT COUNT(*) FROM classes`).Scan(&count)
	return count, err
}

// formatUptime converts a duration to a human-readable string like "2h 15m" or "1d 2h 15m".
func formatUptime(d time.Duration) string {
	totalMinutes := int64(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatBytes converts bytes to a human-readable string (KB, MB, GB).
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
