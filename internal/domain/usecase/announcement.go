package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type AnnouncementUseCase struct {
	announcementRepo *repository.AnnouncementRepo
	roomRepo         *repository.RoomRepo
}

func NewAnnouncementUseCase(announcementRepo *repository.AnnouncementRepo, roomRepo *repository.RoomRepo) *AnnouncementUseCase {
	return &AnnouncementUseCase{announcementRepo: announcementRepo, roomRepo: roomRepo}
}

func (uc *AnnouncementUseCase) Create(roomID, authorID int64, title, content string, isPinned, isSystemWide bool) (*entity.Announcement, error) {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	if room.OwnerID != authorID {
		return nil, errors.ErrForbidden
	}
	a := &entity.Announcement{
		RoomID:       roomID,
		AuthorID:     authorID,
		Title:        title,
		Content:      content,
		IsPinned:     isPinned,
		IsSystemWide: isSystemWide,
	}
	if err := uc.announcementRepo.Create(a); err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}
	return a, nil
}

func (uc *AnnouncementUseCase) GetByID(id int64) (*entity.Announcement, error) {
	return uc.announcementRepo.GetByID(id)
}

func (uc *AnnouncementUseCase) ListByRoom(roomID int64) ([]*entity.Announcement, int64, error) {
	return uc.announcementRepo.ListByRoom(roomID, 1, 100)
}

func (uc *AnnouncementUseCase) checkOwnership(id, actorID int64, role string) error {
	if role == "admin" {
		return nil
	}
	a, err := uc.announcementRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}
	room, err := uc.roomRepo.GetByID(a.RoomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID {
		return errors.ErrForbidden
	}
	return nil
}

func (uc *AnnouncementUseCase) Update(id, actorID int64, role, title, content string) error {
	if err := uc.checkOwnership(id, actorID, role); err != nil {
		return err
	}
	a, err := uc.announcementRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}
	if title != "" {
		a.Title = title
	}
	if content != "" {
		a.Content = content
	}
	return uc.announcementRepo.Update(a)
}

func (uc *AnnouncementUseCase) Delete(id, actorID int64, role string) error {
	if err := uc.checkOwnership(id, actorID, role); err != nil {
		return err
	}
	return uc.announcementRepo.Delete(id)
}

func (uc *AnnouncementUseCase) TogglePin(id, actorID int64, role string) error {
	if err := uc.checkOwnership(id, actorID, role); err != nil {
		return err
	}
	a, err := uc.announcementRepo.GetByID(id)
	if err != nil {
		return err
	}
	return uc.announcementRepo.SetPinned(id, !a.IsPinned)
}
