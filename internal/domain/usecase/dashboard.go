package usecase

import (
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type DashboardUseCase struct {
	userRepo      *repository.UserRepo
	roomRepo      *repository.RoomRepo
	sessionRepo   *repository.SessionRepo
	recordingRepo *repository.RecordingRepo
}

func NewDashboardUseCase(
	userRepo *repository.UserRepo,
	roomRepo *repository.RoomRepo,
	sessionRepo *repository.SessionRepo,
	recordingRepo *repository.RecordingRepo,
) *DashboardUseCase {
	return &DashboardUseCase{
		userRepo:      userRepo,
		roomRepo:      roomRepo,
		sessionRepo:   sessionRepo,
		recordingRepo: recordingRepo,
	}
}

func (uc *DashboardUseCase) Stats() (map[string]int64, error) {
	userCount, _ := uc.userRepo.Count()
	roomCount, _ := uc.roomRepo.Count()
	sessionCount, _ := uc.sessionRepo.Count()
	return map[string]int64{
		"users":    userCount,
		"rooms":    roomCount,
		"sessions": sessionCount,
	}, nil
}
