package middleware

import (
	"log/slog"
	"sync"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/adapter/repository/sqlite"
	"github.com/labstack/echo/v4"
)

// auditQueue holds log entries to be written asynchronously
type auditQueue struct {
	mu     sync.Mutex
	ch     chan *entity.ActivityLog
	repo   *repository.ActivityLogRepo
	quit   chan struct{}
}

func newAuditQueue(repo *repository.ActivityLogRepo, bufferSize int) *auditQueue {
	q := &auditQueue{
		ch:   make(chan *entity.ActivityLog, bufferSize),
		repo: repo,
		quit: make(chan struct{}),
	}
	go q.process()
	return q
}

func (q *auditQueue) enqueue(log *entity.ActivityLog) {
	select {
	case q.ch <- log:
	default:
		slog.Warn("audit queue full, dropping log entry")
	}
}

func (q *auditQueue) process() {
	for {
		select {
		case log := <-q.ch:
			if err := q.repo.Create(log); err != nil {
				slog.Error("failed to create audit log", "error", err)
			}
		case <-q.quit:
			return
		}
	}
}

func (q *auditQueue) close() {
	close(q.quit)
}

// Global async audit queue
var asyncAuditQueue *auditQueue

// InitAsyncAudit initializes the async audit logging queue.
// Call this once at startup.
func InitAsyncAudit(logRepo *repository.ActivityLogRepo) {
	asyncAuditQueue = newAuditQueue(logRepo, 1000)
}

// AuditLog logs admin actions asynchronously to the activity_log table.
func AuditLog(logRepo *repository.ActivityLogRepo) echo.MiddlewareFunc {
	if asyncAuditQueue == nil {
		InitAsyncAudit(logRepo)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			// Only log state-changing methods
			method := c.Request().Method
			if method != "POST" && method != "PUT" && method != "DELETE" {
				return err
			}

			// Skip health and auth endpoints
			path := c.Request().URL.Path
			if path == "/api/v1/health" || path == "/api/v1/auth/login" || path == "/api/v1/auth/register" {
				return err
			}

			role, _ := c.Get("role").(string)

			if role != "admin" && role != "owner" {
				return err
			}

			userID, _ := getUserID(c)
			action := mapAction(method, path)

			log := &entity.ActivityLog{
				UserID:     userID,
				Action:     action,
				EntityType: "api",
				Details:    path,
				IPAddress:  c.RealIP(),
			}

			// Enqueue asynchronously (non-blocking)
			asyncAuditQueue.enqueue(log)

			return err
		}
	}
}

// mapAction converts HTTP method + path to a semantic action name.
func mapAction(method, path string) string {
	// Strip /api/v1 prefix
	p := path
	if len(p) > 8 && p[:8] == "/api/v1" {
		p = p[8:]
	}

	switch {
	// Users
	case p == "/admin/users" && method == "POST":
		return "create_user"
	case p == "/admin/users/batch-delete" && method == "POST":
		return "batch_delete_users"
	case len(p) > 13 && p[:13] == "/admin/users/" && method == "PUT":
		return "update_user"
	case len(p) > 13 && p[:13] == "/admin/users/" && method == "DELETE":
		return "delete_user"

	// Rooms
	case p == "/admin/rooms" && method == "POST":
		return "create_room"
	case len(p) > 12 && p[:12] == "/admin/rooms/" && method == "PUT":
		return "update_room"
	case len(p) > 12 && p[:12] == "/admin/rooms/" && method == "DELETE":
		return "delete_room"
	case len(p) > 12 && p[:12] == "/rooms/" && method == "POST":
		return "add_room_user"
	case len(p) > 12 && p[:12] == "/rooms/" && method == "DELETE":
		return "remove_room_user"

	// Classes
	case p == "/classes" && method == "POST":
		return "create_class"
	case len(p) > 9 && p[:9] == "/classes/" && method == "PUT":
		return "update_class"
	case len(p) > 9 && p[:9] == "/classes/" && method == "DELETE":
		return "delete_class"

	// Sessions
	case p == "/sessions" && method == "POST":
		return "create_session"
	case len(p) > 12 && p[:12] == "/sessions/" && method == "POST":
		return "session_action"
	case len(p) > 12 && p[:12] == "/sessions/" && method == "DELETE":
		return "delete_session"

	// Settings
	case p == "/admin/settings" && method == "PUT":
		return "update_settings"

	// Webhooks
	case p == "/admin/webhooks" && method == "POST":
		return "create_webhook"
	case len(p) > 16 && p[:16] == "/admin/webhooks/" && method == "PUT":
		return "update_webhook"
	case len(p) > 16 && p[:16] == "/admin/webhooks/" && method == "DELETE":
		return "delete_webhook"

	// Files
	case len(p) > 10 && p[:10] == "/sessions/" && method == "POST":
		return "upload_file"
	case len(p) > 7 && p[:7] == "/files/" && method == "DELETE":
		return "delete_file"

	// Recordings
	case len(p) > 10 && p[:10] == "/sessions/" && method == "POST":
		return "upload_recording"

	default:
		return method
	}
}

func getUserID(c echo.Context) (int64, bool) {
	id, ok := c.Get("user_id").(int64)
	return id, ok
}
