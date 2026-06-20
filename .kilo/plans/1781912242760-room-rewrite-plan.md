# IRoom Rewrite Plan

## Decisions Made
- **Merge Class into Room**: Rooms are the single concept. Class entity, usecase, handler, and routes are removed.
- **DB as source of truth for WebRTC rooms**: `sessions` table gets a `room_id` column referencing `rooms(id)`. WebRTC room manager is seeded from DB and persists room state through the session lifecycle.
- **Slug generation**: Extracted to a shared `internal/pkg/slug` package (deduplicates the two identical implementations in `class.go` and `room.go`).

## Architecture After Rewrite

```
domain/entity/       — Room, Session, User (Class/ClassStudent removed)
domain/usecase/      — RoomUseCase, SessionUseCase (ClassUseCase removed)
adapter/handler/     — RoomHandler, SessionHandler (ClassHandler removed)
adapter/repository/  — RoomRepo, SessionRepo (ClassRepo removed)
internal/pkg/slug/   — Shared slug generator
internal/webrtc/     — RoomManager persists via RoomRepo
```

## Step-by-Step Tasks

### 1. Database Migration
- Create migration `020_unify_rooms.sql`:
  - Add `room_id INTEGER REFERENCES rooms(id)` column to `sessions` table.
  - Backfill: set `room_id` from existing `class_id` where possible (or leave NULL for old sessions).
  - Add index `idx_sessions_room_id ON sessions(room_id)`.
  - Keep `class_id` column for backward compat (do NOT drop yet — zero-downtime safety).

### 2. Entity Layer
- **`domain/entity/class.go`**: Delete.
- **`domain/entity/room.go`**: Add `MaxUsers`, `InviteCode`, `IsArchived` fields (migrating the useful Class fields into Room). Add `RoomUser.Role` comment documenting access levels.
- **`domain/entity/session.go`**: Add `RoomID int64` field with `json:"room_id" db:"room_id"`. Keep `ClassID` for backward compat.

### 3. Shared Slug Package
- Create `internal/pkg/slug/slug.go`:
  - `Generate(name string) string` — extracted from the identical `generateSlug` in `class.go` and `generateRoomSlug` in `room.go`.
  - Includes Persian digit replacement, special char stripping, hyphen collapsing.
- Update `domain/usecase/room.go` to call `slug.Generate` instead of local `generateRoomSlug`.

### 4. Repository Layer
- **`adapter/repository/sqlite/class.go`**: Delete.
- **`adapter/repository/sqlite/room.go`**:
  - `Create`: Accept new fields (MaxUsers, InviteCode).
  - `GetByID`/`GetBySlug`/`ListAll`/`ListByUser`: Select new fields.
  - `Update`: Update new fields.
  - `AddUser`: Keep, but also store `access` level (1=student, 2=presenter, 3=operator) — currently only stores role text.
  - Add `GetByInviteCode(code string)` (moved from ClassRepo).
  - Add `UpdateInviteCode(classID, code string)` (moved from ClassRepo).
- **`adapter/repository/sqlite/session.go`**:
  - `Create`: Accept and store `room_id`.
  - `GetByID`/`ListAll`/`ListByClass`: Select `room_id`.
  - Add `ListByRoom(roomID int64)` — replaces `ListByClass` as primary query.
  - Keep `ListByClass` as alias calling `ListByRoom` for backward compat.

### 5. Use Case Layer
- **`domain/usecase/class.go`**: Delete.
- **`domain/usecase/room.go`**:
  - `Create`: Accept `maxStudents` and `inviteCode` params.
  - `Update`: Accept `maxStudents` and `inviteCode` params.
  - `AddUser`: Accept `access int` param (1-3).
  - Add `RegenerateCode(roomID, actorID int64, role string) (string, error)`.
  - Add `JoinByCode(code string) (*entity.Room, error)`.
  - Add `UpdateUserAccess(roomID, userID, actorID int64, role string, access int) error`.
  - Add `RemoveUser(roomID, userID, actorID int64, role string) error` — with ownership check.
  - Add `GetUserRooms(userID int64) ([]entity.Room, error)`.
  - **Authorization**: Every mutation method must verify `actorID` is room owner or admin. Return `ErrForbidden` otherwise.
  - Remove `GetActiveSessionCount` — replace with DB query in SessionUseCase.
- **`domain/usecase/session.go`**:
  - `Create`: Accept `roomID` instead of `classID`.
  - `Start`: Generate room name as `fmt.Sprintf("room-%d", roomID)` using the session's `RoomID`.
  - `checkPermission`: Look up room via `roomRepo` instead of `classRepo`. Check `room.OwnerID == userID`.
  - Remove `classRepo` dependency; add `roomRepo`.

### 6. WebRTC Integration
- **`internal/webrtc/room.go`**:
  - `RoomManager.GetOrCreateRoom`: Accept `roomRepo *repository.RoomRepo` (or interface). On room creation, load `max_users` from `room_settings` via repo. If no settings found, default to 50.
  - `RoomManager.DeleteRoom`: No change needed.
  - Add `RoomManager.LoadFromDB(roomRepo)` — on server start, queries all live sessions and pre-creates in-memory rooms. Called from `main.go`.
- **`internal/webrtc/signaling.go`**:
  - `HandleOffer`: After SDP exchange, check `room_settings` for `allow_student_video`, `allow_student_audio`, `allow_student_screen_share` before forwarding tracks. If student tries to share video and `allow_student_video` is false, reject the track.
  - `HandleLeave`: No change.
  - `MuteParticipant`/`KickParticipant`: Already checks teacher/admin role — keep.
- **`adapter/handler/webrtc.go`**:
  - `GetJoinInfo`: Return `room_id` (from session) instead of `sessionID` as the WebRTC room key.
  - Add `canShareMedia(participantRole, roomSettings) bool` helper for server-side media permission enforcement.

### 7. Handler Layer
- **`adapter/handler/class.go`**: Delete.
- **`adapter/handler/room.go`**:
  - `Create`: Accept `max_students` and `invite_code` in request body.
  - `Update`: Accept `max_students` and `invite_code` in request body.
  - `AddUser`: Accept `access` (int) in request body.
  - Add `RegenerateCode` handler.
  - Add `JoinByCode` handler.
  - Add `UpdateUserAccess` handler.
  - Add `RemoveUser` handler (with ownership check via usecase).
  - Add `GetUserRooms` handler.
  - **Authorization**: Every mutation handler must extract `userID` and `role` from context and pass to usecase. Return 403 if usecase returns `ErrForbidden`.
  - `GetInfo`: Also return `max_users` and `invite_code`.
- **`adapter/handler/helpers.go`**: Add `ErrForbidden` sentinel error to `internal/pkg/errors` (or use middleware). Return 403 consistently.

### 8. Routing (main.go)
- Remove all `/api/v1/classes` routes.
- Remove all `/api/v1/admin/classes` routes.
- Remove `/api/v1/users/:id/rooms` route (was class-based).
- Add new room routes:
  - `POST /api/v1/rooms/:id/regenerate-code` → `roomHandler.RegenerateCode`
  - `POST /api/v1/rooms/join/:code` → `roomHandler.JoinByCode`
  - `PUT /api/v1/rooms/:id/users/:userId` → `roomHandler.UpdateUserAccess`
  - `DELETE /api/v1/rooms/:id/users/:userId` → `roomHandler.RemoveUser`
  - `GET /api/v1/users/:id/rooms` → `roomHandler.GetUserRooms`
- Update admin room routes to match.
- Remove `classUC`, `classHandler`, `classRepo` wiring.
- Add `roomRepo` to `sessionUC` wiring.
- Add `roomRepo` to `dashboardUC` wiring (already uses roomRepo).
- Call `signaling.GetRoomManager().LoadFromDB(roomRepo)` after wiring.

### 9. Error Handling
- Create `internal/pkg/errors/errors.go` with sentinel errors:
  - `ErrNotFound`, `ErrForbidden`, `ErrValidation`, `ErrConflict`
- Update all usecase methods to return specific sentinets instead of `fmt.Errorf("خطا...")`.
- Update handlers to map sentinels to HTTP status codes (404, 403, 400, 409).

### 10. Tests
- Update `internal/webrtc/room_test.go`: No structural changes needed — tests are already solid.
- Add `internal/pkg/slug/slug_test.go`: Test Persian digit replacement, special char stripping, empty name fallback.
- Add `domain/usecase/room_test.go`: Test authorization (non-owner cannot update/delete), slug uniqueness, user access levels.

## Files Deleted
- `domain/entity/class.go`
- `domain/usecase/class.go`
- `adapter/handler/class.go`
- `adapter/repository/sqlite/class.go`

## Files Created
- `internal/pkg/slug/slug.go`
- `internal/pkg/errors/errors.go`
- `internal/pkg/slug/slug_test.go`
- `domain/usecase/room_test.go`
- `internal/database/migrations/020_unify_rooms.sql`

## Files Modified
- `domain/entity/room.go` — Add Class fields
- `domain/entity/session.go` — Add RoomID
- `domain/usecase/room.go` — Merge class logic, add auth
- `domain/usecase/session.go` — Use room_id
- `adapter/repository/sqlite/room.go` — Add class features
- `adapter/repository/sqlite/session.go` — Add room_id support
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
