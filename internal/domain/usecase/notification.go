package usecase

import (
	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type NotificationUseCase struct {
	notificationRepo *repository.NotificationRepo
}

func NewNotificationUseCase(notificationRepo *repository.NotificationRepo) *NotificationUseCase {
	return &NotificationUseCase{notificationRepo: notificationRepo}
}

func (uc *NotificationUseCase) Create(userID int64, notifType, title, content string) error {
	n := &entity.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Content: content,
	}
	return uc.notificationRepo.Create(n)
}

func (uc *NotificationUseCase) List(userID int64, page, perPage int) ([]*entity.Notification, error) {
	return uc.notificationRepo.ListByUser(userID, perPage, (page-1)*perPage)
}

func (uc *NotificationUseCase) UnreadCount(userID int64) (int64, error) {
	return uc.notificationRepo.CountUnread(userID)
}

func (uc *NotificationUseCase) MarkRead(id, userID int64) error {
	return uc.notificationRepo.MarkRead(id, userID)
}

func (uc *NotificationUseCase) MarkAllRead(userID int64) error {
	return uc.notificationRepo.MarkAllRead(userID)
}
