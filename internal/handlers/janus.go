package handlers

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type JanusHandler struct {
	sessionRepo *repository.SessionRepo
	janusSvc    *services.JanusService
}

func NewJanusHandler(sessionRepo *repository.SessionRepo, janusSvc *services.JanusService) *JanusHandler {
	return &JanusHandler{
		sessionRepo: sessionRepo,
		janusSvc:    janusSvc,
	}
}

func (h *JanusHandler) GetJoinInfo(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	session, err := h.sessionRepo.GetByID(sessionID)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if session.Status != "live" {
		return response.BadRequest(c, "جلسه در حال برگزاری نیست")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	return response.Success(c, map[string]interface{}{
		"ws_url":  h.janusSvc.GetWSURL(),
		"room_id": sessionID,
		"user_id": userID,
		"role":    role,
	})
}

func (h *JanusHandler) MuteParticipant(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه جلسه نامعتبر")
	}

	participantID, err := strconv.ParseInt(c.Param("participant_id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه شرکت‌کننده نامعتبر")
	}

	role := getUserRole(c)
	if role != "admin" && role != "teacher" {
		return response.Forbidden(c, "فقط مدیر و مدرس اجازه دسترسی دارند")
	}

	var req struct {
		Audio bool `json:"audio"`
		Video bool `json:"video"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	session, err := h.sessionRepo.GetByID(sessionID)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if session.JanusSessionID == 0 || session.JanusHandleID == 0 {
		return response.BadRequest(c, "اتاق Janus هنوز ایجاد نشده است")
	}

	err = h.janusSvc.MuteParticipant(
		session.JanusSessionID, session.JanusHandleID,
		sessionID, participantID, req.Audio, req.Video,
	)
	if err != nil {
		return response.InternalError(c, "خطا در بی‌صدا کردن شرکت‌کننده")
	}

	return response.Success(c, map[string]string{"message": "عملیات انجام شد"})
}

func (h *JanusHandler) KickParticipant(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه جلسه نامعتبر")
	}

	participantID, err := strconv.ParseInt(c.Param("participant_id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه شرکت‌کننده نامعتبر")
	}

	role := getUserRole(c)
	if role != "admin" && role != "teacher" {
		return response.Forbidden(c, "فقط مدیر و مدرس اجازه دسترسی دارند")
	}

	session, err := h.sessionRepo.GetByID(sessionID)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if session.JanusSessionID == 0 || session.JanusHandleID == 0 {
		return response.BadRequest(c, "اتاق Janus هنوز ایجاد نشده است")
	}

	err = h.janusSvc.KickParticipant(
		session.JanusSessionID, session.JanusHandleID,
		sessionID, participantID,
	)
	if err != nil {
		return response.InternalError(c, "خطا در حذف شرکت‌کننده")
	}

	return response.Success(c, map[string]string{"message": "شرکت‌کننده حذف شد"})
}
