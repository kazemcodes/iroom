package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/debug"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
)

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// autoEndTimeUnit is the time unit multiplier for auto-end countdowns.
// Defaults to time.Minute in production; tests can override it to time.Millisecond.
var autoEndTimeUnit = time.Minute

type autoEndTimer struct {
	cancel    context.CancelFunc
	sessionID int64
}

type ChatHandler struct {
	messageRepo interface {
		Create(m *entity.Message) error
	}
	userRepo interface {
		GetByID(id int64) (*entity.User, error)
	}
	sessionRepo interface {
		GetRoomBySessionID(sessionID int64) (*entity.Room, error)
		GetAutoEndMinutesBySessionID(sessionID int64) int
		IsSessionLive(sessionID int64) bool
	}
	sessionUC interface {
		AutoEnd(id int64) error
	}
	hub          *services.Hub
	secret       string
	autoEndMu    sync.Mutex
	autoEndTimers map[string]*autoEndTimer // keyed by roomID
}

func NewChatHandler(messageRepo interface {
	Create(m *entity.Message) error
}, userRepo interface {
	GetByID(id int64) (*entity.User, error)
}, sessionRepo interface {
	GetRoomBySessionID(sessionID int64) (*entity.Room, error)
	GetAutoEndMinutesBySessionID(sessionID int64) int
	IsSessionLive(sessionID int64) bool
}, sessionUC interface {
	AutoEnd(id int64) error
}, secret string, hub *services.Hub) *ChatHandler {
	return &ChatHandler{
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		sessionUC:     sessionUC,
		hub:           hub,
		secret:        secret,
		autoEndTimers: make(map[string]*autoEndTimer),
	}
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

	if user, err := h.userRepo.GetByID(claims.UserID); err == nil && user != nil {
		client.DisplayName = user.DisplayName
	}
	if client.DisplayName == "" {
		client.DisplayName = claims.Email
	}

	// Determine if this client is an operator for the room
	room, _ := h.sessionRepo.GetRoomBySessionID(sessionID)
	client.IsOperator = claims.Role == "admin" || claims.Role == "operator" || (room != nil && room.OwnerID == claims.UserID)		// If this is the first operator, broadcast operator_joined and cancel any pending auto-end
	if client.IsOperator && h.hub.GetOperatorConnectionCount(client.RoomID) == 0 {
		debug.Log("operator joined", "user_id", client.UserID, "session_id", sessionID)
		// Cancel any pending auto-end timer for this room
		h.cancelAutoEnd(client.RoomID)
		h.hub.BroadcastToRoom(client.RoomID, "chat", map[string]interface{}{
			"type":    "command",
			"command": "operator_joined",
		}, 0)
	}

	debug.Log("chat WS connect",
		"user_id", client.UserID,
		"display_name", client.DisplayName,
		"session_id", sessionID,
	)

	h.hub.Register(client)

	go h.writePump(client)
	h.readPump(client, sessionID)

	return nil
}

func (h *ChatHandler) readPump(client *services.Client, sessionID int64) {
	defer func() {
		debug.Log("chat WS disconnect", "user_id", client.UserID, "session_id", sessionID)
		// If this is the last operator connection, broadcast operator_left and start auto-end timer
		if client.IsOperator && h.hub.GetOperatorConnectionCount(client.RoomID) == 1 {
			debug.Log("operator left", "user_id", client.UserID, "session_id", sessionID)
			h.hub.BroadcastToRoom(client.RoomID, "chat", map[string]interface{}{
				"type":    "command",
				"command": "operator_left",
			}, 0)
			// Start auto-end countdown if configured
			h.startAutoEnd(client.RoomID, sessionID)
		}
		h.hub.Unregister(client)
		client.Conn.Close()
	}()

	for {
		msgType, raw, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user_id", client.UserID)
			}
			break
		}

		// Binary messages = media chunks — relay to room
		if msgType == websocket.BinaryMessage {
			h.hub.BroadcastBinaryToRoom(strconv.FormatInt(sessionID, 10), client.UserID, raw)
			continue
		}

		debug.Log("chat WS recv", "user_id", client.UserID, "raw", string(raw))

		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Command string `json:"command"`
			Private bool   `json:"private"`
			ReplyTo *struct {
				Sender  string `json:"sender"`
				Content string `json:"content"`
			} `json:"reply_to"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			debug.Log("chat WS unmarshal error", "error", err, "raw", string(raw))
			continue
		}

		if msg.Type == "command" && msg.Command != "" {
			debug.Log("chat command", "user_id", client.UserID, "command", msg.Command)

			broadcast := map[string]interface{}{
				"type":    "command",
				"command": msg.Command,
				"user_id": client.UserID,
			}
			if msg.Command == "kick" {
				var kickMsg struct {
					Command string `json:"command"`
					UserID  string `json:"user_id"`
				}
				if err := json.Unmarshal(raw, &kickMsg); err == nil && kickMsg.UserID != "" {
					broadcast["user_id"] = kickMsg.UserID
				}
			}
			h.hub.BroadcastToRoom(strconv.FormatInt(sessionID, 10), "chat", broadcast, 0)
			continue
		}

		if msg.Type == "whiteboard" {
			var wbMsg struct {
				Action string  `json:"action"`
				Show   *bool   `json:"show"`
				X1     float64 `json:"x1"`
				Y1     float64 `json:"y1"`
				X2     float64 `json:"x2"`
				Y2     float64 `json:"y2"`
				Color  string  `json:"color"`
				Width  float64 `json:"width"`
			}
			if err := json.Unmarshal(raw, &wbMsg); err == nil {
				broadcast := map[string]interface{}{
					"type":   "whiteboard",
					"action": wbMsg.Action,
				}
				if wbMsg.Show != nil {
					broadcast["show"] = *wbMsg.Show
				}
				broadcast["x1"] = wbMsg.X1
				broadcast["y1"] = wbMsg.Y1
				broadcast["x2"] = wbMsg.X2
				broadcast["y2"] = wbMsg.Y2
				broadcast["color"] = wbMsg.Color
				broadcast["width"] = wbMsg.Width
				h.hub.BroadcastToRoom(strconv.FormatInt(sessionID, 10), "chat", broadcast, 0)
			}
			continue
		}

		if msg.Type == "message" && msg.Content != "" {
			if len(msg.Content) > 10000 {
				debug.Log("chat message too long", "user_id", client.UserID, "len", len(msg.Content))
				continue
			}

			chatMsg := &entity.Message{
				SessionID: sessionID,
				UserID:    client.UserID,
				Content:   msg.Content,
				Type:      "text",
				CreatedAt: time.Now(),
			}
			if err := h.messageRepo.Create(chatMsg); err != nil {
				debug.Log("chat message save error", "error", err)
			}

			broadcast := map[string]interface{}{
				"type": "message",
				"message": map[string]interface{}{
					"user_id":           client.UserID,
					"user_display_name": client.DisplayName,
					"content":           msg.Content,
					"created_at":        time.Now().Format(time.RFC3339),
					"is_private":        msg.Private,
				},
			}
			if msg.ReplyTo != nil {
				broadcast["message"].(map[string]interface{})["reply_to"] = msg.ReplyTo
			}

			debug.Log("chat broadcast",
				"user_id", client.UserID,
				"session_id", sessionID,
				"content", msg.Content,
			)
			h.hub.BroadcastToRoom(strconv.FormatInt(sessionID, 10), "chat", broadcast, 0)
		}
	}
}

// startAutoEnd begins a countdown to auto-end the session when the last operator leaves.
func (h *ChatHandler) startAutoEnd(roomID string, sessionID int64) {
	// Only auto-end live sessions
	if !h.sessionRepo.IsSessionLive(sessionID) {
		return
	}

	minutes := h.sessionRepo.GetAutoEndMinutesBySessionID(sessionID)
	if minutes <= 0 {
		return
	}

	h.autoEndMu.Lock()
	// Cancel any existing timer for this room (shouldn't happen, but be safe)
	if existing, ok := h.autoEndTimers[roomID]; ok {
		existing.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.autoEndTimers[roomID] = &autoEndTimer{cancel: cancel, sessionID: sessionID}
	h.autoEndMu.Unlock()

	debug.Log("auto-end started", "room_id", roomID, "session_id", sessionID, "minutes", minutes)

	// Notify users about the auto-end countdown
	h.hub.BroadcastToRoom(roomID, "chat", map[string]interface{}{
		"type":    "command",
		"command": "auto_end_warning",
		"minutes": minutes,
	}, 0)

	go func() {
		select {
		case <-ctx.Done():
			// Cancelled — operator rejoined
			debug.Log("auto-end cancelled", "room_id", roomID, "session_id", sessionID)
		case <-time.After(time.Duration(minutes) * autoEndTimeUnit):
			// Timeout — end the session
			debug.Log("auto-end firing", "room_id", roomID, "session_id", sessionID)
			if err := h.sessionUC.AutoEnd(sessionID); err != nil {
				slog.Error("auto-end session failed", "error", err, "session_id", sessionID)
			}
			// Notify all users that the session has ended
			h.hub.BroadcastToRoom(roomID, "chat", map[string]interface{}{
				"type":    "command",
				"command": "session_ended",
			}, 0)
			// Clean up the timer
			h.autoEndMu.Lock()
			delete(h.autoEndTimers, roomID)
			h.autoEndMu.Unlock()
		}
	}()
}

// cancelAutoEnd cancels a pending auto-end timer for the given room.
func (h *ChatHandler) cancelAutoEnd(roomID string) {
	h.autoEndMu.Lock()
	timer, ok := h.autoEndTimers[roomID]
	if ok {
		timer.cancel()
		delete(h.autoEndTimers, roomID)
	}
	h.autoEndMu.Unlock()

	if ok {
		// Notify users that the auto-end has been cancelled
		h.hub.BroadcastToRoom(roomID, "chat", map[string]interface{}{
			"type":    "command",
			"command": "auto_end_cancelled",
		}, 0)
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

			if len(message) > 1 && message[0] == 1 {
				if err := client.Conn.WriteMessage(websocket.BinaryMessage, message[1:]); err != nil {
					return
				}
				continue
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				next := <-client.Send
				if len(next) > 1 && next[0] == 1 {
					w.Close()
					client.Conn.WriteMessage(websocket.BinaryMessage, next[1:])
					w, _ = client.Conn.NextWriter(websocket.TextMessage)
					continue
				}
				w.Write([]byte{'\n'})
				w.Write(next)
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

