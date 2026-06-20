package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/errors"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type SessionHandler struct {
	sessionUC *usecase.SessionUseCase
}

func NewSessionHandler(sessionUC *usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{sessionUC: sessionUC}
}

func (h *SessionHandler) Create(c echo.Context) error {
	var req struct {
		RoomID      int64  `json:"room_id"`
		ClassID     int64  `json:"class_id"`
		Title       string `json:"title"`
		ScheduledAt string `json:"scheduled_at"`
		Duration    int    `json:"duration"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	roomID := req.RoomID
	if roomID == 0 {
		roomID = req.ClassID
	}

	session, err := h.sessionUC.Create(roomID, req.Title, req.ScheduledAt, req.Duration)
	if err != nil {
		if err == errors.ErrNotFound {
			return response.NotFound(c, "اتاق یافت نشد")
		}
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, session)
}

func (h *SessionHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	session, err := h.sessionUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	return response.Success(c, session)
}

func (h *SessionHandler) List(c echo.Context) error {
	roomID, _ := strconv.ParseInt(c.QueryParam("room_id"), 10, 64)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	sessions, total, err := h.sessionUC.List(roomID, page, perPage, c.QueryParam("search"))
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items":       sessions,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *SessionHandler) Start(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)

	session, err := h.sessionUC.Start(id, userID, role)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, session)
}

func (h *SessionHandler) End(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)

	if err := h.sessionUC.End(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "جلسه پایان یافت"})
}

func (h *SessionHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)

	if err := h.sessionUC.Delete(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "جلسه حذف شد"})
}

func (h *SessionHandler) GetPublicInfo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	session, err := h.sessionUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"id":     session.ID,
		"title":  session.Title,
		"status": session.Status,
	})
}
