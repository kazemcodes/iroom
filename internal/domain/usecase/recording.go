package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type RecordingUseCase struct {
	recordingRepo *repository.RecordingRepo
	sessionRepo   *repository.SessionRepo
	uploadDir     string
}

func NewRecordingUseCase(recordingRepo *repository.RecordingRepo, sessionRepo *repository.SessionRepo, uploadDir string) *RecordingUseCase {
	return &RecordingUseCase{recordingRepo: recordingRepo, sessionRepo: sessionRepo, uploadDir: uploadDir}
}

func (uc *RecordingUseCase) Upload(sessionID, userID int64, filename, filepath string, filesize int64, duration int) (*entity.Recording, error) {
	r := &entity.Recording{
		SessionID:  sessionID,
		UploadedBy: userID,
		Filename:   filename,
		Filepath:   filepath,
		Filesize:   filesize,
		Duration:   duration,
		Status:     "ready",
	}
	if err := uc.recordingRepo.Create(r); err != nil {
		return nil, fmt.Errorf("failed to upload recording: %w", err)
	}
	return r, nil
}

func (uc *RecordingUseCase) GetByID(id int64) (*entity.Recording, error) {
	return uc.recordingRepo.GetByID(id)
}

func (uc *RecordingUseCase) ListBySession(sessionID int64) ([]entity.Recording, error) {
	return uc.recordingRepo.ListBySession(sessionID)
}

func (uc *RecordingUseCase) ListAll(page, perPage int) ([]entity.Recording, int64, error) {
	return uc.recordingRepo.ListAll(page, perPage, "")
}

func (uc *RecordingUseCase) Delete(id int64) error {
	return uc.recordingRepo.Delete(id)
}

func (uc *RecordingUseCase) Count() (int64, error) {
	return uc.recordingRepo.Count()
}
