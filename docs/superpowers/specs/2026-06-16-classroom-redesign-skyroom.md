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
