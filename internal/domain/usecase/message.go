package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type MessageUseCase struct {
	messageRepo *repository.MessageRepo
}

func NewMessageUseCase(messageRepo *repository.MessageRepo) *MessageUseCase {
	return &MessageUseCase{messageRepo: messageRepo}
}

func (uc *MessageUseCase) Send(sessionID, userID int64, content string) (*entity.Message, error) {
	if len(content) > 10000 {
		return nil, fmt.Errorf("پیام بیش از حد طولانی است")
	}
	m := &entity.Message{
		SessionID: sessionID,
		UserID:    userID,
		Content:   content,
		Type:      "text",
	}
	if err := uc.messageRepo.Create(m); err != nil {
		return nil, fmt.Errorf("خطا در ارسال پیام")
	}
	return m, nil
}

func (uc *MessageUseCase) List(sessionID int64, page, perPage int) ([]entity.Message, error) {
	return uc.messageRepo.ListBySession(sessionID, perPage, (page-1)*perPage)
}
