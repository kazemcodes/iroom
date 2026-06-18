package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type PollUseCase struct {
	pollRepo *repository.PollRepo
}

func NewPollUseCase(pollRepo *repository.PollRepo) *PollUseCase {
	return &PollUseCase{pollRepo: pollRepo}
}

func (uc *PollUseCase) Create(sessionID int64, question, options string) (*entity.Poll, error) {
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

func (uc *PollUseCase) Close(pollID int64) error {
	return uc.pollRepo.Close(pollID)
}
