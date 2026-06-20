package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	roomUC      *usecase.RoomUseCase
	sessionUC   *usecase.SessionUseCase
}

func NewRoomHandler(roomUC *usecase.RoomUseCase, sessionUC *usecase.SessionUseCase) *RoomHandler {
	return &RoomHandler{roomUC: roomUC, sessionUC: sessionUC}
}

func (h *RoomHandler) Create(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		MaxUsers    int    `json:"max_users"`
		InviteCode  string `json:"invite_code"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Name == "" {
		return response.BadRequest(c, "نام اتاق الزامی است")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	room, err := h.roomUC.Create(userID, req.Name, req.Description, req.Color, req.MaxUsers, req.InviteCode)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, room)
}

func (h *RoomHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	room, err := h.roomUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	return response.Success(c, room)
}

func (h *RoomHandler) GetBySlug(c echo.Context) error {
	slug := c.Param("slug")
	room, err := h.roomUC.GetBySlug(slug)
	if err != nil {
		return response.NotFound(c, "اتاق یافت نشد")
	}
	return response.Success(c, room)
}

func (h *RoomHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	rooms, total, err := h.roomUC.List(page, perPage, c.QueryParam("search"))
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]interface{}{
		"items":       rooms,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *RoomHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		Color             string `json:"color"`
		GuestLoginEnabled *bool  `json:"guest_login_enabled"`
		MaxUsers          int    `json:"max_users"`
		InviteCode        string `json:"invite_code"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	guestLogin := true
	if req.GuestLoginEnabled != nil {
		guestLogin = *req.GuestLoginEnabled
	}

	userID, _ := getUserID(c)
	role := getUserRole(c)
	room, err := h.roomUC.Update(id, userID, role, req.Name, req.Description, req.Color, guestLogin, req.MaxUsers, req.InviteCode)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, room)
}

func (h *RoomHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.roomUC.Delete(id, userID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "اتاق حذف شد"})
}

func (h *RoomHandler) AddUser(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	var req struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
		Access int    `json:"access"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	actorID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.roomUC.AddUser(roomID, req.UserID, actorID, req.Role, req.Access, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "کاربر اضافه شد"})
}

func (h *RoomHandler) RemoveUser(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کاربر نامعتبر")
	}
	actorID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.roomUC.RemoveUser(roomID, userID, actorID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "کاربر حذف شد"})
}

func (h *RoomHandler) UpdateUserAccess(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کاربر نامعتبر")
	}
	var req struct {
		Access int `json:"access"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	actorID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.roomUC.UpdateUserAccess(roomID, userID, actorID, role, req.Access); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]int{"access": req.Access})
}

func (h *RoomHandler) GetUsers(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	users, err := h.roomUC.GetUsers(id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, users)
}

func (h *RoomHandler) GetInfo(c echo.Context) error {
	slug := c.Param("slug")
	room, err := h.roomUC.GetBySlug(slug)
	if err != nil {
		return response.NotFound(c, "اتاق یافت نشد")
	}

	userCount, _ := h.roomUC.GetUserCount(room.ID)
	activeSessions, _ := h.sessionUC.CountActiveByRoom(room.ID)

	return response.Success(c, map[string]interface{}{
		"room":            room,
		"user_count":      userCount,
		"active_sessions": activeSessions,
	})
}

func (h *RoomHandler) GetSettings(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	settings, err := h.roomUC.GetSettings(id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, settings)
}

func (h *RoomHandler) UpdateSettings(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		MaxUsers                  int  `json:"max_users"`
		RecordingEnabled          bool `json:"recording_enabled"`
		AllowStudentVideo         bool `json:"allow_student_video"`
		AllowStudentAudio         bool `json:"allow_student_audio"`
		AllowStudentScreenShare   bool `json:"allow_student_screen_share"`
		AllowStudentWhiteboard    bool `json:"allow_student_whiteboard"`
		AllowStudentChat          bool `json:"allow_student_chat"`
		SessionAutoEndMinutes     int  `json:"session_auto_end_minutes"`
		WaitingRoomEnabled        bool `json:"waiting_room_enabled"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	userID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.roomUC.UpdateSettings(id, userID, role, &entity.RoomSettings{
		RoomID:                  id,
		MaxUsers:                req.MaxUsers,
		RecordingEnabled:        req.RecordingEnabled,
		AllowStudentVideo:       req.AllowStudentVideo,
		AllowStudentAudio:       req.AllowStudentAudio,
		AllowStudentScreenShare: req.AllowStudentScreenShare,
		AllowStudentWhiteboard:  req.AllowStudentWhiteboard,
		AllowStudentChat:        req.AllowStudentChat,
		SessionAutoEndMinutes:   req.SessionAutoEndMinutes,
		WaitingRoomEnabled:      req.WaitingRoomEnabled,
	}); err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "تنظیمات بروزرسانی شد"})
}

func (h *RoomHandler) RegenerateCode(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)
	role := getUserRole(c)
	code, err := h.roomUC.RegenerateCode(id, userID, role)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"code": code})
}

func (h *RoomHandler) JoinByCode(c echo.Context) error {
	code := c.Param("code")
	room, err := h.roomUC.JoinByCode(code)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	return response.Success(c, room)
}

func (h *RoomHandler) GetUserRooms(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	authID, _ := getUserID(c)
	role := getUserRole(c)
	if id != authID && role != "admin" {
		return response.Forbidden(c, "دسترسی غیرمجاز")
	}
	rooms, err := h.roomUC.GetUserRooms(id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, rooms)
}
