package handlers

import (
	"strconv"
	"time"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type TicketHandler struct {
	ticketRepo    *repository.TicketRepo
	sessionLogRepo *repository.SessionLogRepo
}

func NewTicketHandler(ticketRepo *repository.TicketRepo, sessionLogRepo *repository.SessionLogRepo) *TicketHandler {
	return &TicketHandler{
		ticketRepo:    ticketRepo,
		sessionLogRepo: sessionLogRepo,
	}
}

func (h *TicketHandler) Create(c echo.Context) error {
	var req models.CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Title == "" {
		return response.BadRequest(c, "عنوان تیکت الزامی است")
	}

	userID := c.Get("user_id").(int64)

	category := req.Category
	if category == "" {
		category = "general"
	}
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	ticket := &models.Ticket{
		UserID:   userID,
		Title:    req.Title,
		Category: category,
		Priority: priority,
		Status:   "open",
	}

	if err := h.ticketRepo.Create(ticket); err != nil {
		return response.InternalError(c, "خطا در ایجاد تیکت")
	}

	if req.Message != "" {
		msg := &models.TicketMessage{
			TicketID: ticket.ID,
			UserID:   userID,
			Content:  req.Message,
			IsAdmin:  false,
		}
		h.ticketRepo.SendMessage(msg)
	}

	created, err := h.ticketRepo.GetByID(ticket.ID)
	if err != nil {
		return response.Created(c, ticket)
	}

	return response.Created(c, created)
}

func (h *TicketHandler) ListMy(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	tickets, total, err := h.ticketRepo.ListByUser(userID, page, perPage)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت تیکت‌ها")
	}
	if tickets == nil {
		tickets = []models.Ticket{}
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      tickets,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *TicketHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	ticket, err := h.ticketRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "تیکت یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	role, _ := c.Get("role").(string)
	if ticket.UserID != userID && role != "admin" {
		return response.Forbidden(c, "دسترسی غیرمجاز")
	}

	messages, err := h.ticketRepo.ListMessages(id)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت پیام‌ها")
	}
	if messages == nil {
		messages = []models.TicketMessage{}
	}

	return response.Success(c, map[string]interface{}{
		"ticket":   ticket,
		"messages": messages,
	})
}

func (h *TicketHandler) Reply(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	ticket, err := h.ticketRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "تیکت یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	role, _ := c.Get("role").(string)
	if ticket.UserID != userID && role != "admin" {
		return response.Forbidden(c, "دسترسی غیرمجاز")
	}

	if ticket.Status == "closed" {
		return response.BadRequest(c, "تیکت بسته شده است")
	}

	var req models.ReplyTicketRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Content == "" {
		return response.BadRequest(c, "محتوا الزامی است")
	}

	msg := &models.TicketMessage{
		TicketID: id,
		UserID:   userID,
		Content:  req.Content,
		IsAdmin:  role == "admin",
	}

	if err := h.ticketRepo.SendMessage(msg); err != nil {
		return response.InternalError(c, "خطا در ارسال پیام")
	}

	if role == "admin" {
		h.ticketRepo.UpdateStatus(id, "answered")
	}

	sent, err := h.ticketRepo.ListMessages(id)
	if err != nil {
		return response.Created(c, msg)
	}
	if len(sent) > 0 {
		return response.Created(c, sent[len(sent)-1])
	}

	return response.Created(c, msg)
}

func (h *TicketHandler) Close(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	ticket, err := h.ticketRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "تیکت یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	role, _ := c.Get("role").(string)
	if ticket.UserID != userID && role != "admin" {
		return response.Forbidden(c, "دسترسی غیرمجاز")
	}

	if err := h.ticketRepo.UpdateStatus(id, "closed"); err != nil {
		return response.InternalError(c, "خطا در بستن تیکت")
	}

	return response.Success(c, map[string]string{"message": "تیکت بسته شد"})
}

// Admin ticket list

type AdminTicketHandler struct {
	ticketRepo *repository.TicketRepo
}

func NewAdminTicketHandler(ticketRepo *repository.TicketRepo) *AdminTicketHandler {
	return &AdminTicketHandler{ticketRepo: ticketRepo}
}

func (h *AdminTicketHandler) ListAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	search := c.QueryParam("search")

	tickets, total, err := h.ticketRepo.ListAll(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت تیکت‌ها")
	}
	if tickets == nil {
		tickets = []models.Ticket{}
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      tickets,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

// Session log handlers

type SessionLogHandler struct {
	sessionLogRepo *repository.SessionLogRepo
}

func NewSessionLogHandler(sessionLogRepo *repository.SessionLogRepo) *SessionLogHandler {
	return &SessionLogHandler{sessionLogRepo: sessionLogRepo}
}

func (h *SessionLogHandler) ListBySession(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	logs, err := h.sessionLogRepo.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت لاگ‌ها")
	}
	if logs == nil {
		logs = []models.SessionLog{}
	}

	return response.Success(c, logs)
}

func (h *SessionLogHandler) LogJoin(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID := c.Get("user_id").(int64)

	log := &models.SessionLog{
		SessionID: sessionID,
		UserID:    userID,
		JoinedAt:  time.Now(),
		IPAddress: c.RealIP(),
	}

	if err := h.sessionLogRepo.Create(log); err != nil {
		return response.InternalError(c, "خطا در ثبت ورود")
	}

	return response.Created(c, log)
}

func (h *SessionLogHandler) LogLeave(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID := c.Get("user_id").(int64)

	log, err := h.sessionLogRepo.GetActiveLog(sessionID, userID)
	if err != nil {
		return response.NotFound(c, "لاگ فعالی یافت نشد")
	}

	now := time.Now()
	duration := int(now.Sub(log.JoinedAt).Seconds())

	if err := h.sessionLogRepo.UpdateLeave(log.ID, now.Format(time.RFC3339), duration); err != nil {
		return response.InternalError(c, "خطا در ثبت خروج")
	}

	return response.Success(c, map[string]string{"message": "خروج ثبت شد"})
}
