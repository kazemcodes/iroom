package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// NotificationHandler handles HTTP requests for user notifications.
// Routes: GET /notifications, GET /notifications/unread-count
//         POST /notifications/:id/read, POST /notifications/read-all
type NotificationHandler struct {
	notificationUC *usecase.NotificationUseCase
}

func NewNotificationHandler(notificationUC *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notificationUC: notificationUC}
}

func (h *NotificationHandler) List(c echo.Context) error {
	userID, _ := getUserID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	notifications, err := h.notificationUC.List(userID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": notifications,
		"total": len(notifications),
	})
}

func (h *NotificationHandler) UnreadCount(c echo.Context) error {
	userID, _ := getUserID(c)
	count, err := h.notificationUC.UnreadCount(userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]int64{"count": count})
}

func (h *NotificationHandler) MarkRead(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := getUserID(c)
	if err := h.notificationUC.MarkRead(id, userID); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "اعلان خوانده شد"})
}

func (h *NotificationHandler) MarkAllRead(c echo.Context) error {
	userID, _ := getUserID(c)
	if err := h.notificationUC.MarkAllRead(userID); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "همه اعلان‌ها خوانده شد"})
}
