# IRoom — UX/UI Gaps & Improvement Analysis

**Date:** 2026-06-16  
**Scope:** Full-stack review of frontend, backend, and infrastructure  
**Goal:** Identify missing features, UX/UI gaps, and actionable improvements

---

## 1. Executive Summary

IRoom is a functional Persian-language online classroom platform with video (LiveKit), whiteboard, chat, file sharing, session management, and a basic admin panel. The foundation is solid, but several **critical UX/UI gaps** and **missing features** prevent it from being production-ready for real classroom scenarios. This document catalogs all findings organized by severity.

---

## 2. Critical Issues (Must Fix)

### 2.1 No Real-Time Notification System
- **Problem:** Users have no way to know when a session starts, when they're enrolled in a class, or when someone replies to their support ticket. The only way to discover anything is to manually refresh the page.
- **Impact:** Breaks the core classroom workflow — students miss sessions.
- **Missing:** WebSocket-based notification hub, browser push notifications, in-app notification bell with unread count.
- **Backend gap:** No notification model/table, no WebSocket broadcast mechanism beyond chat data channels.

### 2.2 No Password Reset / Recovery Flow
- **Problem:** Auth has register/login/refresh but no "forgot password" endpoint or UI. If a user forgets their password, they're permanently locked out.
- **Impact:** Blocking issue for any real deployment.
- **Missing:** `POST /api/v1/auth/forgot-password`, `POST /api/v1/auth/reset-password`, email service integration, reset token storage, frontend forgot-password page.

### 2.3 No Email Verification
- **Problem:** Registration doesn't verify email ownership. Anyone can register with any email.
- **Impact:** Security and spam risk.
- **Missing:** Email verification token, verification endpoint, resend-verification endpoint.

### 2.4 No User Profile Page / Avatar Support
- **Problem:** Users can update display_name and phone via `PUT /api/v1/auth/me`, but there's no dedicated profile page in the UI. The sidebar shows only a letter avatar.
- **Impact:** Users can't see or manage their own profile, upload avatars, or change passwords.
- **Missing:** Profile page route, avatar upload API, change-password endpoint, avatar display throughout the app.

### 2.5 No Pagination on Most List Views
- **Problem:** Classes, sessions, files, and students lists load everything at once with no pagination. The admin recordings and logs pages have pagination, but most user-facing lists don't.
- **Impact:** Performance degrades rapidly with real usage. A class with 100 students loads all at once.
- **Affected pages:** `/classes`, `/sessions`, `/files`, `/classes/[id]` (students tab), `/support`.
- **Backend gap:** Several repository `List` methods don't support offset/limit.

### 2.6 No Error Boundaries or Global Error Handling (Frontend)
- **Problem:** If any page throws an unhandled error, the entire UI crashes with a blank screen. There's no Svelte error boundary, no toast notification system, and no graceful error recovery.
- **Impact:** Single component failure breaks the whole app.
- **Missing:** Root error boundary component, toast/notification store, consistent error feedback UI.

---

## 3. Major UX/UI Gaps

### 3.1 Classroom Experience

| Issue | Details |
|-------|---------|
| **No hand-raise mechanism** | Students can't signal the teacher. Critical for classroom management. |
| **No participant mute/remove (teacher)** | Teachers can't mute disruptive students or remove participants. |
| **No breakout rooms** | Can't split students into smaller groups. |
| **No session timer / countdown** | No visible elapsed time or remaining time during a live session. |
| **No "raise hand" or reaction emojis** | Students have no way to interact non-verbally. |
| **Video grid is static 2-column** | Doesn't adapt to participant count (1 person should be full-screen, 4 should be 2x2, etc.) |
| **No fullscreen mode** | Can't fullscreen the video or whiteboard area. |
| **No picture-in-picture for local video** | Local video is stuck in the bottom-left corner with fixed size. |
| **No bandwidth indicator** | Users don't know if their connection is poor. |
| **No "reconnecting..." state** | When LiveKit disconnects, there's no visual feedback about reconnection attempts. |
| **Whiteboard has no laser pointer** | Teacher can't point at specific areas. |
| **Whiteboard has no undo history sync** | Undo only works locally, not synced to other participants. |
| **Whiteboard canvas is fixed 800x600** | Doesn't adapt to screen size, wastes space on large screens. |
| **No whiteboard export** | Can't save whiteboard as image/PDF. |
| **Chat has no file sharing** | Can't send images or files in chat. |
| **Chat has no @mentions** | Can't tag specific participants. |
| **Chat has no unread indicator** | When chat is closed, users don't know if new messages arrived. |
| **No session recording indicator for participants** | Only the person who started recording sees the indicator. |

### 3.2 Dashboard

| Issue | Details |
|-------|---------|
| **No upcoming session countdown** | Dashboard shows sessions but no "starts in 15 minutes" indicator. |
| **No quick-join for live sessions** | Live sessions aren't prominently highlighted on the dashboard. |
| **No recent activity feed** | No timeline of recent actions (enrollments, messages, file uploads). |
| **No calendar view** | Sessions are listed flat — no weekly/monthly calendar. |
| **Stats are admin-only** | Teachers don't see their own stats (total students, sessions taught, etc.). |
| **No "today's schedule" section** | No quick overview of what's happening today. |

### 3.3 Session Management

| Issue | Details |
|-------|---------|
| **No recurring sessions** | Can't create weekly/biweekly sessions automatically. |
| **No session notes / description** | Sessions only have a title, no description or agenda. |
| **No session join link / invite** | No way to generate a shareable link for a session. |
| **No session attendance report** | After a session ends, teacher can't see who attended and for how long. |
| **No session status transitions for students** | Students can join a live session but can't see if it's ended without refreshing. |
| **Session logs page is bare** | Shows join/leave times but no summary, no export, no visual timeline. |
| **No session capacity indicator** | Can't see how many participants are in a session vs. the max. |

### 3.4 Class Management

| Issue | Details |
|-------|---------|
| **No class code / invite link** | Students can't self-enroll with a code; teacher must search and add each one. |
| **No class archive** | Can't archive old classes — they clutter the list forever. |
| **No class categories or tags** | No way to organize classes by subject, grade, etc. |
| **No student removal from class** | Once enrolled, a student can't be removed from the UI. |
| **No class-level announcements** | Can't post announcements visible to all enrolled students. |
| **No student count on class cards** | Class cards show max students but not current enrollment count. |
| **No class schedule overview** | No calendar view of all sessions in a class. |

### 3.5 File Management

| Issue | Details |
|-------|---------|
| **No file preview** | Can't preview images, PDFs, or documents inline. |
| **No file search** | Can't search within files. |
| **No file folders / organization** | All files are flat in a session. |
| **No file delete** | Once uploaded, files can't be deleted from the UI. |
| **No file sharing between sessions** | Can't reference a file from another session. |
| **No drag-and-drop upload** | Must click the button and browse. |
| **No upload progress indicator** | No progress bar during file upload. |
| **No file type filtering** | Can't filter by image, document, video, etc. |

### 3.6 Support / Tickets

| Issue | Details |
|-------|---------|
| **No ticket search or filtering** | Can't filter by status, priority, or category. |
| **No ticket assignment** | Admins can't assign tickets to specific staff members. |
| **No ticket templates / FAQ** | No self-service help before creating a ticket. |
| **No file attachments in tickets** | Can't attach screenshots or documents to support tickets. |
| **No ticket status change notifications** | User doesn't know when their ticket is answered. |
| **No ticket categories management** | Categories are hardcoded in the frontend. |

### 3.7 Admin Panel

| Issue | Details |
|-------|---------|
| **No user impersonation** | Admins can't log in as a user to debug issues. |
| **No system health dashboard** | No LiveKit server status, disk usage, or active room count. |
| **No bulk user operations** | Can't bulk-create users (CSV import) or bulk-assign roles. |
| **No role management UI** | Can't create custom roles or manage permissions. |
| **No audit log export** | Activity logs can't be exported as CSV/PDF. |
| **No announcement system** | Can't send system-wide announcements to all users. |
| **No backup/restore UI** | No way to trigger database backups from the admin panel. |
| **Settings page has no validation** | Toggle switches don't confirm destructive actions (maintenance mode). |
| **No user activity timeline** | Can't see a specific user's activity history. |
| **No LiveKit room management** | Can't see active LiveKit rooms, force-disconnect participants, or manage recordings at the SFU level. |

---

## 4. Minor UI/UX Improvements

### 4.1 Visual Design

| Issue | Details |
|-------|---------|
| **Inconsistent modal designs** | Some modals have backdrop-blur, some don't. Some have rounded-2xl, some rounded-xl. |
| **No dark mode** | The entire app is light-only. Classroom is dark, but the rest isn't. |
| **No loading skeletons** | Only spinners are used. Skeleton screens would feel faster. |
| **No empty state illustrations** | Empty states are just text. Custom illustrations would be more welcoming. |
| **No micro-interactions** | Buttons lack press feedback, no page transitions, no smooth state changes. |
| **Toast notifications missing** | Success/error feedback is inconsistent — some pages use `alert()`, most show nothing. |
| **Inconsistent date formatting** | Some pages use `toLocaleDateString`, others use custom formats. Should use a unified Jalali date utility. |
| **No tooltips** | Icon buttons in the classroom have `title` attributes but no styled tooltips. |
| **Sidebar doesn't collapse** | The sidebar is always full-width. A collapsed icon-only mode would give more content space. |
| **No breadcrumb navigation** | Users can't see their navigation path or easily go back to parent pages. |
| **Mobile sidebar animation is abrupt** | The translate animation needs easing curve refinement. |
| **No keyboard shortcuts** | Power users can't use keyboard shortcuts for common actions (mute, video toggle, etc.). |
| **No confirmation for destructive actions** | Some delete actions have `confirm()` dialogs, others don't. Should use a consistent confirmation modal. |
| **RTL inconsistencies** | Some icons and layouts don't flip correctly for RTL (e.g., the recording button icon, certain flex layouts). |

### 4.2 Accessibility

| Issue | Details |
|-------|---------|
| **No ARIA labels** | Most interactive elements lack ARIA attributes. |
| **No focus management in modals** | Opening a modal doesn't trap focus. |
| **No skip-to-content link** | Keyboard users must tab through the entire sidebar. |
| **Color contrast issues** | Some gray text on gray backgrounds fails WCAG AA. |
| **No screen reader announcements** | Dynamic content changes (new chat messages, session status changes) aren't announced. |
| **No reduced motion support** | Animations don't respect `prefers-reduced-motion`. |

### 4.3 Performance

| Issue | Details |
|-------|---------|
| **No image optimization** | No lazy loading for images, no WebP support. |
| **No code splitting per route** | SvelteKit handles this automatically, but heavy components (Whiteboard with Fabric.js) should be dynamically imported. |
| **No API response caching** | Every navigation re-fetches all data. Should use SvelteKit's load function caching or a client-side cache. |
| **No optimistic updates** | Actions like sending a chat message or creating a class wait for server confirmation before updating UI. |
| **No virtual scrolling** | Long lists (sessions, files, logs) render all DOM nodes. |

---

## 5. Missing Features (New Capabilities)

### 5.1 Core Classroom
1. **Polls & Quizzes** — Teacher creates real-time polls, students vote, results shown live
2. **Screen annotation** — Draw on top of shared screen
3. **Virtual backgrounds** — Blur or replace background in video
4. **Closed captions / transcription** — Live transcription of speech
5. **Multi-screen sharing** — Share multiple windows simultaneously
6. **Session playback with whiteboard replay** — Replay recordings with synchronized whiteboard state

### 5.2 Collaboration
1. **Shared notes** — Collaborative text document per session
2. **Assignment submission** — Students submit assignments, teachers grade them
3. **Gradebook** — Track student performance across sessions
4. **Attendance tracking** — Automatic attendance based on session join/leave
5. **Student progress dashboard** — Per-student analytics for teachers

### 5.3 Communication
1. **Direct messaging** — User-to-user private messages outside of sessions
2. **Class discussion forum** — Async discussion board per class
3. **Email notifications** — Configurable email notifications for important events
4. **SMS notifications** — For session reminders (Iranian SMS gateway integration)

### 5.4 Platform
1. **Multi-language support** — i18n infrastructure for Arabic, English, etc.
2. **REST API documentation** — Swagger/OpenAPI docs for the external API
3. **Webhook management UI** — Configure webhooks from the admin panel
4. **Plugin/integration system** — LTI integration with LMS platforms
5. **Mobile app or PWA** — Installable progressive web app with push notifications
6. **Two-factor authentication** — TOTP-based 2FA for enhanced security
7. **Session scheduling with Jalali calendar picker** — Native Jalali date picker instead of HTML date input
8. **Bulk import users from CSV** — For schools importing student rosters
9. **Custom branding** — Allow admins to set logo, colors, and institution name
10. **API rate limiting dashboard** — Visualize and manage rate limits per API key

---

## 6. Backend Gaps

| Gap | Details |
|-----|---------|
| **No refresh token rotation** | Refresh tokens are generated but never invalidated. |
| **No session token blacklist** | No way to revoke tokens before expiry. |
| **No rate limiting on auth endpoints** | Login and register have no brute-force protection. |
| **No file type validation** | Upload accepts any file type — should whitelist safe types. |
| **No file virus scanning** | No ClamAV or similar integration. |
| **No database backup mechanism** | SQLite file can be corrupted; no automated backup. |
| **No request logging middleware** | No structured request/response logging. |
| **No CORS origin validation** | CORS is configured but origins aren't validated against a whitelist. |
| **No health check endpoint details** | Health endpoint exists but doesn't check LiveKit connectivity. |
| **No graceful shutdown** | Server doesn't handle SIGTERM gracefully. |
| **No database connection pooling config** | SQLite WAL is enabled but no busy_timeout or connection limits. |
| **No input sanitization** | User inputs aren't sanitized before database insertion (though parameterized queries prevent SQL injection). |
| **No API versioning strategy** | Currently at v1 but no deprecation/migration strategy. |
| **No webhook delivery retry logic** | Webhooks (if implemented) need retry with exponential backoff. |

---

## 7. Infrastructure & DevOps Gaps

| Gap | Details |
|-----|---------|
| **No CI/CD pipeline** | No GitHub Actions or similar for automated testing/deployment. |
| **No staging environment config** | Only production Docker Compose. |
| **No log aggregation** | Logs go to stdout; no centralized logging. |
| **No monitoring/alerting** | No Prometheus metrics, no uptime monitoring. |
| **No database migration rollback** | Migrations have up but no down migrations. |
| **No SSL certificate renewal monitoring** | Caddy handles this but no alerting if it fails. |
| **No resource limits in Docker** | No memory/CPU limits on containers. |
| **No backup volume for SQLite** | Data volume exists but no backup strategy. |

---

## 8. Recommended Priority Order

### Phase 1: Critical Fixes (Week 1-2)
1. Password reset flow
2. Real-time notification system (WebSocket-based)
3. Pagination on all list views
4. Global error handling with toast notifications
5. User profile page with avatar support
6. Confirmation modals for all destructive actions

### Phase 2: Classroom Essentials (Week 3-4)
1. Hand-raise mechanism
2. Teacher mute/remove controls
3. Session timer and participant count
4. Chat unread indicator
5. Video grid auto-layout
6. Whiteboard responsive canvas + laser pointer
7. File upload progress + drag-and-drop

### Phase 3: Communication & Scheduling (Week 5-6)
1. Email notifications (session start, ticket reply, enrollment)
2. Recurring sessions
3. Class invite codes / self-enrollment
4. Jalali calendar picker for scheduling
5. Session attendance reports
6. Class announcements

### Phase 4: Admin & Platform (Week 7-8)
1. System health dashboard
2. Bulk user import (CSV)
3. API documentation (Swagger)
4. Activity log export
5. User impersonation for admins
6. Two-factor authentication

### Phase 5: Polish & Advanced (Week 9-12)
1. Dark mode
2. PWA support
3. Polls & quizzes
4. Assignment submission
5. Accessibility audit and fixes
6. Performance optimization (caching, lazy loading, optimistic updates)
7. CI/CD pipeline
8. Monitoring and alerting

---

## 9. Summary Statistics

| Category | Count |
|----------|-------|
| Critical Issues | 6 |
| Major UX/UI Gaps | 40+ |
| Minor UI/UX Improvements | 15+ |
| Missing Features | 25+ |
| Backend Gaps | 14 |
| Infrastructure Gaps | 8 |
| **Total Items** | **108+** |

---

*This analysis was generated by reviewing all frontend routes, backend handlers, models, database migrations, configuration files, and the project specification document.*
