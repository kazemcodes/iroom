package handlers

import (
	"crypto/rand"
	"fmt"
	"strconv"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	webhookRepo  *repository.WebhookRepo
	deliveryRepo *repository.WebhookDeliveryRepo
	deliverySvc  *services.WebhookDeliveryService
}

func NewWebhookHandler(webhookRepo *repository.WebhookRepo, deliveryRepo *repository.WebhookDeliveryRepo, deliverySvc *services.WebhookDeliveryService) *WebhookHandler {
	return &WebhookHandler{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		deliverySvc:  deliverySvc,
	}
}

func (h *WebhookHandler) Create(c echo.Context) error {
	var req models.CreateWebhookRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.URL == "" {
		return response.BadRequest(c, "آدرس URL الزامی است")
	}

	if len(req.Events) == 0 {
		return response.BadRequest(c, "حداقل یک رویداد باید انتخاب شود")
	}

	// Validate event types
	validEvents := map[string]bool{
		models.WebhookEventSessionStarted: true,
		models.WebhookEventSessionEnded:   true,
		models.WebhookEventUserRegistered: true,
		models.WebhookEventTicketCreated:  true,
	}
	for _, event := range req.Events {
		if !validEvents[event] {
			return response.BadRequest(c, fmt.Sprintf("رویداد نامعتبر: %s", event))
		}
	}

	// Generate secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return response.InternalError(c, "خطا در تولید کلید امنیتی")
	}
	secret := fmt.Sprintf("%x", secretBytes)

	userID := c.Get("user_id").(int64)

	webhook := &models.Webhook{
		UserID:   userID,
		URL:      req.URL,
		Secret:   secret,
		Events:   req.Events,
		IsActive: true,
	}

	if err := h.webhookRepo.Create(webhook); err != nil {
		return response.InternalError(c, "خطا در ایجاد وب‌هوک")
	}

	// Fetch the created webhook to get all fields
	created, err := h.webhookRepo.GetByID(webhook.ID)
	if err != nil {
		return response.Created(c, webhook)
	}

	return response.Created(c, created)
}

func (h *WebhookHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	webhooks, err := h.webhookRepo.ListByUserID(userID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت وب‌هوک‌ها")
	}

	if webhooks == nil {
		webhooks = []models.Webhook{}
	}

	return response.Success(c, webhooks)
}

func (h *WebhookHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	webhook, err := h.webhookRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "وب‌هوک یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	if webhook.UserID != userID {
		role, _ := c.Get("role").(string)
		if role != "admin" {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	var req models.UpdateWebhookRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.URL != "" {
		webhook.URL = req.URL
	}

	if len(req.Events) > 0 {
		// Validate event types
		validEvents := map[string]bool{
			models.WebhookEventSessionStarted: true,
			models.WebhookEventSessionEnded:   true,
			models.WebhookEventUserRegistered: true,
			models.WebhookEventTicketCreated:  true,
		}
		for _, event := range req.Events {
			if !validEvents[event] {
				return response.BadRequest(c, fmt.Sprintf("رویداد نامعتبر: %s", event))
			}
		}
		webhook.Events = req.Events
	}

	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := h.webhookRepo.Update(webhook); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی وب‌هوک")
	}

	updated, err := h.webhookRepo.GetByID(id)
	if err != nil {
		return response.Success(c, webhook)
	}

	return response.Success(c, updated)
}

func (h *WebhookHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	webhook, err := h.webhookRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "وب‌هوک یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	if webhook.UserID != userID {
		role, _ := c.Get("role").(string)
		if role != "admin" {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	if err := h.webhookRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف وب‌هوک")
	}

	return response.Success(c, map[string]string{"message": "وب‌هوک حذف شد"})
}

func (h *WebhookHandler) ListDeliveries(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	webhook, err := h.webhookRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "وب‌هوک یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	if webhook.UserID != userID {
		role, _ := c.Get("role").(string)
		if role != "admin" {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	deliveries, total, err := h.deliveryRepo.ListByWebhookID(id, page, perPage)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت لاگ‌ها")
	}

	if deliveries == nil {
		deliveries = []models.WebhookDelivery{}
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      deliveries,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *WebhookHandler) Test(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	webhook, err := h.webhookRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "وب‌هوک یافت نشد")
	}

	userID := c.Get("user_id").(int64)
	if webhook.UserID != userID {
		role, _ := c.Get("role").(string)
		if role != "admin" {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	h.deliverySvc.SendTestEvent(*webhook)

	return response.Success(c, map[string]string{"message": "رویداد تست ارسال شد"})
}
