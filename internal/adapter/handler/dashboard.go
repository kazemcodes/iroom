package handler

import (
	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	dashboardUC *usecase.DashboardUseCase
}

func NewDashboardHandler(dashboardUC *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{dashboardUC: dashboardUC}
}

func (h *DashboardHandler) Stats(c echo.Context) error {
	stats, err := h.dashboardUC.Stats()
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, stats)
}
