package handler

import (
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
			// Validate numeric settings
			if k == "max_users_per_room" || k == "max_file_size_mb" || k == "session_auto_end_minutes" || k == "session_timeout_minutes" {
				if val < 0 {
					return response.BadRequest(c, "مقدار نمی‌تواند منفی باشد")
				}
				if val == 0 && (k == "session_timeout_minutes" || k == "max_users_per_room") {
					return response.BadRequest(c, "مقدار نمی‌تواند صفر باشد")
				}
			}
			settings[k] = string(rune(int(val)))
		}
	}

	if err := h.settingsUC.Update(settings); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "تنظیمات بروزرسانی شد"})
}
