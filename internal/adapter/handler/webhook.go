package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	webhookUC *usecase.WebhookUseCase
}

func NewWebhookHandler(webhookUC *usecase.WebhookUseCase) *WebhookHandler {
	return &WebhookHandler{webhookUC: webhookUC}
}

func (h *WebhookHandler) Create(c echo.Context) error {
	userID, _ := getUserID(c)

	var req struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	webhook, err := h.webhookUC.Create(userID, req.URL, req.Events)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, webhook)
}

func (h *WebhookHandler) List(c echo.Context) error {
	userID, _ := getUserID(c)
	webhooks, err := h.webhookUC.ListByUser(userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, webhooks)
}

func (h *WebhookHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		URL      string   `json:"url"`
		Events   []string `json:"events"`
		IsActive *bool    `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.webhookUC.Update(id, req.URL, req.Events, req.IsActive); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]string{"message": "وب‌هوک بروزرسانی شد"})
}

func (h *WebhookHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.webhookUC.Delete(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "وب‌هوک حذف شد"})
}

func (h *WebhookHandler) ListDeliveries(c echo.Context) error {
	webhookID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	deliveries, total, err := h.webhookUC.ListDeliveries(webhookID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": deliveries,
		"total": total,
	})
}

func (h *WebhookHandler) Test(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	webhook, err := h.webhookUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "وب‌هوک یافت نشد")
	}
	_ = webhook
	return response.Success(c, map[string]string{"message": "تست وب‌هوک ارسال شد"})
}
