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
	CheckOrigin: func(r *http.Request) bool { return true },
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
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients[sessionID] {
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)
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
	fileRepo *repository.FileRepo
	uploadDir string
}

func NewFileHandler(fileRepo *repository.FileRepo, uploadDir string) *FileHandler {
	return &FileHandler{fileRepo: fileRepo, uploadDir: uploadDir}
}

func (h *FileHandler) Upload(c echo.Context) error {
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

	userID := c.Get("user_id").(int64)
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

func (h *FileHandler) Download(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	file, err := h.fileRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "فایل یافت نشد")
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
