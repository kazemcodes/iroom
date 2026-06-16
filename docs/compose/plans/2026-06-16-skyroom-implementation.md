# IRoom Skyroom-Style Implementation Plan

> **For agentic workers:** Use compose:subagent or compose:execute to implement this plan task-by-task.

**Goal:** Transform IRoom into a Skyroom-parity Iranian online classroom platform with popup classroom, Skyroom-style admin panel, Persian utilities, and comprehensive UX fixes.

**Architecture:** SvelteKit frontend with popup-based classroom (BroadcastChannel for window communication), Skyroom-matching dark theme for classroom, light theme for main app. Go backend with existing API endpoints.

**Tech Stack:** Svelte 5 runes, TailwindCSS v4, livekit-client v2, Fabric.js, Go/Echo, SQLite WAL

---

## Phase 1: Persian Utilities & Global Fixes

### Task 1: Persian Number & Date Utilities

**Files:**
- Create: `web/src/lib/utils/persian.ts`
- Create: `web/src/lib/utils/persian.test.ts`

- [ ] **Step 1: Create persian.ts**

```typescript
// web/src/lib/utils/persian.ts
const persianDigits = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];

export function toPersianNum(n: number | string): string {
  return String(n).replace(/[0-9]/g, (d) => persianDigits[parseInt(d)]);
}

export function toPersianDate(date: Date | string): string {
  const d = new Date(date);
  const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'long', day: 'numeric', calendar: 'persian' };
  try {
    return new Intl.DateTimeFormat('fa-IR', options).format(d);
  } catch {
    return d.toLocaleDateString('fa-IR');
  }
}

export function toPersianDateTime(date: Date | string): string {
  const d = new Date(date);
  try {
    return new Intl.DateTimeFormat('fa-IR', {
      year: 'numeric', month: 'long', day: 'numeric',
      hour: '2-digit', minute: '2-digit', calendar: 'persian'
    }).format(d);
  } catch {
    return d.toLocaleString('fa-IR');
  }
}

export function toPersianTime(date: Date | string): string {
  const d = new Date(date);
  return d.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
}

export function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${toPersianNum(h.toString().padStart(2, '0'))}:${toPersianNum(m.toString().padStart(2, '0'))}:${toPersianNum(s.toString().padStart(2, '0'))}`;
}
```

- [ ] **Step 2: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 2: Toast Notification System

**Files:**
- Create: `web/src/lib/stores/toast.ts`
- Create: `web/src/lib/components/Toast.svelte`
- Modify: `web/src/routes/+layout.svelte` (add Toast component)

- [ ] **Step 1: Create toast store**

```typescript
// web/src/lib/stores/toast.ts
import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
  id: number;
  message: string;
  type: ToastType;
  duration?: number;
}

let nextId = 0;
export const toasts = writable<Toast[]>([]);

export function addToast(message: string, type: ToastType = 'info', duration = 3000) {
  const id = nextId++;
  toasts.update(t => [...t, { id, message, type, duration }]);
  if (duration > 0) {
    setTimeout(() => removeToast(id), duration);
  }
}

export function removeToast(id: number) {
  toasts.update(t => t.filter(toast => toast.id !== id));
}
```

- [ ] **Step 2: Create Toast.svelte component**

```svelte
<script lang="ts">
  import { toasts, removeToast, type Toast } from '$lib/stores/toast';

  const typeStyles: Record<string, string> = {
    success: 'border-r-green-500 bg-green-50 text-green-700',
    error: 'border-r-red-500 bg-red-50 text-red-700',
    info: 'border-r-blue-500 bg-blue-50 text-blue-700',
    warning: 'border-r-amber-500 bg-amber-50 text-amber-700',
  };
</script>

<div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[9999] flex flex-col gap-2 pointer-events-none">
  {#each $toasts as toast (toast.id)}
    <div
      class="pointer-events-auto px-5 py-3 rounded-lg border-r-4 shadow-lg font-medium text-sm {typeStyles[toast.type]} animate-slide-up"
      role="alert"
    >
      {toast.message}
    </div>
  {/each}
</div>

<style>
  .animate-slide-up {
    animation: slideUp 0.2s ease-out;
  }
  @keyframes slideUp {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
```

- [ ] **Step 3: Add Toast to layout**

In `web/src/routes/+layout.svelte`, add at the end of the template:
```svelte
<script lang="ts">
  import Toast from '$lib/components/Toast.svelte';
</script>

<!-- existing content -->
{@render children()}

<Toast />
```

- [ ] **Step 4: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Phase 2: Classroom Popup Architecture

### Task 3: Classroom Window Manager

**Files:**
- Create: `web/src/lib/classroom/ClassroomWindow.ts`
- Create: `web/src/lib/classroom/ClassroomBridge.ts`

- [ ] **Step 1: Create ClassroomWindow.ts**

```typescript
// web/src/lib/classroom/ClassroomWindow.ts

const POPUP_FEATURES = [
  'width=1100', 'height=700', 'resizable=yes',
  'scrollbars=no', 'toolbar=no', 'menubar=no',
  'location=no', 'status=no'
].join(',');

class ClassroomWindowManager {
  private popup: Window | null = null;

  open(sessionId: string, sessionTitle: string): boolean {
    if (this.popup && !this.popup.closed) {
      this.popup.focus();
      return false;
    }

    const width = 1100;
    const height = 700;
    const left = Math.round((screen.width - width) / 2) + 50;
    const top = Math.round((screen.height - height) / 2);

    this.popup = window.open(
      `/classroom/popup/${sessionId}`,
      `iroom_classroom_${sessionId}`,
      `${POPUP_FEATURES},left=${left},top=${top}`
    );

    if (!this.popup) {
      window.location.href = `/classroom/${sessionId}`;
      return false;
    }

    const checkClosed = setInterval(() => {
      if (this.popup?.closed) {
        clearInterval(checkClosed);
        this.popup = null;
      }
    }, 500);

    return true;
  }

  close() { this.popup?.close(); this.popup = null; }
  isOpen(): boolean { return this.popup !== null && !this.popup.closed; }
}

export const classroomWindow = new ClassroomWindowManager();
```

- [ ] **Step 2: Create ClassroomBridge.ts**

```typescript
// web/src/lib/classroom/ClassroomBridge.ts
export type ClassroomMessage =
  | { type: 'classroom-closed' }
  | { type: 'chat-notification'; count: number }
  | { type: 'session-ended' }
  | { type: 'participant-joined'; name: string }
  | { type: 'participant-left'; name: string }
  | { type: 'toggle-mute' }
  | { type: 'toggle-video' };

let channel: BroadcastChannel | null = null;

export function initBridge(onMessage?: (msg: ClassroomMessage) => void) {
  channel = new BroadcastChannel('iroom_classroom');
  if (onMessage) channel.onmessage = (e) => onMessage(e.data);
  return channel;
}

export function sendBridge(msg: ClassroomMessage) {
  channel?.postMessage(msg);
}

export function closeBridge() {
  channel?.close();
  channel = null;
}
```

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 4: Classroom Popup Route & Main Layout

**Files:**
- Create: `web/src/routes/(app)/classroom/popup/[id]/+page.svelte`
- Modify: `web/src/routes/(app)/classroom/[id]/+page.svelte` (becomes launcher)

- [ ] **Step 1: Create popup page**

The popup page is a compact 3-column layout without the main app sidebar. It should contain the same LiveKit connection logic as the current classroom page but with the Skyroom-style layout (top bar, 3 columns, control bar).

Key elements:
- Dark theme background (#1a1a2e)
- Top bar with session title, connection status, elapsed timer
- 3-column: Video/Whiteboard (flex) | Chat (280px) | Participants (220px)
- Bottom control bar with mic/video/share/whiteboard/record/leave buttons
- No sidebar, no main app header

- [ ] **Step 2: Convert current classroom to launcher**

The `/classroom/[id]` page becomes a simple launcher that shows session info and an "Open Classroom" button that calls `classroomWindow.open()`. If popup is blocked, show inline fallback.

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 5: Update Join Buttons

**Files:**
- Modify: `web/src/routes/(app)/sessions/+page.svelte`
- Modify: `web/src/routes/(app)/classes/[id]/+page.svelte`
- Modify: `web/src/routes/(app)/admin/rooms/+page.svelte`

- [ ] **Step 1: Update session join button**

Change the "پیوستن" button to use `classroomWindow.open(session.id, session.title)` instead of direct navigation.

- [ ] **Step 2: Update class session join button**

Same change in class detail page.

- [ ] **Step 3: Update admin rooms join button**

Same change in admin rooms page.

- [ ] **Step 4: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Phase 3: Admin Panel Redesign

### Task 6: Admin Dashboard with Live Stats

**Files:**
- Modify: `web/src/routes/(app)/admin/+page.svelte`

- [ ] **Step 1: Redesign dashboard page**

Replace current tab-based admin with Skyroom-style dashboard:
- Stats cards row (users, active rooms, today's sessions, recordings)
- Live rooms section showing active sessions with participant count
- System health section (DB size, uptime)
- Activity feed from activity_logs table
- Use Persian numbers via `toPersianNum()`

- [ ] **Step 2: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 7: Admin Sidebar Navigation

**Files:**
- Modify: `web/src/routes/(app)/admin/+layout.svelte` (create if needed)
- Modify: `web/src/routes/+layout.svelte` (sidebar nav items)

- [ ] **Step 1: Add admin-specific sidebar items**

Add navigation items: Dashboard, Users, Rooms, Sessions, Recordings, Tickets, Logs, Settings — each with Persian labels and icons.

- [ ] **Step 2: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 8: Admin Rooms Management

**Files:**
- Modify: `web/src/routes/(app)/admin/rooms/+page.svelte`

- [ ] **Step 1: Redesign rooms page**

Skyroom-style room cards with:
- Room name, teacher, student count, session count
- Status badge (active/inactive)
- Quick actions: join, view sessions, edit, delete
- Search and filter

- [ ] **Step 2: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Phase 4: Critical UX Fixes

### Task 9: Pagination on List Views

**Files:**
- Modify: `web/src/routes/(app)/classes/+page.svelte`
- Modify: `web/src/routes/(app)/sessions/+page.svelte`
- Modify: `web/src/routes/(app)/files/+page.svelte`

- [ ] **Step 1: Add pagination to classes list**

Add page navigation, per-page selector, and page info display.

- [ ] **Step 2: Add pagination to sessions list**

Same pattern.

- [ ] **Step 3: Add pagination to files list**

Same pattern.

- [ ] **Step 4: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 10: Confirmation Modals for Destructive Actions

**Files:**
- Create: `web/src/lib/components/ConfirmModal.svelte`
- Modify: All pages with delete buttons

- [ ] **Step 1: Create ConfirmModal.svelte**

Reusable confirmation modal with title, message, confirm/cancel buttons.

- [ ] **Step 2: Replace all confirm() calls with ConfirmModal**

Replace native `confirm()` dialogs with the styled modal across all pages.

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

### Task 11: Whiteboard Responsive Canvas

**Files:**
- Modify: `web/src/lib/components/Whiteboard.svelte`

- [ ] **Step 1: Make canvas responsive**

Remove fixed 800x600 dimensions. Use container dimensions. Add resize observer.

- [ ] **Step 2: Add laser pointer tool**

Add laser pointer tool that shows a red dot for other participants (via DataChannels).

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Phase 5: Backend Enhancements

### Task 12: Session Logs with Duration Tracking

**Files:**
- Modify: `web/src/routes/(app)/sessions/[id]/logs/+page.svelte`
- Modify: `internal/handlers/ticket.go` (session log handlers)

- [ ] **Step 1: Improve session logs display**

Show participant join/leave times, total duration, and visual timeline.

- [ ] **Step 2: Add CSV export for logs**

Add export button that downloads session logs as CSV.

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build && cd .. && export PATH=$PATH:$HOME/go-sdk/go/bin && GOTOOLCHAIN=local go build -o server ./cmd/server`
Expected: SUCCESS

---

### Task 13: Ticket System Improvements

**Files:**
- Modify: `web/src/routes/(app)/support/+page.svelte`
- Modify: `web/src/routes/(app)/support/[id]/+page.svelte`

- [ ] **Step 1: Add ticket filtering**

Add status, priority, and category filters to ticket list.

- [ ] **Step 2: Add ticket search**

Add search functionality to ticket list.

- [ ] **Step 3: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Execution Order

1. Task 1 (Persian utilities) — foundation
2. Task 2 (Toast system) — foundation
3. Task 3 (Classroom window) — foundation
4. Task 4 (Popup route) — depends on 3
5. Task 5 (Join buttons) — depends on 3
6. Task 6 (Admin dashboard) — independent
7. Task 7 (Admin sidebar) — independent
8. Task 8 (Admin rooms) — depends on 7
9. Task 9 (Pagination) — independent
10. Task 10 (Confirmation modals) — independent
11. Task 11 (Whiteboard) — independent
12. Task 12 (Session logs) — independent
13. Task 13 (Tickets) — independent

Tasks 6-13 can be parallelized across subagents.
