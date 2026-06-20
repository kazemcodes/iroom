package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type WebhookUseCase struct {
	webhookRepo  *repository.WebhookRepo
	deliveryRepo *repository.WebhookDeliveryRepo
}

func NewWebhookUseCase(webhookRepo *repository.WebhookRepo, deliveryRepo *repository.WebhookDeliveryRepo) *WebhookUseCase {
	return &WebhookUseCase{webhookRepo: webhookRepo, deliveryRepo: deliveryRepo}
}

func (uc *WebhookUseCase) Create(userID int64, url string, events []string) (*entity.Webhook, error) {
	w := &entity.Webhook{
		UserID:   userID,
		URL:      url,
		Events:   events,
		IsActive: true,
	}
	if err := uc.webhookRepo.Create(w); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد وب‌هوک")
	}
	return w, nil
}

func (uc *WebhookUseCase) GetByID(id int64) (*entity.Webhook, error) {
	return uc.webhookRepo.GetByID(id)
}

func (uc *WebhookUseCase) ListByUser(userID int64) ([]entity.Webhook, error) {
	return uc.webhookRepo.ListByUserID(userID)
}

func (uc *WebhookUseCase) ListAll() ([]entity.Webhook, error) {
	return uc.webhookRepo.ListAll()
}

func (uc *WebhookUseCase) Update(id, userID int64, role, url string, events []string, isActive *bool) error {
	w, err := uc.webhookRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}
	if role != "admin" && w.UserID != userID {
		return errors.ErrForbidden
	}
	if url != "" {
		w.URL = url
	}
	if events != nil {
		w.Events = events
	}
	if isActive != nil {
		w.IsActive = *isActive
	}
	return uc.webhookRepo.Update(w)
}

func (uc *WebhookUseCase) Delete(id, userID int64, role string) error {
	w, err := uc.webhookRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}
	if role != "admin" && w.UserID != userID {
		return errors.ErrForbidden
	}
	return uc.webhookRepo.Delete(id)
}

func (uc *WebhookUseCase) ListDeliveries(webhookID, userID int64, role string, page, perPage int) ([]entity.WebhookDelivery, int64, error) {
	w, err := uc.webhookRepo.GetByID(webhookID)
	if err != nil {
		return nil, 0, errors.ErrNotFound
	}
	if role != "admin" && w.UserID != userID {
		return nil, 0, errors.ErrForbidden
	}
	return uc.deliveryRepo.ListByWebhookID(webhookID, page, perPage)
}
