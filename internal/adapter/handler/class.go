package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// ClassHandler handles HTTP requests for class management.
// Routes: GET/POST /classes, GET/PUT/DELETE /classes/:id
//         POST /classes/:id/enroll, DELETE /classes/:id/users/:userId
//         GET /classes/:id/students, GET /classes/:id/url
//         POST /classes/join/:code, GET /users/:id/rooms
type ClassHandler struct {
	classUC *usecase.ClassUseCase
}

func NewClassHandler(classUC *usecase.ClassUseCase) *ClassHandler {
	return &ClassHandler{classUC: classUC}
}

func (h *ClassHandler) Create(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		MaxStudents int    `json:"max_students"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	userID, _ := getUserID(c)
	class, err := h.classUC.Create(userID, req.Name, req.Description, req.Color, req.MaxStudents)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, class)
}

func (h *ClassHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.Success(c, class)
}

func (h *ClassHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	classes, total, err := h.classUC.List(0, page, perPage, c.QueryParam("search"))
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items":      classes,
		"total":      total,
		"page":       page,
		"per_page":   perPage,
		"total_pages": (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *ClassHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		MaxStudents int    `json:"max_students"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	userID, _ := getUserID(c)
	role := getUserRole(c)
	class, err := h.classUC.Update(id, userID, req.Name, req.Description, req.Color, req.MaxStudents, role)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, class)
}

func (h *ClassHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.classUC.Delete(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "کلاس حذف شد"})
}

func (h *ClassHandler) Enroll(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		StudentID int64 `json:"student_id"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.classUC.Enroll(id, req.StudentID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "دانش‌آموز اضافه شد"})
}

func (h *ClassHandler) RemoveUser(c echo.Context) error {
	classID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	actorID, _ := getUserID(c)
	role := getUserRole(c)

	if err := h.classUC.RemoveUser(classID, userID, actorID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "کاربر از کلاس حذف شد"})
}

func (h *ClassHandler) UpdateUserAccess(c echo.Context) error {
	classID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	actorID, _ := getUserID(c)
	role := getUserRole(c)

	var req struct {
		Access int `json:"access"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.classUC.UpdateUserAccess(classID, userID, actorID, role, req.Access); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]int{"access": req.Access})
}

func (h *ClassHandler) GetStudents(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	students, err := h.classUC.GetStudents(id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, students)
}

func (h *ClassHandler) GetURL(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	url := h.classUC.GetURL(id)
	return response.Success(c, map[string]string{"url": url})
}

func (h *ClassHandler) GetUserRooms(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rooms, err := h.classUC.GetUserRooms(id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, rooms)
}

func (h *ClassHandler) JoinByCode(c echo.Context) error {
	code := c.Param("code")
	class, err := h.classUC.JoinByCode(code)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	return response.Success(c, class)
}

func (h *ClassHandler) RegenerateCode(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := getUserID(c)
	role := getUserRole(c)

	code, err := h.classUC.RegenerateCode(id, userID, role)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"code": code})
}
