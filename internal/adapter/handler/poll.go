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
	sessionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Question string `json:"question"`
		Options  string `json:"options"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	poll, err := h.pollUC.Create(sessionID, req.Question, req.Options)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, poll)
}

func (h *PollHandler) List(c echo.Context) error {
	sessionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	polls, err := h.pollUC.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, polls)
}

func (h *PollHandler) Vote(c echo.Context) error {
	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
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
	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	results, err := h.pollUC.GetResults(pollID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, results)
}

func (h *PollHandler) Close(c echo.Context) error {
	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.pollUC.Close(pollID); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "نظرسنجی بسته شد"})
}
