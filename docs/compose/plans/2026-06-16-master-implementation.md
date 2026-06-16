# IRoom — Master Implementation Plan

> **For agentic workers:** Use compose:subagent or compose:execute to implement this plan task-by-task.

**Goal:** Complete the Skyroom-parity implementation of IRoom. All foundation code is written. Remaining work: finish incomplete features, apply Persian numbers + Jalali dates to ALL pages, apply Skyroom visual design tokens, add missing backend endpoints, and verify everything builds.

**Current State:** 
- ✅ Persian utilities (persian.ts)
- ✅ Toast system (toast.ts + Toast.svelte)
- ✅ ClassroomWindow + ClassroomBridge
- ✅ Classroom popup route (3-column layout, dark theme, controls)
- ✅ Classroom launcher page (with inline fallback)
- ✅ ConfirmModal component
- ✅ SettingsPopup component
- ✅ Admin dashboard (stats, live rooms, system health, activity feed)
- ✅ Admin users page (with pagination, filters, modals)
- ✅ Admin rooms page (cards, search, filter, create modal)
- ✅ Admin sessions page (with pagination, filters)
- ✅ Admin tickets page (with filters, detail modal, reply)
- ✅ Admin recordings page
- ✅ Admin logs page
- ✅ Admin settings page
- ✅ Profile page
- ✅ Forgot password page
- ✅ Whiteboard (responsive, laser pointer)
- ✅ JalaliDatePicker (basic)

**Remaining Work:**
- 🔄 JalaliDatePicker needs proper Jalali calendar (currently just shows Gregorian with Persian digits)
- 🔄 Some pages still use Western numbers instead of `toPersianNum()`
- 🔄 Some pages use `toLocaleDateString('fa-IR')` instead of proper Jalali
- 🔄 Backend health endpoint needs enhancement
- 🔄 Session logs page needs improvement
- 🔄 Files page needs pagination
- 🔄 Classes list page needs pagination
- 🔄 Sessions list page needs pagination
- 🔄 Apply consistent Skyroom design tokens across all pages
- 🔄 Build verification

---

## Phase 1: Fix JalaliDatePicker

### Task 1: Proper Jalali Date Picker

**Files:**
- Modify: `web/src/lib/components/JalaliDatePicker.svelte`

The current implementation just wraps `<input type="date">` with a Persian label. It needs to be a proper Jalali calendar picker.

- [ ] **Step 1: Implement proper Jalali date picker**

Replace the current implementation with a dropdown-style Jalali calendar:
- Month names: فروردین، اردیبهشت، خرداد، تیر، مرداد، شهریور، مهر، آبان، آذر، دی، بهمن، اسفند
- Week starts on Saturday (شنبه)
- Day names: ش، ی، د، س، چ، پ، ج
- Navigation: previous/next month, year selector
- Output: ISO date string (YYYY-MM-DD) for form submission
- Display: Persian date format (۱۴۰۳/۰۳/۲۵)

Use the `date-fns-jalali` library (already in package.json) for date conversion.

- [ ] **Step 2: Verify build**

Run: `cd web && npm run build`
Expected: SUCCESS

---

## Phase 2: Backend Health Endpoint

### Task 2: Enhanced Health Endpoint

**Files:**
- Modify: `cmd/server/main.go` or `internal/handlers/handlers.go`

- [ ] **Step 1: Add health endpoint with system info**

The health endpoint should return:
```json
{
  "status": "ok",
  "uptime": "2h 15m",
  "db_size": "125 MB",
  "livekit_status": "connected",
  "active_rooms": 3,
  "total_users": 123,
  "total_sessions": 456,
  "total_classes": 12
}
```

- [ ] **Step 2: Verify build**

Run: `cd /home/kazem/StudioProjects/iroom && go build -o server ./cmd/server`
Expected: SUCCESS

---

## Phase 3: Apply Persian Numbers to ALL Pages

### Task 3: Dashboard Page

**Files:**
- Modify: `web/src/routes/(app)/dashboard/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all number displays with toPersianNum()**
- [ ] **Step 3: Replace date formatting with toPersianDate()**
- [ ] **Step 4: Verify build**

### Task 4: Classes List Page

**Files:**
- Modify: `web/src/routes/(app)/classes/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Add pagination (page navigation, per-page selector)**
- [ ] **Step 3: Replace all numbers with toPersianNum()**
- [ ] **Step 4: Replace dates with toPersianDate()**
- [ ] **Step 5: Verify build**

### Task 5: Class Detail Page

**Files:**
- Modify: `web/src/routes/(app)/classes/[id]/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Replace dates with toPersianDate()**
- [ ] **Step 4: Use JalaliDatePicker for session date input**
- [ ] **Step 5: Verify build**

### Task 6: Sessions List Page

**Files:**
- Modify: `web/src/routes/(app)/sessions/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Add pagination**
- [ ] **Step 3: Replace all numbers with toPersianNum()**
- [ ] **Step 4: Replace dates with toPersianDate()**
- [ ] **Step 5: Verify build**

### Task 7: Files Page

**Files:**
- Modify: `web/src/routes/(app)/files/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Add pagination**
- [ ] **Step 3: Replace all numbers with toPersianNum()**
- [ ] **Step 4: Replace dates with toPersianDate()**
- [ ] **Step 5: Add drag-and-drop upload**
- [ ] **Step 6: Add upload progress indicator**
- [ ] **Step 7: Verify build**

### Task 8: Support Pages

**Files:**
- Modify: `web/src/routes/(app)/support/+page.svelte`
- Modify: `web/src/routes/(app)/support/[id]/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Replace dates with toPersianDateTime()**
- [ ] **Step 4: Verify build**

### Task 9: Profile Page

**Files:**
- Modify: `web/src/routes/(app)/profile/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Verify build**

### Task 10: Auth Pages

**Files:**
- Modify: `web/src/routes/auth/+page.svelte`
- Modify: `web/src/routes/auth/forgot-password/+page.svelte`

- [ ] **Step 1: Import toPersianNum where needed**
- [ ] **Step 2: Verify build**

### Task 11: Session Logs Page

**Files:**
- Modify: `web/src/routes/(app)/sessions/[id]/logs/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Replace dates with toPersianDateTime()**
- [ ] **Step 4: Add visual timeline for participant join/leave**
- [ ] **Step 5: Add CSV export button**
- [ ] **Step 6: Verify build**

### Task 12: Recordings Pages

**Files:**
- Modify: `web/src/routes/(app)/recordings/[id]/+page.svelte`
- Modify: `web/src/routes/(app)/admin/recordings/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Replace dates with toPersianDate()**
- [ ] **Step 4: Verify build**

### Task 13: Admin Logs Page

**Files:**
- Modify: `web/src/routes/(app)/admin/logs/+page.svelte`

- [ ] **Step 1: Import toPersianNum**
- [ ] **Step 2: Replace all numbers with toPersianNum()**
- [ ] **Step 3: Replace dates with toPersianDateTime()**
- [ ] **Step 4: Add CSV export**
- [ ] **Step 5: Verify build**

---

## Phase 4: Skyroom Visual Design Polish

### Task 14: Update app.css with Skyroom Design Tokens

**Files:**
- Modify: `web/src/app.css`

- [ ] **Step 1: Add Skyroom color variables**
- [ ] **Step 2: Add Skyroom scrollbar styles**
- [ ] **Step 3: Add Skyroom animation keyframes**
- [ ] **Step 4: Update component classes to match Skyroom spacing**
- [ ] **Step 5: Verify build**

### Task 15: Update Main Layout Sidebar

**Files:**
- Modify: `web/src/routes/+layout.svelte`

- [ ] **Step 1: Apply Skyroom sidebar colors (#1a1a2e bg, #2a2a4a borders)**
- [ ] **Step 2: Apply Skyroom active link style**
- [ ] **Step 3: Apply Skyroom user info section**
- [ ] **Step 4: Verify build**

---

## Phase 5: Final Build & Verification

### Task 16: Full Build Verification

- [ ] **Step 1: Frontend build**

Run: `cd web && npm run build`
Expected: SUCCESS with no errors

- [ ] **Step 2: Backend build**

Run: `cd /home/kazem/StudioProjects/iroom && go build -o server ./cmd/server`
Expected: SUCCESS with no errors

- [ ] **Step 3: Run existing tests**

Run: `cd /home/kazem/StudioProjects/iroom && go test ./internal/handlers/ -v`
Expected: All tests pass

- [ ] **Step 4: Docker build**

Run: `cd /home/kazem/StudioProjects/iroom && docker-compose build`
Expected: SUCCESS

---

## Execution Order

**Parallel Group A (can run simultaneously):**
- Task 1 (JalaliDatePicker)
- Task 2 (Health endpoint)
- Task 3 (Dashboard)
- Task 4 (Classes list)
- Task 5 (Class detail)

**Parallel Group B (can run simultaneously, after A):**
- Task 6 (Sessions list)
- Task 7 (Files)
- Task 8 (Support)
- Task 9 (Profile)
- Task 10 (Auth)

**Parallel Group C (can run simultaneously, after B):**
- Task 11 (Session logs)
- Task 12 (Recordings)
- Task 13 (Admin logs)
- Task 14 (app.css)
- Task 15 (Main layout)

**Final:**
- Task 16 (Build verification)
