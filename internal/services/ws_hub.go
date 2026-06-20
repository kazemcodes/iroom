package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a single WebSocket connection for a user
type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	UserID      int64
	Email       string
	DisplayName string
	Role        string
	RoomID      string
}

// BroadcastMessage represents a message to be broadcast to clients
type BroadcastMessage struct {
	UserID  int64  // 0 = broadcast to all
	RoomID  string // empty = not room-specific
	Type    string // "notification", "presence", "chat"
	Payload []byte
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			slog.Info("client registered", "user_id", client.UserID, "email", client.Email)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			slog.Info("client unregistered", "user_id", client.UserID, "email", client.Email)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

// handleBroadcast processes a broadcast message
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.UserID != 0 {
		// Send to specific user
		if clients, ok := h.clients[msg.UserID]; ok {
			for client := range clients {
				select {
				case client.Send <- msg.Payload:
				default:
					// Client buffer full, skip
				}
			}
		}
	} else {
		// Broadcast to all users
		for _, clients := range h.clients {
			for client := range clients {
				select {
				case client.Send <- msg.Payload:
				default:
					// Client buffer full, skip
				}
			}
		}
	}
}

// BroadcastToUser sends a message to all connections of a specific user
func (h *Hub) BroadcastToUser(userID int64, msgType string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	})
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Type:    msgType,
		Payload: data,
	}
}

// BroadcastToAll sends a message to all connected users
func (h *Hub) BroadcastToAll(msgType string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	})
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID:  0,
		Type:    msgType,
		Payload: data,
	}
}

// BroadcastToRoom sends a message to all users in a specific room
func (h *Hub) BroadcastToRoom(roomID string, msgType string, payload interface{}, excludeUserID int64) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"room_id": roomID,
		"payload": payload,
	})
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	sent := 0
	for userID, clients := range h.clients {
		if userID == excludeUserID {
			continue
		}
		for client := range clients {
			if client.RoomID != "" && client.RoomID != roomID {
				continue
			}
			select {
			case client.Send <- data:
				sent++
			default:
			}
		}
	}
	slog.Debug("BroadcastToRoom", "room_id", roomID, "type", msgType, "sent_to", sent, "total_clients", h.getClientCount())
}

// GetOnlineUsers returns a list of currently connected user IDs
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline checks if a user has at least one active connection
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetClientCount returns the total number of connected clients
func (h *Hub) GetClientCount() int {
	return h.getClientCount()
}

func (h *Hub) getClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// BroadcastBinaryToRoom sends a binary message to all users in a room except the sender
func (h *Hub) BroadcastBinaryToRoom(roomID string, excludeUserID int64, data []byte) {
	msg := make([]byte, 1+len(data))
	msg[0] = 1
	copy(msg[1:], data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, clients := range h.clients {
		if userID == excludeUserID {
			continue
		}
		for client := range clients {
			if client.RoomID == roomID {
				select {
				case client.Send <- msg:
				default:
				}
			}
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	slog.Info("hub register", "user_id", client.UserID, "room_id", client.RoomID, "display_name", client.DisplayName)
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	slog.Info("hub unregister", "user_id", client.UserID, "room_id", client.RoomID)
	h.unregister <- client
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user_id", c.UserID)
			}
			break
		}
		// We don't expect messages from clients on this hub
		// This hub is for server-initiated notifications/presence only
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleWS handles WebSocket connections for notifications and presence
func (h *Hub) HandleWS(secret string) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenStr := c.QueryParam("token")
		if tokenStr == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "token required")
		}

		claims, err := jwt.Validate(secret, tokenStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			slog.Error("websocket upgrade failed", "error", err)
			return err
		}

		client := &Client{
			Hub:    h,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}

		h.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()

		return nil
	}
}
