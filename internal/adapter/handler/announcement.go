package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type AnnouncementHandler struct {
	announcementUC *usecase.AnnouncementUseCase
}

func NewAnnouncementHandler(announcementUC *usecase.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUC: announcementUC}
}

func (h *AnnouncementHandler) Create(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)

	var req struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		IsPinned     bool   `json:"is_pinned"`
		IsSystemWide bool   `json:"is_system_wide"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	announcement, err := h.announcementUC.Create(roomID, userID, req.Title, req.Content, req.IsPinned, req.IsSystemWide)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Created(c, announcement)
}

func (h *AnnouncementHandler) List(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	announcements, _, err := h.announcementUC.ListByRoom(roomID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, announcements)
}

func (h *AnnouncementHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	role := getUserRole(c)
	if err := h.announcementUC.Update(id, userID, role, req.Title, req.Content); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "اعلان بروزرسانی شد"})
}

func (h *AnnouncementHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.announcementUC.Delete(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "اعلان حذف شد"})
}

func (h *AnnouncementHandler) Pin(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.announcementUC.TogglePin(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "سنجاق تغییر کرد"})
}
