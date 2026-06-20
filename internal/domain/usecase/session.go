package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type SessionUseCase struct {
	sessionRepo *repository.SessionRepo
	roomRepo    *repository.RoomRepo
}

func NewSessionUseCase(sessionRepo *repository.SessionRepo, roomRepo *repository.RoomRepo) *SessionUseCase {
	return &SessionUseCase{sessionRepo: sessionRepo, roomRepo: roomRepo}
}

func (uc *SessionUseCase) Create(roomID int64, title, scheduledAt string, duration int) (*entity.Session, error) {
	if _, err := uc.roomRepo.GetByID(roomID); err != nil {
		return nil, errors.ErrNotFound
	}
	s := &entity.Session{
		RoomID:    roomID,
		Title:     title,
		Duration:  duration,
		Status:    entity.SessionScheduled,
	}
	if err := uc.sessionRepo.Create(s); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return s, nil
}

func (uc *SessionUseCase) GetByID(id int64) (*entity.Session, error) {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return s, nil
}

func (uc *SessionUseCase) List(roomID int64, page, perPage int, search string) ([]entity.Session, int64, error) {
	if roomID > 0 {
		sessions, err := uc.sessionRepo.ListByRoom(roomID)
		return sessions, int64(len(sessions)), err
	}
	return uc.sessionRepo.ListAll(page, perPage, search)
}

func (uc *SessionUseCase) checkPermission(s *entity.Session, userID int64, role string) error {
	if role == "admin" || role == "owner" {
		return nil
	}
	room, err := uc.roomRepo.GetByID(s.RoomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID == userID {
		return nil
	}
	return errors.ErrForbidden
}

func (uc *SessionUseCase) Start(id, userID int64, role string) (*entity.Session, error) {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return nil, err
	}

	roomName := fmt.Sprintf("room-%d", s.RoomID)
	if err := uc.sessionRepo.UpdateStatus(id, "live", roomName); err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	s.Status = entity.SessionLive
	s.LivekitRoom = roomName
	return s, nil
}

func (uc *SessionUseCase) End(id, userID int64, role string) error {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return err
	}

	return uc.sessionRepo.UpdateStatus(id, "ended", "")
}

func (uc *SessionUseCase) Delete(id, userID int64, role string) error {
	s, err := uc.sessionRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}

	if err := uc.checkPermission(s, userID, role); err != nil {
		return err
	}

	return uc.sessionRepo.Delete(id)
}

func (uc *SessionUseCase) Count() (int64, error) {
	return uc.sessionRepo.Count()
}

func (uc *SessionUseCase) CountActiveByRoom(roomID int64) (int, error) {
	return uc.sessionRepo.CountActiveByRoom(roomID)
}
