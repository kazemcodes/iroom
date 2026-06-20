# IRoom Rewrite Plan

## Decisions Made
- **Merge Class into Room**: Rooms are the single concept. Class entity, usecase, handler, and repository are removed. Class-specific features (invite codes, access levels, slug) migrate into Room.
- **DB as source of truth for WebRTC rooms**: `sessions` table gets a `room_id` column referencing `rooms(id)`. WebRTC room manager loads room state from DB on startup.
- **Slug generation**: Extracted to a shared `internal/pkg/slug` package (deduplicates the two identical implementations in `class.go` and `room.go`).
- **Announcements**: `Announcement.ClassID` becomes `Announcement.RoomID`. The `classRepo` field in `AnnouncementUseCase` is dead code — removed. Routes change from `/classes/:id/announcements` to `/rooms/:id/announcements`.
- **UserUseCase.GetUserRooms**: Moved into `RoomUseCase.GetUserRooms`. `UserUseCase` no longer needs `classRepo`.
- **RecordingUseCase/FileUseCase**: `classRepo` field is dead code — removed from both.
- **Backward compat**: `sessions.class_id` column is NOT dropped. Old sessions without `room_id` still work.

## Architecture After Rewrite

```
domain/entity/       — Room, Session, User (Class/ClassStudent removed, Announcement.RoomID)
domain/usecase/      — RoomUseCase, SessionUseCase (ClassUseCase removed)
adapter/handler/     — RoomHandler, SessionHandler (ClassHandler removed)
adapter/repository/  — RoomRepo, SessionRepo (ClassRepo removed)
internal/pkg/slug/   — Shared slug generator
internal/pkg/errors/ — Sentinel errors
internal/webrtc/     — RoomManager loads from DB on startup
```

## Step-by-Step Tasks

### 1. Database Migration
Create `020_unify_rooms.sql`:
- Add `room_id INTEGER REFERENCES rooms(id)` to `sessions`.
- Backfill: `UPDATE sessions SET room_id = class_id WHERE room_id IS NULL`.
- Add index `idx_sessions_room_id`.
- Add `room_id` column to `announcements` table (backfill from `class_id`).
- Rename `announcements.class_id` → keep column, add `room_id`, backfill. (SQLite doesn't support RENAME COLUMN in older versions; add new column + backfill.)
- Add `access INTEGER DEFAULT 1` to `room_users` table (for presenter/operator levels).
- Keep `sessions.class_id` for backward compat.

### 2. Entity Layer
- **Delete** `domain/entity/class.go`.
- **Modify** `domain/entity/room.go`: Add `MaxUsers int`, `InviteCode string`, `IsArchived bool` fields.
- **Modify** `domain/entity/session.go`: Add `RoomID int64` with `json:"room_id" db:"room_id"`. Keep `ClassID`.
- **Modify** `domain/entity/models.go`: Rename `Announcement.ClassID` → `RoomID` (update db tag to `room_id`).

### 3. Shared Slug Package
Create `internal/pkg/slug/slug.go`:
- `Generate(name string) string` — extracted from the identical `generateSlug` in `class.go` and `generateRoomSlug` in `room.go`.
- Includes Persian digit replacement, special char stripping, hyphen collapsing, empty-name fallback.
Create `internal/pkg/slug/slug_test.go` with tests for Persian chars, special chars, empty input.

### 4. Error Package
Create `internal/pkg/errors/errors.go`:
- Sentinel errors: `ErrNotFound`, `ErrForbidden`, `ErrValidation`, `ErrConflict`.
- Helper `HTTPStatus(error) int` mapping sentinels to status codes.

### 5. Repository Layer
- **Delete** `adapter/repository/sqlite/class.go`.
- **Modify** `adapter/repository/sqlite/room.go`:
  - `Create`: Accept and store `MaxUsers`, `InviteCode`.
  - `GetByID`/`GetBySlug`/`ListAll`/`ListByUser`: Select new fields.
  - `Update`: Update new fields.
  - `AddUser`: Store `access` level (new column).
  - Add `GetByInviteCode(code string) (*entity.Room, error)`.
  - Add `UpdateInviteCode(roomID int64, code string) error`.
  - Add `GetByUserID(userID int64) ([]entity.Room, error)` (moved from ClassRepo).
  - Add `RemoveUser(roomID, userID int64) error` (already exists).
  - Add `UpdateUserAccess(roomID, userID int64, access int) error`.
- **Modify** `adapter/repository/sqlite/session.go`:
  - `Create`: Accept and store `room_id`.
  - `GetByID`/`ListAll`: Select `room_id`.
  - Add `ListByRoom(roomID int64) ([]entity.Session, error)`.
  - Keep `ListByClass` as backward-compat alias for `ListByRoom`.
- **Modify** `adapter/repository/sqlite/announcement.go`:
  - Replace all `class_id` references with `room_id` in SQL queries.
  - Update `entity.Announcement` field references from `ClassID` to `RoomID`.

### 6. Use Case Layer
- **Delete** `domain/usecase/class.go`.
- **Modify** `domain/usecase/room.go`:
  - Remove local `generateRoomSlug`, use `slug.Generate`.
  - `Create`: Accept `maxUsers int`, `inviteCode string`.
  - `Update`: Accept `maxUsers int`, `inviteCode string`.
  - `AddUser`: Accept `access int` param.
  - Add `RegenerateCode(roomID, actorID int64, role string) (string, error)`.
  - Add `JoinByCode(code string) (*entity.Room, error)`.
  - Add `UpdateUserAccess(roomID, userID, actorID int64, role string, access int) error`.
  - Add `RemoveUser(roomID, userID, actorID int64, role string) error` with ownership check.
  - Add `GetUserRooms(userID int64) ([]entity.Room, error)`.
  - **Authorization**: Every mutation verifies `actorID == room.OwnerID || role == "admin"`. Return `errors.ErrForbidden` otherwise.
  - Remove `GetActiveSessionCount` — callers use `SessionUseCase.CountByRoom` instead.
- **Modify** `domain/usecase/session.go`:
  - `Create`: Accept `roomID int64` (replace `classID`).
  - `checkPermission`: Use `roomRepo.GetByID(s.RoomID)` instead of `classRepo.GetByID(s.ClassID)`. Check `room.OwnerID == userID`.
  - Replace `classRepo` field with `roomRepo`.
  - Add `CountByRoom(roomID int64) (int, error)` for active session counting.
- **Modify** `domain/usecase/announcement.go`:
  - Remove `classRepo` field (dead code — never used).
  - Change `Create` param from `classID` to `roomID`.
  - Change `ListByClass` to `ListByRoom`.
- **Modify** `domain/usecase/user.go`:
  - Remove `classRepo` field.
  - Remove `GetUserRooms` method (moved to RoomUseCase).
- **Modify** `domain/usecase/recording.go`:
  - Remove `classRepo` field (dead code — never used).
- **Modify** `domain/usecase/file.go`:
  - Remove `classRepo` field (dead code — never used).

### 7. WebRTC Integration
- **Modify** `internal/webrtc/room.go`:
  - Add `LoadFromDB` method to `RoomManager`: queries all sessions with `room_id IS NOT NULL` via a provided `RoomLoader` interface, creates in-memory rooms with `maxSize` from room settings.
  - Define `RoomLoader` interface: `GetLiveSessions() []LiveSession` where `LiveSession` has `RoomID`, `MaxUsers`.
- **Modify** `internal/webrtc/signaling.go**:
  - `HandleOffer`: After creating room via `GetOrCreateRoom`, load room settings. For each incoming track from a student, check `allow_student_video`/`allow_student_audio` before broadcasting. Reject screen share if `allow_student_screen_share` is false.
- **Modify** `adapter/handler/webrtc.go`:
  - `GetJoinInfo`: Look up session to get `room_id`, return that as the WebRTC room key instead of `sessionID`.

### 8. Handler Layer
- **Delete** `adapter/handler/class.go`.
- **Modify** `adapter/handler/room.go`:
  - `Create`: Accept `max_students` and `invite_code` in request body.
  - `Update`: Accept `max_students` and `invite_code` in request body.
  - `AddUser`: Accept `access` (int) in request body.
  - Add `RegenerateCode` handler.
  - Add `JoinByCode` handler.
  - Add `UpdateUserAccess` handler.
  - Add `RemoveUser` handler (with ownership check via usecase).
  - Add `GetUserRooms` handler.
  - **Authorization**: Every mutation handler extracts `userID` and `role` from context, passes to usecase. Returns 403 on `errors.ErrForbidden`.
  - `GetInfo`: Also return `max_users` and `invite_code`.
  - Replace `GetActiveSessionCount` call with `sessionUC.CountByRoom`.
- **Modify** `adapter/handler/announcement.go` (if exists as separate file) or inline handlers in main.go:
  - Change route param from `classID` to `roomID`.
  - Update handler to pass `roomID` to usecase.

### 9. Routing (main.go)
- Remove all `/api/v1/classes` routes.
- Remove all `/api/v1/admin/classes` routes.
- Remove `/api/v1/users/:id/rooms` route (was class-based, now on room handler).
- Change `/api/v1/classes/:id/announcements` → `/api/v1/rooms/:id/announcements`.
- Add new room routes:
  - `POST /api/v1/rooms/:id/regenerate-code`
  - `POST /api/v1/rooms/join/:code`
  - `PUT /api/v1/rooms/:id/users/:userId`
  - `DELETE /api/v1/rooms/:id/users/:userId`
  - `GET /api/v1/users/:id/rooms`
- Update admin routes to match (remove admin class routes, add admin room routes for new endpoints).
- Remove `classUC`, `classHandler`, `classRepo` wiring.
- Add `roomRepo` to `sessionUC` wiring.
- Remove `classRepo` from `announcementUC`, `recordingUC`, `fileUC`, `userUC` wiring.
- Call `signaling.GetRoomManager().LoadFromDB(...)` after wiring.

### 10. Tests
- `internal/pkg/slug/slug_test.go`: Test Persian digit replacement, special char stripping, empty name fallback, hyphen collapsing.
- `domain/usecase/room_test.go`: Test authorization (non-owner cannot update/delete), slug uniqueness, user access levels, invite code flow.
- Existing `internal/webrtc/room_test.go`: No changes needed.

## Files Deleted
- `domain/entity/class.go`
- `domain/usecase/class.go`
- `adapter/handler/class.go`
- `adapter/repository/sqlite/class.go`

## Files Created
- `internal/pkg/slug/slug.go`
- `internal/pkg/slug/slug_test.go`
- `internal/pkg/errors/errors.go`
- `domain/usecase/room_test.go`
- `internal/database/migrations/020_unify_rooms.sql`

## Files Modified
- `domain/entity/room.go` — Add MaxUsers, InviteCode, IsArchived fields
- `domain/entity/session.go` — Add RoomID field
- `domain/entity/models.go` — Announcement.ClassID → RoomID
- `domain/usecase/room.go` — Merge class logic, add auth, use slug package
- `domain/usecase/session.go` — Use room_id, roomRepo
- `domain/usecase/announcement.go` — roomID param, remove classRepo
- `domain/usecase/user.go` — Remove classRepo, remove GetUserRooms
- `domain/usecase/recording.go` — Remove classRepo
- `domain/usecase/file.go` — Remove classRepo
- `adapter/repository/sqlite/room.go` — Add class features, access level
- `adapter/repository/sqlite/session.go` — Add room_id support
- `adapter/repository/sqlite/announcement.go` — room_id instead of class_id
- `adapter/handler/room.go` — Merge class handlers, add auth
- `adapter/handler/webrtc.go` — Use room_id, enforce permissions
- `internal/webrtc/room.go` — DB-backed room loading
- `internal/webrtc/signaling.go` — Media permission checks
- `cmd/server/main.go` — Route cleanup, wiring

## Validation Plan
1. `go build ./...` — must compile cleanly.
2. `go test ./...` — all existing + new tests pass.
3. Manual: Create room → add user → create session → start session → join via WebRTC → verify media permissions enforced.
4. Manual: Non-owner cannot update/delete room (403).
5. Manual: Guest login with slug works, auto-enrolls in room.
6. Old `class_id` sessions still work (backward compat column retained).
