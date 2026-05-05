# Dashboard Theme Toggle + PDF Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a light/dark mode toggle (persisted via localStorage) and a PDF export button (browser print, content-only) to the oxAudit web dashboard tab bar.

**Architecture:** All changes are in the single embedded HTML string in `internal/dashboard/ui.go`. No new Go code, no new dependencies. CSS variables drive theming — a `.light` class on `<html>` overrides the dark defaults. `@media print` CSS hides the sidebar and tab bar for PDF output.

**Tech Stack:** Vanilla JS, CSS custom properties, `window.print()`, `localStorage`. Go (only for rebuilding the binary).

---

## File map

| File | Change |
|------|--------|
| `internal/dashboard/ui.go` | All changes — CSS additions, HTML additions, JS additions within the embedded `dashboardHTML` string |

No other files are touched.

---

### Task 1: Add `@media print` CSS

**Files:**
- Modify: `internal/dashboard/ui.go` — inside the `<style>` block, before the closing `</style>` tag

- [ ] **Step 1: Locate the insertion point**

Open `internal/dashboard/ui.go`. Find the line that reads:
```css
    ::-webkit-scrollbar-thumb:hover { background: #3a4055; }
```
This is the last CSS rule before `</style>`. You will insert new CSS immediately after this line.

- [ ] **Step 2: Add the print media query**

Insert the following block after the scrollbar hover rule and before `</style>`:

```css
    /* ── Print / PDF export ── */
    @media print {
      html, body { height: auto; overflow: visible; display: block; }
      #sidebar  { display: none !important; }
      #tab-bar  { display: none !important; }
      #main     { width: 100%; height: auto; overflow: visible; }
      #content  { overflow: visible; padding: 24px 48px; }
    }
```

- [ ] **Step 3: Build and verify the file compiles**

```bash
go build ./...
```

Expected: no output (success). If there's a syntax error, the Go string literal likely has an unescaped backtick — check for that.

- [ ] **Step 4: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add @media print CSS for PDF export"
```

---

### Task 2: Add `.theme-btn` and `.pdf-btn` CSS styles

**Files:**
- Modify: `internal/dashboard/ui.go` — inside the `<style>` block, after the `@media print` block added in Task 1

- [ ] **Step 1: Insert the control styles**

Immediately after the `@media print` block (still before `</style>`), add:

```css
    /* ── Tab bar controls ── */
    #tab-bar-controls {
      display: flex; align-items: center; gap: 8px;
      margin-left: auto; flex-shrink: 0; padding-right: 4px;
    }
    .theme-btn {
      width: 30px; height: 30px;
      display: flex; align-items: center; justify-content: center;
      border-radius: 6px; cursor: pointer;
      background: var(--header-bg);
      border: 1px solid var(--border);
      font-size: 14px;
      transition: background .12s;
      flex-shrink: 0; user-select: none;
    }
    .theme-btn:hover { background: var(--accent-soft); }
    .pdf-btn {
      padding: 5px 12px;
      background: var(--accent);
      color: #fff;
      border: none;
      border-radius: 6px;
      font-size: 12px; font-weight: 600;
      cursor: pointer;
      white-space: nowrap;
      flex-shrink: 0;
      transition: opacity .12s;
      line-height: 1;
    }
    .pdf-btn:hover { opacity: .85; }
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add CSS for theme toggle and PDF export button"
```

---

### Task 3: Add `html.light` CSS variable overrides

**Files:**
- Modify: `internal/dashboard/ui.go` — inside the `<style>` block, after the `#tab-bar-controls` / button styles added in Task 2

- [ ] **Step 1: Insert the light mode variable overrides**

Immediately after the `.pdf-btn:hover` rule (still before `</style>`), add:

```css
    /* ── Light mode ── */
    html.light {
      --bg:           #ffffff;
      --sidebar-bg:   #f9fafb;
      --header-bg:    #f3f4f6;
      --border:       #e5e7eb;
      --text:         #111827;
      --text-muted:   #6b7280;
      --text-dim:     #9ca3af;
      --accent:       #6366f1;
      --accent-soft:  rgba(99,102,241,0.10);
      --accent-hover: #4f46e5;
      --code-bg:      #f3f4f6;
    }
    html.light ::-webkit-scrollbar-thumb { background: #d1d5db; }
    html.light ::-webkit-scrollbar-thumb:hover { background: #9ca3af; }
    html.light .tab.active {
      background: #ffffff;
      border-color: #e5e7eb;
      border-bottom-color: #ffffff;
    }
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add html.light CSS variable overrides"
```

---

### Task 4: Add light mode markdown content overrides

**Files:**
- Modify: `internal/dashboard/ui.go` — inside the `<style>` block, after the `html.light .tab.active` rule added in Task 3

- [ ] **Step 1: Insert the markdown overrides**

Immediately after the `html.light .tab.active` rule (still before `</style>`), add:

```css
    html.light #md h1 { color: #111827; border-bottom-color: #e5e7eb; }
    html.light #md h2 { color: #1e293b; border-bottom-color: #e5e7eb; }
    html.light #md h3 { color: #374151; }
    html.light #md h4 { color: #4b5563; }
    html.light #md p  { color: #374151; }
    html.light #md ul, html.light #md ol { color: #374151; }
    html.light #md strong { color: #111827; }
    html.light #md em    { color: #6b7280; }
    html.light #md a     { color: #4f46e5; }
    html.light #md blockquote {
      background: rgba(99,102,241,.07);
      border-left-color: #6366f1;
      color: #4b5563;
    }
    html.light #md blockquote p { color: #4b5563; }
    html.light #md hr { border-top-color: #e5e7eb; }
    html.light #md code { background: #f3f4f6; color: #4f46e5; }
    html.light #md pre  { background: #f3f4f6; border-color: #e5e7eb; }
    html.light #md pre code { background: none; color: #1e293b; }
    html.light #md th { background: #f3f4f6; color: #111827; border-color: #e5e7eb; }
    html.light #md td { color: #374151; border-color: #e5e7eb; }
    html.light #md tr:nth-child(even) td { background: rgba(0,0,0,.02); }
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add light mode markdown content overrides"
```

---

### Task 5: Add controls HTML to the tab bar

**Files:**
- Modify: `internal/dashboard/ui.go` — the `#tab-bar` div in the HTML body

- [ ] **Step 1: Locate the tab bar div**

Find this line in `ui.go`:
```html
<div id="tab-bar"></div>
```

- [ ] **Step 2: Add the controls wrapper inside the tab bar**

Replace that line with:

```html
<div id="tab-bar">
  <div id="tab-bar-controls">
    <div id="theme-toggle" class="theme-btn" title="Toggle light/dark mode">
      <span id="theme-icon">🌙</span>
    </div>
    <button class="pdf-btn" onclick="window.print()">↓ Export PDF</button>
  </div>
</div>
```

Note: the controls wrapper uses `margin-left: auto` (set in `#tab-bar-controls` CSS) to push itself to the right. Tab elements are prepended by `renderTabs()` and will naturally appear to the left of the controls.

- [ ] **Step 3: Update `renderTabs` to prepend tabs instead of clearing the whole bar**

Find the `renderTabs` function in the `<script>` block:

```js
  function renderTabs(files) {
    tabBar.textContent = '';
    files.forEach(function (file) {
```

Replace it with:

```js
  function renderTabs(files) {
    document.querySelectorAll('.tab').forEach(function (el) { el.remove(); });
    var controls = document.getElementById('tab-bar-controls');
    files.forEach(function (file) {
```

And find the line inside `renderTabs` that appends each tab:
```js
      tabBar.appendChild(el);
```

Replace it with:
```js
      tabBar.insertBefore(el, controls);
```

This ensures tabs are always inserted before the controls wrapper, so the toggle and PDF button stay right-aligned even as tabs change.

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add theme toggle and PDF export button to tab bar"
```

---

### Task 6: Add theme toggle JS and stored theme application

**Files:**
- Modify: `internal/dashboard/ui.go` — the `<script>` block

- [ ] **Step 1: Apply stored theme before first render**

Find the opening of the IIFE in the script block:
```js
(function () {
  'use strict';

  var tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
```

Insert the stored-theme application immediately after `'use strict';` and before the `var tz` line:

```js
  // Apply stored theme immediately to avoid flash of wrong mode.
  if (localStorage.getItem('oxaudit-theme') === 'light') {
    document.documentElement.classList.add('light');
  }
```

- [ ] **Step 2: Add `toggleTheme` function**

Find the `boot` function declaration:
```js
  function boot() {
```

Insert the `toggleTheme` function immediately before it:

```js
  function toggleTheme() {
    var isLight = document.documentElement.classList.toggle('light');
    document.getElementById('theme-icon').textContent = isLight ? '☀️' : '🌙';
    localStorage.setItem('oxaudit-theme', isLight ? 'light' : 'dark');
  }

```

- [ ] **Step 3: Wire the toggle button's click handler and sync the icon on boot**

Find the `boot` function body. Locate the opening line of the fetch call:
```js
    fetch('/api/runs')
```

Insert these two lines immediately before that fetch call, at the start of `boot()`:

```js
    document.getElementById('theme-toggle').addEventListener('click', toggleTheme);
    if (document.documentElement.classList.contains('light')) {
      document.getElementById('theme-icon').textContent = '☀️';
    }
```

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/ui.go
git commit -m "feat(dashboard): add theme toggle JS and localStorage persistence"
```

---

### Task 7: Manual verification

**Files:**
- None modified — this is a verification-only task.

- [ ] **Step 1: Run the dashboard**

```bash
./oxaudit dashboard
```

Expected output:
```
[oxaudit] Dashboard: http://localhost:<port>
```

The browser should open automatically.

- [ ] **Step 2: Verify dark mode default**

On first load (clear localStorage if needed: DevTools → Application → Local Storage → delete `oxaudit-theme`), confirm:
- Background is dark (`#0f1117`)
- Theme icon shows 🌙
- Tab bar controls are visible on the right

- [ ] **Step 3: Toggle to light mode**

Click the 🌙 button. Confirm:
- Background switches to white
- Sidebar switches to `#f9fafb`
- Icon changes to ☀️
- Markdown content (open any `.md` tab) is readable — dark text on white

- [ ] **Step 4: Verify persistence**

Refresh the page. Confirm light mode is restored (no flash of dark).

- [ ] **Step 5: Toggle back to dark**

Click ☀️. Confirm dark mode restored. Refresh — dark mode persists.

- [ ] **Step 6: Test PDF export in light mode**

Switch to light mode. Open a tab with content (e.g. Action Plan). Click "↓ Export PDF". Confirm:
- Browser print dialog opens
- Print preview shows **no sidebar** and **no tab bar**
- Content spans full width
- Background is white (light mode)

- [ ] **Step 7: Test PDF export in dark mode**

Switch to dark mode. Click "↓ Export PDF". Confirm:
- Print preview shows no sidebar, no tab bar
- Background is dark

- [ ] **Step 8: Final commit (if any fixes were needed)**

If you made any small fixes during verification, commit them:

```bash
git add internal/dashboard/ui.go
git commit -m "fix(dashboard): address verification findings for theme/PDF"
```

If no fixes were needed, skip this step.
