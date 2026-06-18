package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// AnnouncementHandler handles HTTP requests for class announcements.
// Routes: POST /classes/:id/announcements, GET /classes/:id/announcements
//         PUT /announcements/:id, DELETE /announcements/:id, POST /announcements/:id/pin
type AnnouncementHandler struct {
	announcementUC *usecase.AnnouncementUseCase
}

func NewAnnouncementHandler(announcementUC *usecase.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUC: announcementUC}
}

func (h *AnnouncementHandler) Create(c echo.Context) error {
	classID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
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

	announcement, err := h.announcementUC.Create(classID, userID, req.Title, req.Content, req.IsPinned, req.IsSystemWide)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, announcement)
}

func (h *AnnouncementHandler) List(c echo.Context) error {
	classID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	announcements, _, err := h.announcementUC.ListByClass(classID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, announcements)
}

func (h *AnnouncementHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.announcementUC.Update(id, req.Title, req.Content); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "اعلان بروزرسانی شد"})
}

func (h *AnnouncementHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.announcementUC.Delete(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "اعلان حذف شد"})
}

func (h *AnnouncementHandler) Pin(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.announcementUC.TogglePin(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "سنجاق تغییر کرد"})
}
