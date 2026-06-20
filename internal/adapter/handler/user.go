package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	users, total, err := h.userUC.List(page, perPage, c.QueryParam("search"))
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": users,
		"total": total,
	})
}

func (h *UserHandler) Create(c echo.Context) error {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
		Role        string `json:"role"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	callerRole := getUserRole(c)
	user, err := h.userUC.Create(req.Email, req.Password, req.DisplayName, req.Phone, req.Role, callerRole)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Created(c, user)
}

func (h *UserHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
		Role        string `json:"role"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	callerRole := getUserRole(c)
	if err := h.userUC.Update(id, req.DisplayName, req.Phone, req.Role, req.IsActive, callerRole); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "کاربر بروزرسانی شد"})
}

func (h *UserHandler) Deactivate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	callerRole := getUserRole(c)
	if err := h.userUC.Delete(id, callerRole); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "کاربر غیرفعال شد"})
}

func (h *UserHandler) BatchDelete(c echo.Context) error {
	var req struct {
		Users []int64 `json:"users"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	callerRole := getUserRole(c)
	success, failure := h.userUC.BatchDelete(req.Users, callerRole)
	return response.Success(c, map[string]int{
		"success": success,
		"failure": failure,
	})
}

func (h *UserHandler) ResetPassword(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Password == "" {
		return response.BadRequest(c, "رمز عبور الزامی است")
	}

	callerRole := getUserRole(c)
	if err := h.userUC.ResetPassword(id, req.Password, callerRole); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "رمز عبور بروزرسانی شد"})
}

func (h *UserHandler) ChangeOwnPassword(c echo.Context) error {
	userID, _ := getUserID(c)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return response.BadRequest(c, "رمز عبور فعلی و جدید الزامی است")
	}

	if err := h.userUC.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "رمز عبور تغییر کرد"})
}
