package handler

import (
	"fmt"
	"strings"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type ClassURLHandler struct {
	classUC *usecase.ClassUseCase
	userUC  *usecase.UserUseCase
}

func NewClassURLHandler(classUC *usecase.ClassUseCase, userUC *usecase.UserUseCase) *ClassURLHandler {
	return &ClassURLHandler{classUC: classUC, userUC: userUC}
}

// ResolveSlug resolves a class slug to its join URL.
// Skyroom-style: /user-name/class-name/
func (h *ClassURLHandler) ResolveSlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return response.BadRequest(c, "شناسه کلاس نامعتبر")
	}

	class, err := h.classUC.GetBySlug(slug)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	teacher, err := h.userUC.GetByID(class.TeacherID)
	teacherSlug := "admin"
	if err == nil && teacher != nil {
		// Use email prefix as username
		teacherSlug = strings.Split(teacher.Email, "@")[0]
	}

	return response.Success(c, map[string]interface{}{
		"id":   class.ID,
		"name": class.Name,
		"url":  fmt.Sprintf("/%s/%s", teacherSlug, class.Slug),
	})
}
