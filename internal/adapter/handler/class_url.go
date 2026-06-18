package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type ClassURLHandler struct {
	classUC *usecase.ClassUseCase
}

func NewClassURLHandler(classUC *usecase.ClassUseCase) *ClassURLHandler {
	return &ClassURLHandler{classUC: classUC}
}

// ResolveSlug resolves a class slug to its ID and redirects to the join page.
// Used for Skyroom-style URLs: /ch-{org}/{slug}/
func (h *ClassURLHandler) ResolveSlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return response.BadRequest(c, "شناسه کلاس نامعتبر")
	}

	class, err := h.classUC.GetBySlug(slug)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	return response.Success(c, map[string]interface{}{
		"id":   class.ID,
		"name": class.Name,
		"url":  "/classroom/join/" + strconv.FormatInt(class.ID, 10),
	})
}
