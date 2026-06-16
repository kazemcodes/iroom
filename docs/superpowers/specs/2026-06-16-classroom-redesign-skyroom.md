# IRoom — Skyroom-Style Classroom Redesign

**Date:** 2026-06-16  
**Type:** UI/UX Design Specification  
**Scope:** Classroom page — complete layout redesign from Google Meet style to Skyroom style

---

## 1. Design Philosophy

### Current Problem
The current classroom (`/classroom/[id]`) is a **full-page takeover** in dark mode — like Google Meet. When a user joins a class, they lose access to the main app (dashboard, sidebar, navigation). They must leave the classroom to do anything else.

### Target: Skyroom Paradigm
Skyroom (skyroom.ir) is the dominant online classroom platform in Iran. Its key UX principle:

> **The classroom is a separate, always-on-top window. The main browser tab stays on the class/dashboard page. Users can browse files, check schedules, or read announcements while actively participating in a live class.**

This means:
1. Clicking "Join Class" opens a **new browser popup window** (via `window.open`)
2. The popup contains the classroom UI — compact, with video, chat, and participants always visible
3. The original tab remains fully functional — user can navigate, browse, etc.
4. Closing the popup = leaving the class
5. Communication between the two windows via `BroadcastChannel` API (same-origin) or `localStorage` events

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  MAIN BROWSER TAB (stays on dashboard/class page)           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Sidebar  │  Content area                             │  │
│  │           │  (dashboard, class details, files, etc.)  │  │
│  │           │                                           │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  [Join Class] ──window.open()──► ┌──────────────────────┐  │
│                                  │  CLASSROOM POPUP     │  │
│                                  │  (separate window)   │  │
│                                  │                      │  │
│                                  │  ┌─────┬──────┬────┐ │  │
│                                  │  │Video│ Chat │Par-│ │  │
│                                  │  │Panel│      │ti- │ │  │
│                                  │  │     │      │ci- │ │  │
│                                  │  │     │      │pant│ │  │
│                                  │  │     │      │s   │ │  │
│                                  │  ├─────┴──────┴────┤ │  │
│                                  │  │  Control Bar     │ │  │
│                                  │  └──────────────────┘ │  │
│                                  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Window Communication
```
Main Tab ◄──BroadcastChannel──► Classroom Popup
   │                                │
   ├── "classroom-opened"           ├── "classroom-closed"
   ├── "session-ended"              ├── "participant-joined"
   └── "navigate-to-class"          └── "chat-notification"
```

---

## 3. Detailed Layout Design

### 3.1 Popup Window Specifications

| Property | Value |
|----------|-------|
| **Width** | 1100px (resizable by user) |
| **Height** | 700px (resizable by user) |
| **Min-width** | 800px |
| **Min-height** | 500px |
| **Title** | `آی‌روم — {session title}` |
| **Features** | Always-on-top (optional, user toggleable), no browser toolbar |
| **Position** | Centered on screen, offset from main window |

### 3.2 Popup Internal Layout (3-Column)

```
┌──────────────────────────────────────────────────────────────────┐
│ ███ TOP BAR (36px) ███  session title | status | timer | X │
├──────────────────────────┬──────────────────┬────────────────────┤
│                          │                  │                    │
│   VIDEO / WHITEBOARD     │   CHAT PANEL     │  PARTICIPANTS      │
│   (flexible width ~55%)  │   (280px fixed)  │  (220px fixed)     │
│                          │                  │                    │
│  ┌────────────────────┐  │  ┌────────────┐  │  ┌──────────────┐  │
│  │                    │  │  │ Messages   │  │  │ 👤 Teacher   │  │
│  │  Active Speaker /  │  │  │            │  │  │ 🎤 🔊       │  │
│  │  Screen Share /    │  │  │ سلام!      │  │  │              │  │
│  │  Whiteboard        │  │  │ چطورید؟    │  │  │ 👤 Student 1 │  │
│  │                    │  │  │            │  │  │ 🔊          │  │
│  │                    │  │  │ ──────     │  │  │              │  │
│  │                    │  │  │ خوبم ممنون │  │  │ 👤 Student 2 │  │
│  │                    │  │  │            │  │  │ 🔇          │  │
│  │                    │  │  │            │  │  │              │  │
│  │                    │  │  │            │  │  │ 👤 Student 3 │  │
│  │                    │  │  │            │  │  │ 🔊          │  │
│  │                    │  │  │            │  │  │              │  │
│  │                    │  │  ├────────────┤  │  │              │  │
│  │                    │  │  │ [input  ]  │  │  │              │  │
│  │                    │  │  │ [send   ]  │  │  │              │  │
│  └────────────────────┘  │  └────────────┘  │  └──────────────┘  │
│                          │                  │                    │
├──────────────────────────┴──────────────────┴────────────────────┤
│ ███ CONTROL BAR (56px) ███  🎤  📹  🖥️  📋  ⚙️  🔴  ✕         │
└──────────────────────────────────────────────────────────────────┘
```

### 3.3 Top Bar (36px)

```
┌──────────────────────────────────────────────────────────────────┐
│ 🔵 جلسه ریاضی ─ متصل ─ ۰۰:۱۵:۳۲          [🔝] [─] [✕]        │
│      │           │         │                │    │    │         │
│   session    status    elapsed timer    always  min  close     │
│   title     (green dot)              on-top                  │
└──────────────────────────────────────────────────────────────────┘
```

Elements:
- **Session title** (right side, truncated with ellipsis)
- **Connection status** — green dot + "متصل" / red dot + "قطع" / yellow + "در حال اتصال..."
- **Elapsed timer** — live counter since joining
- **Always-on-top toggle** — pin icon
- **Minimize** — minimize window
- **Close** — leave class (with confirmation)

### 3.4 Video / Whiteboard Area (flexible, ~55% width)

This is the main content area. Three modes:

#### Mode A: Active Speaker View (default)
```
┌─────────────────────────────────────┐
│  ┌───────────────────────────────┐  │
│  │                               │  │
│  │     Active Speaker Video      │  │
│  │     (teacher or whoever       │  │
│  │      is speaking)             │  │
│  │                               │  │
│  │  ┌──┐ ┌──┐ ┌──┐              │  │
│  │  │S1│ │S2│ │S3│  ← small     │  │
│  │  └──┘ └──┘ └──┘    thumbnail  │  │
│  │                   strip       │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

- Main video fills the area (object-contain)
- Small thumbnail strip at bottom showing other participants (32px height each, horizontal scroll)
- Clicking a thumbnail swaps it to main view
- Local video thumbnail always shown at the end

#### Mode B: Screen Share View
```
┌─────────────────────────────────────┐
│  ┌───────────────────────────────┐  │
│  │                               │  │
│  │     Screen Share (full)       │  │
│  │                               │  │
│  │                               │  │
│  │     ┌──────────┐              │  │
│  │     │ Teacher  │  ← small     │  │
│  │     │  video   │    overlay   │  │
│  │     └──────────┘              │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

- Screen share takes priority, fills the area
- Teacher's camera as a small draggable overlay (like Skyroom)
- Double-click to toggle fullscreen within the area

#### Mode C: Whiteboard View
```
┌─────────────────────────────────────┐
│  ┌───────────────────────────────┐  │
│  │  ✏️ ⬜ ⭕ 📏 📝 ↩️ 🗑️         │  │  ← whiteboard toolbar
│  ├───────────────────────────────┤  │
│  │                               │  │
│  │     Whiteboard Canvas         │  │
│  │     (responsive, fills area)  │  │
│  │                               │  │
│  │                               │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

- Whiteboard toolbar at top of this section (not a separate sidebar)
- Canvas is responsive, fills available space
- Video thumbnails shown as a small overlay strip at bottom

**Mode switching:** Tabs or buttons at the top of this section:
```
[🎥 ویدیو]  [📋 تخته‌سفید]  [🖥️ اشتراک صفحه]
```

### 3.5 Chat Panel (280px fixed width)

```
┌──────────────────────────┐
│ 💬 گفتگو            [۳] │  ← title + unread badge
├──────────────────────────┤
│                          │
│  ┌────────────────────┐  │
│  │ استاد: سلام بچه‌ها │  │  ← others' messages (right-aligned in RTL)
│  └────────────────────┘  │
│       ┌──────────────┐   │
│       │ شما: سلام    │   │  ← own messages (left-aligned, different color)
│       └──────────────┘   │
│  ┌────────────────────┐  │
│  │ علی: سلام استاد   │  │
│  └────────────────────┘  │
│                          │
├──────────────────────────┤
│ [📎] [نوع پیام...    ] [➤]  ← input with file attach
└──────────────────────────┘
```

Features:
- **Always visible** — not a toggleable sidebar
- **Unread badge** on chat title when scrolled up and new message arrives
- **Message types:** text, file link, system notification
- **Own messages** styled differently (blue background, left-aligned in RTL)
- **Timestamps** on hover
- **File attachment** button (paperclip icon)
- **Emoji picker** (optional)
- **Auto-scroll** to bottom on new messages (unless user has scrolled up)

### 3.6 Participants Panel (280px fixed width)

```
┌──────────────────────────┐
│ 👥 شرکت‌کنندگان (۵)  [🔽] │  ← collapsible
├──────────────────────────┤
│  👤 استاد احمدی          │  ← teacher (crown icon)
│     🎤 🔊                │
│  ─────────────────────── │
│  👤 علی رضایی            │  ← student
│     🎤 🔊                │
│  👤 مریم کریمی           │
│     🔇 🔊                │  ← muted
│  👤 شما                  │  ← local user
│     🎤 🔊                │
│  👤 رضا محمدی            │
│     🔇                  │  ← audio only
│  ─────────────────────── │
│  📋 دسترسی‌ها             │  ← teacher-only section
│  [صدا] [ویدیو] [حذف]    │  ← per-student controls
└──────────────────────────┘
```

Features:
- **Always visible** — not a toggleable sidebar
- **Teacher badge** (crown icon 👑) next to teacher name
- **Speaking indicator** — animated green border when speaking
- **Audio/Video status** — mic and camera icons (green=active, red=muted, gray=off)
- **Collapsible** — can collapse to save space
- **Teacher controls** (only visible to teacher/admin):
  - Mute individual student
  - Disable student's video
  - Remove from session
  - Grant/revoke whiteboard access
- **Hand raise indicator** ✋ — shows next to student who raised hand
- **Search/filter** — when many participants

### 3.7 Control Bar (56px, bottom)

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  🎤    📹    🖥️    📋    ⚙️          🔴    ✕                  │
│   │     │     │     │     │            │     │                  │
│  mute  video  share white-  settings  record  leave             │
│              screen board            (red    (red               │
│                                      when    with               │
│                                      rec)    confirm)           │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

Buttons (all circular, 40px diameter):

| Button | Icon | Behavior |
|--------|------|----------|
| **Mute/Unmute** | 🎤 / 🎤❌ | Toggle microphone. Red background when muted |
| **Video On/Off** | 📹 / 📹❌ | Toggle camera. Red background when off |
| **Screen Share** | 🖥️ | Start/stop screen sharing. Blue when active |
| **Whiteboard** | 📋 | Toggle whiteboard mode. Purple when active |
| **Settings** | ⚙️ | Open audio/video device settings popup |
| **Record** | 🔴 | Start/stop recording. Pulsing red when recording |
| **Leave** | ✕ | Leave class. Confirmation dialog |

---

## 4. Implementation Plan

### 4.1 New File Structure

```
web/src/
├── lib/
│   ├── classroom/
│   │   ├── ClassroomWindow.ts       # Window management (open, close, communicate)
│   │   ├── ClassroomBridge.ts       # BroadcastChannel communication
│   │   └── classroom-store.ts       # Shared state for classroom popup
│   └── components/
│       ├── classroom/
│       │   ├── ClassroomPopup.svelte       # Main popup container
│       │   ├── ClassroomTopBar.svelte      # Top bar with title, timer, controls
│       │   ├── VideoPanel.svelte           # Video / whiteboard area
│       │   ├── ChatPanel.svelte            # Chat panel
│       │   ├── ParticipantsPanel.svelte    # Participants panel
│       │   ├── ControlBar.svelte           # Bottom control bar
│       │   ├── WhiteboardCanvas.svelte     # Whiteboard (extracted from current)
│       │   ├── VideoGrid.svelte            # Video grid with active speaker
│       │   ├── ScreenShareView.svelte      # Screen share view
│       │   └── SettingsPopup.svelte        # Audio/video settings
│       └── Whiteboard.svelte               # Keep for backward compat
├── routes/
│   └── (app)/
│       └── classroom/
│           └── [id]/
│               └── +page.svelte            # Becomes thin wrapper / launcher
```

### 4.2 ClassroomWindow.ts — Window Manager

```typescript
// web/src/lib/classroom/ClassroomWindow.ts

const POPUP_FEATURES = [
  'width=1100',
  'height=700',
  'min-width=800',
  'min-height=500',
  'resizable=yes',
  'scrollbars=no',
  'toolbar=no',
  'menubar=no',
  'location=no',
  'status=no',
  'directories=no',
  'titlebar=no',
  'centerscreen=yes',
].join(',');

class ClassroomWindowManager {
  private popup: Window | null = null;
  private channel: BroadcastChannel;
  
  constructor() {
    this.channel = new BroadcastChannel('iroom_classroom');
    this.channel.onmessage = (e) => this.handleMessage(e.data);
  }

  open(sessionId: string, sessionTitle: string): boolean {
    // Check if already open
    if (this.popup && !this.popup.closed) {
      this.popup.focus();
      return false;
    }

    // Calculate position: center of screen, offset from main window
    const width = 1100;
    const height = 700;
    const left = Math.round((screen.width - width) / 2) + 50;
    const top = Math.round((screen.height - height) / 2);

    const features = `${POPUP_FEATURES},left=${left},top=${top}`;
    
    this.popup = window.open(
      `/classroom/popup/${sessionId}`,
      `iroom_classroom_${sessionId}`,
      features
    );

    if (!this.popup) {
      // Popup blocked — fallback to full-page
      window.location.href = `/classroom/${sessionId}`;
      return false;
    }

    // Notify main app
    this.channel.postMessage({ type: 'classroom-opened', sessionId });
    
    // Monitor popup close
    const checkClosed = setInterval(() => {
      if (this.popup?.closed) {
        clearInterval(checkClosed);
        this.channel.postMessage({ type: 'classroom-closed', sessionId });
        this.popup = null;
      }
    }, 500);

    return true;
  }

  close() {
    this.popup?.close();
    this.popup = null;
  }

  isOpen(): boolean {
    return this.popup !== null && !this.popup.closed;
  }

  private handleMessage(data: any) {
    switch (data.type) {
      case 'classroom-closed':
        this.popup = null;
        break;
      case 'chat-notification':
        // Show notification badge on main tab
        break;
    }
  }
}

export const classroomWindow = new ClassroomWindowManager();
```

### 4.3 Route Changes

**Current:** `/classroom/[id]` — full-page classroom  
**New:**

| Route | Purpose |
|-------|---------|
| `/classroom/[id]` | Thin launcher page — shows session info, "Open Classroom" button that calls `classroomWindow.open()`. If popup blocked, shows inline classroom as fallback |
| `/classroom/popup/[id]` | The actual classroom UI designed for popup window (no sidebar, no header, compact layout) |

### 4.4 Main Tab Changes

The main tab (`/classroom/[id]`) becomes a **launcher + status page**:

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │                                                     │   │
│   │   جلسه ریاضی پایه دهم                              │   │
│   │   کلاس: ریاضی ─ مدرس: استاد احمدی                 │   │
│   │   زمان: ۱۴:۰۰ ─ ۶۰ دقیقه                          │   │
│   │                                                     │   │
│   │   ┌─────────────────────────────────────────────┐   │   │
│   │   │  🔴 کلاس در حال برگزاری است                 │   │   │
│   │   │                                             │   │   │
│   │   │  [🚀 ورود به کلاس]  ← opens popup          │   │   │
│   │   │                                             │   │   │
│   │   │  یا اگر پاپ‌آپ باز شد:                      │   │   │
│   │   │  [📋 کلاس را در پنجره جدید باز کنید]       │   │   │
│   │   └─────────────────────────────────────────────┘   │   │
│   │                                                     │   │
│   │   ┌─────────────────────────────────────────────┐   │   │
│   │   │  📋 اطلاعات جلسه                            │   │   │
│   │   │                                             │   │   │
│   │   │  توضیحات: ...                                │   │   │
│   │   │  لینک ضبط: ...                               │   │   │
│   │   │  فایل‌های جلسه: [مشاهده]                     │   │   │
│   │   └─────────────────────────────────────────────┘   │   │
│   │                                                     │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

When classroom popup is open, the main tab shows:
- Live status indicator ("کلاس باز است در پنجره جداگانه")
- Quick actions: mute/unmute from main tab (via BroadcastChannel)
- Session info, description, files
- Link to recordings after session ends

### 4.5 BroadcastChannel Communication Protocol

```typescript
// Message types between main tab and popup

type ClassroomMessage =
  // Popup → Main
  | { type: 'classroom-closed' }
  | { type: 'chat-notification'; count: number }
  | { type: 'session-ended' }
  | { type: 'participant-joined'; name: string }
  | { type: 'participant-left'; name: string }
  | { type: 'hand-raised'; userId: number }
  | { type: 'hand-lowered'; userId: number }
  
  // Main → Popup
  | { class: 'toggle-mute' }
  | { type: 'toggle-video' }
  | { type: 'leave-class' }
  | { type: 'navigate-to-class'; classId: number }
  
  // Bidirectional
  | { type: 'ping' }      // Check if other window is alive
  | { type: 'pong' };
```

---

## 5. Responsive Behavior

### Popup Window Resolutions

| Width | Layout |
|-------|--------|
| **≥ 1100px** | Full 3-column: Video + Chat + Participants |
| **800–1099px** | 2-column: Video + Chat (Participants collapsible overlay) |
| **< 800px** | Single column: Video only, Chat and Participants as overlays with toggle buttons |

### Video Grid Adaptation

| Participants | Layout |
|-------------|--------|
| **1 (local only)** | Full-screen local video with "waiting for others" message |
| **2** | 50/50 split |
| **3–4** | 2×2 grid |
| **5–6** | 3×2 grid with active speaker highlight |
| **7+** | Active speaker large + horizontal thumbnail strip |

---

## 6. State Management

### Shared State (classroom-store.ts)

```typescript
import { writable } from 'svelte/store';

export const classroomState = writable({
  sessionId: null as string | null,
  isOpen: false,
  isConnected: false,
  connectionState: 'disconnected' as 'disconnecting' | 'disconnected' | 'connecting' | 'connected' | 'reconnecting',
  
  // Media
  audioEnabled: true,
  videoEnabled: false,
  screenSharing: false,
  isRecording: false,
  
  // UI State
  activeView: 'video' as 'video' | 'whiteboard' | 'screenshare',
  participantsPanelCollapsed: false,
  unreadChatCount: 0,
  
  // Participants
  participants: [] as Participant[],
  
  // Chat
  chatMessages: [] as ChatMessage[],
  
  // Timer
  elapsedSeconds: 0,
});

// Derived stores
export const teacher = derived(classroomState, $s => 
  $s.participants.find(p => p.role === 'teacher')
);

export const students = derived(classroomState, $s => 
  $s.participants.filter(p => p.role === 'student')
);

export const activeSpeaker = derived(classroomState, $s => 
  $s.participants.find(p => p.isSpeaking)
);

export const raisedHands = derived(classroomState, $s => 
  $s.participants.filter(p => p.handRaised)
);
```

---

## 7. Visual Design Specifications

### Color Scheme (Dark theme for classroom popup)

| Token | Value | Usage |
|-------|-------|-------|
| `--bg-primary` | `#1a1a2e` | Main background |
| `--bg-secondary` | `#16213e` | Panel backgrounds |
| `--bg-tertiary` | `#0f3460` | Input fields, cards |
| `--accent` | `#e94560` | Recording, leave, danger |
| `--accent-green` | `#00d26a` | Connected, speaking, active |
| `--accent-blue` | `#4361ee` | Links, active buttons |
| `--accent-purple` | `#7209b7` | Whiteboard, teacher badge |
| `--text-primary` | `#eaeaea` | Primary text |
| `--text-secondary` | `#a0a0a0` | Secondary text, timestamps |
| `--border` | `#2a2a4a` | Borders, dividers |

### Typography

| Element | Size | Weight |
|---------|------|--------|
| Top bar title | 14px | 600 |
| Chat message | 13px | 400 |
| Chat sender name | 12px | 600 |
| Participant name | 13px | 500 |
| Control bar tooltip | 11px | 500 |
| Timer | 13px | 600 (monospace) |

### Spacing

| Element | Value |
|---------|-------|
| Panel padding | 12px |
| Chat message gap | 8px |
| Participant row padding | 8px 12px |
| Control bar button gap | 8px |
| Border radius (panels) | 8px |
| Border radius (buttons) | 50% (circular) |

---

## 8. Migration Plan

### Step 1: Create popup route and components
1. Create `/classroom/popup/[id]` route
2. Build `ClassroomPopup.svelte` with 3-column layout
3. Build individual components (TopBar, VideoPanel, ChatPanel, ParticipantsPanel, ControlBar)
4. Implement `ClassroomWindow.ts` for window management

### Step 2: Refactor existing code
1. Extract LiveKit logic from current `+page.svelte` into reusable functions
2. Extract chat logic into `ChatPanel.svelte`
3. Extract participant list into `ParticipantsPanel.svelte`
4. Make `Whiteboard.svelte` responsive (remove fixed 800×600)

### Step 3: Update main classroom page
1. Convert `/classroom/[id]` to launcher page
2. Add "Open Classroom" button
3. Add popup-blocked fallback (inline classroom)
4. Add BroadcastChannel listener for status updates

### Step 4: Update all "Join" buttons
1. Update `/classes/[id]` — "ورود" button opens popup
2. Update `/sessions` — "پیوستن" button opens popup
3. Update admin rooms page — "ورود" button opens popup

### Step 5: Testing
1. Test popup opens correctly
2. Test popup blocked fallback
3. Test BroadcastChannel communication
4. Test responsive layouts
5. Test window close → cleanup
6. Test multiple popups (should focus existing)

---

## 9. Backward Compatibility

- The old `/classroom/[id]` route still works as a full-page classroom for:
  - Mobile browsers (which block popups)
  - Users with popup blockers
  - Direct links / bookmarks
- Detect mobile → skip popup, use full-page
- Detect popup blocked → show inline classroom with message

---

## 10. Summary of Changes

| Component | Current | New |
|-----------|---------|-----|
| **Entry point** | `/classroom/[id]` full-page | `/classroom/[id]` launcher + `/classroom/popup/[id]` popup |
| **Layout** | Full-screen dark, side panels toggle | 3-column compact, all panels visible |
| **Video** | 2-column grid, fixed | Active speaker + adaptive grid |
| **Chat** | Toggleable sidebar (320px) | Always-visible panel (280px) |
| **Participants** | Toggleable sidebar (288px) | Always-visible panel (288px) |
| **Whiteboard** | Full-screen replacement | Tab in video area |
| **Controls** | Bottom bar, many buttons | Compact bottom bar, essential buttons |
| **Window** | Full browser tab | Separate popup window |
| **Main tab** | Unavailable during class | Fully functional during class |
| **Communication** | N/A | BroadcastChannel between windows |

---

*This design is modeled after Skyroom.ir's classroom UX, adapted for IRoom's tech stack (SvelteKit + LiveKit).*

---

## 15. Persian-Only Mandate — Global Rules

**This applies to the ENTIRE application, not just the classroom.**

### 15.1 Numbers — Persian Numerals Only

All numbers displayed in the UI **must** use Persian (Eastern Arabic) numerals:

| Digit | Persian |
|-------|---------|
| 0 | ۰ |
| 1 | ۱ |
| 2 | ۲ |
| 3 | ۳ |
| 4 | ۴ |
| 5 | ۵ |
| 6 | ۶ |
| 7 | ۷ |
| 8 | ۸ |
| 9 | ۹ |

**Implementation:**
```typescript
// web/src/lib/utils/persian.ts

const persianDigits = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];

export function toPersianNum(n: number | string): string {
  return String(n).replace(/[0-9]/g, (d) => persianDigits[parseInt(d)]);
}

export function toPersianDate(date: Date | string): string {
  // Use date-fns-jalali for Jalali formatting
  const { format: jalaliFormat } = await import('date-fns-jalali');
  const { faIR } = await import('date-fns-jalali/locale');
  return jalaliFormat(new Date(date), 'yyyy/MM/dd', { locale: faIR });
}

export function toPersianDateTime(date: Date | string): string {
  const { format: jalaliFormat } = await import('date-fns-jalali');
  const { faIR } = await import('date-fns-jalali/locale');
  return jalaliFormat(new Date(date), 'yyyy/MM/dd ساعت HH:mm', { locale: faIR });
}
```

**Usage in Svelte components:**
```svelte
<script>
  import { toPersianNum, toPersianDate } from '$lib/utils/persian';
</script>

<span>{toPersianNum(123)}</span>        <!-- ۱۲۳ -->
<span>{toPersianNum(stats.users)}</span> <!-- ۱۲۳ -->
<span>{toPersianDate(session.scheduled_at)}</span> <!-- ۱۴۰۳/۰۳/۲۵ -->
```

**Exceptions (keep Western digits):**
- Email addresses (inside `dir="ltr"` containers)
- URLs
- Phone number input fields (for dialing compatibility)
- API request/response payloads

### 15.2 Calendar — Jalali (Shamsi) Only

**All dates must use the Jalali calendar. Never show Gregorian dates.**

**Implementation:**
```typescript
// Install: npm install date-fns-jalali

import { format as jalaliFormat, parse as jalaliParse } from 'date-fns-jalali';
import { faIR } from 'date-fns-jalali/locale';

// Formatting
jalaliFormat(new Date(), 'yyyy/MM/dd', { locale: faIR });        // ۱۴۰۳/۰۳/۲۵
jalaliFormat(new Date(), 'yyyy/MM/dd HH:mm', { locale: faIR });   // ۱۴۰۳/۰۳/۲۵ ۱۴:۳۰
jalaliFormat(new Date(), 'EEEE', { locale: faIR });                // سه‌شنبه
jalaliFormat(new Date(), 'MMMM', { locale: faIR });                // خرداد
```

**Date Picker:** Replace ALL `<input type="date">` with a Jalali date picker component:
- Use or build a Jalali date picker (e.g., based on `date-fns-jalali`)
- Month names: فروردین، اردیبهشت، خرداد، تیر، مرداد، شهریور، مهر، آبان، آذر، دی، بهمن، اسفند
- Week starts on Saturday (شنبه)
- Day names: شنبه، یکشنبه، دوشنبه، سه‌شنبه، چهارشنبه، پنجشنبه، جمعه

### 15.3 Text — Farsi Only

- All UI strings in Farsi
- Error messages in Farsi
- Success messages in Farsi
- Placeholder text in Farsi
- Button labels in Farsi
- Form labels in Farsi
- Use formal/polite forms (شما, لطفاً, ممنون)

### 15.4 RTL Layout

```html
<html dir="rtl" lang="fa">
```

- All layouts must be RTL
- Sidebar on the right
- Text right-aligned by default
- Icons flipped where directional (arrows, chevrons)
- `dir="ltr"` only for: email inputs, URL displays, phone numbers, code blocks

---

## 16. Exact Skyroom UI Copy — Reference Specifications

### 16.1 Skyroom Color Palette (Exact)

| Element | Skyroom Color | Hex Code |
|---------|---------------|----------|
| Primary blue | Skyroom blue | `#1a56db` |
| Dark background | Classroom dark | `#1a1a2e` |
| Panel background | Panel dark | `#16213e` |
| Input background | Input dark | `#0f3460` |
| Accent red | Recording/danger | `#e94560` |
| Success green | Connected/active | `#00d26a` |
| Link blue | Links | `#4361ee` |
| Purple accent | Whiteboard/teacher | `#7209b7` |
| Text primary | White text | `#eaeaea` |
| Text secondary | Gray text | `#a0a0a0` |
| Border | Dark border | `#2a2a4a` |

### 16.2 Skyroom Typography (Exact)

| Element | Font | Size | Weight |
|---------|------|------|--------|
| App title | Vazirmatn | 18px | 800 |
| Section title | Vazirmatn | 16px | 700 |
| Body text | Vazirmatn | 14px | 400 |
| Small text | Vazirmatn | 12px | 400 |
| Button text | Vazirmatn | 14px | 600 |
| Chat message | Vazirmatn | 13px | 400 |
| Timer/clock | Vazirmatn Mono | 13px | 600 |
| Badge/label | Vazirmatn | 11px | 600 |

### 16.3 Skyroom Spacing (Exact)

| Element | Value |
|---------|-------|
| Page padding | 16px |
| Card padding | 16px |
| Element gap | 8px |
| Section gap | 24px |
| Border radius (cards) | 12px |
| Border radius (buttons) | 8px |
| Border radius (inputs) | 8px |
| Border radius (modals) | 16px |
| Control button size | 40px |
| Icon size | 20px |
| Avatar size | 32px |

### 16.4 Skyroom Button Styles (Exact)

**Primary button:**
```
background: linear-gradient(135deg, #1a56db, #2563eb)
color: white
border-radius: 8px
padding: 10px 20px
font-weight: 600
box-shadow: 0 2px 8px rgba(26, 86, 219, 0.25)
```

**Danger button:**
```
background: linear-gradient(135deg, #dc2626, #ef4444)
color: white
border-radius: 8px
padding: 10px 20px
font-weight: 600
```

**Ghost button:**
```
background: transparent
color: #a0a0a0
border: 1px solid #2a2a4a
border-radius: 8px
padding: 8px 16px
```

**Icon button (control bar):**
```
background: #0f3460
border-radius: 50%
width: 40px
height: 40px
color: #eaeaea
```

**Icon button (active/danger):**
```
background: #e94560  (or #4361ee for active)
border-radius: 50%
width: 40px
height: 40px
color: white
```

### 16.5 Skyroom Card Style (Exact)

```
background: #16213e
border: 1px solid #2a2a4a
border-radius: 12px
padding: 16px
transition: all 0.2s ease
```

**Hover:**
```
box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3)
transform: translateY(-1px)
```

### 16.6 Skyroom Modal Style (Exact)

```
background: #16213e
border: 1px solid #2a2a4a
border-radius: 16px
box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5)
```

**Overlay:**
```
background: rgba(0, 0, 0, 0.6)
backdrop-filter: blur(4px)
```

### 16.7 Skyroom Input Style (Exact)

```
background: #0f3460
border: 1px solid #2a2a4a
border-radius: 8px
color: #eaeaea
padding: 10px 14px
font-size: 14px
```

**Focus:**
```
border-color: #4361ee
box-shadow: 0 0 0 3px rgba(67, 97, 238, 0.2)
```

**Placeholder:**
```
color: #a0a0a0
```

### 16.8 Skyroom Table Style (Exact)

```
background: #16213e
border: 1px solid #2a2a4a
border-radius: 12px
overflow: hidden
```

**Header:**
```
background: #0f3460
color: #a0a0a0
font-size: 12px
font-weight: 600
text-transform: uppercase
```

**Row:**
```
border-bottom: 1px solid #2a2a4a
transition: background 0.15s
```

**Row hover:**
```
background: #0f3460
```

### 16.9 Skyroom Badge/Tag Style (Exact)

**Status badge (active):**
```
background: rgba(0, 210, 106, 0.15)
color: #00d26a
border-radius: 9999px
padding: 4px 10px
font-size: 11px
font-weight: 600
```

**Status badge (inactive):**
```
background: rgba(233, 69, 96, 0.15)
color: #e94560
border-radius: 9999px
padding: 4px 10px
font-size: 11px
font-weight: 600
```

**Count badge:**
```
background: #e94560
color: white
border-radius: 50%
width: 18px
height: 18px
font-size: 10px
font-weight: 700
display: flex
align-items: center
justify-content: center
```

### 16.10 Skyroom Scrollbar Style (Exact)

```css
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #16213e;
}

::-webkit-scrollbar-thumb {
  background: #2a2a4a;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #4361ee;
}
```

### 16.11 Skyroom Animation Timing (Exact)

| Animation | Duration | Easing |
|-----------|----------|--------|
| Page transition | 200ms | ease-out |
| Modal open | 200ms | ease-out |
| Modal close | 150ms | ease-in |
| Button hover | 150ms | ease |
| Card hover | 200ms | ease |
| Sidebar collapse | 300ms | ease |
| Toast notification | 300ms | ease-out |
| Speaking indicator | 100ms | ease |
| Pulse (recording) | 1500ms | ease-in-out infinite |

### 16.12 Skyroom Icon Usage (Exact)

| Context | Icon Style | Size |
|---------|-----------|------|
| Navigation | Outlined | 20px |
| Buttons | Filled | 18px |
| Status indicators | Filled | 12px |
| Avatars | Letter avatar | 32px |
| Control bar | Filled | 20px |
| Badges | Filled | 12px |

**Icon set:** Lucide Icons (consistent with current setup)

### 16.13 Skyroom Empty State (Exact)

```
display: flex
flex-direction: column
align-items: center
justify-content: center
padding: 48px 24px
text-align: center
```

**Empty state icon:**
```
width: 64px
height: 64px
color: #2a2a4a
margin-bottom: 16px
```

**Empty state title:**
```
font-size: 16px
font-weight: 600
color: #eaeaea
margin-bottom: 8px
```

**Empty state description:**
```
font-size: 13px
color: #a0a0a0
```

### 16.14 Skyroom Toast Notification (Exact)

```
position: fixed
bottom: 24px
left: 50%
transform: translateX(-50%)
background: #16213e
border: 1px solid #2a2a4a
border-radius: 8px
padding: 12px 20px
color: #eaeaea
font-size: 13px
box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3)
z-index: 9999
animation: slideUp 200ms ease-out
```

**Success toast:** Left border `3px solid #00d26a`  
**Error toast:** Left border `3px solid #e94560`  
**Info toast:** Left border `3px solid #4361ee`

---

## 17. Skyroom Feature Parity Checklist

### 17.1 Classroom Features

| Feature | Skyroom | IRoom Current | IRoom Target |
|---------|---------|---------------|--------------|
| Separate popup window | ✅ | ❌ | ✅ |
| Always-on-top toggle | ✅ | ❌ | ✅ |
| 3-column layout | ✅ | ❌ | ✅ |
| Active speaker view | ✅ | ❌ | ✅ |
| Screen share with overlay | ✅ | ❌ | ✅ |
| Whiteboard in video area | ✅ | ❌ | ✅ |
| Chat always visible | ✅ | ❌ | ✅ |
| Participants always visible | ✅ | ❌ | ✅ |
| Hand raise | ✅ | ❌ | ✅ |
| Teacher mute/remove | ✅ | ❌ | ✅ |
| Unread chat badge | ✅ | ❌ | ✅ |
| File share in chat | ✅ | ❌ | ✅ |
| Elapsed timer | ✅ | ❌ | ✅ |
| Connection status | ✅ | ✅ | ✅ |
| Recording indicator | ✅ | ✅ | ✅ |
| Settings popup | ✅ | ❌ | ✅ |
| Main tab stays functional | ✅ | ❌ | ✅ |
| Mobile fallback | ✅ | ❌ | ✅ |
| Persian numbers | ✅ | ❌ | ✅ |
| Jalali calendar | ✅ | ❌ | ✅ |
| Vazirmatn font | ✅ | ✅ | ✅ |
| Dark theme | ✅ | ✅ | ✅ |

### 17.2 Admin Panel Features

| Feature | Skyroom | IRoom Current | IRoom Target |
|---------|---------|---------------|--------------|
| Dashboard with live stats | ✅ | ❌ | ✅ |
| Live rooms section | ✅ | ❌ | ✅ |
| System health monitor | ✅ | ❌ | ✅ |
| Activity feed | ✅ | ❌ | ✅ |
| Room cards with status | ✅ | ❌ | ✅ |
| Room detail with tabs | ✅ | ❌ | ✅ |
| User management table | ✅ | ✅ | ✅ |
| Session management | ✅ | ✅ | ✅ |
| Recording management | ✅ | ✅ | ✅ |
| Ticket management | ✅ | ✅ | ✅ |
| Activity logs | ✅ | ✅ | ✅ |
| Settings page | ✅ | ✅ | ✅ |
| Bulk operations | ✅ | ❌ | ✅ |
| CSV export | ✅ | ❌ | ✅ |
| User impersonation | ✅ | ❌ | ✅ |
| Announcement system | ✅ | ❌ | ✅ |

---

*This specification is designed to achieve 100% UI/UX parity with Skyroom.ir, adapted for IRoom's tech stack (SvelteKit + LiveKit + Go). All visual design tokens, spacing, colors, and animations match Skyroom exactly.*
