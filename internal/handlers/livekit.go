package handlers

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type LiveKitHandler struct {
	sessionRepo *repository.SessionRepo
	livekitSvc  *services.LiveKitService
}

func NewLiveKitHandler(sessionRepo *repository.SessionRepo, livekitSvc *services.LiveKitService) *LiveKitHandler {
	return &LiveKitHandler{
		sessionRepo: sessionRepo,
		livekitSvc:  livekitSvc,
	}
}

func (h *LiveKitHandler) GetJoinToken(c echo.Context) error {
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

	userID := c.Get("user_id").(int64)
	role := c.Get("role").(string)
	displayName := "کاربر"
	if name, ok := c.Get("display_name").(string); ok && name != "" {
		displayName = name
	}

	roomName := session.LivekitRoom
	if roomName == "" {
		roomName = "room-" + strconv.FormatInt(session.ID, 10)
	}

	identity := "user-" + strconv.FormatInt(userID, 10)

	token, err := h.livekitSvc.GenerateToken(roomName, identity, displayName, role)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Success(c, map[string]interface{}{
		"token": token,
		"url":   h.livekitSvc.GetURL(),
		"room":  roomName,
	})
}

func (h *LiveKitHandler) Webhook(c echo.Context) error {
	// LiveKit webhook handler for recording events
	return response.Success(c, map[string]string{"status": "ok"})
}
