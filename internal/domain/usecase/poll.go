package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type PollUseCase struct {
	pollRepo    *repository.PollRepo
	sessionRepo *repository.SessionRepo
	roomRepo    *repository.RoomRepo
}

func NewPollUseCase(pollRepo *repository.PollRepo, sessionRepo *repository.SessionRepo, roomRepo *repository.RoomRepo) *PollUseCase {
	return &PollUseCase{pollRepo: pollRepo, sessionRepo: sessionRepo, roomRepo: roomRepo}
}

func (uc *PollUseCase) Create(sessionID, actorID int64, role, question, options string) (*entity.Poll, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	room, err := uc.roomRepo.GetByID(session.RoomID)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return nil, errors.ErrForbidden
	}
	p := &entity.Poll{
		SessionID: sessionID,
		Question:  question,
		Options:   options,
		IsActive:  true,
	}
	if err := uc.pollRepo.Create(p); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد نظرسنجی")
	}
	return p, nil
}

func (uc *PollUseCase) ListBySession(sessionID int64) ([]*entity.Poll, error) {
	return uc.pollRepo.ListBySession(sessionID)
}

func (uc *PollUseCase) Vote(pollID, userID int64, optionIndex int) error {
	vote := &entity.PollVote{
		PollID:      pollID,
		UserID:      userID,
		OptionIndex: optionIndex,
	}
	return uc.pollRepo.Vote(vote)
}

func (uc *PollUseCase) GetResults(pollID int64) (*entity.PollResults, error) {
	return uc.pollRepo.GetResults(pollID)
}

func (uc *PollUseCase) Close(pollID, actorID int64, role string) error {
	poll, err := uc.pollRepo.GetByID(pollID)
	if err != nil {
		return errors.ErrNotFound
	}
	session, err := uc.sessionRepo.GetByID(poll.SessionID)
	if err != nil {
		return errors.ErrNotFound
	}
	room, err := uc.roomRepo.GetByID(session.RoomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return errors.ErrForbidden
	}
	return uc.pollRepo.Close(pollID)
}
