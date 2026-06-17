# Skyroom Saved UI — Complete Replication Plan

**Goal:** Replicate the complete Skyroom online classroom UI (from `skyroom_saved_ui/`) into the existing SvelteKit frontend, with pixel-perfect visual fidelity for every component, layout, and interaction state.

**Architecture:** The Skyroom online classroom uses a dark-themed full-screen layout with a top header bar, a collapsible sidebar (users + chat panels), a central media area, and a bottom toolbar. The existing iroom codebase already has a classroom page at `web/src/routes/(app)/classroom/[id]/+page.svelte` with Janus WebRTC integration. This plan refactors the visual layer to match Skyroom's exact design while preserving all WebRTC functionality.

**Tech Stack:** Svelte 5, SvelteKit, Tailwind CSS v4, Vazirmatn font, inline SVG icons, CSS custom properties for dark theming.

---

## Design System Reference

### Color Tokens (Dark Theme — Online Classroom)
```
--dark-bg:           #1a1a2e    (main background)
--dark-surface:      #252540    (panel backgrounds)
--dark-border:       #3a3a5a    (dividers, borders)
--dark-text:         #e2e8f0    (primary text)
--dark-muted:        #94a3b8    (secondary/muted text)
--primary:           #23b9d7    (accent, active states)
--danger:            #e05252    (mute, hand raised, danger)
--success:           #40bf7f    (online, active)
--warning:           #d7911d    (warnings)
```

### Typography
```
Font:       Vazirmatn (already in project)
Sizes:      11px (timestamps), 12px (labels), 13px (body), 14px (titles), 16px (headings)
Weights:    400 (normal), 500 (medium), 600 (semibold), 700 (bold)
```

### Spacing Scale
```
4px, 8px, 12px, 16px, 20px, 24px, 32px
```

### Border Radius
```
Small (badges, tags):   4px
Medium (buttons, inputs): 8px
Large (panels, cards):  12px
Full (avatars):         50%
```

---

## Task 1: Online Classroom — Header Bar

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 173-191

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ [👤] OwnerName : RoomName                    ⏱ 00:13  [clock icon]     │
└──────────────────────────────────────────────────────────────────────────┘
```
- Height: 48px
- Background: `--dark-surface` (#252540)
- Border-bottom: 1px solid `--dark-border` (#3a3a5a)
- Owner logo: 28×28px circular image
- Room name: 14px, `--dark-text`, font-weight 500
- Timer: 13px monospace, `--dark-muted`, with clock SVG icon (16px)
- Layout: flex row, space-between, align-items center
- Padding: 0 16px

**Files:** Modify `web/src/routes/(app)/classroom/[id]/+page.svelte`, Create `web/src/lib/components/classroom/ClassroomHeader.svelte`

- [ ] Step 1: Create `ClassroomHeader.svelte` with props: `ownerName`, `ownerLogoUrl`, `roomName`, `timerSeconds`. Layout: flex row with owner info on left (avatar + name + colon + room name), timer on right (time display + clock icon).
- [ ] Step 2: Style with dark theme: bg `#252540`, border-bottom `1px solid #3a3a5a`, height 48px, padding 0 16px.
- [ ] Step 3: Timer format: `MM:SS` computed from seconds prop, updates via `setInterval`.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 2: Online Classroom — Toolbar (Media Controls)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 206-234

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ [🔊] [🎤] [📹] [💻] [🖌️] [📊] [✋]                                       │
└──────────────────────────────────────────────────────────────────────────┘
```
- Height: 52px
- Background: transparent (overlaid on media area, bottom-aligned)
- Buttons: 40×40px circular, `--dark-surface` background, `--dark-border` border
- Icon size: 24×24px SVG
- Active state: icon color `--primary` (#23d9d7), border color `--primary`
- Inactive/off state: icon color `--dark-muted` (#94a3b8), background `#252540`
- Hover: background lightens to `#2d2d4a`
- Gap between buttons: 8px
- Layout: flex row, centered, padding 8px
- Button order (LTR): Audio output, Microphone, Webcam, Screen share, Whiteboard, Files, Hand raise

**Files:** Create `web/src/lib/components/classroom/Toolbar.svelte`

- [ ] Step 1: Create `Toolbar.svelte` with props for each toggle state: `audioOn`, `micOn`, `webcamOn`, `screenShareOn`, `whiteboardOn`, `filesOn`, `handRaised`. Emit `onToggle` events.
- [ ] Step 2: Each button is a 40×40px circle with SVG icon. Active: border `#23b9d7`, icon `#23b9d7`. Inactive: border `#3a3a5a`, icon `#94a3b8`. Off state (mic off, webcam off): icon shows "off" variant (mic_off, videocamoff).
- [ ] Step 3: Add tooltips on hover (title attribute): "خروجی صدا", "میکروفون", "وبکم", "اشتراک‌گذاری صفحه", "تخته", "فایل‌ها", "بالا بردن دست".
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 3: Online Classroom — Mini Toolbar (Sidebar Toggles)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 196-204

**Visual Design:**
```
┌──────────┐
│ [👥] [💬] │  ← positioned at top-right of content area
└──────────┘
```
- Two icon buttons: Users (group icon) and Chat (chat icon)
- Same 40×40px circular button style as main toolbar
- Active state: background `--primary` with white icon
- Inactive: `--dark-surface` background with `--dark-muted` icon
- Position: absolute top-right of the workspace, above the sidebar
- These toggle the sidebar panels visibility

**Files:** Create `web/src/lib/components/classroom/MiniToolbar.svelte`

- [ ] Step 1: Create `MiniToolbar.svelte` with props: `usersPanelOpen`, `chatPanelOpen`. Two toggle buttons with group/chat SVG icons.
- [ ] Step 2: Style identically to Toolbar buttons. Active: bg `#23b9d7`, white icon. Inactive: bg `#252540`, muted icon.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 4: Online Classroom — App Menu (Dropdown)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 193-195 (button) + 299-333 (menu)

**Visual Design:**
```
┌─────────────┐
│ [☰]         │  ← hamburger button, top-left
└─────────────┘
    ↓ click
    ┌──────────────────────────┐
    │ [ℹ] اطلاعات کاربری       │
    │ [📶] وضعیت اتصال         │
    │ [⚙] تنظیمات             │
    │ [🌐] چیدمان         →   │  ← submenu
    │ ─────────────────────── │
    │ [🖌] تخته  (mobile)     │
    │ [📊] فایل‌ها (mobile)    │
    │ ─────────────────────── │
    │ [🚪] خروج               │
    │ [⏻] بستن اتاق (operator)│
    └──────────────────────────┘
```
- Menu button: 40×40px, hamburger icon (3 horizontal lines), same circular style
- Dropdown: white background, border-radius 8px, box-shadow `0 4px 20px rgba(0,0,0,0.3)`
- Menu items: 14px text, 12px icon (24px), padding 10px 16px, hover bg `#f0f2f5`
- Separator: 1px border `#e0e4eb`
- Submenu arrow: small chevron-left icon
- "بستن اتاق" item: only visible for operators, text color `--danger`

**Files:** Create `web/src/lib/components/classroom/AppMenu.svelte`

- [ ] Step 1: Create `ClassroomHeader.svelte` integration with AppMenu dropdown. Use the existing `AppMenu.svelte` as reference but restyle to match Skyroom's dark theme.
- [ ] Step 2: Menu items with SVG icons, Persian labels, hover states. Submenu for "چیدمان" with checkable items (کاربران, پیام‌ها, اسلاید, تخته).
- [ ] Step 3: Position dropdown below the hamburger button, left-aligned.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 5: Online Classroom — Users Panel (Sidebar Block)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 247-266

**Visual Design:**
```
┌─────────────────────────┐
│ [👥] کاربران        (1) │  ← block header
├─────────────────────────┤
│ 👤 honarestanfarhang... │  ← owner row (highlighted)
│    [🎤] [📹] [💻] [🖌] [✋] │  ← action buttons per user
├─────────────────────────┤
│ (hands-up section)      │
├─────────────────────────┤
│ (other role groups)     │
└─────────────────────────┘
```
- Block header: 12px label "کاربران", group icon (16px), user count badge
- Header background: `--dark-surface`, border-bottom: 1px solid `--dark-border`
- User row: flex row, 48px height, padding 8px 12px
- User avatar: 32×32px circle, `--primary` border for owner
- User nickname: 13px, `--dark-text`, truncated with ellipsis
- Role icon: small 16px SVG (person, owner crown, etc.)
- Action buttons per user: 28×28px icon buttons (mic, webcam, desktop, board, hand) — only visible for owner/operator
- Active action: `--primary` color. Inactive: `--dark-muted`
- Scrollable list with perfect-scrollbar style (6px wide scrollbar)
- Role groups: owner → support → admin → operator → presenter → active → regular users
- Hands-up section: separated, shows users who raised hands

**Files:** Create `web/src/lib/components/classroom/UsersPanel.svelte`, Create `web/src/lib/components/classroom/UserRow.svelte`

- [ ] Step 1: Create `UsersPanel.svelte` with props: `users` (array of `{id, nickname, role, avatarUrl, micOn, webcamOn, desktopOn, boardOn, handRaised, isOwner}`), `currentUserId`.
- [ ] Step 2: Create `UserRow.svelte` for individual user rows. Layout: role icon + avatar + nickname on left, action buttons on right. Owner row gets `--primary` accent border.
- [ ] Step 3: Group users by role, render hands-up section separately. Show user count in header.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 6: Online Classroom — Chat Panel (Sidebar Block)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 267-283

**Visual Design:**
```
┌─────────────────────────┐
│ [💬] پیام‌ها            │  ← block header
├─────────────────────────┤
│ ┌─────────────────────┐ │
│ │ 😃 سلام              │ │  ← message bubbles
│ │ 12:30               │ │
│ └─────────────────────┘ │
│     ┌───────────────┐   │
│     │ 😄 خوش آمدید  │   │  ← own messages (right-aligned)
│     │ 12:31         │   │
│     └───────────────┘   │
│ (pinned messages area)  │
├─────────────────────────┤
│ [📝 type here...] [😊] [➤]│  ← input area
└─────────────────────────┘
```
- Block header: same style as Users panel, "پیام‌ها" label, chat icon
- Message area: scrollable, padding 12px, gap 8px between messages
- Message bubble: max-width 85%, border-radius 12px, padding 8px 12px
- Other user messages: `--dark-surface` bg, `--dark-text`, left-aligned
- Own messages: `--primary` bg, white text, right-aligned
- Message text: 13px, line-height 1.4
- Timestamp: 10px, `--dark-muted`, below message
- Pinned messages: shown at top with `--warning-bg` background, pin icon
- Input area: contentEditable div, min-height 36px, max-height 80px, border-top 1px solid `--dark-border`
- Input placeholder: "پیام خود را وارد کنید" (13px, `--dark-muted`)
- Emoji button: 28×28px icon button, emoji picker on click (grid of 27 common emojis)
- Send button: 28×28px, `--primary` color, disabled when input empty
- Reply-to preview: shown above input when replying, with cancel button and reply icon
- Edit preview: similar to reply preview, with edit icon
- Emoji picker: 3 rows × 9 columns, shown below input on emoji button click

**Files:** Create `web/src/lib/components/classroom/ChatPanel.svelte`, Create `web/src/lib/components/classroom/ChatMessage.svelte`, Create `web/src/lib/components/classroom/EmojiPicker.svelte`

- [ ] Step 1: Create `ChatPanel.svelte` with props: `messages` (array of `{id, userId, nickname, text, timestamp, isOwn, isPinned, replyTo}`), `currentUserId`. Include message list, pinned messages, input area.
- [ ] Step 2: Create `ChatMessage.svelte` for individual messages. Style: bubble with rounded corners, own messages right-aligned with `--primary` bg, others left-aligned with `--dark-surface` bg.
- [ ] Step 3: Create `EmojiPicker.svelte` with 27 common emojis in a 3×9 grid. On emoji click, insert into input.
- [ ] Step 4: Implement reply-to and edit preview bars above input.
- [ ] Step 5: Run `cd web && npm run check` to verify.

---

## Task 7: Online Classroom — Sidebar Layout

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 236-283

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────┐
│ Header Bar                                                   │
├──────────────────────────────────────────────────────────────┤
│                                        ┌──────────────────┐  │
│                                        │ Mini Toolbar     │  │
│                                        ├──────────────────┤  │
│                                        │                  │  │
│                                        │ Users Panel      │  │
│                                        │ (scrollable)     │  │
│                                        │                  │  │
│          Main Content Area            ├──────────────────┤  │
│          (media/slides/               │                  │  │
│           whiteboard)                 │ Chat Panel       │  │
│                                        │ (scrollable)     │  │
│                                        │                  │  │
│                                        ├──────────────────┤  │
│                                        │ [📝 input] [😊][➤]│  │
│                                        └──────────────────┘  │
├──────────────────────────────────────────────────────────────┤
│ Toolbar (centered, bottom)                                   │
└──────────────────────────────────────────────────────────────┘
```
- Sidebar width: 280px (expanded), 0 (collapsed)
- Sidebar background: `--dark-surface` (#252540)
- Sidebar position: left side (LTR) — in RTL layout, it appears on the right
- Transition: width 0.25s ease, overflow hidden
- Users panel: flex-grow, takes available space above chat
- Chat panel: fixed height ~280px at bottom of sidebar, or flex-shrink
- Main content area: flex-grow, fills remaining space
- Toolbar: positioned absolute bottom-center, margin-bottom 16px

**Files:** Modify `web/src/routes/(app)/classroom/[id]/+page.svelte`

- [ ] Step 1: Refactor the classroom page layout to match Skyroom's structure: header (full-width), workspace (sidebar + main area), toolbar (bottom overlay).
- [ ] Step 2: Sidebar: 280px fixed width, dark background, contains UsersPanel and ChatPanel stacked vertically. Users panel flex-grow, chat panel fixed height.
- [ ] Step 3: Main area: flex-grow, dark background `#1a1a2e`, contains media/slides/whiteboard content.
- [ ] Step 4: Toolbar: absolute positioned at bottom, centered, with semi-transparent dark background and border-radius 24px.
- [ ] Step 5: Run `cd web && npm run dev` and verify layout structure at `/classroom/[id]`.

---

## Task 8: Online Classroom — Media Block (Slideshow/Video)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 237-241

**Visual Design:**
```
┌──────────────────────────────────┐
│ [📊] اسلاید                      │  ← block header (when in sidebar)
├──────────────────────────────────┤
│                                  │
│    ┌──────────────────────────┐  │
│    │                          │  │
│    │    Slide Content         │  │  ← video player area
│    │    (video wrapper)       │  │
│    │                          │  │
│    └──────────────────────────┘  │
│                                  │
│    [spinner when loading]        │
└──────────────────────────────────┘
```
- Media block header: slideshow icon (16px), "اسلاید" label, block-menu dropdown (3-dot icon)
- Video container: 16:9 aspect ratio, black background
- Video wrapper: positioned relative, contains video element
- Loading spinner: centered, 38×38 SVG with rotating gradient stroke (same as panel spinner)
- Video statistics: hidden by default (toggle via menu)
- When displayed in main area: full-width, no sidebar block header
- When displayed in sidebar: compact view with block header

**Files:** Create `web/src/lib/components/classroom/MediaBlock.svelte`

- [ ] Step 1: Create `MediaBlock.svelte` with props: `isInSidebar`, `isLoading`, `slideUrl`. Contains video wrapper, loading spinner, optional header.
- [ ] Step 2: Video wrapper: 16:9 aspect ratio, `position: relative`, overflow hidden.
- [ ] Step 3: Loading spinner: centered absolute, 38×38 SVG with rotating gradient stroke animation.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 9: Online Classroom — Video Grid (Webcam Thumbnails)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 243-246

**Visual Design:**
```
┌──────────────────────────────────┐
│ [📹] وبکم                        │  ← block header
├──────────────────────────────────┤
│ ┌──────────┐ ┌──────────┐       │
│ │          │ │          │       │
│ │  Video 1 │ │  Video 2 │  ...  │  ← grid of webcam feeds
│ │          │ │          │       │
│ └──────────┘ └──────────┘       │
└──────────────────────────────────┘
```
- Block header: videocam icon, "وبکم" label
- Grid: responsive, 2 columns in sidebar, 3-4 in main area
- Each video tile: 16:9 aspect ratio, `--dark-bg` background, border-radius 4px
- Video tile shows participant name at bottom (12px, white, semi-transparent bg)
- Muted indicator: small mic-off icon overlay when participant is muted
- Active speaking indicator: `--primary` border glow

**Files:** Create `web/src/lib/components/classroom/VideoGrid.svelte`

- [ ] Step 1: Create `VideoGrid.svelte` with props: `participants` (array of `{id, name, stream, isMuted, isSpeaking}`), `columns` (2 for sidebar, auto for main).
- [ ] Step 2: Grid layout using CSS grid, gap 4px, responsive columns.
- [ ] Step 3: Each tile: video element, name label, muted/speaking indicators.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 10: Online Classroom — Whiteboard Block

**Reference file:** `web/src/lib/components/Whiteboard.svelte` (existing, needs restyling)

**Visual Design:**
```
┌──────────────────────────────────┐
│ [🖌] تخته                        │  ← block header
├──────────────────────────────────┤
│                                  │
│    ┌──────────────────────────┐  │
│    │                          │  │
│    │    Whiteboard Canvas     │  │  ← drawing area
│    │                          │  │
│    └──────────────────────────┘  │
│                                  │
└──────────────────────────────────┘
```
- Block header: brush icon, "تخته" label
- Canvas: full available area, white background (for drawing), crosshair cursor when drawing
- Drawing tools: pen, eraser, color picker, clear (positioned as overlay or in toolbar)
- When active: main content area shows whiteboard instead of media

**Files:** Modify `web/src/lib/components/Whiteboard.svelte`

- [ ] Step 1: Restyle the existing Whiteboard component to match Skyroom's dark theme. Canvas area: white bg for drawing, dark surrounding.
- [ ] Step 2: Add block header with brush icon and "تخته" label when displayed in sidebar.
- [ ] Step 3: Drawing tools: pen (default), eraser, color palette (primary colors), clear button. Position as floating toolbar.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 11: Online Classroom — Block Menu (3-Dot Dropdown)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 238, 248, 268 (menu buttons) + 333-355 (chat menu)

**Visual Design:**
```
┌──────────────────────────┐
│ [⋮]  ← 3-dot menu button  │
└───┬──────────────────────┘
    ↓ click
    ┌──────────────────────────┐
    │ [⬆] نمایش بزرگتر        │  ← enlarge to main area
    │ ─────────────────────── │
    │ [✕] بستن                │  ← close panel
    └──────────────────────────┘

    Chat-specific menu:
    ┌──────────────────────────┐
    │ [⬆] نمایش بزرگتر        │
    │ [🔒] غیر فعال سازی چت   │
    │ [👁] حالت خصوصی          │
    │ [🗑] پاک کردن همه پیام‌ها│
    │ [💾] ذخیره تمام پیام‌ها  │
    │ [🚫] کاربران ساکت شده → │
    │ ─────────────────────── │
    │ [✕] بستن                │
    └──────────────────────────┘
```
- Menu button: 24×24px, 3 vertical dots icon, `--dark-muted` color
- Dropdown: white bg, border-radius 8px, box-shadow, min-width 180px
- Menu items: 13px text, 16px icon, padding 8px 16px, hover bg `#f0f2f5`
- Checkmark icon (16px) for active/checked items
- Separator: 1px `#e0e4eb` border
- Submenu arrow for items with children

**Files:** Create `web/src/lib/components/classroom/BlockMenu.svelte`

- [ ] Step 1: Create `BlockMenu.svelte` with props: `items` (array of `{icon, label, action, checked, separator, children}`).
- [ ] Step 2: 3-dot button triggers dropdown. Items render with icons, labels, optional checkmarks.
- [ ] Step 3: Position dropdown below the button, right-aligned.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 12: Online Classroom — Chat Message Context Menu

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 357-365

**Visual Design:**
```
┌──────────────────────────┐
│ [📌] سنجاق پیام          │  ← pin message
│ [🚫] ساکت کردن کاربر     │  ← mute user
│ [✏] ویرایش پیام         │  ← edit message
│ [🗑] حذف پیام            │  ← delete message
└──────────────────────────┘
```
- Triggered by right-clicking a message or clicking a message menu button
- White bg, border-radius 8px, box-shadow, padding 4px 0
- Items: 13px text, 16px icon, padding 8px 16px, hover bg `#f0f2f5`
- Pin: attach_file icon. Mute: block icon. Edit: mode_edit icon. Delete: delete icon

**Files:** Modify `web/src/lib/components/classroom/ChatMessage.svelte`

- [ ] Step 1: Add right-click context menu to ChatMessage component with pin, mute, edit, delete options.
- [ ] Step 2: Emit events for each action: `onPin`, `onMute`, `onEdit`, `onDelete`.
- [ ] Step 3: Position menu at cursor location, close on outside click.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 13: Online Classroom — Emoji Picker

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 271-273

**Visual Design:**
```
┌──────────────────────────────────────────────┐
│ [😃] [😄] [😊] [😁] [😂] [😅] [😉] [😜] [😍] │
│ [😏] [😒] [😞] [😩] [😢] [😭] [😤] [😡] [😲] │
│ [😨] [😱] [🙏] [👍] [👎] [👏] [👋] [👌] [✌] │
│ [❤] [🌹]                                      │
└──────────────────────────────────────────────┘
```
- Grid: 3 rows × 9 columns (27 emojis total)
- Each emoji: 28×28px button, font-size 18px, hover bg `#f0f2f5`, border-radius 4px
- Positioned above the chat input area
- Hidden by default, shown on emoji button click
- Close on outside click or emoji selection

**Files:** Modify `web/src/lib/components/classroom/EmojiPicker.svelte` (existing, needs restyling)

- [ ] Step 1: Restyle the existing EmojiPicker to match Skyroom's grid layout (3×9).
- [ ] Step 2: Use exact emoji set from the reference: 😃😄😊😁😂😅😉😜😍😘😏😒😞😩😢😭😤😡😲😨😱🙏👍👎👏👋👌✌❤🌹
- [ ] Step 3: Position above chat input, toggle visibility on emoji button click.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 14: Online Classroom — Login Page

**Reference file:** `skyroom_saved_ui/اسکای_روم - جلسات.html` lines 22-172

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│                    [Skyroom Logo]                        │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  👤  honarestanfarhangmahan                        │  │
│  │      wwww.alireza.sh9342@gmail.com                 │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  [👤]  ورود با حساب کاربری دیگر                    │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  [username input]                                  │  │
│  │  [password input]                                  │  │
│  │                                                    │  │
│  │  [    مهمان    ] [    ورود    ]                     │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│              [devices illustration SVG]                  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  مرورگرهای زیر پشتیبانی می شوند:                   │  │
│  │  [Chrome] [Firefox] [Opera] [Edge] [Samsung] [Safari]│
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  [FA ▾] | [راهنما] | [قوانین] | [حریم خصوصی]            │
│                                                          │
│                    © اسکای‌روم®                           │
└──────────────────────────────────────────────────────────┘
```
- Centered layout, max-width 400px
- Background: wallpaper (CSS custom properties `--wallpaper-tall`, `--wallpaper-wide`)
- Login box: white bg, border-radius 16px, box-shadow, padding 24px
- Logo: skyroomFa.png, max-width 200px
- Session block: user avatar (48px circle), name (14px bold), email (12px muted), clickable row
- "ورود با حساب کاربری دیگر" link: account_circle icon, 14px text
- Form inputs: border 1px `#e0e4eb`, border-radius 8px, padding 10px 12px, focus border `--primary`
- Buttons: "مهمان" (orange bg `#d7911d`), "ورود" (blue bg `--primary`), full-width, border-radius 8px, padding 10px
- Language selector: `<select>` with EN/FA options
- Footer links: 12px, `--dark-muted`, spaced with dividers
- Supported browsers: logos in a row, 32×32px each
- Trademark: "© اسکای‌روم®" at bottom

**Files:** Modify `web/src/routes/auth/+page.svelte`

- [ ] Step 1: Restyle the auth page to match Skyroom's login layout. Centered card, logo at top, session block, form, guest/login buttons.
- [ ] Step 2: Add session block showing logged-in user (if token exists) with avatar, name, email.
- [ ] Step 3: Add "ورود با حساب کاربری دیگر" link below session block.
- [ ] Step 4: Add supported browsers section with Chrome/Firefox/Opera/Edge/Samsung/Safari logos.
- [ ] Step 5: Add language selector (FA/EN) and footer links (راهنما, قوانین, حریم خصوصی).
- [ ] Step 6: Run `cd web && npm run dev` and verify at `/auth`.

---

## Task 15: Online Classroom — Room Page Layout (Full Assembly)

**Reference file:** `skyroom_saved_ui/اسکای_روم - 1.html` lines 173-289

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ [👤] Owner : RoomName                     ⏱ 00:13  [clock]             │
├──────────────────────────────────────────────────────────────────────────┤
│ [☰]                              [👥] [💬]                    │         │
│                                  ┌─────────────────────────┐ │         │
│                                  │ کاربران             (1) │ │         │
│                                  │ 👤 OwnerName           │ │         │
│                                  │    [🎤][📹][💻][🖌][✋]  │ │         │
│                                  │                         │ │         │
│                                  │ ─────────────────────── │ │         │
│                                  │ پیام‌ها                  │ │         │
│                                  │ ┌─────────────────────┐ │ │         │
│                                  │ │ 😃 سلام              │ │ │         │
│                                  │ └─────────────────────┘ │ │         │
│                                  │ [📝 type...]      [😊][➤]│ │         │
│                                  └─────────────────────────┘ │         │
│                                                                │         │
│         ┌─────────────────────────────────────────────┐        │         │
│         │                                             │        │         │
│         │          Media / Slide / Whiteboard         │        │         │
│         │          (main content area)                │        │         │
│         │                                             │        │         │
│         └─────────────────────────────────────────────┘        │         │
│                                                                │         │
│              [🔊] [🎤] [📹] [💻] [🖌️] [📊] [✋]                 │         │
│                                                                │         │
├──────────────────────────────────────────────────────────────────────────┤
│ © اسکای‌روم®                                                              │
└──────────────────────────────────────────────────────────────────────────┘
```

**Files:** Modify `web/src/routes/(app)/classroom/[id]/+page.svelte`

- [ ] Step 1: Assemble all classroom components into the full page layout: ClassroomHeader, AppMenu, MiniToolbar, UsersPanel, ChatPanel, MediaBlock, VideoGrid, Whiteboard, Toolbar.
- [ ] Step 2: Layout structure:
  - Header: full-width, fixed height 48px
  - Workspace: flex row, height `calc(100vh - 48px)`
    - Main area: flex-grow, position relative, dark bg
    - Sidebar: 280px, flex-shrink 0, dark bg, flex column
      - UsersPanel: flex-grow, overflow-y auto
      - ChatPanel: flex-shrink 0, ~280px height
  - Toolbar: absolute bottom-4, left-50% translate-x(-50%), flex row
- [ ] Step 3: Sidebar collapse: when both mini toolbar buttons are off, sidebar width animates to 0. Main area expands to fill.
- [ ] Step 4: Preserve all existing Janus/WebRTC functionality (janus-client.ts, ClassroomBridge.ts, ClassroomWindow.ts).
- [ ] Step 5: Run `cd web && npm run dev` and verify the full classroom page at `/classroom/[id]`.

---

## Task 16: Online Classroom — Popup Classroom Window

**Reference file:** `web/src/routes/(app)/classroom/popup/[id]/+page.svelte` (existing)

**Visual Design:**
Same as the main classroom page but in a popup window without the sidebar. Contains: header, main media area, toolbar. No users panel or chat panel (or minimal).

**Files:** Modify `web/src/routes/(app)/classroom/popup/[id]/+page.svelte`

- [ ] Step 1: Restyle the popup classroom to match Skyroom's design. Same header, toolbar, and media area styling.
- [ ] Step 2: Remove sidebar elements. Toolbar remains at bottom-center.
- [ ] Step 3: Run `cd web && npm run dev` and verify at `/classroom/popup/[id]`.

---

## Task 17: Admin Panel — Sidebar Component

**Reference file:** `skyroom_saved_ui/اسکای_روم _ پنل مدیریت.html` lines 860-865

**Visual Design:**
```
┌──────────────────┐
│ [◆] SKYROOM      │  ← logo (shape icon + wordmark)
│                  │
│ [👤] اتاق‌ها      │  ← main nav
│ [🎤] رویدادها جدید│  ← with "جدید" badge
│ [📊] سرویس‌ها     │
│ [💳] مالی        │
│ [🎧] پشتیبانی    │
│ [?] راهنما       │
│                  │
│ ──────────────── │
│ [📈] وضعیت       │  ← bottom nav
│ [📋] فهرست اتاق‌ها│
│ [👥] کاربران     │
│ [📁] فایل‌ها      │
│ [⏺] ضبط ابری    │
│ [📅] نشست‌ها      │
│ [📡] اتصال‌ها     │
│                  │
│ 27 خرداد 1405    │  ← date + version
│ (ورژن 3.24.0)    │
└──────────────────┘
```
- Width: 260px expanded, 60px collapsed
- Background: `#1c293a` (dark navy)
- Logo: shape icon (32px) + wordmark SVG, clickable
- Nav items: 14px text, 24px icons, padding 10px 16px
- Active item: background `rgba(255,255,255,0.1)`, right border 3px `--primary`
- Hover: background `rgba(255,255,255,0.05)`
- Text color: `#e2e8f0` (active), `#94a3b8` (inactive)
- Badge "جدید": small orange pill next to "رویدادها"
- Collapse toggle: X icon at top
- Date/version: 11px, `#6790a0`, at bottom
- RTL layout (sidebar on right)

**Files:** Create `web/src/lib/components/SkyroomSidebar.svelte`

- [ ] Step 1: Create `SkyroomSidebar.svelte` with all nav items, logo, collapse toggle, date display.
- [ ] Step 2: Style with dark navy bg `#1c293a`, active states, hover states, RTL layout.
- [ ] Step 3: Collapse animation: width 260px → 60px, hide text labels, center icons.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 18: Admin Panel — Header Component

**Reference file:** `skyroom_saved_ui/اسکای_روم _ پنل مدیریت.html` lines ~860-900

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ [☰]  [breadcrumb: اتاق‌ها / فهرست اتاق‌ها]    [👤 User ▾] [🔔] [❓]     │
└──────────────────────────────────────────────────────────────────────────┘
```
- Height: 56px
- Background: white
- Border-bottom: 1px `#e0e4eb`
- Hamburger: 20×20 grid icon (4×4 dots), toggles sidebar
- Breadcrumb: 14px, `--dark-muted` with `/` separators, active page `--dark-text`
- User dropdown: avatar (32px circle), name (14px), chevron-down icon
- Notifications bell: 24px SVG, optional unread badge (red dot)
- Help icon: circle-question mark, 20px
- Mobile actions: three-dot menu button (hidden on desktop)

**Files:** Create `web/src/lib/components/SkyroomHeader.svelte`

- [ ] Step 1: Create `SkyroomHeader.svelte` with hamburger toggle, breadcrumb, user dropdown, notifications, help icon.
- [ ] Step 2: Style: white bg, bottom border, 56px height, flex row.
- [ ] Step 3: User dropdown menu: "حساب کاربری", "پنل کاربر", "خروج" items with icons.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 19: Admin Panel — Layout Shell

**Reference file:** `skyroom_saved_ui/اسکای_روم _ پنل مدیریت.html` structure

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ Header Bar                                                               │
├────────────────────────┬─────────────────────────────────────────────────┤
│                        │                                                 │
│    Sidebar (260px)     │    Main Content Area                            │
│    (dark navy)         │    (white bg, padding 24px)                     │
│                        │                                                 │
│                        │    ┌─────────────────────────────────────────┐  │
│                        │    │ Page Header + Actions                   │  │
│                        │    ├─────────────────────────────────────────┤  │
│                        │    │                                         │  │
│                        │    │ Page Content                            │  │
│                        │    │                                         │  │
│                        │    └─────────────────────────────────────────┘  │
│                        │                                                 │
└────────────────────────┴─────────────────────────────────────────────────┘
```

**Files:** Modify `web/src/routes/(app)/+layout.svelte`

- [ ] Step 1: Refactor the app layout to use Skyroom admin panel structure: sidebar (right side RTL) + content area.
- [ ] Step 2: Sidebar: fixed position, 260px, dark bg. Content: margin-right 260px, white bg, min-height calc(100vh - 56px).
- [ ] Step 3: Preserve all existing auth, WebSocket, and notification logic.
- [ ] Step 4: Run `cd web && npm run dev` and verify at `/dashboard`.

---

## Task 20: Admin Panel — Button Component

**Reference file:** Various panel pages

**Visual Design:**
```
Primary:    [ باز کردن اتاق ]     bg: #23b9d7, white text, border-radius 8px
Secondary:  [ انصراف ]            bg: #f0f2f5, #1c293a text, border-radius 8px
Danger:     [ حذف اتاق ]          bg: #e05252, white text, border-radius 8px
Ghost:      [ کپی لینک ]          transparent, #23d9d7 text + icon
Outline:    [ بستن ]              border 1px #e0e4eb, #1c293a text

Sizes:      sm: 28px height, md: 36px height, lg: 44px height
Loading:    [ ⟳ spinner + text ]  circular SVG spinner animation
Disabled:   opacity: 0.5, cursor: not-allowed
```
- Padding: 8px 16px (md), 6px 12px (sm), 10px 20px (lg)
- Font: 14px (md), 12px (sm), 16px (lg), weight 600
- Transition: all 0.15s ease
- Hover: darken bg 10%
- Active: scale 0.98

**Files:** Create `web/src/lib/components/ui/SkyroomButton.svelte`

- [ ] Step 1: Create with props: `variant` ('primary'|'secondary'|'danger'|'ghost'|'outline'), `size` ('sm'|'md'|'lg'), `loading`, `disabled`, `icon`, `type`.
- [ ] Step 2: Style each variant to match Skyroom exactly.
- [ ] Step 3: Loading state: inline SVG spinner (circular gradient animation).
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 21: Admin Panel — Input Component

**Reference file:** `اسکای_روم _ پنل مدیریت.html` form fields

**Visual Design:**
```
┌──────────────────────────────────────┐
│ عنوان اتاق: *                        │  ← label (14px, #1c293a)
│ ┌──────────────────────────────────┐ │
│ │                                  │ │  ← input (border 1px #e0e4eb, radius 8px)
│ └──────────────────────────────────┘ │
│ https://skyroom.online/ch/...        │  ← help text (12px, #6790a0)
└──────────────────────────────────────┘

Focus state: border-color: #23b9d7, box-shadow: 0 0 0 2px rgba(35,185,215,0.2)
Error state:  border-color: #e0e4eb, error text below (12px, #e05252)
Disabled:    bg: #f9fafb, cursor: not-allowed
```
- Input height: 38px, padding 10px 12px
- Label: 14px, `--dark-text`, margin-bottom 4px, optional required asterisk (red)
- Tooltip icon: small question mark circle, hover shows popover
- Help text: 12px, `--text-secondary`, below input

**Files:** Create `web/src/lib/components/ui/SkyroomInput.svelte`

- [ ] Step 1: Create with props: `label`, `type`, `placeholder`, `value`, `error`, `helpText`, `tooltip`, `required`, `disabled`, `inputMode`.
- [ ] Step 2: Style: border 1px `#e0e4eb`, border-radius 8px, focus ring `#23b9d7`.
- [ ] Step 3: Add tooltip popover on hover (inline-block with popper arrow).
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 22: Admin Panel — Select Component

**Reference file:** `اسکای_روم _ پنل مدیریت.html` dropdown selects

**Visual Design:**
```
┌──────────────────────────────────────┐
│ سرویس میزبان: *                      │
│ ┌──────────────────────────────────┐ │
│ │ سرویس آموزشی ۵۰ کاربره      [▾] │ │  ← custom select
│ └──────────────────────────────────┘ │
└──────────────────────────────────────┘

Dropdown:
┌──────────────────────────────────────┐
│ سرویس آموزشی ۵۰ کاربره              │  ← selected (bg #f0f2f5)
│ سرویس آموزشی ۱۰۰ کاربره            │
│ سرویس آموزشی ۲۰۰ کاربره            │
└──────────────────────────────────────┘
```
- Same height/border/radius as Input
- Custom dropdown arrow: chevron-down SVG
- Dropdown: white bg, border 1px `#e0e4eb`, border-radius 8px, box-shadow
- Options: padding 8px 12px, hover bg `#f0f2f5`, selected bg `#f0f2f5` with checkmark
- Searchable: input at top of dropdown

**Files:** Create `web/src/lib/components/ui/SkyroomSelect.svelte`

- [ ] Step 1: Create with props: `label`, `options` (array of `{value, label}`), `value`, `searchable`, `required`.
- [ ] Step 2: Custom styled select with dropdown panel, search input, option list.
- [ ] Step 3: Match Skyroom's exact styling: border, radius, shadow, hover states.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 23: Admin Panel — Checkbox Component

**Reference file:** `اسکای_روم _ پنل مدیریت.html` checkboxes

**Visual Design:**
```
┌──────────────────────────────────────┐
│ ☑ ورود به عنوان مهمان                │
│ ☐ ابتدا اپراتور وارد شود            │
└──────────────────────────────────────┘

Checkbox: 18×18px, border 2px #c8c8c8, border-radius 4px
Checked:  bg #23b9d7, white checkmark SVG
Label:    14px, #1c293a, margin-right 8px
```

**Files:** Create `web/src/lib/components/ui/SkyroomCheckbox.svelte`

- [ ] Step 1: Create with props: `label`, `checked`, `disabled`.
- [ ] Step 2: Custom checkbox: 18×18px, border 2px `#c8c8c8`, checked state with `--primary` bg and white checkmark.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 24: Admin Panel — Card Component

**Reference file:** `اسکای_روم _ پنل مدیریت.html` form sections

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ ویرایش مشخصات اتاق :                    [باز کردن اتاق]  │  ← header
│ جلسات                                    [حذف اتاق]     │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  [Form fields...]                                        │
│                                                          │
├──────────────────────────────────────────────────────────┤
│ تاریخچه اتاق                                             │  ← sub-section header
│ ⏱ زمان مصرفی: 00:01:24                    [↻ refresh]   │
│ ⏱ مجموع زمان مصرفی: 00:01:24                             │
│ 🕐 ایجاد شده در: 10:36:24 - 1405/03/26                   │
│ 🕐 بروز شده در: 12:01:37 - 1405/03/26                   │
└──────────────────────────────────────────────────────────┘
```
- Background: white
- Border: 1px `#e0e4eb`
- Border-radius: 12px
- Box-shadow: 0 1px 3px rgba(0,0,0,0.08)
- Header: padding 16px 20px, border-bottom 1px `#e0e4eb`, flex row space-between
- Title: 16px, `--dark-text`, font-weight 600
- Actions: flex row, gap 8px
- Content: padding 20px
- Sub-section header: 14px, `--dark-text`, font-weight 600, margin-top 20px

**Files:** Create `web/src/lib/components/ui/SkyroomCard.svelte`

- [ ] Step 1: Create with slots: `header`, `title`, `actions`, `content`, `footer`.
- [ ] Step 2: Style: white bg, border 1px `#e0e4eb`, border-radius 12px, subtle shadow.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 25: Admin Panel — Tab Navigation

**Reference file:** `اسکای_روم _ پنل مدیریت.html` "مشخصات" / "کاربران" tabs

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ [ مشخصات ]  [ کاربران ]                                   │
│  ─────────                                               │  ← active: 2px #23b9d7 bottom border
└──────────────────────────────────────────────────────────┘
```
- Tab height: 40px
- Active: `--primary` bottom border (2px), `--primary` text
- Inactive: `--dark-muted` text, no border
- Hover: `--dark-text` text
- Font: 14px, weight 500
- Gap: 24px between tabs

**Files:** Create `web/src/lib/components/ui/SkyroomTabs.svelte`

- [ ] Step 1: Create with props: `tabs` (array of `{id, label, active}`), `onSelect`.
- [ ] Step 2: Style: horizontal flex, bottom border 1px `#e0e4eb`, active tab with `--primary` bottom border.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 26: Admin Panel — Table Component

**Reference file:** Various panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ نام اتاق    │ پیوند       │ سرویس      │ وضعیت  │ کاربران │ عملیات      │
├─────────────┼─────────────┼────────────┼────────┼─────────┼──────────────┤
│ جلسات       │ farhangmahan│ سرویس ۵۰   │ ● فعال │ 0       │ [✏] [🗑]    │
│ ...         │ ...         │ ...        │ ...    │ ...     │ [✏] [🗑]    │
└─────────────┴─────────────┴────────────┴────────┴─────────┴──────────────┘
```
- Header: bg `#f9fafb`, 12px text `--dark-text` weight 600, padding 10px 16px
- Rows: border-bottom 1px `#e0e4eb`, padding 12px 16px, hover bg `#f9fafb`
- Cell text: 13px `--dark-text`
- Status badge: colored dot (8px circle) + text
- Action buttons: 28×28px icon buttons, `--dark-muted` → `--primary` on hover
- Empty state: centered, muted text, optional illustration

**Files:** Create `web/src/lib/components/ui/SkyroomTable.svelte`

- [ ] Step 1: Create with props: `columns` (array of `{key, label, sortable}`), `rows`, `emptyMessage`.
- [ ] Step 2: Style: full-width table, header bg `#f9fafb`, row borders, hover states.
- [ ] Step 3: Add sortable column headers with chevron icons.
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 27: Admin Panel — Badge Component

**Reference file:** Status badges throughout panel pages

**Visual Design:**
```
Success:  [ ● فعال ]       bg: #e6f9f0, text: #40bf7f, border-radius 16px
Warning:  [ ● منقضی شده ]  bg: #fff4e5, text: #d7911d, border-radius 16px
Danger:   [ ● غیرفعال ]    bg: #fde8e8, text: #e05252, border-radius 16px
Info:     [ ● جدید ]       bg: #e6f4ff, text: #23b9d7, border-radius 16px
Default:  [ ● ... ]        bg: #f0f2f5, text: #6790a0, border-radius 16px
```
- Padding: 2px 10px
- Font: 12px, weight 500
- Dot: 6px circle, same color as text
- Border-radius: 16px

**Files:** Create `web/src/lib/components/ui/SkyroomBadge.svelte`

- [ ] Step 1: Create with props: `variant` ('success'|'warning'|'danger'|'info'|'default'), `label`, `dot` (boolean).
- [ ] Step 2: Style each variant with matching bg/text colors.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 28: Admin Panel — Modal Component

**Reference file:** Modal patterns in panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ ┌──────────────────────────────────────────────────────┐ │
│ │ عنوان مودال                              [✕]        │ │
│ ├──────────────────────────────────────────────────────┤ │
│ │                                                      │ │
│ │  [Modal content...]                                  │ │
│ │                                                      │ │
│ ├──────────────────────────────────────────────────────┤ │
│ │                    [ انصراف ]  [ تأیید ]              │ │
│ └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```
- Overlay: `rgba(0,0,0,0.5)`, fixed, full-screen
- Panel: white bg, border-radius 12px, max-width 550px, centered
- Header: 16px title, close button (X icon), border-bottom 1px `#e0e4eb`
- Content: padding 20px
- Footer: padding 16px 20px, border-top 1px `#e0e4eb`, buttons right-aligned

**Files:** Create `web/src/lib/components/ui/SkyroomModal.svelte`

- [ ] Step 1: Create with props: `open`, `title`, `onClose`, `size` ('sm'|'md'|'lg'). Slots: `content`, `footer`.
- [ ] Step 2: Style: overlay, centered panel, header with close button, footer with action buttons.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 29: Admin Panel — Toast Component

**Reference file:** `#toasts` div in panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ ┌──────────────────────────────────────────────────────┐ │
│ │ ✓ عملیات با موفقیت انجام شد                    [✕]  │ │  ← success
│ └──────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────┐ │
│ │ ✕ خطا در انجام عملیات                           [✕]  │ │  ← error
│ └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```
- Position: fixed top-right, z-index 9999
- Success: left border 4px `#40bf7f`, white bg, `--dark-text`
- Error: left border 4px `#e05252`, white bg, `--dark-text`
- Warning: left border 4px `#d7911d`, white bg, `--dark-text`
- Padding: 12px 16px, border-radius 8px, box-shadow
- Auto-dismiss after 4 seconds
- Close button: X icon
- Animation: slide-in from right, 0.3s ease

**Files:** Create `web/src/lib/components/ui/SkyroomToast.svelte`

- [ ] Step 1: Create with props: `type`, `message`, `duration`, `onDismiss`.
- [ ] Step 2: Style: positioned top-right, colored left border, close button.
- [ ] Step 3: Integrate with existing toast store (`web/src/lib/stores/toast.ts`).
- [ ] Step 4: Run `cd web && npm run check` to verify.

---

## Task 30: Admin Panel — Spinner Component

**Reference file:** Loading spinners throughout panel pages

**Visual Design:**
```
    ╭───────╮
   ╱    •    ╲      ← 38×38 SVG with rotating gradient stroke
  │  ↻       │
   ╲         ╱
    ╰───────╯

Sizes: sm: 16px, md: 24px, lg: 38px
Color: currentColor (inherits from parent)
Animation: 0.9s linear infinite rotation
```

**Files:** Create `web/src/lib/components/ui/SkyroomSpinner.svelte`

- [ ] Step 1: Create with props: `size` ('sm'|'md'|'lg'), `color`.
- [ ] Step 2: SVG with gradient stroke, rotating animation. Exact replica of Skyroom's spinner.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 31: Admin Panel — Info Row Component

**Reference file:** `اسکای_روم _ پنل مدیریت.html` "تاریخچه اتاق" section

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ ⏱  زمان مصرفی                        00:01:24    [↻]   │
├──────────────────────────────────────────────────────────┤
│ ⏱  مجموع زمان مصرفی                  00:01:24           │
├──────────────────────────────────────────────────────────┤
│ 🕐 ایجاد شده در              10:36:24 - 1405/03/26       │
├──────────────────────────────────────────────────────────┤
│ 🕐 بروز شده در                12:01:37 - 1405/03/26       │
└──────────────────────────────────────────────────────────┘
```
- Icon: 16px, `--dark-muted`
- Label: 13px, `--dark-text`, flex-grow
- Value: 13px, `--dark-text`, weight 500
- Action button (optional): 24×24px icon button
- Border-bottom: 1px `#e0e4eb` between rows
- Padding: 10px 0

**Files:** Create `web/src/lib/components/ui/SkyroomInfoRow.svelte`

- [ ] Step 1: Create with props: `icon` (SVG path), `label`, `value`, `actionLabel`, `onAction`.
- [ ] Step 2: Style: flex row, icon + label + value + optional action, border-bottom separator.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 32: Admin Panel — Page Header Component

**Reference file:** Page headers throughout panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ اتاق‌ها                                    [+ اتاق جدید] │
│ subtitle (optional)                                      │
└──────────────────────────────────────────────────────────┐
```
- Title: 18px, `--dark-text`, weight 600
- Subtitle: 13px, `--text-secondary`
- Actions: right-aligned buttons
- Margin-bottom: 20px

**Files:** Create `web/src/lib/components/ui/SkyroomPageHeader.svelte`

- [ ] Step 1: Create with props: `title`, `subtitle`, `onRefresh`. Slot: `actions`.
- [ ] Step 2: Style: flex row, title left, actions right.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 33: Admin Panel — Search Component

**Reference file:** Search inputs in panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│ 🔍  [جستجو...                    ]                       │
└──────────────────────────────────────────────────────────┘
```
- Height: 36px
- Border: 1px `#e0e4eb`, border-radius 8px
- Search icon: 16px, `--dark-muted`, left side
- Input: flex-grow, no border (inherits from container)
- Clear button: X icon, appears when input has value
- Focus: border-color `--primary`

**Files:** Create `web/src/lib/components/ui/SkyroomSearch.svelte`

- [ ] Step 1: Create with props: `placeholder`, `value`, `onSearch`, `debounce` (ms).
- [ ] Step 2: Style: bordered container with search icon, embedded input.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 34: Admin Panel — Dropdown Menu Component

**Reference file:** User menu, action menus in panel pages

**Visual Design:**
```
┌──────────────────────────┐
│ [👤] حساب کاربری          │
│ [⊞] پنل کاربر            │
│ [🚪] خروج                │
└──────────────────────────┘
```
- Min-width: 180px
- White bg, border 1px `#e0e4eb`, border-radius 8px, box-shadow
- Items: 13px text, 16px icon, padding 8px 16px
- Hover: bg `#f0f2f5`
- Separator: 1px `#e0e4eb` border
- Position: below trigger, right-aligned

**Files:** Create `web/src/lib/components/ui/SkyroomDropdown.svelte`

- [ ] Step 1: Create with props: `open`, `onClose`, `position`. Slots: `trigger`, `items`.
- [ ] Step 2: Style: white bg, border, shadow, hover states.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 35: Admin Panel — Pagination Component

**Reference file:** Table pagination in panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│  [‹]  [1]  [2]  [3]  ...  [10]  [›]                     │
│          ───                                              │
│     active: bg #23b9d7, white text                        │
└──────────────────────────────────────────────────────────┘
```
- Button size: 32×32px
- Border: 1px `#e0e4eb`, border-radius 4px
- Active: bg `--primary`, white text
- Inactive: white bg, `--dark-text`
- Disabled: opacity 0.4
- Gap: 4px

**Files:** Create `web/src/lib/components/ui/SkyroomPagination.svelte`

- [ ] Step 1: Create with props: `currentPage`, `totalPages`, `onPageChange`.
- [ ] Step 2: Style: numbered buttons with active state, prev/next arrows.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 36: Admin Panel — Empty State Component

**Reference file:** Empty table states in panel pages

**Visual Design:**
```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│                    [📭 icon]                              │
│                                                          │
│              هنوز موردی ثبت نشده است                      │
│                                                          │
│          [+ ایجاد مورد جدید]                              │
│                                                          │
└──────────────────────────────────────────────────────────┘
```
- Centered, max-width 400px
- Icon: 48px, `--dark-muted`
- Title: 16px, `--dark-text`, weight 500
- Description: 13px, `--text-secondary`
- Action button: optional, below text

**Files:** Create `web/src/lib/components/ui/SkyroomEmptyState.svelte`

- [ ] Step 1: Create with props: `icon`, `title`, `description`, `actionLabel`, `onAction`.
- [ ] Step 2: Style: centered layout, muted colors.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 37: Refactor Admin Pages to Skyroom Style

**Files:** All admin pages in `web/src/routes/(app)/admin/`

- [ ] Step 1: Refactor `admin/rooms/+page.svelte` — Use `SkyroomPageHeader`, `SkyroomSearch`, `SkyroomTable`, `SkyroomBadge`, `SkyroomButton`. Match Skyroom's rooms list layout.
- [ ] Step 2: Refactor `admin/channel/[id]/+page.svelte` — Use `SkyroomTabs`, `SkyroomInput`, `SkyroomSelect`, `SkyroomCheckbox`, `SkyroomCard`, `SkyroomInfoRow`. Match the exact form layout.
- [ ] Step 3: Refactor `admin/users/+page.svelte` — Table with user avatars, role badges, status indicators.
- [ ] Step 4: Refactor `admin/sessions/+page.svelte` — Session table with room/time/duration columns.
- [ ] Step 5: Refactor `admin/files/+page.svelte` — File list with name/size/date columns.
- [ ] Step 6: Refactor `admin/recordings/+page.svelte` — Recording table with download/delete actions.
- [ ] Step 7: Refactor `admin/settings/+page.svelte` — Tabbed form sections with inputs and checkboxes.
- [ ] Step 8: Refactor `admin/tickets/+page.svelte` — Ticket table with status/priority badges.
- [ ] Step 9: Refactor `admin/logs/+page.svelte` — Log table with level badges and date filter.
- [ ] Step 10: Refactor `admin/channels/+page.svelte` — Channel list table.
- [ ] Step 11: Refactor `admin/+page.svelte` — Dashboard with stat cards and activity list.
- [ ] Step 12: Run `cd web && npm run build` to verify all pages compile.

---

## Task 38: Refactor User-Facing Pages to Skyroom Style

**Files:** User pages in `web/src/routes/(app)/`

- [ ] Step 1: Refactor `dashboard/+page.svelte` — Stats cards row, recent activity, quick actions.
- [ ] Step 2: Refactor `classes/+page.svelte` — Class cards grid with instructor/schedule/participants.
- [ ] Step 3: Refactor `classes/[id]/+page.svelte` — Class detail with tabs (sessions, participants, materials).
- [ ] Step 4: Refactor `sessions/+page.svelte` — Session cards with room/time/duration/status.
- [ ] Step 5: Refactor `profile/+page.svelte` — User info card, form fields, tabbed sections.
- [ ] Step 6: Refactor `support/+page.svelte` — Ticket list with create form and status filters.
- [ ] Step 7: Refactor `files/+page.svelte` — File grid/list with upload area.
- [ ] Step 8: Refactor `chat/[id]/+page.svelte` — Chat interface (if applicable).
- [ ] Step 9: Run `cd web && npm run build` to verify all pages compile.

---

## Task 39: SVG Icon System

**Files:** Create `web/src/lib/assets/skyroom-icons.svg`, Create `web/src/lib/components/ui/SkyroomIcon.svelte`

- [ ] Step 1: Extract all unique SVG icon paths from the saved HTML files. Create an SVG sprite file with `<symbol>` elements for each icon: menu, group, chat, volume_up, volume_off, mic, mic_off, videocam, videocamoff, laptop, brush, slideshow, hand, settings, exit, delete, edit, reply, send, emoji, more_vert, info_outline, network_check, web, lock, aspect_ratio, file_download, block, clear, done, refresh, account_circle, access_time, cached, disconnected, person, power_settings_new, shape-specific icons, check, close, chevron-down, chevron-left, chevron-right, search, notifications, help, dashboard, book-open, video-camera, folder, credit-card, support, chart-bar, calendar, wifi, link, copy, external-link, upload, download, add, remove, warning, error, success.
- [ ] Step 2: Create `SkyroomIcon.svelte` component: `<script>let { name, size = 24, color = 'currentColor' } = $props();</script>` → `<svg width={size} height={size} fill={color}><use href="/lib/assets/skyroom-icons.svg#{name}" /></svg>`.
- [ ] Step 3: Run `cd web && npm run check` to verify.

---

## Task 40: Final Integration and Visual QA

**Files:** All modified files

- [ ] Step 1: Run `cd web && npm run build` to ensure the entire project compiles without errors.
- [ ] Step 2: Run `cd web && npm run dev` and systematically navigate through every page, comparing against the corresponding saved HTML reference.
- [ ] Step 3: Fix any visual discrepancies: spacing, colors, font sizes, border radii, shadows, hover states, active states.
- [ ] Step 4: Test responsive behavior at mobile widths (375px, 768px) and verify sidebar collapse, mobile menu overlay, and touch-friendly tap targets.
- [ ] Step 5: Test the classroom page specifically: verify all toolbar buttons, sidebar panels, chat functionality, and media area render correctly.
- [ ] Step 6: Run `cd web && npm run check` one final time to ensure type safety.
