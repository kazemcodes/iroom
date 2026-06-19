package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/adapter/repository/sqlite"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type ActivityLogHandler struct {
	logRepo *repository.ActivityLogRepo
}

func NewActivityLogHandler(logRepo *repository.ActivityLogRepo) *ActivityLogHandler {
	return &ActivityLogHandler{logRepo: logRepo}
}

func (h *ActivityLogHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 30
	}

	logs, total, err := h.logRepo.List(page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items":       logs,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": (total + int64(perPage) - 1) / int64(perPage),
	})
}
