package handlers

import (
	"strconv"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	notificationRepo *repository.NotificationRepo
}

func NewNotificationHandler(notificationRepo *repository.NotificationRepo) *NotificationHandler {
	return &NotificationHandler{notificationRepo: notificationRepo}
}

func (h *NotificationHandler) List(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage
	notifications, err := h.notificationRepo.ListByUser(userID, perPage, offset)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت اعلان‌ها")
	}
	if notifications == nil {
		notifications = []*models.Notification{}
	}

	total, _ := h.notificationRepo.CountByUser(userID)
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return response.Success(c, models.PaginatedResponse{
		Items:      notifications,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

func (h *NotificationHandler) UnreadCount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	count, err := h.notificationRepo.CountUnread(userID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت تعداد اعلان‌ها")
	}

	return response.Success(c, map[string]int64{"unread_count": count})
}

func (h *NotificationHandler) MarkRead(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	if err := h.notificationRepo.MarkRead(id, userID); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی اعلان")
	}

	return response.Success(c, map[string]string{"message": "اعلان خوانده شد"})
}

func (h *NotificationHandler) MarkAllRead(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	if err := h.notificationRepo.MarkAllRead(userID); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی اعلان‌ها")
	}

	return response.Success(c, map[string]string{"message": "همه اعلان‌ها خوانده شدند"})
}
