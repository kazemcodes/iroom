package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type RecordingHandler struct {
	recordingRepo *repository.RecordingRepo
	uploadDir     string
}

func NewRecordingHandler(recordingRepo *repository.RecordingRepo, uploadDir string) *RecordingHandler {
	return &RecordingHandler{recordingRepo: recordingRepo, uploadDir: uploadDir}
}

func (h *RecordingHandler) Upload(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل ارائه نشده")
	}

	src, err := file.Open()
	if err != nil {
		return response.InternalError(c, "خطا در خواندن فایل")
	}
	defer src.Close()

	recDir := filepath.Join(h.uploadDir, "recordings", strconv.FormatInt(sessionID, 10))
	if err := os.MkdirAll(recDir, 0755); err != nil {
		return response.InternalError(c, "خطا در ایجاد پوشه")
	}

	dstPath := filepath.Join(recDir, file.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return response.InternalError(c, "خطا در ذخیره فایل")
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return response.InternalError(c, "خطا در کپی فایل")
	}

	duration := 0
	if d := c.FormValue("duration"); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			duration = v
		}
	}

	userID := c.Get("user_id").(int64)
	rec := &models.Recording{
		SessionID:  sessionID,
		UploadedBy: userID,
		Filename:   file.Filename,
		Filepath:   dstPath,
		Filesize:   written,
		Duration:   duration,
		Status:     "ready",
	}

	if err := h.recordingRepo.Create(rec); err != nil {
		return response.InternalError(c, "خطا در ذخیره اطلاعات")
	}

	return response.Created(c, rec)
}

func (h *RecordingHandler) Download(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	rec, err := h.recordingRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "ضبط یافت نشد")
	}

	return c.File(rec.Filepath)
}

func (h *RecordingHandler) ListBySession(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	recordings, err := h.recordingRepo.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت ضبط‌ها")
	}
	if recordings == nil {
		recordings = []models.Recording{}
	}
	return response.Success(c, recordings)
}
