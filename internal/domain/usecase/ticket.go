package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type TicketUseCase struct {
	ticketRepo *repository.TicketRepo
}

func NewTicketUseCase(ticketRepo *repository.TicketRepo) *TicketUseCase {
	return &TicketUseCase{ticketRepo: ticketRepo}
}

func (uc *TicketUseCase) Create(userID int64, title, category, priority, content string) (*entity.Ticket, error) {
	t := &entity.Ticket{
		UserID:   userID,
		Title:    title,
		Category: category,
		Status:   "open",
		Priority: priority,
	}
	if err := uc.ticketRepo.Create(t); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد تیکت")
	}
	if content != "" {
		msg := &entity.TicketMessage{
			TicketID: t.ID,
			UserID:   userID,
			Content:  content,
			IsAdmin:  false,
		}
		uc.ticketRepo.SendMessage(msg)
	}
	return t, nil
}

func (uc *TicketUseCase) GetByID(id int64) (*entity.Ticket, error) {
	return uc.ticketRepo.GetByID(id)
}

func (uc *TicketUseCase) ListMy(userID int64, page, perPage int) ([]entity.Ticket, int64, error) {
	return uc.ticketRepo.ListByUser(userID, page, perPage)
}

func (uc *TicketUseCase) ListAll(page, perPage int, search, status string) ([]entity.Ticket, int64, error) {
	return uc.ticketRepo.ListAll(page, perPage, search)
}

func (uc *TicketUseCase) Reply(ticketID, userID int64, content string, isAdmin bool) error {
	msg := &entity.TicketMessage{
		TicketID: ticketID,
		UserID:   userID,
		Content:  content,
		IsAdmin:  isAdmin,
	}
	return uc.ticketRepo.SendMessage(msg)
}

func (uc *TicketUseCase) Close(ticketID int64) error {
	return uc.ticketRepo.UpdateStatus(ticketID, "closed")
}

func (uc *TicketUseCase) GetMessages(ticketID int64) ([]entity.TicketMessage, error) {
	return uc.ticketRepo.ListMessages(ticketID)
}
