package usecase

import (
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type DashboardUseCase struct {
	userRepo      *repository.UserRepo
	classRepo     *repository.ClassRepo
	sessionRepo   *repository.SessionRepo
	recordingRepo *repository.RecordingRepo
	logRepo       *repository.ActivityLogRepo
}

func NewDashboardUseCase(
	userRepo *repository.UserRepo,
	classRepo *repository.ClassRepo,
	sessionRepo *repository.SessionRepo,
	recordingRepo *repository.RecordingRepo,
	logRepo *repository.ActivityLogRepo,
) *DashboardUseCase {
	return &DashboardUseCase{
		userRepo:      userRepo,
		classRepo:     classRepo,
		sessionRepo:   sessionRepo,
		recordingRepo: recordingRepo,
		logRepo:       logRepo,
	}
}

func (uc *DashboardUseCase) Stats() (map[string]int64, error) {
	userCount, _ := uc.userRepo.Count()
	classCount, _ := uc.classRepo.Count()
	sessionCount, _ := uc.sessionRepo.Count()
	return map[string]int64{
		"users":    userCount,
		"classes":  classCount,
		"sessions": sessionCount,
	}, nil
}
