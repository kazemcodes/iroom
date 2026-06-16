package handlers

import (
	"time"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/hash"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type ExternalHandler struct {
	userRepo    *repository.UserRepo
	classRepo   *repository.ClassRepo
	sessionRepo *repository.SessionRepo
	apiKey      string
}

func NewExternalHandler(userRepo *repository.UserRepo, classRepo *repository.ClassRepo, sessionRepo *repository.SessionRepo, apiKey string) *ExternalHandler {
	return &ExternalHandler{
		userRepo:    userRepo,
		classRepo:   classRepo,
		sessionRepo: sessionRepo,
		apiKey:      apiKey,
	}
}

func (h *ExternalHandler) CreateUser(c echo.Context) error {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Phone       string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return response.BadRequest(c, "ایمیل، رمز عبور و نام الزامی هستند")
	}

	if req.Role == "" {
		req.Role = "student"
	}

	hashedPassword, err := hash.Hash(req.Password)
	if err != nil {
		return response.InternalError(c, "خطای داخلی")
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         req.Role,
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := h.userRepo.Create(user); err != nil {
		return response.BadRequest(c, "ایمیل تکراری است")
	}

	return response.Created(c, user)
}

func (h *ExternalHandler) CreateClass(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TeacherID   int64  `json:"teacher_id"`
		MaxStudents int    `json:"max_students"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Name == "" {
		return response.BadRequest(c, "نام کلاس الزامی است")
	}

	if req.TeacherID == 0 {
		req.TeacherID = 1
	}
	if req.MaxStudents <= 0 {
		req.MaxStudents = 30
	}

	class := &models.Class{
		TeacherID:   req.TeacherID,
		Name:        req.Name,
		Description: req.Description,
		Color:       "#3B82F6",
		MaxStudents: req.MaxStudents,
	}

	if err := h.classRepo.Create(class); err != nil {
		return response.InternalError(c, "خطا در ایجاد کلاس")
	}

	return response.Created(c, class)
}

func (h *ExternalHandler) CreateSession(c echo.Context) error {
	var req struct {
		ClassID     int64  `json:"class_id"`
		Title       string `json:"title"`
		ScheduledAt string `json:"scheduled_at"`
		Duration    int    `json:"duration"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.ClassID == 0 || req.Title == "" {
		return response.BadRequest(c, "شناسه کلاس و عنوان الزامی هستند")
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return response.BadRequest(c, "تاریخ نامعتبر")
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 60
	}

	session := &models.Session{
		ClassID:     req.ClassID,
		Title:       req.Title,
		ScheduledAt: scheduledAt,
		Duration:    duration,
		Status:      "scheduled",
	}

	if err := h.sessionRepo.Create(session); err != nil {
		return response.InternalError(c, "خطا در ایجاد جلسه")
	}

	return response.Created(c, session)
}

func (h *ExternalHandler) Status(c echo.Context) error {
	return response.Success(c, map[string]interface{}{
		"status":    "healthy",
		"version":   "0.1.0",
		"timestamp": time.Now().Unix(),
	})
}

func (h *ExternalHandler) Stats(c echo.Context) error {
	userCount, _ := h.userRepo.Count()
	classCount, _ := h.classRepo.Count()
	sessionCount, _ := h.sessionRepo.Count()

	return response.Success(c, map[string]interface{}{
		"users":    userCount,
		"classes":  classCount,
		"sessions": sessionCount,
	})
}

type WebhookEntry struct {
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Active  bool     `json:"active"`
}

var webhookStore = map[string]*WebhookEntry{}

func (h *ExternalHandler) HandleWebhook(c echo.Context) error {
	var req struct {
		Action string `json:"action"`
		URL    string `json:"url"`
		Events []string `json:"events"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	switch req.Action {
	case "register":
		if req.URL == "" {
			return response.BadRequest(c, "URL الزامی است")
		}
		webhookStore[req.URL] = &WebhookEntry{
			URL:    req.URL,
			Events: req.Events,
			Active: true,
		}
		return response.Created(c, map[string]string{"message": "webhook ثبت شد"})

	case "unregister":
		delete(webhookStore, req.URL)
		return response.Success(c, map[string]string{"message": "webhook حذف شد"})

	case "list":
		var list []*WebhookEntry
		for _, v := range webhookStore {
			list = append(list, v)
		}
		return response.Success(c, list)

	default:
		return response.BadRequest(c, "عملیات نامعتبر")
	}
}
