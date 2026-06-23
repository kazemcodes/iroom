package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type ClassroomHandler struct {
	hub *services.Hub
}

func NewClassroomHandler(hub *services.Hub) *ClassroomHandler {
	return &ClassroomHandler{hub: hub}
}

func (h *ClassroomHandler) GetJoinInfo(c echo.Context) error {
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

func (h *ClassroomHandler) GetParticipants(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return response.BadRequest(c, "شناسه اتاق نامعتبر")
	}

	clients := h.hub.GetRoomClients(roomID)
	type participant struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Role        string `json:"role"`
		IsMuted     bool   `json:"is_muted"`
		IsVideoOff  bool   `json:"is_video_off"`
		IsScreenSharing bool `json:"is_screen_sharing"`
	}

	result := make([]participant, 0, len(clients))
	for _, cl := range clients {
		result = append(result, participant{
			ID:   strconv.FormatInt(cl.ID, 10),
			Name: cl.DisplayName,
			Role: cl.Role,
		})
	}

	return response.Success(c, result)
}
