# API Documentation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require a JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are obtained via `/auth/login` or `/auth/guest-login`.

## Response Format

All responses follow this structure:

```json
{
  "success": true,
  "data": { ... },
  "error": "error message"
}
```

Paginated responses include:
```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "per_page": 20,
    "total_pages": 5
  }
}
```

## Rate Limits

| Endpoint Group | Limit | Window |
|---------------|-------|--------|
| General | 100 req | 1 min |
| Auth endpoints | 100 req | 1 min |
| WebSocket | 30 req | 1 min |

---

## Public Endpoints (No Auth)

### Health Check
```
GET /api/v1/health
```
Returns server status, uptime, database size, active rooms, total users.

**Response:**
```json
{
  "status": "ok",
  "uptime": "2h 15m",
  "db_size": "12 MB",
  "webrtc_status": "pion_builtin",
  "active_rooms": 3,
  "total_users": 45
}
```

### Public Session Info
```
GET /api/v1/sessions/:id/info
```
Returns basic session info without authentication. Used by the join page.

**Response:**
```json
{
  "id": 4,
  "title": "ریاضی پایه دهم",
  "status": "live"
}
```

---

## Auth Endpoints

### Register
```
POST /api/v1/auth/register
```
Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure123",
  "display_name": "نام کاربر",
  "phone": "09120000000"
}
```

**Response:**
```json
{
  "user": { "id": 1, "email": "user@example.com", "role": "student" },
  "tokens": { "access_token": "...", "refresh_token": "...", "expires_in": 900 }
}
```

### Login
```
POST /api/v1/auth/login
```
Authenticate with email and password.

**Request:**
```json
{
  "email": "admin@iroom.local",
  "password": "admin123"
}
```

**Response:** Same as Register.

### Guest Login
```
POST /api/v1/auth/guest-login
```
Create a temporary guest account for joining a class session. Session must be "live".

**Request:**
```json
{
  "session_id": 4,
  "display_name": "دانش‌آموز"
}
```

**Response:** Same as Register (guest user with auto-generated email).

### Refresh Token
```
POST /api/v1/auth/refresh
```
Get new access/refresh tokens from a valid refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiJ9..."
}
```

**Response:**
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 900
}
```

### Create Login URL
```
POST /api/v1/auth/create-login-url
```
Generate a direct login URL for a room (no account needed).

**Request:**
```json
{
  "room_id": 1,
  "user_id": "student-123",
  "nickname": "علی",
  "access": 1,
  "concurrent": 1,
  "ttl": 3600,
  "language": "fa"
}
```

**Response:**
```json
{
  "url": "/classroom/join/1?token=eyJ...&nickname=علی&access=1"
}
```

### Forgot Password
```
POST /api/v1/auth/forgot-password
```
Send password reset link to email.

### Reset Password
```
POST /api/v1/auth/reset-password
```
Reset password with token.

---

## Protected Endpoints (JWT Required)

### Profile
```
GET  /api/v1/auth/me              — Get current user info
PUT  /api/v1/auth/me              — Update profile
POST /api/v1/auth/change-password — Change password
POST /api/v1/auth/avatar         — Upload avatar
```

### Classes
```
GET    /api/v1/classes                    — List classes (paginated)
POST   /api/v1/classes                    — Create class
GET    /api/v1/classes/:id                — Get class by ID
PUT    /api/v1/classes/:id                — Update class
DELETE /api/v1/classes/:id                — Delete class
POST   /api/v1/classes/:id/enroll         — Enroll student
DELETE /api/v1/classes/:id/users/:userId  — Remove student from class
PUT    /api/v1/classes/:id/users/:userId  — Update student access level
GET    /api/v1/classes/:id/students       — List enrolled students
GET    /api/v1/classes/:id/url            — Get class join URL
POST   /api/v1/classes/:id/regenerate-code — Regenerate invite code
POST   /api/v1/classes/join/:code         — Join class by invite code
GET    /api/v1/users/:id/rooms            — Get user's rooms
```

### Sessions
```
GET    /api/v1/sessions              — List sessions (paginated, filterable by class_id)
POST   /api/v1/sessions              — Create session
GET    /api/v1/sessions/:id          — Get session by ID
POST   /api/v1/sessions/:id/start    — Start session (teacher/admin only)
POST   /api/v1/sessions/:id/end      — End session (teacher/admin only)
DELETE /api/v1/sessions/:id          — Delete session (teacher/admin only)
```

### WebRTC (Classroom)
```
GET    /api/v1/sessions/:id/classroom              — Get join info (room_id, user_id, role)
POST   /api/v1/sessions/:id/classroom/offer        — Send SDP offer, receive answer
POST   /api/v1/sessions/:id/classroom/candidate    — Send ICE candidate
DELETE /api/v1/sessions/:id/classroom/:userId       — Leave room
GET    /api/v1/sessions/:id/classroom/participants  — List connected participants
POST   /api/v1/sessions/:id/classroom/mute/:participantId — Mute participant (teacher/admin)
POST   /api/v1/sessions/:id/classroom/kick/:participantId — Kick participant (teacher/admin)
```

### Messages
```
GET  /api/v1/sessions/:id/messages — List messages (paginated)
POST /api/v1/sessions/:id/messages — Send message
```

### Files
```
POST   /api/v1/sessions/:id/files      — Upload file
GET    /api/v1/sessions/:id/files      — List files
GET    /api/v1/files/:id/download      — Download file
DELETE /api/v1/files/:id               — Delete file
```

### Recordings
```
POST /api/v1/sessions/:id/recordings      — Upload recording
GET  /api/v1/sessions/:id/recordings      — List recordings
GET  /api/v1/recordings/:id/download      — Download recording
```

### Announcements
```
POST   /api/v1/classes/:id/announcements — Create announcement
GET    /api/v1/classes/:id/announcements — List announcements
PUT    /api/v1/announcements/:id         — Update announcement
DELETE /api/v1/announcements/:id         — Delete announcement
POST   /api/v1/announcements/:id/pin     — Toggle pin
```

### Polls
```
POST /api/v1/sessions/:id/polls      — Create poll
GET  /api/v1/sessions/:id/polls      — List polls
POST /api/v1/polls/:id/vote          — Vote in poll
GET  /api/v1/polls/:id/results       — Get poll results
POST /api/v1/polls/:id/close         — Close poll
```

### Tickets
```
POST /api/v1/tickets          — Create ticket
GET  /api/v1/tickets          — List my tickets (paginated)
GET  /api/v1/tickets/:id      — Get ticket by ID
POST /api/v1/tickets/:id/reply — Reply to ticket
POST /api/v1/tickets/:id/close — Close ticket
```

### Notifications
```
GET  /api/v1/notifications             — List notifications (paginated)
GET  /api/v1/notifications/unread-count — Get unread count
POST /api/v1/notifications/:id/read    — Mark as read
POST /api/v1/notifications/read-all    — Mark all as read
```

### WebSocket
```
WS /ws/sessions/:id — Real-time chat (rate limited: 30/min)
```

---

## Admin Endpoints (JWT + Admin Role)

### Dashboard
```
GET /api/v1/admin/dashboard/stats — System statistics
```

### Users
```
GET    /api/v1/admin/users              — List users (paginated)
POST   /api/v1/admin/users              — Create user
POST   /api/v1/admin/users/batch-delete — Delete multiple users
PUT    /api/v1/admin/users/:id          — Update user
DELETE /api/v1/admin/users/:id          — Deactivate user
```

### Classes (Admin)
```
GET    /api/v1/admin/classes            — List all classes
POST   /api/v1/admin/classes            — Create class
PUT    /api/v1/admin/classes/:id        — Update class
DELETE /api/v1/admin/classes/:id        — Delete class
```

### Sessions (Admin)
```
GET    /api/v1/admin/sessions           — List all sessions
GET    /api/v1/admin/sessions/:id       — Get session by ID
DELETE /api/v1/admin/sessions/:id       — Delete session
```

### Recordings (Admin)
```
GET    /api/v1/admin/recordings         — List all recordings
DELETE /api/v1/admin/recordings/:id     — Delete recording
```

### Settings
```
GET  /api/v1/admin/settings  — Get all settings
PUT  /api/v1/admin/settings  — Update settings
```

### Webhooks
```
POST   /api/v1/admin/webhooks            — Create webhook
GET    /api/v1/admin/webhooks            — List webhooks
PUT    /api/v1/admin/webhooks/:id        — Update webhook
DELETE /api/v1/admin/webhooks/:id        — Delete webhook
GET    /api/v1/admin/webhooks/:id/deliveries — List deliveries
POST   /api/v1/admin/webhooks/:id/test   — Test webhook
```

### Logs
```
GET /api/v1/admin/logs — List activity logs
```

### Tickets (Admin)
```
GET /api/v1/admin/tickets — List all tickets
```

---

## Middleware

### Auth (`middleware/auth.go`)

**`Auth(secret)`** — JWT authentication middleware.
- Extracts `Authorization: Bearer <token>` header
- Validates JWT signature and expiry
- Sets `user_id`, `email`, `role` in context
- Returns 401 if token is missing, invalid, or expired

**`AdminOnly()`** — Admin role check.
- Must be used after `Auth()` middleware
- Returns 403 if role is not "admin"

**`TeacherOrAdmin()`** — Teacher or admin role check.
- Returns 403 if role is not "admin" or "teacher"

**`APIKeyAuth(validKey)`** — API key authentication.
- Checks `X-API-Key` header or `api_key` query parameter
- Used for external API integrations

### CORS (`middleware/cors.go`)

**`CORS()`** — Cross-Origin Resource Sharing.
- Allows all origins (`*`)
- Allows GET, POST, PUT, DELETE, OPTIONS
- Allows Content-Type and Authorization headers
- Handles OPTIONS preflight requests

### Maintenance Mode (`middleware/maintenance.go`)

**`MaintenanceMode(db, jwtSecret)`** — System maintenance mode.
- Checks `maintenance_mode` setting in database (cached 30s)
- Returns 503 if maintenance mode is enabled
- Allows through: health check, login, register, admin users
- Skips check for admin role users

### Rate Limiting (`middleware/ratelimit.go`)

**`RateLimit(limit, window)`** — Per-IP rate limiter.
- Tracks requests per IP address
- Returns 429 when limit exceeded
- Default: 100 requests per minute

**`AuthRateLimit()`** — Auth-specific rate limit (100/min).

**`APIKeyRateLimit()`** — API key rate limit (120/min).

---

## Error Codes

| Code | Meaning |
|------|---------|
| 400 | Bad request (invalid data) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not found |
| 429 | Too many requests (rate limit) |
| 500 | Internal server error |
| 503 | Service unavailable (maintenance mode) |
