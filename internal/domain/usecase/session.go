package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type SessionUseCase struct {
	sessionRepo *repository.SessionRepo
	classRepo   *repository.ClassRepo
}

func NewSessionUseCase(sessionRepo *repository.SessionRepo, classRepo *repository.ClassRepo) *SessionUseCase {
	return &SessionUseCase{sessionRepo: sessionRepo, classRepo: classRepo}
}

func (uc *SessionUseCase) Create(classID int64, title, scheduledAt string, duration int) (*entity.Session, error) {
	s := &entity.Session{
		ClassID:     classID,
		Title:       title,
		Duration:    duration,
		Status:      entity.SessionScheduled,
	}
	if err := uc.sessionRepo.Create(s); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد جلسه")
	}
	return s, nil
}

func (uc *SessionUseCase) GetByID(id int64) (*entity.Session, error) {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("جلسه یافت نشد")
	}
	return s, nil
}

func (uc *SessionUseCase) List(classID int64, page, perPage int, search string) ([]entity.Session, int64, error) {
	if classID > 0 {
		sessions, err := uc.sessionRepo.ListByClass(classID)
		return sessions, int64(len(sessions)), err
	}
	return uc.sessionRepo.ListAll(page, perPage, search)
}

func (uc *SessionUseCase) checkPermission(s *entity.Session, userID int64, role string) error {
	if role == "admin" || role == "owner" {
		return nil
	}
	class, err := uc.classRepo.GetByID(s.ClassID)
	if err != nil {
		return nil
	}
	if class.TeacherID == userID {
		return nil
	}
	return fmt.Errorf("شما اجازه انجام این عملیات را ندارید")
}

func (uc *SessionUseCase) Start(id, userID int64, role string) (*entity.Session, error) {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("جلسه یافت نشد")
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return nil, err
	}

	roomName := fmt.Sprintf("room-%d", s.ID)
	if err := uc.sessionRepo.UpdateStatus(id, "live", roomName); err != nil {
		return nil, fmt.Errorf("خطا در شروع جلسه")
	}

	s.Status = entity.SessionLive
	s.LivekitRoom = roomName
	return s, nil
}

func (uc *SessionUseCase) End(id, userID int64, role string) error {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("جلسه یافت نشد")
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return err
	}

	return uc.sessionRepo.UpdateStatus(id, "ended", "")
}

func (uc *SessionUseCase) Delete(id, userID int64, role string) error {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("جلسه یافت نشد")
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return err
	}

	return uc.sessionRepo.Delete(id)
}

func (uc *SessionUseCase) Count() (int64, error) {
	return uc.sessionRepo.Count()
}
