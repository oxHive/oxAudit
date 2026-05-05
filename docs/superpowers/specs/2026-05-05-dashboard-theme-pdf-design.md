# Dashboard: Light/Dark Mode Toggle + PDF Export

**Date:** 2026-05-05
**Status:** Approved

---

## Overview

Add two features to the oxAudit web dashboard (`internal/dashboard/ui.go`):

1. **Light/dark mode toggle** — switch between the existing dark theme and a new crisp-white light theme. Persisted across sessions via `localStorage`.
2. **PDF export** — export the current page content (no sidebar) as a PDF using the browser's native print dialog. The exported PDF reflects whichever theme is currently active.

Both controls live on the right side of the existing `#tab-bar`.

---

## Scope

**Single file change:** `internal/dashboard/ui.go` (the embedded HTML string).
**No new dependencies.** The existing CDN scripts (marked, DOMPurify) are unchanged.

---

## Theme Toggle

### Control placement
A pill-style toggle button is added to the right end of `#tab-bar`, to the left of the PDF export button. It shows 🌙 when dark mode is active and ☀️ when light mode is active.

### Implementation
- Toggle adds/removes the `.light` class on `<html>`.
- CSS variables under `html.light` override the dark defaults defined on `:root`.
- On page load, the saved preference is read from `localStorage` (`oxaudit-theme`). If `light` is stored, the `.light` class is applied immediately (before render) to avoid a flash of dark mode.
- Clicking the toggle updates the class and writes the new value to `localStorage`.

### Light mode palette (crisp white)
| Variable         | Dark value  | Light value |
|-----------------|-------------|-------------|
| `--bg`          | `#0f1117`   | `#ffffff`   |
| `--sidebar-bg`  | `#161b27`   | `#f9fafb`   |
| `--header-bg`   | `#1a2035`   | `#f3f4f6`   |
| `--border`      | `#252d3d`   | `#e5e7eb`   |
| `--text`        | `#e2e8f0`   | `#111827`   |
| `--text-muted`  | `#7a8499`   | `#6b7280`   |
| `--text-dim`    | `#4a5568`   | `#9ca3af`   |
| `--accent`      | `#6366f1`   | `#6366f1`   |
| `--accent-soft` | `rgba(99,102,241,0.15)` | `rgba(99,102,241,0.10)` |
| `--accent-hover`| `#818cf8`   | `#4f46e5`   |
| `--code-bg`     | `#1a2035`   | `#f3f4f6`   |

Additional light mode overrides for markdown content:
- `#md h1`, `h2`, `h3`, `h4` — dark slate colors (`#111827`, `#1e293b`, `#374151`, `#4b5563`)
- `#md p`, `ul`, `ol`, `li` — `#374151`
- `#md strong` — `#111827`
- `#md em` — `#6b7280`
- `#md a` — `#4f46e5`
- `#md code` — background `#f3f4f6`, color `#4f46e5`
- `#md pre` — background `#f3f4f6`, border `#e5e7eb`
- `#md pre code` — `#1e293b`
- `#md th` — background `#f3f4f6`, color `#111827`
- `#md td` — `#374151`
- `#md tr:nth-child(even) td` — `rgba(0,0,0,.02)`
- `#md blockquote` — background `rgba(99,102,241,.07)`, text `#4b5563`
- Scrollbar thumb — `#d1d5db`

Tab bar active tab in light mode: background `#ffffff`, border `#e5e7eb`, border-bottom-color `#ffffff`.

---

## PDF Export

### Control placement
A button labelled "↓ Export PDF" sits at the right end of `#tab-bar`, to the right of the theme toggle.

### Implementation
- Clicking calls `window.print()`.
- `@media print` CSS rules handle layout:
  - `#sidebar` — `display: none`
  - `#tab-bar` — `display: none`
  - `#main` — `width: 100%; height: auto; overflow: visible`
  - `#content` — `overflow: visible; padding: 0`
  - `body` — `overflow: visible; height: auto; display: block`
  - `html, body` — `height: auto`
- No forced color override — the current CSS variables (dark or light) carry through into the print output naturally, so the PDF matches the active theme.
- No additional print-specific font or margin overrides needed; browser defaults handle page margins.

---

## Controls HTML & JS

### Tab bar additions (appended to `#tab-bar` in `renderTabs`)
The theme toggle and PDF button are injected once into the DOM on `boot()`, not recreated on each `renderTabs` call. This keeps them stable and avoids duplicating event listeners.

```html
<!-- injected once on boot, appended to #tab-bar -->
<div id="theme-toggle" class="theme-btn" onclick="toggleTheme()" title="Toggle light/dark mode">
  <span id="theme-icon">🌙</span>
</div>
<button id="pdf-btn" class="pdf-btn" onclick="window.print()">↓ Export PDF</button>
```

### Toggle logic
```js
function toggleTheme() {
  var isLight = document.documentElement.classList.toggle('light');
  document.getElementById('theme-icon').textContent = isLight ? '☀️' : '🌙';
  localStorage.setItem('oxaudit-theme', isLight ? 'light' : 'dark');
}

// On boot, before rendering:
(function applyStoredTheme() {
  if (localStorage.getItem('oxaudit-theme') === 'light') {
    document.documentElement.classList.add('light');
    // icon updated after DOM ready
  }
})();
```

---

## Tab bar control styles

```css
.theme-btn {
  width: 30px; height: 30px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 6px; cursor: pointer;
  background: var(--header-bg);
  border: 1px solid var(--border);
  font-size: 14px;
  transition: background .12s;
  flex-shrink: 0;
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
}
.pdf-btn:hover { opacity: .85; }
```

---

## Testing checklist

- [ ] Dark mode is the default on first load (no localStorage entry)
- [ ] Clicking the toggle switches to light mode; icon changes to ☀️
- [ ] Refreshing the page restores the saved theme (no flash)
- [ ] Clicking toggle again restores dark mode; icon changes to 🌙
- [ ] All markdown content (headings, tables, code blocks, blockquotes) is readable in light mode
- [ ] Clicking "↓ Export PDF" opens the browser print dialog
- [ ] Print preview shows content only (no sidebar, no tab bar)
- [ ] Print preview in dark mode has dark background
- [ ] Print preview in light mode has white background
- [ ] PDF export works on the Action Plan, Overview, Findings, and Remediation tabs
