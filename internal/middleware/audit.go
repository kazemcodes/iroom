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

			log := &entity.ActivityLog{
				UserID:     userID,
				Action:     method,
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

func getUserID(c echo.Context) (int64, bool) {
	id, ok := c.Get("user_id").(int64)
	return id, ok
}
