package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type ClassroomHandler struct {
	hub      *services.Hub
	userRepo interface {
		UpdateRole(id int64, role string) error
	}
}

func NewClassroomHandler(hub *services.Hub, userRepo interface {
	UpdateRole(id int64, role string) error
}) *ClassroomHandler {
	return &ClassroomHandler{hub: hub, userRepo: userRepo}
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
		ID              string `json:"id"`
		Name            string `json:"name"`
		Role            string `json:"role"`
		IsMuted         bool   `json:"is_muted"`
		IsVideoOff      bool   `json:"is_video_off"`
		IsScreenSharing bool   `json:"is_screen_sharing"`
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

func (h *ClassroomHandler) ChangeRole(c echo.Context) error {
	sessionID := c.Param("id")
	targetIDStr := c.Param("userId")

	role := getUserRole(c)
	if role != "admin" && role != "owner" && role != "operator" {
		return response.Forbidden(c, "شما اجازه تغییر نقش ندارید")
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده نامعتبر")
	}

	validRoles := map[string]bool{"operator": true, "presenter": true, "user": true}
	if !validRoles[req.Role] {
		return response.BadRequest(c, "نقش نامعتبر")
	}

	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کاربر نامعتبر")
	}

	if err := h.userRepo.UpdateRole(targetID, req.Role); err != nil {
		return response.InternalError(c, err.Error())
	}

	h.hub.UpdateClientRole(targetID, req.Role)

	broadcast := map[string]interface{}{
		"type":      "command",
		"command":   "role_change",
		"user_id":   getUserIDOrZero(c),
		"target_id": targetIDStr,
		"role":      req.Role,
	}
	h.hub.BroadcastToRoom(sessionID, "chat", broadcast, 0)

	return response.Success(c, map[string]string{"role": req.Role})
}

func getUserIDOrZero(c echo.Context) int64 {
	id, ok := getUserID(c)
	if !ok {
		return 0
	}
	return id
}
