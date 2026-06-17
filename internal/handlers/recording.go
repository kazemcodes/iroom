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
	sessionRepo   *repository.SessionRepo
	classRepo     *repository.ClassRepo
	uploadDir     string
}

func NewRecordingHandler(recordingRepo *repository.RecordingRepo, sessionRepo *repository.SessionRepo, classRepo *repository.ClassRepo, uploadDir string) *RecordingHandler {
	return &RecordingHandler{recordingRepo: recordingRepo, sessionRepo: sessionRepo, classRepo: classRepo, uploadDir: uploadDir}
}

func (h *RecordingHandler) Upload(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	if role != "admin" {
		session, err := h.sessionRepo.GetByID(sessionID)
		if err != nil {
			return response.NotFound(c, "جلسه یافت نشد")
		}
		class, err := h.classRepo.GetByID(session.ClassID)
		if err != nil || class.TeacherID != userID {
			return response.Forbidden(c, "شما اجازه آپلود ضبط در این جلسه را ندارید")
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل ارائه نشده")
	}

	if file.Size > 500<<20 {
		return response.BadRequest(c, "حجم فایل بیش از حد مجاز است (حداکثر ۵۰۰ مگابایت)")
	}

	src, err := file.Open()
	if err != nil {
		return response.InternalError(c, "خطا در خواندن فایل")
	}
	defer src.Close()

	safeName := filepath.Base(file.Filename)
	if safeName == "." || safeName == ".." || safeName == "/" {
		return response.BadRequest(c, "نام فایل نامعتبر")
	}

	recDir := filepath.Join(h.uploadDir, "recordings", strconv.FormatInt(sessionID, 10))
	if err := os.MkdirAll(recDir, 0755); err != nil {
		return response.InternalError(c, "خطا در ایجاد پوشه")
	}

	dstPath := filepath.Join(recDir, safeName)
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

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	if rec.UploadedBy != userID && role != "admin" {
		session, err := h.sessionRepo.GetByID(rec.SessionID)
		if err != nil {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
		class, err := h.classRepo.GetByID(session.ClassID)
		if err != nil {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
		if class.TeacherID != userID && !h.classRepo.IsEnrolled(class.ID, userID) {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
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
