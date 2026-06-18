package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type FileUseCase struct {
	fileRepo    *repository.FileRepo
	sessionRepo *repository.SessionRepo
	classRepo   *repository.ClassRepo
	uploadDir   string
}

func NewFileUseCase(fileRepo *repository.FileRepo, sessionRepo *repository.SessionRepo, classRepo *repository.ClassRepo, uploadDir string) *FileUseCase {
	return &FileUseCase{fileRepo: fileRepo, sessionRepo: sessionRepo, classRepo: classRepo, uploadDir: uploadDir}
}

func (uc *FileUseCase) Upload(sessionID, userID int64, filename, filepath string, filesize int64) (*entity.File, error) {
	f := &entity.File{
		SessionID:  sessionID,
		UploadedBy: userID,
		Filename:   filename,
		Filepath:   filepath,
		Filesize:   filesize,
	}
	if err := uc.fileRepo.Create(f); err != nil {
		return nil, fmt.Errorf("خطا در آپلود فایل")
	}
	return f, nil
}

func (uc *FileUseCase) GetByID(id int64) (*entity.File, error) {
	return uc.fileRepo.GetByID(id)
}

func (uc *FileUseCase) ListBySession(sessionID int64) ([]entity.File, error) {
	return uc.fileRepo.ListBySession(sessionID)
}

func (uc *FileUseCase) Delete(id int64) error {
	return uc.fileRepo.Delete(id)
}
