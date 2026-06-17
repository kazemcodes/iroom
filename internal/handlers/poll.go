package handlers

import (
	"database/sql"
	"strconv"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type PollHandler struct {
	pollRepo   *repository.PollRepo
	sessionRepo *repository.SessionRepo
	logRepo    *repository.ActivityLogRepo
}

func NewPollHandler(pollRepo *repository.PollRepo, sessionRepo *repository.SessionRepo, logRepo *repository.ActivityLogRepo) *PollHandler {
	return &PollHandler{
		pollRepo:   pollRepo,
		sessionRepo: sessionRepo,
		logRepo:    logRepo,
	}
}

func (h *PollHandler) log(userID int64, action, entityType string, entityID int64, details, ip string) {
	h.logRepo.Create(&models.ActivityLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  ip,
	})
}

// Create creates a new poll for a session (teacher/admin only)
func (h *PollHandler) Create(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه جلسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	// Get class to check teacher (also verifies session exists)
	class, err := h.sessionRepo.GetClassBySessionID(sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "جلسه یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت کلاس")
	}

	if class.TeacherID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه ایجاد نظرسنجی در این جلسه را ندارید")
	}

	var req models.CreatePollRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Question == "" {
		return response.BadRequest(c, "سوال الزامی است")
	}

	if len(req.Options) < 2 {
		return response.BadRequest(c, "حداقل دو گزینه الزامی است")
	}

	poll := &models.Poll{
		SessionID: sessionID,
		Question:  req.Question,
		Options:   req.Options,
		IsActive:  true,
	}

	if err := h.pollRepo.Create(poll); err != nil {
		return response.InternalError(c, "خطا در ایجاد نظرسنجی")
	}

	h.log(userID, "create_poll", "poll", poll.ID, req.Question, c.RealIP())

	return response.Created(c, poll)
}

// List lists all polls for a session
func (h *PollHandler) List(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه جلسه نامعتبر")
	}

	polls, err := h.pollRepo.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت نظرسنجی‌ها")
	}
	if polls == nil {
		polls = []*models.Poll{}
	}

	return response.Success(c, polls)
}

// Vote casts a vote on a poll
func (h *PollHandler) Vote(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نظرسنجی نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	// Get poll to verify it exists and is active
	poll, err := h.pollRepo.GetByID(pollID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "نظرسنجی یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت نظرسنجی")
	}

	if !poll.IsActive {
		return response.BadRequest(c, "این نظرسنجی بسته شده است")
	}

	var req models.VoteRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	// Validate option index
	if req.OptionIndex < 0 || req.OptionIndex >= len(poll.Options) {
		return response.BadRequest(c, "گزینه انتخاب شده نامعتبر است")
	}

	vote := &models.PollVote{
		PollID:      pollID,
		UserID:      userID,
		OptionIndex: req.OptionIndex,
	}

	if err := h.pollRepo.Vote(vote); err != nil {
		return response.InternalError(c, "خطا در ثبت رأی")
	}

	h.log(userID, "vote_poll", "poll", pollID, strconv.Itoa(req.OptionIndex), c.RealIP())

	return response.Success(c, map[string]interface{}{
		"message":      "رأی شما ثبت شد",
		"poll_id":      pollID,
		"option_index": req.OptionIndex,
	})
}

// Results returns the results of a poll
func (h *PollHandler) Results(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نظرسنجی نامعتبر")
	}

	// Verify poll exists
	_, err = h.pollRepo.GetByID(pollID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "نظرسنجی یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت نظرسنجی")
	}

	results, err := h.pollRepo.GetResults(pollID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت نتایج")
	}

	return response.Success(c, results)
}

// Close closes a poll (teacher/admin only)
func (h *PollHandler) Close(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نظرسنجی نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	// Get poll to verify it exists
	poll, err := h.pollRepo.GetByID(pollID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "نظرسنجی یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت نظرسنجی")
	}

	// Get session and class to check teacher
	class, err := h.sessionRepo.GetClassBySessionID(poll.SessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت کلاس")
	}

	if class.TeacherID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه بستن این نظرسنجی را ندارید")
	}

	if err := h.pollRepo.Close(pollID); err != nil {
		return response.InternalError(c, "خطا در بستن نظرسنجی")
	}

	h.log(userID, "close_poll", "poll", pollID, "", c.RealIP())

	return response.Success(c, map[string]interface{}{
		"id":      pollID,
		"message": "نظرسنجی بسته شد",
	})
}
