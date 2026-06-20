package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type PollHandler struct {
	pollUC *usecase.PollUseCase
}

func NewPollHandler(pollUC *usecase.PollUseCase) *PollHandler {
	return &PollHandler{pollUC: pollUC}
}

func (h *PollHandler) Create(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	actorID, _ := getUserID(c)
	role := getUserRole(c)

	var req struct {
		Question string `json:"question"`
		Options  string `json:"options"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	poll, err := h.pollUC.Create(sessionID, actorID, role, req.Question, req.Options)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Created(c, poll)
}

func (h *PollHandler) List(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	polls, err := h.pollUC.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, polls)
}

func (h *PollHandler) Vote(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)

	var req struct {
		OptionIndex int `json:"option_index"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.pollUC.Vote(pollID, userID, req.OptionIndex); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "رای ثبت شد"})
}

func (h *PollHandler) Results(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	results, err := h.pollUC.GetResults(pollID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, results)
}

func (h *PollHandler) Close(c echo.Context) error {
	pollID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	actorID, _ := getUserID(c)
	role := getUserRole(c)
	if err := h.pollUC.Close(pollID, actorID, role); err != nil {
		return response.Forbidden(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "نظرسنجی بسته شد"})
}
