package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

// RecordingHandler handles HTTP requests for session recordings.
// Routes: POST /sessions/:id/recordings, GET /sessions/:id/recordings
//         GET /recordings/:id/download, DELETE /recordings/:id
type RecordingHandler struct {
	recordingUC *usecase.RecordingUseCase
}

func NewRecordingHandler(recordingUC *usecase.RecordingUseCase) *RecordingHandler {
	return &RecordingHandler{recordingUC: recordingUC}
}

func (h *RecordingHandler) Upload(c echo.Context) error {
	sessionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := getUserID(c)

	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل ارسال نشده")
	}
	defer file.Close()

	r, err := h.recordingUC.Upload(sessionID, userID, header.Filename, "", header.Size, 0)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, r)
}

func (h *RecordingHandler) ListBySession(c echo.Context) error {
	sessionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordings, err := h.recordingUC.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, recordings)
}

func (h *RecordingHandler) ListAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	recordings, total, err := h.recordingUC.ListAll(page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": recordings,
		"total": total,
	})
}

func (h *RecordingHandler) Download(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recording, err := h.recordingUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "ضبط یافت نشد")
	}
	return c.File(recording.Filepath)
}

func (h *RecordingHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.recordingUC.Delete(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "ضبط حذف شد"})
}
