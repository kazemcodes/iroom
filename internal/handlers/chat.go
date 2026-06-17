package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		host := r.Host
		return origin == "http://"+host || origin == "https://"+host
	},
}

type ChatHub struct {
	mu      sync.RWMutex
	clients map[int64]map[*websocket.Conn]bool
}

var chatHub = &ChatHub{
	clients: make(map[int64]map[*websocket.Conn]bool),
}

func (h *ChatHub) Add(sessionID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[sessionID] == nil {
		h.clients[sessionID] = make(map[*websocket.Conn]bool)
	}
	h.clients[sessionID][conn] = true
}

func (h *ChatHub) Remove(sessionID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[sessionID] != nil {
		delete(h.clients[sessionID], conn)
		if len(h.clients[sessionID]) == 0 {
			delete(h.clients, sessionID)
		}
	}
}

func (h *ChatHub) Broadcast(sessionID int64, msg interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, _ := json.Marshal(msg)
	for conn := range h.clients[sessionID] {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(h.clients[sessionID], conn)
		}
	}
	if len(h.clients[sessionID]) == 0 {
		delete(h.clients, sessionID)
	}
}

type ChatHandler struct {
	messageRepo *repository.MessageRepo
	secret      string
}

func NewChatHandler(messageRepo *repository.MessageRepo, secret string) *ChatHandler {
	return &ChatHandler{messageRepo: messageRepo, secret: secret}
}

func (h *ChatHandler) HandleWS(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return response.Unauthorized(c, "توکن ارائه نشده")
	}

	claims, err := jwt.Validate(h.secret, tokenStr)
	if err != nil {
		return response.Unauthorized(c, "توکن نامعتبر")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetReadLimit(10240)

	chatHub.Add(sessionID, conn)
	defer chatHub.Remove(sessionID, conn)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		if msg.Type == "message" && msg.Content != "" {
			if len(msg.Content) > 10000 {
				continue
			}
			chatMsg := &models.Message{
				SessionID: sessionID,
				UserID:    claims.UserID,
				Content:   msg.Content,
				Type:      "text",
			}
			if err := h.messageRepo.Create(chatMsg); err != nil {
				slog.Error("failed to save message", "error", err)
				continue
			}

			chatHub.Broadcast(sessionID, map[string]interface{}{
				"type":    "message",
				"message": chatMsg,
			})
		}
	}

	return nil
}

type FileHandler struct {
	fileRepo    *repository.FileRepo
	sessionRepo *repository.SessionRepo
	classRepo   *repository.ClassRepo
	uploadDir   string
}

func NewFileHandler(fileRepo *repository.FileRepo, sessionRepo *repository.SessionRepo, classRepo *repository.ClassRepo, uploadDir string) *FileHandler {
	return &FileHandler{fileRepo: fileRepo, sessionRepo: sessionRepo, classRepo: classRepo, uploadDir: uploadDir}
}

func (h *FileHandler) Upload(c echo.Context) error {
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
		if err != nil {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
		if class.TeacherID != userID && !h.classRepo.IsEnrolled(class.ID, userID) {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل ارائه نشده")
	}

	if file.Size > 50<<20 {
		return response.BadRequest(c, "حجم فایل بیش از حد مجاز است (حداکثر ۵۰ مگابایت)")
	}

	// Validate file type by extension
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true, // images
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true, // documents
		".mp4": true, ".avi": true, ".mov": true, ".mkv": true, ".webm": true, // videos
		".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true, // archives
		".mp3": true, ".wav": true, ".ogg": true, ".flac": true, // audio
	}
	if !allowedExts[ext] {
		return response.BadRequest(c, "نوع فایل مجاز نیست")
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

	sessionDir := filepath.Join(h.uploadDir, strconv.FormatInt(sessionID, 10))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return response.InternalError(c, "خطا در ایجاد پوشه")
	}

	dstPath := filepath.Join(sessionDir, safeName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return response.InternalError(c, "خطا در ذخیره فایل")
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return response.InternalError(c, "خطا در کپی فایل")
	}

	fileModel := &models.File{
		SessionID:  sessionID,
		UploadedBy: userID,
		Filename:   file.Filename,
		Filepath:   dstPath,
		Filesize:   written,
	}

	if err := h.fileRepo.Create(fileModel); err != nil {
		return response.InternalError(c, "خطا در ذخیره اطلاعات فایل")
	}

	return response.Created(c, fileModel)
}

func (h *FileHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	file, err := h.fileRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "فایل یافت نشد")
	}

	// Check ownership or admin role
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)
	if file.UploadedBy != userID && role != "admin" {
		return response.Forbidden(c, "شما اجازه حذف این فایل را ندارید")
	}

	// Delete physical file
	if err := os.Remove(file.Filepath); err != nil && !os.IsNotExist(err) {
		return response.InternalError(c, "خطا در حذف فایل")
	}

	// Delete database record
	if err := h.fileRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف رکورد فایل")
	}

	return response.Success(c, map[string]string{"message": "فایل با موفقیت حذف شد"})
}

func (h *FileHandler) Download(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	file, err := h.fileRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "فایل یافت نشد")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	if file.UploadedBy != userID && role != "admin" {
		session, err := h.sessionRepo.GetByID(file.SessionID)
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

	return c.File(file.Filepath)
}

func (h *FileHandler) ListBySession(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	files, err := h.fileRepo.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت فایل‌ها")
	}
	if files == nil {
		files = []models.File{}
	}
	return response.Success(c, files)
}
