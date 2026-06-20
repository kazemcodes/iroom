package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/services"
	"github.com/iroom/iroom/internal/webrtc"
	"github.com/labstack/echo/v4"
)

type WebRTCHandler struct {
	signaling *webrtc.SignalingServer
	hub       *services.Hub
}

func NewWebRTCHandler(signaling *webrtc.SignalingServer, hub *services.Hub) *WebRTCHandler {
	return &WebRTCHandler{signaling: signaling, hub: hub}
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

	resp := map[string]interface{}{
		"room_id": strconv.FormatInt(sessionID, 10),
		"user_id": strconv.FormatInt(userID, 10),
		"role":    role,
	}

	if stunURL := h.signaling.GetSTUNURL(); stunURL != "" {
		resp["stun_url"] = stunURL
	}

	if turnURL := h.signaling.GetTURNURL(); turnURL != "" {
		resp["turn_url"] = turnURL
		if turnSecret := h.signaling.GetTURNSecret(); turnSecret != "" {
			turnUsername, turnCredential := generateTurnCredentials(userID, turnSecret)
			resp["turn_username"] = turnUsername
			resp["turn_credential"] = turnCredential
		}
	}

	return response.Success(c, resp)
}

func generateTurnCredentials(userID int64, secret string) (string, string) {
	return strconv.FormatInt(userID, 10), secret
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

	clients := h.hub.GetRoomClients(roomID)
	result := make([]webrtc.ParticipantInfo, 0, len(clients))
	for _, cl := range clients {
		result = append(result, webrtc.ParticipantInfo{
			ID:   strconv.FormatInt(cl.ID, 10),
			Name: cl.DisplayName,
			Role: cl.Role,
		})
	}
	if result == nil {
		result = []webrtc.ParticipantInfo{}
	}

	return response.Success(c, result)
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

	room := h.signaling.GetRoom(roomID)
	if room == nil {
		return response.NotFound(c, "اتاق یافت نشد")
	}

	isMuted, ok := room.ToggleParticipantMute(participantID)
	if !ok {
		return response.NotFound(c, "شرکت‌کننده یافت نشد")
	}

	h.signaling.BroadcastToRoom(roomID, map[string]interface{}{
		"type":     "participant_muted",
		"user_id":  participantID,
		"is_muted": isMuted,
	})

	return response.Success(c, map[string]interface{}{
		"is_muted": isMuted,
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

	h.signaling.BroadcastToRoom(roomID, map[string]interface{}{
		"type":    "participant_kicked",
		"user_id": participantID,
	})

	h.signaling.GetRoomManager().RemoveParticipant(roomID, participantID)

	return response.Success(c, map[string]string{"message": "شرکت‌کننده حذف شد"})
}
