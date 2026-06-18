# API Architecture Redesign — Skyroom-Compatible

## Skyroom API Structure (from webservice.html)

Skyroom uses a **single endpoint** with `action` field:
```
POST /skyroom/api/{api-key}
{ "action": "actionName", "params": {...} }
```

Response format:
```json
{ "ok": true, "result": ... }
{ "ok": false, "error_code": 14, "error_message": "..." }
```

## Skyroom API Functions

### 1. Services (سرویس‌ها)
| Function | Description | iroom Equivalent |
|----------|-------------|------------------|
| `getServices` | List services | ❌ Not needed (iroom doesn't have services) |

### 2. Rooms (اتاق‌ها) → Maps to Classes + Sessions
| Function | Description | Current iroom | Status |
|----------|-------------|---------------|--------|
| `getRooms` | List rooms | `GET /classes` | ✅ Exists |
| `countRooms` | Count rooms | `GET /admin/dashboard/stats` | ⚠️ Separate |
| `getRoom` | Get room by id/name | `GET /classes/:id` | ✅ Exists |
| `getRoomUrl` | Get room URL | ❌ Missing | **NEW** |
| `createRoom` | Create room | `POST /classes` | ✅ Exists |
| `updateRoom` | Update room | `PUT /classes/:id` | ✅ Exists |
| `deleteRoom` | Delete room | `DELETE /classes/:id` | ✅ Exists |
| `getRoomUsers` | Get room users | `GET /classes/:id/students` | ⚠️ Different name |
| `addRoomUsers` | Add users to room | `POST /classes/:id/enroll` | ⚠️ Different name |
| `removeRoomUsers` | Remove users from room | ❌ Missing | **NEW** |
| `updateRoomUser` | Update user access | ❌ Missing | **NEW** |

### 3. Users (کاربران)
| Function | Description | Current iroom | Status |
|----------|-------------|---------------|--------|
| `getUsers` | List users | `GET /admin/users` | ⚠️ Admin-only |
| `countUsers` | Count users | `GET /admin/dashboard/stats` | ⚠️ Separate |
| `getUser` | Get user by id/username | `GET /auth/me` | ⚠️ Self only |
| `createUser` | Create user | `POST /admin/users` | ⚠️ Admin-only |
| `updateUser` | Update user | `PUT /admin/users/:id` | ⚠️ Admin-only |
| `deleteUser` | Delete user | `DELETE /admin/users/:id` | ⚠️ Admin-only |
| `deleteUsers` | Delete multiple users | ❌ Missing | **NEW** |
| `getUserRooms` | Get user's rooms | ❌ Missing | **NEW** |
| `addUserRooms` | Add rooms to user | ❌ Missing | **NEW** |
| `removeUserRooms` | Remove rooms from user | ❌ Missing | **NEW** |
| `createLoginUrl` | Direct login URL | ❌ Missing | **NEW** |

### 4. Classroom (added by iroom, not in Skyroom)
| Endpoint | Description | Status |
|----------|-------------|--------|
| `POST /sessions/:id/classroom/offer` | WebRTC signaling | ✅ iroom-specific |
| `POST /sessions/:id/classroom/candidate` | ICE candidate | ✅ iroom-specific |
| `GET /sessions/:id/classroom/participants` | List participants | ✅ iroom-specific |
| `POST /sessions/:id/classroom/mute/:id` | Mute participant | ✅ iroom-specific |
| `POST /sessions/:id/classroom/kick/:id` | Kick participant | ✅ iroom-specific |

### 5. Sessions (جلسات) — iroom-specific, not in Skyroom
| Endpoint | Description | Status |
|----------|-------------|--------|
| `GET /sessions` | List sessions | ✅ Exists |
| `POST /sessions` | Create session | ✅ Exists |
| `POST /sessions/:id/start` | Start session | ✅ Exists |
| `POST /sessions/:id/end` | End session | ✅ Exists |

### 6. Chat (پیام‌ها) — iroom-specific
| Endpoint | Description | Status |
|----------|-------------|--------|
| `GET /sessions/:id/messages` | List messages | ✅ Exists |
| `POST /sessions/:id/messages` | Send message | ✅ Exists |
| `WS /ws/sessions/:id` | Real-time chat | ✅ Exists |

### 7. Files, Recordings, Tickets, Announcements, Polls — iroom-specific
All exist. No changes needed.

## New Endpoints to Add

### Room URL
```go
GET /api/v1/classes/:id/url
// Returns the join URL for a class/session
```

### Room Users (rename existing + add missing)
```go
GET    /api/v1/classes/:id/users          // was /students
POST   /api/v1/classes/:id/users          // was /enroll
DELETE /api/v1/classes/:id/users/:userId  // NEW: remove user from class
PUT    /api/v1/classes/:id/users/:userId  // NEW: update user access level
```

### User Rooms
```go
GET /api/v1/users/:id/rooms              // NEW: rooms a user belongs to
```

### Batch Operations
```go
DELETE /api/v1/admin/users/batch         // NEW: delete multiple users
```

### Direct Login URL (Skyroom-style)
```go
POST /api/v1/auth/create-login-url
// Body: { room_id, user_id, nickname, access, concurrent, ttl }
// Returns: { url: "https://..." }
```

## Response Format Standardization

Current iroom:
```json
{ "success": true, "data": {...} }
{ "success": false, "error": "message" }
```

Skyroom:
```json
{ "ok": true, "result": ... }
{ "ok": false, "error_code": 14, "error_message": "..." }
```

**Decision**: Keep iroom's format (it's cleaner for frontend). Don't change response format — just add missing endpoints.

## Route Cleanup

### Remove/Old (currently broken or unused)
- `GET /admin/activity-logs` — 404 in logs
- `POST /sessions/:id/logs/join` — not used by frontend
- `POST /sessions/:id/logs/leave` — not used by frontend

### Consolidate
- Merge `classHandler.GetStudents` → `classHandler.GetUsers` (rename)
- Merge `classHandler.Enroll` → `classHandler.AddUsers` (rename)

## Implementation Priority

1. **Fix broken endpoints** (activity-logs 404)
2. **Add missing room-user endpoints** (remove, update access)
3. **Add user-rooms endpoint** (get rooms for user)
4. **Add createLoginUrl** (direct join link generation)
5. **Add batch delete users**
6. **Add room URL endpoint**
7. **Clean up unused endpoints**
