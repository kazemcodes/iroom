package handler

import (
	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles HTTP requests for authentication operations.
// Routes: POST /auth/register, /auth/login, /auth/refresh, /auth/guest-login, /auth/create-login-url
//         GET /auth/me (protected)
type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return response.BadRequest(c, "فیلدهای الزامی خالی هستند")
	}

	user, tokens, err := h.authUC.Register(req.Email, req.Password, req.DisplayName, req.Phone)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Email == "" || req.Password == "" {
		return response.BadRequest(c, "ایمیل و رمز عبور الزامی است")
	}

	user, tokens, err := h.authUC.Login(req.Email, req.Password)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	tokens, err := h.authUC.Refresh(req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, tokens)
}

func (h *AuthHandler) GuestLogin(c echo.Context) error {
	var req struct {
		SessionID   int64  `json:"session_id"`
		DisplayName string `json:"display_name"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	user, tokens, err := h.authUC.GuestLogin(req.SessionID, req.DisplayName)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) CreateLoginURL(c echo.Context) error {
	var req struct {
		RoomID     int64  `json:"room_id"`
		UserID     string `json:"user_id"`
		Nickname   string `json:"nickname"`
		Access     int    `json:"access"`
		Concurrent int    `json:"concurrent"`
		TTL        int    `json:"ttl"`
		Language   string `json:"language"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	url, err := h.authUC.CreateLoginURL(req.RoomID, req.UserID, req.Nickname, req.Access, req.Concurrent, req.TTL, req.Language)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]string{"url": url})
}

func (h *AuthHandler) RoomGuestLogin(c echo.Context) error {
	var req struct {
		RoomSlug    string `json:"room_slug"`
		DisplayName string `json:"display_name"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	user, tokens, err := h.authUC.RoomGuestLogin(req.RoomSlug, req.DisplayName)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	user, err := h.authUC.GetUserByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	return response.Success(c, user)
}
