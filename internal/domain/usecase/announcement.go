package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type AnnouncementUseCase struct {
	announcementRepo *repository.AnnouncementRepo
	classRepo        *repository.ClassRepo
}

func NewAnnouncementUseCase(announcementRepo *repository.AnnouncementRepo, classRepo *repository.ClassRepo) *AnnouncementUseCase {
	return &AnnouncementUseCase{announcementRepo: announcementRepo, classRepo: classRepo}
}

func (uc *AnnouncementUseCase) Create(classID, authorID int64, title, content string, isPinned, isSystemWide bool) (*entity.Announcement, error) {
	a := &entity.Announcement{
		ClassID:      classID,
		AuthorID:     authorID,
		Title:        title,
		Content:      content,
		IsPinned:     isPinned,
		IsSystemWide: isSystemWide,
	}
	if err := uc.announcementRepo.Create(a); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد اعلان")
	}
	return a, nil
}

func (uc *AnnouncementUseCase) GetByID(id int64) (*entity.Announcement, error) {
	return uc.announcementRepo.GetByID(id)
}

func (uc *AnnouncementUseCase) ListByClass(classID int64) ([]*entity.Announcement, int64, error) {
	return uc.announcementRepo.ListByClass(classID, 1, 100)
}

func (uc *AnnouncementUseCase) Update(id int64, title, content string) error {
	a, err := uc.announcementRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("اعلان یافت نشد")
	}
	if title != "" {
		a.Title = title
	}
	if content != "" {
		a.Content = content
	}
	return uc.announcementRepo.Update(a)
}

func (uc *AnnouncementUseCase) Delete(id int64) error {
	return uc.announcementRepo.Delete(id)
}

func (uc *AnnouncementUseCase) TogglePin(id int64) error {
	a, err := uc.announcementRepo.GetByID(id)
	if err != nil {
		return err
	}
	return uc.announcementRepo.SetPinned(id, !a.IsPinned)
}
