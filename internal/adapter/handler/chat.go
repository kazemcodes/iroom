package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		host := r.Host
		return origin == "http://"+host || origin == "https://"+host
	},
}

type ChatHandler struct {
	messageRepo interface {
		Create(m *entity.Message) error
	}
	hub    *services.Hub
	secret string
}

func NewChatHandler(messageRepo interface {
	Create(m *entity.Message) error
}, secret string) *ChatHandler {
	return &ChatHandler{messageRepo: messageRepo, hub: services.NewHub(), secret: secret}
}

func (h *ChatHandler) GetHub() *services.Hub {
	return h.hub
}

func (h *ChatHandler) HandleWS(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid session id"})
	}

	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return c.JSON(401, map[string]string{"error": "token required"})
	}

	claims, err := jwt.Validate(h.secret, tokenStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "invalid token"})
	}

	conn, err := chatUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return err
	}

	client := &services.Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RoomID: strconv.FormatInt(sessionID, 10),
	}

	h.hub.Register(client)

	go h.writePump(client)
	h.readPump(client, sessionID)

	return nil
}

func (h *ChatHandler) readPump(client *services.Client, sessionID int64) {
	defer func() {
		h.hub.Unregister(client)
		client.Conn.Close()
	}()

	for {
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user_id", client.UserID)
			}
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

			chatMsg := &entity.Message{
				SessionID: sessionID,
				UserID:    client.UserID,
				Content:   msg.Content,
				Type:      "text",
			}
			h.messageRepo.Create(chatMsg)

			broadcast := map[string]interface{}{
				"type": "message",
				"message": map[string]interface{}{
					"user_id":           client.UserID,
					"user_display_name": client.Email,
					"content":           msg.Content,
					"created_at":        "",
				},
			}
			data, _ := json.Marshal(broadcast)
			h.hub.BroadcastToRoom(strconv.FormatInt(sessionID, 10), "chat", data, 0)
		}
	}
}

func (h *ChatHandler) writePump(client *services.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
