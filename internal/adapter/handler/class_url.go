package handler

import (
	"fmt"
	"strings"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type ClassURLHandler struct {
	roomUC *usecase.RoomUseCase
	userUC *usecase.UserUseCase
}

func NewClassURLHandler(roomUC *usecase.RoomUseCase, userUC *usecase.UserUseCase) *ClassURLHandler {
	return &ClassURLHandler{roomUC: roomUC, userUC: userUC}
}

// ResolveSlug resolves a room slug to its join URL.
// Skyroom-style: /user-name/room-name/
func (h *ClassURLHandler) ResolveSlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return response.BadRequest(c, "شناسه اتاق نامعتبر")
	}

	room, err := h.roomUC.GetBySlug(slug)
	if err != nil {
		return response.NotFound(c, "اتاق یافت نشد")
	}

	teacher, err := h.userUC.GetByID(room.OwnerID)
	teacherSlug := "admin"
	if err == nil && teacher != nil {
		teacherSlug = strings.Split(teacher.Email, "@")[0]
	}

	return response.Success(c, map[string]interface{}{
		"id":   room.ID,
		"name": room.Name,
		"url":  fmt.Sprintf("/%s/%s", teacherSlug, room.Slug),
	})
}
