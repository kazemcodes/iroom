# IRoom — Session Summary (June 20, 2026)

## What Was Done

This session focused on **merging the Class entity into Room**, fixing critical bugs across the stack, and building real-time classroom features (chat, whiteboard, video/audio, screen sharing).

### 1. Class → Room Merge (Commit 459979db)

- Removed the `Class` entity entirely — merged invite codes, access levels, and slugs into `Room`
- Added `room_id` to sessions and announcements (kept `class_id` for backward compat)
- Created `slug` package for URL-friendly room identifiers
- Created shared `errors` package
- Updated all handlers/routes from `/classes` to `/rooms`
- Added migration `020_unify_rooms.sql` for schema changes

### 2. Critical Bug Fixes

| Bug | Root Cause | Fix |
|-----|-----------|-----|
| `/room/slug` returns "اتاق یافت نشد" | `GetBySlug` returned room directly, but frontend expected `{room: ...}` wrapper | Wrapped response in `{"room": room}` |
| Room slugs missing for old rooms | Migration 020 ran without slug backfill | Added migration `021_backfill_room_slugs.sql` |
| POST /sessions 500 | `class_id` had `NOT NULL REFERENCES classes(id)` FK — new sessions passed room ID as class_id which doesn't exist in classes table | Migration `022_fix_sessions_class_id.sql` recreated sessions table without FK |
| POST /sessions 500 (part 2) | `scheduledAt` string never parsed into `time.Time` — zero value caused INSERT failure | Added RFC3339/time parsing with `time.Now()` fallback |
| All admin POST/PUT/DELETE 403 | CSRF middleware rejected all mutations (no CSRF token sent by SPA with JWT auth) | Skip CSRF when `Authorization` header is present |
| Admin logs 404 | Frontend calls `/admin/logs`, backend registered `/admin/activity-logs` | Added `/logs` route alias |
| Webhook update/delete 403 | Use case checked `w.UserID != userID` — admin couldn't manage other users' webhooks | Added `role` parameter, admin bypass |
| Webhook list empty for admin | `ListByUser` filtered by current user ID only | Admin now uses `ListAll()` |
| Session `room_id`/`class_id` NULL crash | `int64` can't scan NULL — old sessions had NULL room_id | Added `COALESCE(room_id, 0)` to all session queries |

### 3. Chat System (Real-Time WebSocket)

- Created `ChatHandler` with WebSocket hub for real-time messaging
- Fixed `Hub.Run()` never being called (hub channels were blocked forever)
- Fixed `CheckOrigin` rejecting WebSocket upgrades through Vite proxy
- Fixed `writePump` batching multiple JSON messages with `\n` separator — frontend's single `JSON.parse` failed silently
- Fixed double-marshal bug: `BroadcastToRoom` base64-encoded `[]byte` payloads
- Added `DisplayName` lookup on WebSocket connect
- Added chat history loading on join (last 50 messages)
- Added reply-to feature (UI + backend)
- Added `IROOM_DEBUG=1` env var for debug logging

### 4. Real-Time Classroom Features

**Whiteboard sync:**
- Drawing strokes sent via WebSocket (`type: "whiteboard"`)
- Clear actions synced
- Whiteboard open/close state synced across users
- Only admin/teacher can draw (permission check)

**All settings synced via WebSocket:**
- Chat disable/enable (admin broadcasts to all)
- Private message mode
- Clear all messages
- Chat/users panel toggle
- Layout changes

**Raise hand:** Synced via `hand_up`/`hand_down` commands

**Kick:** Admin can kick users — sends `kick` command, target disconnects and redirects

### 5. Video/Audio/Screen Sharing (WebRTC)

- Fixed `onRemoteStream` never firing — server sent `track_added` data channel messages before DC was open → fixed by using `streamId` (set to userID) for direct identification
- Added separate rendering for screen share (full-screen) vs webcam (small overlay)
- Added TURN servers (OpenRelay) for faster connection
- Added detailed WebRTC debug logging in Pion client

### 6. Frontend UI

- Chat messages: left-aligned, full-width gray bubbles, sender name on all messages
- Reply feature with context menu icon on each message bubble
- Chat panel: expand/collapse, disabled state, private mode indicators
- Admin panel: logs page fixed, webhooks fully functional
- Room page: all features working (join, chat, whiteboard, video)

### 7. DevOps

- Added `.air.toml` for Go live-reload (`IROOM_DEBUG=1 air`)
- Updated `README.md` with both manual and air build instructions

---

## What Was Learned

1. **WebSocket + Vite proxy**: The Vite dev proxy doesn't always forward WebSocket upgrades reliably. In dev mode, connecting directly to port 8080 is more reliable.

2. **Svelte 5 `$state` reactivity**: Mutating objects inside `$state` arrays doesn't trigger re-renders. Must reassign the entire array (`remoteStreams = [...remoteStreams, updated]`).

3. **WebRTC data channel timing**: Server-side data channel `SendText` can silently fail if the DC isn't open yet. Need to check `ReadyState()` or use `OnOpen` callback.

4. **WebRTC track identification**: Using `streamId = userID` on `NewTrackLocalStaticRTP` is more reliable than data channel signaling for mapping tracks to participants.

5. **SQLite NULL scanning**: `int64` can't scan NULL — always use `COALESCE(column, 0)` for nullable integer columns when the entity uses plain `int64`.

6. **CSRF + JWT SPAs**: JWT-authenticated SPAs don't need CSRF protection (Bearer tokens can't be sent cross-origin via forms). Skip CSRF when `Authorization` header is present.

7. **Migration idempotency**: If a migration already ran and a new column was added later, the old migration's `ALTER TABLE ADD COLUMN` will fail if the column exists. New migrations must handle this.

8. **Frontend WebSocket batching**: The `writePump` pattern of writing multiple JSON objects separated by `\n` in one frame breaks `JSON.parse`. Send each message as its own WebSocket frame instead.

---

## Files Changed (Key)

### Backend (Go)
- `cmd/server/main.go` — Routes, use case wiring, ICE config
- `internal/adapter/handler/chat.go` — WebSocket chat handler
- `internal/adapter/handler/room.go` — Room CRUD + slug lookup
- `internal/adapter/handler/session.go` — Session create with class_id fallback
- `internal/adapter/handler/webhook.go` — Admin bypass for webhook ops
- `internal/adapter/repository/sqlite/room.go` — Slug name fallback
- `internal/adapter/repository/sqlite/session.go` — COALESCE for nullable columns
- `internal/domain/entity/models.go` — DisplayName on Message
- `internal/domain/usecase/session.go` — Time parsing, class_id compat
- `internal/domain/usecase/webhook.go` — Role-based bypass
- `internal/middleware/csrf.go` — Authorization header bypass
- `internal/pkg/debug/debug.go` — Debug logging package
- `internal/webrtc/signaling.go` — Data channel OnOpen for track info
- `internal/database/migrations/021_backfill_room_slugs.sql`
- `internal/database/migrations/022_fix_sessions_class_id.sql`

### Frontend (Svelte)
- `web/src/routes/room/[slug]/+page.svelte` — Chat, whiteboard, video, all sync
- `web/src/lib/components/classroom/ChatPanel.svelte` — Reply, full-width bubbles
- `web/src/lib/classroom/pion-client.ts` — TURN servers, ontrack streamId fallback
