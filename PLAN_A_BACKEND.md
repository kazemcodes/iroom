# Plan A: Backend Security Hardening

**Agent: Backend Go**
**Files touched: handlers.go, chat.go, announcement.go, external.go, livekit.go, ticket.go, poll.go, ws_hub.go, middleware/auth.go, cmd/server/main.go**
**No overlap with Plan B**

---

## Task A1: Safe Type Assertions in All Handlers (Critical)
**File: `internal/handlers/handlers.go`**

Replace ALL unsafe `c.Get("user_id").(int64)` and `c.Get("role").(string)` with safe helper calls. The helpers `getUserID()` and `getUserRole()` already exist at ~line 566-574.

Pattern to replace everywhere:
```go
// BEFORE (panics on nil):
userID := c.Get("user_id").(int64)

// AFTER (safe):
userID, ok := c.Get("user_id").(int64)
if !ok {
    return response.Unauthorized(c, "احراز هویت نامعتبر")
}
```

There are ~55 instances across these files:
- `handlers.go` (~25 instances)
- `ticket.go` (~7 instances)
- `notification.go` (~4 instances)
- `chat.go` (~4 instances)
- `poll.go` (~3 instances)
- `webhook.go` (~6 instances)
- `announcement.go` (~4 instances)
- `recording.go` (~2 instances)
- `livekit.go` (~1 instance)

**Important**: Some handlers already check role and can use `getUserRole()` instead of `c.Get("role").(string)`. Use the safe pattern for ALL of them.

---

## Task A2: Chat WebSocket Message Length Limit (Important)
**File: `internal/handlers/chat.go`**

Add a max message length check in the WebSocket chat handler's read loop (~line 97-128). Before saving to database, validate:
```go
if len(msg.Content) > 10000 {
    continue // skip oversized messages
}
```

Also add `conn.SetReadLimit(10240)` before the read loop starts (after line 91, after `conn` is created).

---

## Task A3: Announcement List Authorization (Important)
**File: `internal/handlers/announcement.go`**

In `AnnouncementHandler.List` (~line 87), add authorization check before returning announcements:
```go
// After getting classID from params, verify user has access:
role := getUserRole(c)
if role != "admin" {
    class, err := h.classRepo.GetByID(classID)
    if err != nil {
        return response.NotFound(c, "کلاس یافت نشد")
    }
    userID := getUserID(c)  // need to get from context safely first
    if class.TeacherID != userID && !h.classRepo.IsEnrolled(class.ID, userID) {
        return response.Forbidden(c, "دسترسی غیرمجاز")
    }
}
```

This requires `AnnouncementHandler` to have access to `classRepo`. Check if it already has it. If not, add it to the struct and update `NewAnnouncementHandler` in `main.go`.

---

## Task A4: External API Input Validation (Important)
**File: `internal/handlers/external.go`**

Add validation to `CreateUser` (~line 30):
```go
if len(req.Password) < 6 {
    return response.BadRequest(c, "رمز عبور باید حداقل ۶ کاراکتر باشد")
}
```

Add validation to `CreateSession` (~line 108) to verify class exists:
```go
_, err := h.classRepo.GetByID(req.ClassID)
if err != nil {
    return response.NotFound(c, "کلاس یافت نشد")
}
```

Fix `CreateClass` default teacher ID (~line 86):
```go
if req.TeacherID == 0 {
    return response.BadRequest(c, "شناسه مدرس الزامی است")
}
```
Remove the `req.TeacherID = 1` fallback.

---

## Task A5: Chat Broadcast Dead Connection Cleanup (Important)
**File: `internal/handlers/chat.go`**

In `ChatHub.Broadcast` (~line 54), check `WriteMessage` error and remove dead connections:
```go
func (h *ChatHub) Broadcast(sessionID int64, msg interface{}) {
    h.mu.Lock() // upgrade to write lock
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
```

---

## Task A6: LiveKit Handler Safe Type Assertion + Auth on Webhook (Important)
**File: `internal/handlers/livekit.go`**

Fix line 39-40 safe type assertion:
```go
userID, ok := c.Get("user_id").(int64)
if !ok {
    return response.Unauthorized(c, "احراز هویت نامعتبر")
}
role, _ := c.Get("role").(string)
```

Also need to set `display_name` in the Auth middleware. Read `middleware/auth.go` and add after line 31:
```go
// In the Auth middleware, after setting role, fetch display_name from DB if needed
// OR better: add display_name to JWT claims
```

Simplest fix: fetch user from DB in the LiveKit handler and use the display name from there instead of `c.Get("display_name")`.

---

## Task A7: CreateUser Validation + DashboardStats Error Handling (Important)
**File: `internal/handlers/handlers.go`**

In `AdminHandler.CreateUser` (~line 654), add validation:
```go
if req.Email == "" || req.Password == "" || req.DisplayName == "" {
    return response.BadRequest(c, "ایمیل، رمز عبور و نام الزامی هستند")
}
if len(req.Password) < 6 {
    return response.BadRequest(c, "رمز عبور باید حداقل ۶ کاراکتر باشد")
}
```

In `DashboardStats` (~line 615), check errors:
```go
userCount, err := h.userRepo.Count()
if err != nil {
    return response.InternalError(c, "خطا در دریافت آمار")
}
// same for other counts
```

---

## Task A8: Unauthenticated Endpoint Security (Important)
**File: `cmd/server/main.go`**

Add API key auth to the external webhook receiver (~line 205):
```go
e.POST("/api/v1/webhooks", middleware.APIKeyAuth(cfg.External.APIKey), externalHandler.HandleWebhook)
```

Add rate limiting to WebSocket endpoints (~line 167, 170):
```go
wsGroup := e.Group("/ws")
wsGroup.Use(middleware.RateLimit(30, time.Minute))
wsGroup.GET("/sessions/:id", chatHandler.HandleWS)
wsGroup.GET("", wsHub.HandleWS(cfg.JWT.Secret))
```

---

## Task A9: Admin Handler perPage Cap in Remaining Files
**File: `internal/handlers/ticket.go`, `internal/handlers/webhook.go`, `internal/handlers/announcement.go`**

Add `if perPage > 100 { perPage = 100 }` after each `perPage` default in:
- `ticket.go` ListMy (~line 86) and ListAll in AdminTicketHandler (~line 233)
- `webhook.go` ListDeliveries (~line 211)
- `announcement.go` List (~line 98)

---

## Task A10: BroadcastToRoom Room Filtering (Important)
**File: `internal/services/ws_hub.go`**

In `BroadcastToRoom` (~line 156), the `roomID` parameter is unused for filtering. Add room tracking to the Client struct and filter by room when broadcasting. Minimal fix: add a `RoomID` field to Client and set it when joining, then filter in BroadcastToRoom.

---

## Verification

After all changes:
```bash
go build ./...
go vet ./...
```
