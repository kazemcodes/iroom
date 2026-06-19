# Room Feature Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Merge کلاس‌ها and جلسات into a unified "room" concept where admins create persistent rooms, manage users per-room, and share join links. No self-signup for ordinary users.

**Architecture:** New `rooms` table replaces both `classes` and `class_students`. Sessions remain as live meetings within rooms. Guest login becomes room-scoped with a toggle per room.

**Tech Stack:** Go (Echo), SQLite, SvelteKit 5, Tailwind CSS

---

## Task 1: Database Migration — Rooms Table

**Covers:** Data model redesign

**Files:**
- Create: `internal/database/migrations/018_rooms.sql`

- [ ] **Step 1: Create migration file**

```sql
-- 018_rooms.sql
-- New rooms table replacing classes + class_students

CREATE TABLE IF NOT EXISTS rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    color TEXT DEFAULT '#3B82F6',
    slug TEXT UNIQUE,
    guest_login_enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS room_users (
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'student' CHECK(role IN ('teacher', 'student')),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_rooms_slug ON rooms(slug);
CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_room_users_user ON room_users(user_id);
```

- [ ] **Step 2: Verify migration runs**

Run the server and check that the tables are created. No manual step needed — migrations auto-run on startup via `database.New()`.

---

## Task 2: Backend — Room Entity

**Covers:** Data model redesign

**Files:**
- Create: `internal/domain/entity/room.go`

- [ ] **Step 1: Create Room entity**

```go
package entity

import "time"

// Room represents a persistent virtual classroom.
// Replaces the old Class entity. Rooms are owned by admins and contain
// assigned users and live sessions.
//
// Business rules:
//   - Only admins can create/modify rooms
//   - guest_login_enabled controls whether guests can join via link
//   - slug is URL-friendly identifier for /room/:slug links
type Room struct {
    ID                 int64     `json:"id" db:"id"`
    OwnerID            int64     `json:"owner_id" db:"owner_id"`
    Name               string    `json:"name" db:"name"`
    Description        string    `json:"description" db:"description"`
    Color              string    `json:"color" db:"color"`
    Slug               string    `json:"slug" db:"slug"`
    GuestLoginEnabled  bool      `json:"guest_login_enabled" db:"guest_login_enabled"`
    CreatedAt          time.Time `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// RoomUser maps users to rooms with roles.
type RoomUser struct {
    RoomID int64  `json:"room_id" db:"room_id"`
    UserID int64  `json:"user_id" db:"user_id"`
    Role   string `json:"role" db:"role"` // "teacher" or "student"
}
```

---

## Task 3: Backend — Room Repository

**Covers:** Data model redesign

**Files:**
- Create: `internal/adapter/repository/sqlite/room.go`

- [ ] **Step 1: Create RoomRepo**

```go
package repository

import (
    "database/sql"
    "github.com/iroom/iroom/internal/domain/entity"
)

type RoomRepo struct {
    db *sql.DB
}

func NewRoomRepo(db *sql.DB) *RoomRepo {
    return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *entity.Room) error {
    result, err := r.db.Exec(
        `INSERT INTO rooms (owner_id, name, description, color, slug, guest_login_enabled) VALUES (?, ?, ?, ?, ?, ?)`,
        room.OwnerID, room.Name, room.Description, room.Color, room.Slug, room.GuestLoginEnabled,
    )
    if err != nil {
        return err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    room.ID = id
    return nil
}

func (r *RoomRepo) GetByID(id int64) (*entity.Room, error) {
    room := &entity.Room{}
    err := r.db.QueryRow(
        `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE id = ?`, id,
    ).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return room, nil
}

func (r *RoomRepo) GetBySlug(slug string) (*entity.Room, error) {
    room := &entity.Room{}
    err := r.db.QueryRow(
        `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE slug = ?`, slug,
    ).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return room, nil
}

func (r *RoomRepo) ListAll(page, perPage int, search string) ([]entity.Room, int64, error) {
    var total int64
    countQuery := `SELECT COUNT(*) FROM rooms WHERE 1=1`
    args := []interface{}{}
    if search != "" {
        countQuery += ` AND (name LIKE ? OR description LIKE ?)`
        s := "%" + search + "%"
        args = append(args, s, s)
    }
    if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    offset := (page - 1) * perPage
    query := `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE 1=1`
    if search != "" {
        query += ` AND (name LIKE ? OR description LIKE ?)`
    }
    query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
    args = append(args, perPage, offset)

    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var rooms []entity.Room
    for rows.Next() {
        var room entity.Room
        if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt); err != nil {
            return nil, 0, err
        }
        rooms = append(rooms, room)
    }
    return rooms, total, nil
}

func (r *RoomRepo) ListByUser(userID int64) ([]entity.Room, error) {
    rows, err := r.db.Query(
        `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at
         FROM rooms WHERE owner_id = ? OR id IN (SELECT room_id FROM room_users WHERE user_id = ?)
         ORDER BY id DESC`, userID, userID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var rooms []entity.Room
    for rows.Next() {
        var room entity.Room
        if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt); err != nil {
            return nil, err
        }
        rooms = append(rooms, room)
    }
    return rooms, nil
}

func (r *RoomRepo) Update(room *entity.Room) error {
    _, err := r.db.Exec(
        `UPDATE rooms SET name = ?, description = ?, color = ?, guest_login_enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
        room.Name, room.Description, room.Color, room.GuestLoginEnabled, room.ID,
    )
    return err
}

func (r *RoomRepo) Delete(id int64) error {
    _, err := r.db.Exec(`DELETE FROM rooms WHERE id = ?`, id)
    return err
}

func (r *RoomRepo) Count() (int64, error) {
    var count int64
    err := r.db.QueryRow(`SELECT COUNT(*) FROM rooms`).Scan(&count)
    return count, err
}

// Room users

func (r *RoomRepo) AddUser(roomID, userID int64, role string) error {
    _, err := r.db.Exec(
        `INSERT OR REPLACE INTO room_users (room_id, user_id, role) VALUES (?, ?, ?)`,
        roomID, userID, role,
    )
    return err
}

func (r *RoomRepo) RemoveUser(roomID, userID int64) error {
    _, err := r.db.Exec(`DELETE FROM room_users WHERE room_id = ? AND user_id = ?`, roomID, userID)
    return err
}

func (r *RoomRepo) GetUsers(roomID int64) ([]entity.User, error) {
    rows, err := r.db.Query(
        `SELECT u.id, u.email, u.display_name, u.role, u.is_active, u.created_at, u.updated_at
         FROM users u JOIN room_users ru ON u.id = ru.user_id WHERE ru.room_id = ?`, roomID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []entity.User
    for rows.Next() {
        var u entity.User
        if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, nil
}

func (r *RoomRepo) IsUserInRoom(roomID, userID int64) bool {
    var count int
    r.db.QueryRow(`SELECT COUNT(*) FROM room_users WHERE room_id = ? AND user_id = ?`, roomID, userID).Scan(&count)
    return count > 0
}

func (r *RoomRepo) GetUserCount(roomID int64) (int, error) {
    var count int
    err := r.db.QueryRow(`SELECT COUNT(*) FROM room_users WHERE room_id = ?`, roomID).Scan(&count)
    return count, err
}
```

---

## Task 4: Backend — Room Use Case

**Covers:** Data model redesign, room management logic

**Files:**
- Create: `internal/domain/usecase/room.go`

- [ ] **Step 1: Create RoomUseCase**

```go
package usecase

import (
    "fmt"
    "strings"
    "time"
    "unicode"

    "github.com/iroom/iroom/internal/domain/entity"
    repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type RoomUseCase struct {
    roomRepo   *repository.RoomRepo
    userRepo   *repository.UserRepo
    sessionRepo *repository.SessionRepo
}

func NewRoomUseCase(roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, sessionRepo *repository.SessionRepo) *RoomUseCase {
    return &RoomUseCase{roomRepo: roomRepo, userRepo: userRepo, sessionRepo: sessionRepo}
}

func generateRoomSlug(name string) string {
    slug := strings.ToLower(name)
    replacements := map[string]string{
        " ": "-", "‌": "", "۰": "0", "۱": "1", "۲": "2", "۳": "3",
        "۴": "4", "۵": "5", "۶": "6", "۷": "7", "۸": "8", "۹": "9",
    }
    for k, v := range replacements {
        slug = strings.ReplaceAll(slug, k, v)
    }
    var result []rune
    for _, r := range slug {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
            result = append(result, r)
        }
    }
    slug = string(result)
    for strings.Contains(slug, "--") {
        slug = strings.ReplaceAll(slug, "--", "-")
    }
    slug = strings.Trim(slug, "-")
    if slug == "" {
        slug = fmt.Sprintf("room-%d", time.Now().UnixMilli())
    }
    return slug
}

func (uc *RoomUseCase) Create(ownerID int64, name, description, color string) (*entity.Room, error) {
    slug := generateRoomSlug(name)
    room := &entity.Room{
        OwnerID:           ownerID,
        Name:              name,
        Description:       description,
        Color:             color,
        Slug:              slug,
        GuestLoginEnabled: true,
    }
    if err := uc.roomRepo.Create(room); err != nil {
        return nil, fmt.Errorf("خطا در ایجاد اتاق")
    }
    return room, nil
}

func (uc *RoomUseCase) GetByID(id int64) (*entity.Room, error) {
    return uc.roomRepo.GetByID(id)
}

func (uc *RoomUseCase) GetBySlug(slug string) (*entity.Room, error) {
    return uc.roomRepo.GetBySlug(slug)
}

func (uc *RoomUseCase) List(page, perPage int, search string) ([]entity.Room, int64, error) {
    return uc.roomRepo.ListAll(page, perPage, search)
}

func (uc *RoomUseCase) Update(id int64, name, description, color string, guestLoginEnabled bool) (*entity.Room, error) {
    room, err := uc.roomRepo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("اتاق یافت نشد")
    }
    if name != "" {
        room.Name = name
    }
    if description != "" {
        room.Description = description
    }
    if color != "" {
        room.Color = color
    }
    room.GuestLoginEnabled = guestLoginEnabled
    if err := uc.roomRepo.Update(room); err != nil {
        return nil, fmt.Errorf("خطا در بروزرسانی")
    }
    return room, nil
}

func (uc *RoomUseCase) Delete(id int64) error {
    return uc.roomRepo.Delete(id)
}

func (uc *RoomUseCase) AddUser(roomID, userID int64, role string) error {
    if role == "" {
        role = "student"
    }
    return uc.roomRepo.AddUser(roomID, userID, role)
}

func (uc *RoomUseCase) RemoveUser(roomID, userID int64) error {
    return uc.roomRepo.RemoveUser(roomID, userID)
}

func (uc *RoomUseCase) GetUsers(roomID int64) ([]entity.User, error) {
    users, err := uc.roomRepo.GetUsers(roomID)
    if err != nil {
        return nil, fmt.Errorf("خطا در دریافت کاربران")
    }
    if users == nil {
        users = []entity.User{}
    }
    return users, nil
}

func (uc *RoomUseCase) GetUserCount(roomID int64) (int, error) {
    return uc.roomRepo.GetUserCount(roomID)
}

func (uc *RoomUseCase) GetActiveSessionCount(roomID int64) (int, error) {
    sessions, err := uc.sessionRepo.ListByClass(roomID)
    if err != nil {
        return 0, err
    }
    count := 0
    for _, s := range sessions {
        if s.Status == entity.SessionLive {
            count++
        }
    }
    return count, nil
}

func (uc *RoomUseCase) IsUserInRoom(roomID, userID int64) bool {
    return uc.roomRepo.IsUserInRoom(roomID, userID)
}

func (uc *RoomUseCase) GetShareURL(slug string) string {
    return fmt.Sprintf("/room/%s", slug)
}
```

---

## Task 5: Backend — Room Handler

**Covers:** API endpoints for room management

**Files:**
- Create: `internal/adapter/handler/room.go`

- [ ] **Step 1: Create RoomHandler**

```go
package handler

import (
    "strconv"

    "github.com/iroom/iroom/internal/domain/usecase"
    "github.com/iroom/iroom/internal/pkg/response"
    "github.com/labstack/echo/v4"
)

type RoomHandler struct {
    roomUC *usecase.RoomUseCase
}

func NewRoomHandler(roomUC *usecase.RoomUseCase) *RoomHandler {
    return &RoomHandler{roomUC: roomUC}
}

func (h *RoomHandler) Create(c echo.Context) error {
    var req struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        Color       string `json:"color"`
    }
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "داده‌های نامعتبر")
    }
    if req.Name == "" {
        return response.BadRequest(c, "نام اتاق الزامی است")
    }

    userID, _ := getUserID(c)
    room, err := h.roomUC.Create(userID, req.Name, req.Description, req.Color)
    if err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Created(c, room)
}

func (h *RoomHandler) GetByID(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return response.BadRequest(c, "شناسه نامعتبر")
    }
    room, err := h.roomUC.GetByID(id)
    if err != nil {
        return response.NotFound(c, err.Error())
    }
    return response.Success(c, room)
}

func (h *RoomHandler) GetBySlug(c echo.Context) error {
    slug := c.Param("slug")
    room, err := h.roomUC.GetBySlug(slug)
    if err != nil {
        return response.NotFound(c, "اتاق یافت نشد")
    }
    return response.Success(c, room)
}

func (h *RoomHandler) List(c echo.Context) error {
    page, _ := strconv.Atoi(c.QueryParam("page"))
    perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
    if page < 1 {
        page = 1
    }
    if perPage < 1 {
        perPage = 20
    }

    rooms, total, err := h.roomUC.List(page, perPage, c.QueryParam("search"))
    if err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, map[string]interface{}{
        "items":       rooms,
        "total":       total,
        "page":        page,
        "per_page":    perPage,
        "total_pages": (total + int64(perPage) - 1) / int64(perPage),
    })
}

func (h *RoomHandler) Update(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return response.BadRequest(c, "شناسه نامعتبر")
    }

    var req struct {
        Name               string `json:"name"`
        Description        string `json:"description"`
        Color              string `json:"color"`
        GuestLoginEnabled  *bool  `json:"guest_login_enabled"`
    }
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "داده‌های نامعتبر")
    }

    guestLogin := true
    if req.GuestLoginEnabled != nil {
        guestLogin = *req.GuestLoginEnabled
    }

    room, err := h.roomUC.Update(id, req.Name, req.Description, req.Color, guestLogin)
    if err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, room)
}

func (h *RoomHandler) Delete(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return response.BadRequest(c, "شناسه نامعتبر")
    }
    if err := h.roomUC.Delete(id); err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, map[string]string{"message": "اتاق حذف شد"})
}

func (h *RoomHandler) AddUser(c echo.Context) error {
    roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    var req struct {
        UserID int64  `json:"user_id"`
        Role   string `json:"role"`
    }
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "داده‌های نامعتبر")
    }
    if err := h.roomUC.AddUser(roomID, req.UserID, req.Role); err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, map[string]string{"message": "کاربر اضافه شد"})
}

func (h *RoomHandler) RemoveUser(c echo.Context) error {
    roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    userID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
    if err := h.roomUC.RemoveUser(roomID, userID); err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, map[string]string{"message": "کاربر حذف شد"})
}

func (h *RoomHandler) GetUsers(c echo.Context) error {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    users, err := h.roomUC.GetUsers(id)
    if err != nil {
        return response.InternalError(c, err.Error())
    }
    return response.Success(c, users)
}

func (h *RoomHandler) GetInfo(c echo.Context) error {
    slug := c.Param("slug")
    room, err := h.roomUC.GetBySlug(slug)
    if err != nil {
        return response.NotFound(c, "اتاق یافت نشد")
    }

    userCount, _ := h.roomUC.GetUserCount(room.ID)
    activeSessions, _ := h.roomUC.GetActiveSessionCount(room.ID)

    return response.Success(c, map[string]interface{}{
        "room":            room,
        "user_count":      userCount,
        "active_sessions": activeSessions,
    })
}
```

---

## Task 6: Backend — Wire Routes in main.go

**Covers:** API routing

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add room imports and initialization**

In `main.go`, after the existing repository/usecase/handler initializations (around line 100-133), add:

```go
roomRepo := sqliteRepo.NewRoomRepo(db)
roomUC := usecase.NewRoomUseCase(roomRepo, userRepo, sessionRepo)
roomHandler := handler.NewRoomHandler(roomUC)
```

- [ ] **Step 2: Add public room routes** (after auth group, around line 174)

```go
// Public room info (no auth required)
e.GET("/api/v1/rooms/slug/:slug", roomHandler.GetInfo)
```

- [ ] **Step 3: Add protected room routes** (after classes section, around line 204)

```go
// Rooms
api.GET("/rooms", roomHandler.List)
api.GET("/rooms/:id", roomHandler.GetByID)
api.POST("/rooms", roomHandler.Create)
api.PUT("/rooms/:id", roomHandler.Update)
api.DELETE("/rooms/:id", roomHandler.Delete)
api.GET("/rooms/:id/users", roomHandler.GetUsers)
api.POST("/rooms/:id/users", roomHandler.AddUser)
api.DELETE("/rooms/:id/users/:userId", roomHandler.RemoveUser)
```

- [ ] **Step 4: Add admin room routes** (in admin group, around line 289)

```go
admin.GET("/rooms", roomHandler.List)
admin.POST("/rooms", roomHandler.Create)
admin.PUT("/rooms/:id", roomHandler.Update)
admin.DELETE("/rooms/:id", roomHandler.Delete)
admin.GET("/rooms/:id/users", roomHandler.GetUsers)
admin.POST("/rooms/:id/users", roomHandler.AddUser)
admin.DELETE("/rooms/:id/users/:userId", roomHandler.RemoveUser)
```

- [ ] **Step 5: Update guest login to be room-scoped**

Modify `internal/domain/usecase/auth.go` `GuestLogin` method to accept roomID instead of sessionID, and check `room.GuestLoginEnabled`:

```go
func (uc *AuthUseCase) GuestLoginByRoom(roomID int64, displayName string) (*entity.User, map[string]interface{}, error) {
    // This will be a new method; keep old GuestLogin for backward compat
}
```

Actually — simpler approach: add a new endpoint `POST /auth/room-guest-login` that accepts `room_slug` and `display_name`, checks room exists and guest_login_enabled, creates guest user, and returns tokens.

- [ ] **Step 6: Add room guest login handler**

In `internal/adapter/handler/auth.go`, add:

```go
func (h *AuthHandler) RoomGuestLogin(c echo.Context) error {
    var req struct {
        RoomSlug    string `json:"room_slug"`
        DisplayName string `json:"display_name"`
    }
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "داده‌های نامعتبر")
    }
    user, tokens, err := h.authUC.RoomGuestLogin(req.RoomSlug, req.DisplayName)
    if err != nil {
        return response.BadRequest(c, err.Error())
    }
    return response.Success(c, map[string]interface{}{
        "user":   user,
        "tokens": tokens,
    })
}
```

And in `internal/domain/usecase/auth.go`, add the `RoomGuestLogin` method that:
1. Looks up room by slug
2. Checks `GuestLoginEnabled`
3. Creates guest user with `guest_{roomID}_{timestamp}@iroom.local`
4. Returns tokens

And in `cmd/server/main.go`, add to the auth group:

```go
authGroup.POST("/room-guest-login", authHandler.RoomGuestLogin)
```

---

## Task 7: Frontend — Room Type

**Covers:** Frontend type definitions

**Files:**
- Modify: `web/src/lib/types.ts`

- [ ] **Step 1: Add Room type**

Add after the `Class` interface:

```typescript
export interface Room {
    id: number;
    owner_id: number;
    name: string;
    description: string;
    color: string;
    slug: string;
    guest_login_enabled: boolean;
    created_at: string;
    updated_at: string;
}
```

---

## Task 8: Frontend — Room Join Page `/room/[slug]`

**Covers:** Guest/admin join flow

**Files:**
- Create: `web/src/routes/room/[slug]/+page.svelte`

- [ ] **Step 1: Create room join page**

This page replaces `/classroom/join/[id]`. It shows:
- Room info (name, description)
- If guest login enabled: name input + join button
- If user has credentials: email/password login form
- Error if guest login disabled

```svelte
<script lang="ts">
    import { page } from '$app/state';
    import { goto } from '$app/navigation';
    import { api } from '$lib/api';
    import { auth } from '$lib/stores';
    import { onMount } from 'svelte';
    import type { User, Tokens, Room } from '$lib/types';

    let room = $state<Room | null>(null);
    let loading = $state(true);
    let displayName = $state('');
    let email = $state('');
    let password = $state('');
    let joinMode = $state<'guest' | 'login'>('guest');
    let actionLoading = $state(false);
    let error = $state('');

    const slug = $derived(page.params.slug!);

    onMount(async () => {
        const res = await api.get<any>('/rooms/slug/' + slug);
        if (res.success && res.data) {
            room = res.data.room;
            joinMode = room?.guest_login_enabled ? 'guest' : 'login';
        }
        loading = false;
    });

    async function handleGuestJoin() {
        if (!displayName.trim()) { error = 'لطفاً نام خود را وارد کنید'; return; }
        actionLoading = true; error = '';
        const res = await api.post<{ user: User; tokens: Tokens }>('/auth/room-guest-login', {
            room_slug: slug,
            display_name: displayName.trim()
        });
        if (!res.success) { error = res.error || 'خطا در ورود'; actionLoading = false; return; }
        auth.login(res.data!.user, res.data!.tokens);
        goto(`/room/${slug}/live`);
    }

    async function handleLogin() {
        if (!email || !password) { error = 'ایمیل و رمز عبور الزامی است'; return; }
        actionLoading = true; error = '';
        const res = await api.post<{ user: User; tokens: Tokens }>('/auth/login', { email, password });
        if (!res.success) { error = res.error || 'خطا در ورود'; actionLoading = false; return; }
        auth.login(res.data!.user, res.data!.tokens);
        goto(`/room/${slug}/live`);
    }
</script>

<div class="min-h-screen flex items-center justify-center px-4" style="background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);">
    <div class="w-full max-w-[400px] rounded-2xl p-6" style="background: #16213e; border: 1px solid #2a2a4a;">
        {#if loading}
            <div class="text-center py-8"><div class="animate-spin w-8 h-8 border-3 border-[#23b9d7] border-t-transparent rounded-full mx-auto"></div></div>
        {:else if room}
            <div class="text-center mb-6">
                <div class="w-14 h-14 rounded-xl mx-auto mb-3 flex items-center justify-center text-white font-bold text-xl" style="background: {room.color};">
                    {room.name.charAt(0)}
                </div>
                <h1 class="text-xl font-bold" style="color: #eaeaea;">{room.name}</h1>
                {#if room.description}
                    <p class="text-sm mt-1" style="color: #94a3b8;">{room.description}</p>
                {/if}
            </div>

            {#if error}
                <div class="mb-4 px-4 py-3 rounded-lg text-sm text-center" style="background: rgba(224,82,82,0.08); color: #e05252;">{error}</div>
            {/if}

            {#if joinMode === 'guest'}
                <form onsubmit={(e) => { e.preventDefault(); handleGuestJoin(); }} class="space-y-3">
                    <div>
                        <label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">نام شما</label>
                        <input type="text" bind:value={displayName} placeholder="نام خود را وارد کنید" dir="auto" required
                            class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" />
                    </div>
                    <button type="submit" disabled={actionLoading}
                        class="w-full py-2.5 rounded-lg text-sm font-semibold text-white" style="background: #23b9d7;">
                        {actionLoading ? 'در حال پیوستن...' : 'پیوستن به اتاق'}
                    </button>
                </form>
                <div class="mt-3 text-center">
                    <button onclick={() => joinMode = 'login'} class="text-xs" style="color: #6790a0;">ورود با حساب کاربری</button>
                </div>
            {:else}
                <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }} class="space-y-3">
                    <div>
                        <label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">ایمیل</label>
                        <input type="email" bind:value={email} placeholder="ایمیل" required
                            class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" />
                    </div>
                    <div>
                        <label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">رمز عبور</label>
                        <input type="password" bind:value={password} placeholder="رمز عبور" required
                            class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" />
                    </div>
                    <button type="submit" disabled={actionLoading}
                        class="w-full py-2.5 rounded-lg text-sm font-semibold text-white" style="background: #23b9d7;">
                        {actionLoading ? 'در حال ورود...' : 'ورود'}
                    </button>
                </form>
                {#if room.guest_login_enabled}
                    <div class="mt-3 text-center">
                        <button onclick={() => joinMode = 'guest'} class="text-xs" style="color: #6790a0;">ورود مهمان</button>
                    </div>
                {/if}
            {/if}
        {:else}
            <div class="text-center py-8"><p style="color: #e05252;">اتاق یافت نشد</p></div>
        {/if}
    </div>
</div>
```

---

## Task 9: Frontend — Room Live Page `/room/[slug]/live`

**Covers:** Classroom view

**Files:**
- Create: `web/src/routes/room/[slug]/live/+page.svelte`

- [ ] **Step 1: Create room live page**

This is a thin wrapper that loads the room, fetches the active session, and redirects to the existing classroom popup. Copy the structure from `classroom/popup/[id]/+page.svelte` but:
1. On mount: fetch room by slug
2. Find or create active session for this room
3. Load classroom UI with that session

For now, redirect to the existing popup if a session exists:

```svelte
<script lang="ts">
    import { page } from '$app/state';
    import { goto } from '$app/navigation';
    import { api } from '$lib/api';
    import { onMount } from 'svelte';

    const slug = $derived(page.params.slug!);

    onMount(async () => {
        const res = await api.get<any>('/rooms/slug/' + slug);
        if (!res.success || !res.data) { return; }
        const room = res.data.room;
        // Find active session or redirect to sessions list
        const sessRes = await api.get<any[]>('/sessions', { search: room.name });
        if (sessRes.success && sessRes.data?.items?.length > 0) {
            const live = sessRes.data.items.find((s: any) => s.status === 'live');
            if (live) {
                goto(`/classroom/popup/${live.id}`);
                return;
            }
        }
        // No active session — show room info
    });
</script>

<div class="min-h-screen flex items-center justify-center" style="background: #121822;">
    <div class="text-center">
        <p style="color: #8a8a96;">اتاق فعالی وجود ندارد</p>
        <a href="/" class="mt-4 inline-block px-4 py-2 rounded-lg text-sm" style="background: #23b9d7; color: white;">بازگشت</a>
    </div>
</div>
```

---

## Task 10: Frontend — Update Admin Rooms Page

**Covers:** Admin room management UI

**Files:**
- Modify: `web/src/routes/(app)/admin/rooms/+page.svelte`

- [ ] **Step 1: Update to use Room API**

Replace the existing admin/rooms page to use the new `/rooms` API instead of `/classes`. Key changes:
1. Fetch from `/rooms` instead of `/classes`
2. Show room users count
3. Add guest login toggle in create/edit modal
4. Add "manage users" button per room
5. Add "copy share link" button

The existing page already has a good structure. Update:
- `loadRooms()` to call `/rooms`
- Create modal to include `guest_login_enabled` toggle
- Add a "Users" button that navigates to `/admin/rooms/:id/users`
- Add a "Copy Link" button that copies `/room/{slug}` to clipboard

---

## Task 11: Frontend — Room Users Management Page

**Covers:** Per-room user assignment

**Files:**
- Create: `web/src/routes/(app)/admin/rooms/[id]/users/+page.svelte`

- [ ] **Step 1: Create room users page**

```svelte
<script lang="ts">
    import { page } from '$app/state';
    import { api } from '$lib/api';
    import { onMount } from 'svelte';
    import type { User, Room } from '$lib/types';
    import { toPersianNum } from '$lib/utils/persian';

    const roomId = $derived(Number(page.params.id));
    let room = $state<Room | null>(null);
    let users = $state<User[]>([]);
    let allUsers = $state<User[]>([]);
    let loading = $state(true);
    let showAddModal = $state(false);
    let selectedUserId = $state(0);

    onMount(loadData);

    async function loadData() {
        loading = true;
        const [roomRes, usersRes, allRes] = await Promise.all([
            api.get<Room>(`/rooms/${roomId}`),
            api.get<User[]>(`/rooms/${roomId}/users`),
            api.get<{ items: User[] }>('/admin/users', { per_page: '1000' })
        ]);
        if (roomRes.success) room = roomRes.data;
        if (usersRes.success && Array.isArray(usersRes.data)) users = usersRes.data;
        if (allRes.success && allRes.data) allUsers = allRes.data.items || [];
        loading = false;
    }

    async function addUser() {
        if (!selectedUserId) return;
        await api.post(`/rooms/${roomId}/users`, { user_id: selectedUserId, role: 'student' });
        selectedUserId = 0;
        showAddModal = false;
        await loadData();
    }

    async function removeUser(userId: number) {
        if (!confirm('آیا از حذف این کاربر اطمینان دارید؟')) return;
        await api.delete(`/rooms/${roomId}/users/${userId}`);
        await loadData();
    }

    const availableUsers = $derived(allUsers.filter(u => !users.find(ur => ur.id === u.id)));
</script>

<div class="space-y-5">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="sky-page-title">مدیریت کاربران اتاق</h1>
            <p class="sky-page-subtitle">{room?.name || ''} — {toPersianNum(users.length)} کاربر</p>
        </div>
        <div class="flex gap-2">
            <a href="/admin/rooms" class="sky-btn sky-btn-secondary">بازگشت</a>
            <button onclick={() => showAddModal = true} class="sky-btn sky-btn-primary">افزودن کاربر</button>
        </div>
    </div>

    {#if loading}
        <div class="flex justify-center py-16"><div class="sky-spinner lg"></div></div>
    {:else if users.length === 0}
        <div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">کاربری اختصاص داده نشده</p></div></div>
    {:else}
        <div class="sky-card">
            <table class="sky-table">
                <thead><tr><th>نام</th><th>ایمیل</th><th>عملیات</th></tr></thead>
                <tbody>
                    {#each users as u}
                        <tr>
                            <td class="font-semibold">{u.display_name}</td>
                            <td style="color: var(--color-mystic-sea);">{u.email}</td>
                            <td>
                                <button onclick={() => removeUser(u.id)} class="sky-btn-icon" style="width:32px;height:32px;">
                                    <svg width="15" height="15" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
                                </button>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
</div>

{#if showAddModal}
    <div class="modal-overlay" onclick={() => showAddModal = false} role="button" tabindex="-1">
        <div class="modal-content" onclick={(e) => e.stopPropagation()}>
            <div class="sky-modal-header">
                <h2>افزودن کاربر</h2>
                <button onclick={() => showAddModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
            </div>
            <div class="sky-modal-body">
                <select bind:value={selectedUserId} class="sky-input">
                    <option value={0}>انتخاب کاربر</option>
                    {#each availableUsers as u}
                        <option value={u.id}>{u.display_name} ({u.email})</option>
                    {/each}
                </select>
            </div>
            <div class="sky-modal-footer">
                <button onclick={() => showAddModal = false} class="sky-btn sky-btn-secondary">انصراف</button>
                <button onclick={addUser} disabled={!selectedUserId} class="sky-btn sky-btn-primary">افزودن</button>
            </div>
        </div>
    </div>
{/if}
```

---

## Task 12: Frontend — Update Sidebar Navigation

**Covers:** Navigation merge

**Files:**
- Modify: `web/src/routes/(app)/+layout.svelte`

- [ ] **Step 1: Replace کلاس‌ها and جلسات with اتاق‌ها**

In `navItems` (line 108-115), replace:

```typescript
const navItems = $derived.by(() => [
    { href: '/dashboard', label: 'داشبورد', icon: icons.dashboard },
    { href: '/rooms', label: 'اتاق‌ها', icon: icons.rooms },
    { href: '/files', label: 'فایل‌ها', icon: icons.files },
    { href: '/support', label: 'پشتیبانی', icon: icons.support },
    { href: '/profile', label: 'حساب کاربری', icon: icons.profile },
]);
```

Add a `rooms` icon if not already defined (it is — `icons.rooms` exists at line 102).

---

## Task 13: Frontend — Create `/rooms` Route

**Covers:** User-facing rooms list

**Files:**
- Create: `web/src/routes/(app)/rooms/+page.svelte`

- [ ] **Step 1: Create rooms list page**

This page lists rooms the current user has access to. For admin, show all rooms. For students, show only assigned rooms.

```svelte
<script lang="ts">
    import { auth, isAdmin } from '$lib/stores';
    import { api } from '$lib/api';
    import { onMount } from 'svelte';
    import type { Room } from '$lib/types';
    import { toPersianNum } from '$lib/utils/persian';

    let rooms = $state<Room[]>([]);
    let loading = $state(true);
    let search = $state('');

    onMount(loadRooms);

    async function loadRooms() {
        loading = true;
        const res = await api.get<{ items: Room[]; total: number }>('/rooms', { per_page: '100' });
        if (res.success && res.data) {
            rooms = res.data.items || [];
        }
        loading = false;
    }

    function copyLink(slug: string) {
        navigator.clipboard.writeText(`${window.location.origin}/room/${slug}`);
    }
</script>

<div class="space-y-5">
    <div class="flex items-center justify-between">
        <div>
            <h1 class="sky-page-title">اتاق‌ها</h1>
            <p class="sky-page-subtitle">{toPersianNum(rooms.length)} اتاق</p>
        </div>
    </div>

    {#if loading}
        <div class="flex justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
    {:else if rooms.length === 0}
        <div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">اتاقی یافت نشد</p></div></div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each rooms as room}
                <div class="sky-card p-5">
                    <div class="flex items-center gap-3 mb-3">
                        <div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shrink-0"
                            style="background: {room.color};">
                            {room.name.charAt(0)}
                        </div>
                        <div class="flex-1 min-w-0">
                            <h3 class="font-bold text-sm truncate" style="color: var(--color-midnight-sky);">{room.name}</h3>
                        </div>
                    </div>
                    {#if room.description}
                        <p class="text-xs line-clamp-2 mb-3" style="color: var(--color-mystic-sea);">{room.description}</p>
                    {/if}
                    <div class="flex items-center gap-2">
                        <a href="/room/{room.slug}" class="sky-btn sky-btn-primary flex-1" style="font-size: 12px; padding: 0.45rem;">ورود</a>
                        <button onclick={() => copyLink(room.slug)} class="sky-btn sky-btn-secondary" style="font-size: 12px; padding: 0.45rem;">کپی لینک</button>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
```

---

## Task 14: Cleanup — Remove Old Routes

**Covers:** Remove deprecated routes

**Files:**
- Delete or deprecate: `web/src/routes/(app)/classes/+page.svelte`
- Delete or deprecate: `web/src/routes/(app)/sessions/+page.svelte`

- [ ] **Step 1: Redirect old routes**

Replace contents of `classes/+page.svelte` with redirect:

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    onMount(() => goto('/rooms', { replaceState: true }));
</script>
```

Same for `sessions/+page.svelte`.

---

## Verification

After all tasks are complete:

1. Run `go build ./cmd/server` to verify backend compiles
2. Run `cd web && npm run build` to verify frontend builds
3. Test: Create a room via admin panel
4. Test: Add a user to the room
5. Test: Copy share link, open in incognito — guest join works
6. Test: Disable guest login — error shown on join page
7. Test: Login with assigned user credentials — joins room
8. Test: Sidebar shows "اتاق‌ها" instead of "کلاس‌ها" and "جلسات"
