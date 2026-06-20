package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	messageUC *usecase.MessageUseCase
}

func NewMessageHandler(messageUC *usecase.MessageUseCase) *MessageHandler {
	return &MessageHandler{messageUC: messageUC}
}

func (h *MessageHandler) Send(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)

	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	msg, err := h.messageUC.Send(sessionID, userID, req.Content)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, msg)
}

func (h *MessageHandler) List(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}

	messages, err := h.messageUC.List(sessionID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": messages,
		"total": len(messages),
	})
}
