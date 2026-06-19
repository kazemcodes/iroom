package handler

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// SettingsHandler handles HTTP requests for system settings (admin only).
// Routes: GET /admin/settings, PUT /admin/settings
type SettingsHandler struct {
	settingsUC *usecase.SettingsUseCase
}

func NewSettingsHandler(settingsUC *usecase.SettingsUseCase) *SettingsHandler {
	return &SettingsHandler{settingsUC: settingsUC}
}

func (h *SettingsHandler) Get(c echo.Context) error {
	settings, err := h.settingsUC.GetAll()
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, settings)
}

func (h *SettingsHandler) Update(c echo.Context) error {
	var req map[string]interface{}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	settings := make(map[string]string)
	for k, v := range req {
		switch val := v.(type) {
		case bool:
			if val {
				settings[k] = "true"
			} else {
				settings[k] = "false"
			}
		case string:
			settings[k] = val
		case float64:
			settings[k] = fmt.Sprintf("%d", int(val))
		}
	}

	if err := h.settingsUC.Update(settings); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "تنظیمات بروزرسانی شد"})
}
