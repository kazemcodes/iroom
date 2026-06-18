# IRoom Admin Panel — Comprehensive Audit Report

**Date:** 2026-06-18
**Scope:** Admin panel frontend + backend — bugs, security, UX/UI, efficiency
**Severity Scale:** Critical | High | Medium | Low

---

## 1. Frontend Bugs

### 1.1 Dashboard Stats Double-API Call
**File:** `web/src/routes/(app)/admin/+page.svelte:53-99`
**Severity:** Medium

**Problem:** `loadDashboard()` calls `GET /admin/stats` AND separately calls `/classes`, `/sessions`, `/admin/users`, `/admin/recordings`. The stats endpoint already returns these counts. The fallback logic at lines 66-72 reconstructs stats from the separate calls, meaning 5 API calls on every dashboard load.

**Why it's an issue:** On slow connections, the dashboard takes 2-3x longer than necessary. Users see a loading spinner for too long.

**Suggested Fix:** Use the stats endpoint as the single source of truth. Only fall back to separate calls if stats endpoint fails.

---

### 1.2 User Toggle Active — No Error Feedback
**File:** `web/src/routes/(app)/admin/+page.svelte:159-164`
**Severity:** Low

**Problem:** `toggleUserActive()` silently fails if the API returns an error. The user clicks the toggle, nothing happens visually, and there's no error message.

**Why it's an issue:** Admins may think they deactivated a user when they didn't, leading to security concerns.

**Suggested Fix:** Add error handling with toast/alert feedback.

---

### 1.3 Delete User Modal — Misleading Title
**File:** `web/src/routes/(app)/admin/+page.svelte:648`
**Severity:** Low

**Problem:** The `ConfirmModal` for delete user says "غیرفعال‌سازی کاربر" (deactivate user) but `deleteUser()` calls `DELETE /admin/users/${id}` which permanently deactivates. The message is misleading.

**Why it's an issue:** Admin may think they're soft-deactivating when they're actually hard-deactivating.

**Suggested Fix:** Change title to "حذف کاربر" (delete user) and message to clarify this is permanent.

---

### 1.4 Rooms Page — `getStudentsCount` Returns Wrong Data
**File:** `web/src/routes/(app)/admin/rooms/+page.svelte:57-60`
**Severity:** Medium

**Problem:**
```typescript
function getStudentsCount(classId: number) {
    const classSessions = sessions.filter(s => s.class_id === classId);
    return classSessions.length > 0 ? Math.min(classSessions.length * 5, 30) : 0;
}
```
This calculates student count as `sessions * 5` capped at 30. It should query the `class_students` table.

**Why it's an issue:** Rooms page shows incorrect student counts, confusing admins.

**Suggested Fix:** Add API call to `GET /classes/${id}/students` and use the actual count.

---

### 1.5 Settings Page — No Form Validation
**File:** `web/src/routes/(app)/admin/settings/+page.svelte`
**Severity:** Medium

**Problem:** Settings form has no client-side validation. Admin can enter negative numbers for `max_users_per_room`, zero for `session_timeout_minutes`, etc. The backend also doesn't validate these values.

**Why it's an issue:** Invalid settings can cause runtime errors or security issues (e.g., zero timeout = no timeout).

**Suggested Fix:** Add validation rules (min/max values, required fields) with error messages.

---

### 1.6 Bulk Import CSV Parsing Doesn't Handle Quoted Fields
**File:** `web/src/routes/(app)/admin/users/+page.svelte:148-160`
**Severity:** Medium

**Problem:** `parseCSVPreview()` uses `line.split(',')` which breaks on CSV fields containing commas (e.g., `"Smith, John",john@test.com`).

**Why it's an issue:** Admins importing users with commas in names get garbled data.

**Suggested Fix:** Use a proper CSV parsing library (e.g., PapaParse) or handle quoted fields with regex.

---

### 1.7 Admin Layout — Token Not Validated on Initial Load
**File:** `web/src/routes/(app)/+layout.svelte:58-61`
**Severity:** Medium

**Problem:** `auth.init()` restores state from localStorage but doesn't validate the token against the server. If the token is expired, the user sees the admin shell but all API calls fail with 401. There's a flash of the admin UI before redirect.

**Why it's an issue:** Poor UX — user sees a broken admin panel before being redirected.

**Suggested Fix:** Add a token validation call (`GET /auth/me`) on mount. Show a "verifying..." state until confirmed.

---

### 1.8 Tickets Page — Reply Doesn't Show Errors
**File:** `web/src/routes/(app)/admin/tickets/+page.svelte:51-60`
**Severity:** Low

**Problem:** `sendReply()` silently fails if the API returns an error. No feedback to the admin user.

**Why it's an issue:** Admin doesn't know if the reply was sent or failed.

**Suggested Fix:** Add error handling with toast/alert feedback.

---

## 2. Backend Bugs

### 2.1 Chat WebSocket Doesn't Set `created_at`
**File:** `internal/adapter/handler/chat.go:114-120`
**Severity:** High

**Problem:**
```go
chatMsg := &entity.Message{
    SessionID: sessionID,
    UserID:    client.UserID,
    Content:   msg.Content,
    Type:      "text",
}
```
`CreatedAt` is not set. The database has `DEFAULT CURRENT_TIMESTAMP` so the row gets a timestamp, but the broadcast message at line 128 has `created_at: ""` (empty string). The frontend shows an empty timestamp until page refresh.

**Why it's an issue:** Users see messages with no timestamp in real-time chat.

**Suggested Fix:** Set `CreatedAt: time.Now()` explicitly when creating the message.

---

### 2.2 User Update Doesn't Prevent Role Escalation
**File:** `internal/adapter/handler/user.go:63-81`
**Severity:** Critical (Security)

**Problem:** `Update()` allows any admin to change a user's role to anything, including promoting themselves or others to admin. There's no check to prevent privilege escalation.

**Why it's an issue:** A compromised admin account could create more admin accounts or escalate privileges.

**Suggested Fix:** Add a check: only users with role "owner" can create/modify admins. Prevent self-escalation.

---

### 2.3 Guest Login Creates Persistent Accounts
**File:** `internal/domain/usecase/auth.go:153-154`
**Severity:** Medium

**Problem:**
```go
guestEmail := fmt.Sprintf("guest_%d_%d@iroom.local", sessionID, time.Now().UnixMilli())
hashedPassword, _ := uc.hasher.Hash("guest_no_password")
```
Guest users are created with a dummy password and persist in the database indefinitely.

**Why it's an issue:** Database bloat from thousands of guest accounts. Potential security issue if guest accounts are ever exposed.

**Suggested Fix:** Add a cleanup job that deletes guest accounts older than 24 hours. Or use a separate `guest_users` table with no password.

---

### 2.4 Rate Limiter Memory Leak
**File:** `internal/middleware/ratelimit.go:27-52`
**Severity:** Medium

**Problem:** The `rateLimiter` stores request timestamps in a map keyed by IP. It cleans up expired entries per-IP, but never removes IPs with zero valid requests from the map. Over time, the map grows unbounded.

**Why it's an issue:** Memory grows indefinitely in production with many unique IPs.

**Suggested Fix:** Add a periodic cleanup goroutine that removes stale IPs, or use a TTL cache.

---

### 2.5 Session Start Doesn't Verify Class Existence
**File:** `internal/domain/usecase/session.go:48-69`
**Severity:** Low

**Problem:** `Start()` checks if the user is admin or teacher of the class, but the `classRepo.GetByID` error at line 55 is silently ignored with `if err != nil || class.TeacherID != userID`.

**Why it's an issue:** If the class was deleted, the session still gets started with a reference to a non-existent class.

**Suggested Fix:** Return an explicit error if the class doesn't exist.

---

### 2.6 WebSocket Hub — Room ID Not Set on Chat Client
**File:** `internal/services/ws_hub.go:163-193`, `internal/adapter/handler/chat.go:68-76`
**Severity:** Medium

**Problem:** `BroadcastToRoom()` checks `client.RoomID != roomID` but the chat handler sets `RoomID` on the client struct. If a user connects to multiple sessions, messages could be misdirected or missed.

**Why it's an issue:** Messages may not reach the intended recipients in multi-session scenarios.

**Suggested Fix:** Ensure `RoomID` is correctly set and validated. Consider using a map of roomID -> set of clients.

---

### 2.7 No CSRF Protection
**File:** `internal/middleware/` (missing)
**Severity:** Critical (Security)

**Problem:** No CSRF middleware exists. State-changing operations rely solely on JWT auth. If tokens are ever stored in cookies, CSRF becomes a real vulnerability.

**Why it's an issue:** Potential for cross-site request forgery attacks on state-changing endpoints.

**Suggested Fix:** Add CSRF middleware (e.g., `echo-csrf`) or use double-submit cookie pattern.

---

### 2.8 File Upload — No Server-Side Size Validation
**File:** `internal/adapter/handler/file.go`
**Severity:** Medium

**Problem:** The file upload handler doesn't validate file size on the server side. Frontend limits can be bypassed.

**Why it's an issue:** Malicious clients can upload arbitrarily large files, exhausting disk space.

**Suggested Fix:** Add server-side size validation with a configurable maximum.

---

## 3. Security Vulnerabilities

### 3.1 No Input Sanitization on Display Names (XSS)
**File:** `internal/adapter/handler/user.go`, `internal/adapter/handler/auth.go`
**Severity:** Critical

**Problem:** User display names are stored and rendered without sanitization. An admin could set a display name containing `<script>` tags, which would execute in other users' browsers.

**Why it's an issue:** Stored XSS attack vector.

**Suggested Fix:** Use a sanitization library (e.g., bluemonday for Go) to strip HTML from display names before storage.

---

### 3.2 No Audit Log for Admin Actions
**File:** `internal/adapter/handler/` (multiple files)
**Severity:** High

**Problem:** Admin actions (create user, delete class, change role) are not logged to the activity log. Only auth events are logged.

**Why it's an issue:** No accountability trail for admin actions. Cannot investigate security incidents.

**Suggested Fix:** Add audit logging middleware that logs all state-changing admin actions.

---

### 3.3 JWT Secret — No Rotation Mechanism
**File:** `internal/config/config.go`
**Severity:** Medium

**Problem:** The JWT secret is a single static value. If compromised, all tokens are immediately invalidated.

**Why it's an issue:** No way to rotate keys without invalidating all existing sessions.

**Suggested Fix:** Support multiple keys (current + previous) for graceful rotation.

---

### 3.4 No HTTPS Enforcement
**File:** `cmd/server/main.go`
**Severity:** Medium

**Problem:** The server doesn't redirect HTTP to HTTPS or set HSTS headers.

**Why it's an issue:** JWT tokens sent over HTTP are vulnerable to interception.

**Suggested Fix:** Add HTTPS redirect middleware and HSTS header in production mode.

---

### 3.5 Rate Limiter Uses In-Memory Map (Not Shared Across Instances)
**File:** `internal/middleware/ratelimit.go`
**Severity:** Medium

**Problem:** The rate limiter uses an in-memory map. If the server runs multiple instances (behind a load balancer), rate limits are per-instance, not global.

**Why it's an issue:** Attackers can bypass rate limits by distributing requests across instances.

**Suggested Fix:** Use Redis or a shared data store for rate limiting in production.

---

## 4. UX/UI Improvements

### 4.1 No Loading Skeletons for Data Tables
**Files:** All admin list pages
**Problem:** Tables show a generic spinner. Users don't know what's loading.
**Improvement:** Use skeleton loaders that match the table structure.

---

### 4.2 No Empty State Illustrations
**Files:** All admin list pages
**Problem:** Empty states show plain text like "کاربری یافت نشد".
**Improvement:** Add helpful illustrations and action buttons (e.g., "Create your first user").

---

### 4.3 No Real-Time Dashboard Updates
**File:** `web/src/routes/(app)/admin/+page.svelte`
**Problem:** Dashboard stats only load on page refresh.
**Improvement:** Connect to the WebSocket hub for real-time dashboard updates.

---

### 4.4 No Export Functionality
**Files:** All admin list pages
**Problem:** Admins can't export data to CSV/Excel.
**Improvement:** Add export buttons to all data tables.

---

### 4.5 No Bulk Actions
**Files:** All admin list pages
**Problem:** Admins can only act on one item at a time.
**Improvement:** Add checkboxes and bulk actions (bulk delete, bulk activate).

---

### 4.6 No Responsive Design for Mobile Admin
**Files:** All admin pages
**Problem:** Admin panel is not optimized for mobile/tablet.
**Improvement:** Add responsive breakpoints, collapsible tables for mobile.

---

### 4.7 No Search Highlighting
**Files:** All admin list pages
**Problem:** Search results don't highlight matching text.
**Improvement:** Highlight matching text in search results.

---

### 4.8 No Undo for Destructive Actions
**Files:** All admin pages with delete/deactivate
**Problem:** No way to undo accidental deletions from the UI.
**Improvement:** Add a toast with "Undo" action that temporarily soft-deletes.

---

### 4.9 No Keyboard Shortcuts
**Files:** All admin pages
**Problem:** Admin users can't use keyboard shortcuts.
**Improvement:** Add shortcuts (Ctrl+K for search, Ctrl+N for new).

---

### 4.10 No Breadcrumb Navigation
**File:** `web/src/routes/(app)/+layout.svelte`
**Problem:** Users can't easily navigate back without the sidebar.
**Improvement:** Add breadcrumb navigation.

---

## 5. Efficiency Improvements

### 5.1 Dashboard Stats — N+1 Query Pattern
**File:** `internal/adapter/handler/dashboard.go`
**Problem:** Stats handler likely makes separate queries for each stat.
**Improvement:** Use a single query with COUNT and GROUP BY.

---

### 5.2 No Database Connection Pool Configuration
**File:** `internal/database/db.go`
**Problem:** SQLite uses default connection pool settings.
**Improvement:** Configure SetMaxOpenConns, SetMaxIdleConns, SetConnMaxLifetime.

---

### 5.3 No Response Caching
**File:** `internal/adapter/handler/` (multiple)
**Problem:** Frequently accessed endpoints never cached.
**Improvement:** Add in-memory caching with TTL for read-heavy endpoints.

---

### 5.4 No Compression Middleware
**File:** `cmd/server/main.go`
**Problem:** No gzip/brotli compression. Large JSON responses sent uncompressed.
**Improvement:** Add `echoMiddleware.Gzip()` middleware.

---

### 5.5 Activity Log Writes Synchronously
**File:** `internal/adapter/handler/` (multiple)
**Problem:** Activity log writes happen in the request handler, adding latency.
**Improvement:** Use a channel/queue for async activity logging.

---

### 5.6 No Maximum Page Size Enforcement
**File:** `internal/adapter/repository/sqlite/` (multiple)
**Problem:** If per_page is 0 or very large, queries return all results.
**Improvement:** Enforce maximum page size (e.g., 100) and always paginate.

---

### 5.7 WebSocket Broadcasts Send Full State
**File:** `internal/services/ws_hub.go`
**Problem:** Broadcasts send full payload to all clients.
**Improvement:** Use delta updates — only send what changed.

---

## 6. Implementation Plan

### Phase 1: Critical Security Fixes (Priority 1) ✅ COMPLETED

**Task 1.1: Add XSS Sanitization** ✅
- Files: `internal/adapter/handler/user.go`, `internal/adapter/handler/auth.go`
- Add bluemonday sanitization to display names before storage

**Task 1.2: Add Role Escalation Prevention** ✅
- File: `internal/adapter/handler/user.go`
- Add check: only owners can create admins

**Task 1.3: Add CSRF Middleware** ✅
- File: `internal/middleware/csrf.go` (new)
- Add CSRF protection to state-changing endpoints

**Task 1.4: Add Audit Logging** ✅
- File: `internal/middleware/audit.go` (new)
- Log all admin actions to activity_log

---

### Phase 2: High-Priority Bug Fixes (Priority 2) ✅ COMPLETED

**Task 2.1: Fix Message `created_at`** ✅
- File: `internal/adapter/handler/chat.go:114-120`
- Set `CreatedAt: time.Now()` explicitly

**Task 2.2: Fix Dashboard API Calls** ✅
- File: `web/src/routes/(app)/admin/+page.svelte:53-99`
- Use stats endpoint as single source of truth

**Task 2.3: Fix Student Count** ✅
- File: `web/src/routes/(app)/admin/rooms/+page.svelte:57-60`
- Query actual student count from API

**Task 2.4: Fix Token Validation on Load** ✅
- File: `web/src/routes/(app)/+layout.svelte:58-61`
- Add server-side token validation

---

### Phase 3: Medium-Priority Fixes (Priority 3) ✅ COMPLETED

**Task 3.1: Fix Rate Limiter Memory Leak** ✅
- File: `internal/middleware/ratelimit.go`
- Add stale IP cleanup

**Task 3.2: Add Guest Account Cleanup** ✅
- File: `internal/domain/usecase/auth.go`
- Add cleanup job for old guest accounts

**Task 3.3: Add Settings Validation** ✅
- File: `web/src/routes/(app)/admin/settings/+page.svelte`
- Add client-side and server-side validation

**Task 3.4: Fix CSV Parsing** ✅
- File: `web/src/routes/(app)/admin/users/+page.svelte:148-160`
- Use proper CSV library

---

### Phase 4: UX/UI Improvements (Priority 4) ✅ COMPLETED

**Task 4.1: Add Skeleton Loaders** ✅
- Files: All admin list pages

**Task 4.2: Add Empty State Illustrations** ✅
- Files: All admin list pages

**Task 4.3: Add Real-Time Dashboard** ✅
- File: `web/src/routes/(app)/admin/+page.svelte`

**Task 4.4: Add Export Buttons** ✅
- Files: All admin list pages

**Task 4.5: Add Bulk Actions** ✅
- Files: All admin list pages

---

### Phase 5: Efficiency Improvements (Priority 5) ✅ COMPLETED

**Task 5.1: Optimize Dashboard Stats Query** ✅
- File: `internal/adapter/handler/dashboard.go`

**Task 5.2: Configure Connection Pool** ✅
- File: `internal/database/db.go`

**Task 5.3: Add Compression Middleware** ✅
- File: `cmd/server/main.go`

**Task 5.4: Add Response Caching** ✅
- File: `internal/adapter/handler/` (multiple)

**Task 5.5: Add Async Activity Logging** ✅
- File: `internal/adapter/handler/` (multiple)

---

## Summary

| Priority | Count | Key Issues |
|----------|-------|------------|
| Critical | 4 | XSS, role escalation, CSRF, no audit log |
| High | 3 | Message timestamps, dashboard API, token validation |
| Medium | 5 | Rate limiter, guest cleanup, settings validation, CSV parsing |
| Low (UX) | 10 | Skeletons, empty states, real-time, export, bulk actions |
| Efficiency | 5 | N+1 queries, connection pool, compression, caching, async logging |

---

## Files Reviewed

### Frontend
- `web/src/routes/(app)/admin/+page.svelte`
- `web/src/routes/(app)/admin/users/+page.svelte`
- `web/src/routes/(app)/admin/rooms/+page.svelte`
- `web/src/routes/(app)/admin/settings/+page.svelte`
- `web/src/routes/(app)/admin/tickets/+page.svelte`
- `web/src/routes/(app)/+layout.svelte`
- `web/src/lib/api.ts`
- `web/src/lib/stores.ts`

### Backend
- `internal/adapter/handler/auth.go`
- `internal/adapter/handler/user.go`
- `internal/adapter/handler/class.go`
- `internal/adapter/handler/session.go`
- `internal/adapter/handler/chat.go`
- `internal/adapter/handler/file.go`
- `internal/domain/usecase/auth.go`
- `internal/domain/usecase/session.go`
- `internal/middleware/auth.go`
- `internal/middleware/ratelimit.go`
- `internal/services/ws_hub.go`
- `internal/database/db.go`
- `cmd/server/main.go`
