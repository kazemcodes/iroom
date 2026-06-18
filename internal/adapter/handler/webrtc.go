package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/webrtc"
	"github.com/labstack/echo/v4"
)

type WebRTCHandler struct {
	signaling *webrtc.SignalingServer
}

func NewWebRTCHandler(signaling *webrtc.SignalingServer) *WebRTCHandler {
	return &WebRTCHandler{signaling: signaling}
}

func (h *WebRTCHandler) GetJoinInfo(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	return response.Success(c, map[string]interface{}{
		"room_id": strconv.FormatInt(sessionID, 10),
		"user_id": strconv.FormatInt(userID, 10),
		"role":    role,
	})
}

func (h *WebRTCHandler) HandleOffer(c echo.Context) error {
	return h.signaling.HandleOffer(c)
}

func (h *WebRTCHandler) HandleCandidate(c echo.Context) error {
	return h.signaling.HandleCandidate(c)
}

func (h *WebRTCHandler) HandleLeave(c echo.Context) error {
	return h.signaling.HandleLeave(c)
}

func (h *WebRTCHandler) HandleRoomInfo(c echo.Context) error {
	return h.signaling.HandleRoomInfo(c)
}

func (h *WebRTCHandler) GetParticipants(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return response.BadRequest(c, "شناسه اتاق نامعتبر")
	}

	participants := h.signaling.GetRoomManager().GetRoomParticipants(roomID)
	if participants == nil {
		participants = []webrtc.ParticipantInfo{}
	}

	return response.Success(c, participants)
}

func (h *WebRTCHandler) MuteParticipant(c echo.Context) error {
	roomID := c.Param("id")
	participantID := c.Param("participantId")

	if roomID == "" || participantID == "" {
		return response.BadRequest(c, "پارامترهای نامعتبر")
	}

	role := getUserRole(c)
	if role != "admin" && role != "teacher" {
		return response.Forbidden(c, "فقط مدیر و مدرس اجازه دسترسی دارند")
	}

	participant := h.signaling.GetRoomManager().GetParticipant(roomID, participantID)
	if participant == nil {
		return response.NotFound(c, "شرکت‌کننده یافت نشد")
	}

	participant.IsMuted = !participant.IsMuted

	return response.Success(c, map[string]interface{}{
		"is_muted": participant.IsMuted,
	})
}

func (h *WebRTCHandler) KickParticipant(c echo.Context) error {
	roomID := c.Param("id")
	participantID := c.Param("participantId")

	if roomID == "" || participantID == "" {
		return response.BadRequest(c, "پارامترهای نامعتبر")
	}

	role := getUserRole(c)
	if role != "admin" && role != "teacher" {
		return response.Forbidden(c, "فقط مدیر و مدرس اجازه دسترسی دارند")
	}

	participant := h.signaling.GetRoomManager().GetParticipant(roomID, participantID)
	if participant == nil {
		return response.NotFound(c, "شرکت‌کننده یافت نشد")
	}

	h.signaling.GetRoomManager().RemoveParticipant(roomID, participantID)

	return response.Success(c, map[string]string{"message": "شرکت‌کننده حذف شد"})
}
