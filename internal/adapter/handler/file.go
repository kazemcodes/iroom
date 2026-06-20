package handler

import (
	"strconv"

	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
)

type FileHandler struct {
	fileUC *usecase.FileUseCase
}

func NewFileHandler(fileUC *usecase.FileUseCase) *FileHandler {
	return &FileHandler{fileUC: fileUC}
}

func (h *FileHandler) Upload(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, _ := getUserID(c)

	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل ارسال نشده")
	}
	defer file.Close()

	f, err := h.fileUC.Upload(sessionID, userID, header.Filename, "", header.Size)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, f)
}

func (h *FileHandler) ListBySession(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	files, err := h.fileUC.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Success(c, map[string]interface{}{
		"items": files,
		"total": len(files),
	})
}

func (h *FileHandler) Download(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	file, err := h.fileUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "فایل یافت نشد")
	}
	return c.File(file.Filepath)
}

func (h *FileHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.fileUC.Delete(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Success(c, map[string]string{"message": "فایل حذف شد"})
}
