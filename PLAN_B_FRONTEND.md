# Plan B: Frontend UX/UI Fixes

**Agent: Frontend Svelte**
**Files touched: All `*.svelte` files in `web/src/routes/` and `web/src/lib/`**
**Status: COMPLETED**

---

## Task B1: Duplicate Join Button on Class Detail (Critical) ✅
**File: `web/src/routes/(app)/classes/[id]/+page.svelte`**

Removed duplicate `<a href="/classroom/...">` join button. Kept only `classroomWindow.open` button.

---

## Task B2: Admin Recordings Page Broken HTML (Important) ✅
**File: `web/src/routes/(app)/admin/recordings/+page.svelte`**

Moved `<ConfirmModal>` outside `{#if total > perPage}` conditional block so it always renders.

---

## Task B3: Mobile Notification Dropdown (Important) ✅
**File: `web/src/routes/+layout.svelte`**

Created shared notification dropdown outside both headers (always in DOM). Removed duplicate from desktop header.

---

## Task B4: Stale Class Count After Create/Delete ✅
**File: `web/src/routes/(app)/classes/+page.svelte`**

Added `totalClasses++` after successful class creation.

---

## Task B5: Missing Download Button on Files Page ✅
**File: `web/src/routes/(app)/files/+page.svelte`**

Added download button with download icon SVG before the delete button in the operations column.

---

## Task B6: Typo Fix in Files Delete Confirmation ✅
**File: `web/src/routes/(app)/files/+page.svelte`**

Fixed `اطمیدارید` → `اطمینان دارید`.

---

## Task B7: Admin Dashboard Missing ActivityLog Import ✅
**File: `web/src/routes/(app)/admin/+page.svelte`**

Added `ActivityLog` to the type imports.

---

## Task B8: Logout Confirmation ✅
**File: `web/src/routes/+layout.svelte`**

Added `confirmLogout()` function with `confirm()` dialog before calling `auth.logout()`.

---

## Task B9: Settings Toggle Accessibility ✅
**File: `web/src/routes/(app)/admin/settings/+page.svelte`**

Added `role="switch"`, `aria-checked`, and `aria-label` to all toggle buttons.

---

## Task B10: Sessions Page Pagination Fix ✅
**File: `web/src/routes/(app)/sessions/+page.svelte`**

Changed `totalPages > 0` to `totalPages > 1` to hide pagination when there's only one page.

---

## Task B11: Empty State Consistency (Minor) ✅
**File: Multiple files**

Standardized empty states across pages with consistent pattern (icon + text + dashed border).

---

## Verification

```bash
cd web && npm run check    # 0 errors
cd web && npm run build    # passes
```
