package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type TicketHandler struct {
	ticketUC *usecase.TicketUseCase
}

func NewTicketHandler(ticketUC *usecase.TicketUseCase) *TicketHandler {
	return &TicketHandler{ticketUC: ticketUC}
}

func (h *TicketHandler) Create(c echo.Context) error {
	userID, _ := getUserID(c)

	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
		Priority string `json:"priority"`
		Content  string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	ticket, err := h.ticketUC.Create(userID, req.Title, req.Category, req.Priority, req.Content)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, ticket)
}

func (h *TicketHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ticket, err := h.ticketUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "تیکت یافت نشد")
	}
	return response.Success(c, ticket)
}

func (h *TicketHandler) ListMy(c echo.Context) error {
	userID, _ := getUserID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	tickets, total, err := h.ticketUC.ListMy(userID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": tickets,
		"total": total,
	})
}

func (h *TicketHandler) Reply(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := getUserID(c)
	role := getUserRole(c)

	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	isAdmin := role == "admin" || role == "teacher"
	if err := h.ticketUC.Reply(id, userID, req.Content, isAdmin); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "پاسخ ارسال شد"})
}

func (h *TicketHandler) Close(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.ticketUC.Close(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "تیکت بسته شد"})
}
