# Dashboard Architecture

A breakdown of every tool and technology used to build the oxAudit web dashboard.

---

## Overview

The dashboard is a **zero-dependency, single-binary web app** — no Node.js, no build step, no bundler. The Go binary serves a single HTML page that is compiled directly into the binary as a string constant (`dashboardHTML` in `internal/dashboard/ui.go`). All frontend libraries are loaded from CDNs at runtime.

---

## Backend

### Go standard library — `net/http`

The entire HTTP server is built on Go's standard library with no third-party web framework.

| Component | Role |
|-----------|------|
| `net.Listen` | Opens the TCP port (default `7842`) |
| `http.NewServeMux` | Routes requests to three handlers |
| `http.Serve` | Blocks and serves requests |
| `http.ServeFile` | Streams audit output files to the browser |
| `os.ReadDir` | Scans the output directory for dated run folders |

**Three routes:**

| Route | Handler | Purpose |
|-------|---------|---------|
| `GET /api/runs` | `handleRuns` | Returns JSON list of all audit runs |
| `GET /api/content/{folder}/{file}` | `handleContent` | Serves raw file content from disk |
| `GET /` | `handleUI` | Serves the embedded HTML page |

**Source:** `internal/dashboard/server.go`

### Path traversal protection

`handleContent` explicitly rejects any path segment containing `..` before passing it to `http.ServeFile`, preventing directory traversal attacks.

### Auto browser launch

On `Serve()`, the binary calls the OS-native open command (`open` on macOS, `xdg-open` on Linux, `cmd /c start` on Windows) to open the dashboard in the default browser automatically.

---

## Frontend

### HTML + CSS (vanilla, no framework)

The entire UI is written in plain HTML and CSS embedded in a single Go string constant. No React, Vue, or any other component framework is used.

**Layout:** CSS Flexbox — sidebar + main panel side by side.

**Design system:** CSS custom properties (`--bg`, `--accent`, `--border`, etc.) define the entire color palette. Swapping to light mode is a single class toggle (`html.light`) that overrides all variables.

**Theming:** Dark mode by default. Light mode is toggled by adding/removing the `light` class on `<html>` and persisted to `localStorage` so it survives page reloads.

### JavaScript (vanilla ES5)

All interactivity is written in ES5 (`var`, no arrow functions, no modules) wrapped in an IIFE for scope isolation. No build tool, no TypeScript, no transpilation.

**Key patterns used:**

| Pattern | Where |
|---------|-------|
| `fetch()` | Loads run list and file content from the Go API |
| `Intl.DateTimeFormat` | Formats timestamps in the user's local timezone |
| `localStorage` | Persists the dark/light theme preference |
| `document.createElement` | Builds sidebar run items and tab buttons dynamically |
| DOM event delegation | `click` listeners on each run item and tab |

---

## CDN Libraries (loaded at runtime)

All three libraries are loaded from `cdn.jsdelivr.net` via `<script>` tags. They have no build-time presence in the repo.

### marked.js `v9`

**Purpose:** Converts Markdown text (`.md` audit files) to HTML for rendering in the browser.

**Usage:** `marked.parse(text)` is called on the raw file content returned by `/api/content/`. All four audit output formats — `ACTION-PLAN.md`, `llm-digest.md`, `findings.md`, `remediation-backlog.md` — are Markdown files rendered through marked.

**CDN:** `https://cdn.jsdelivr.net/npm/marked@9/marked.min.js`

### DOMPurify `v3`

**Purpose:** Sanitizes the HTML output of marked.js before injecting it into the DOM, preventing XSS attacks from malicious content that could appear in audit findings.

**Usage:** `DOMPurify.sanitize(raw)` wraps every `marked.parse()` call. The cleaned HTML is then set via `innerHTML`.

**CDN:** `https://cdn.jsdelivr.net/npm/dompurify@3/dist/purify.min.js`

### html2pdf.js `v0.10.1`

**Purpose:** Exports the currently displayed audit report as a PDF, preserving the dark or light theme background.

**Usage:** The "↓ Export PDF" button triggers `html2pdf().set({...}).from(el).save()`, targeting the `#md` content div. It uses html2canvas internally to rasterize the content and jsPDF to produce the PDF file.

**CDN:** `https://cdn.jsdelivr.net/npm/html2pdf.js@0.10.1/dist/html2pdf.bundle.min.js`

---

## Data Flow

```
oxaudit all
    └── writes timestamped folders to output dir (e.g. ~/.config/oxaudit/)
            ├── 2026-05-04T041657Z/
            │   ├── ACTION-PLAN.md
            │   └── exports/
            │       ├── findings.md
            │       ├── findings.csv
            │       └── ...
            └── ...

oxaudit dashboard
    └── Go HTTP server on :7842
            ├── GET /api/runs        → scans output dir, returns JSON
            ├── GET /api/content/... → streams file from disk
            └── GET /               → serves embedded HTML

Browser
    └── fetch /api/runs → render sidebar
    └── click run       → fetch /api/content → marked.parse → DOMPurify.sanitize → innerHTML
    └── click PDF       → html2pdf → download
```

---

## Why This Architecture

| Decision | Reason |
|----------|--------|
| No frontend build step | Single binary distribution — users run `oxaudit dashboard`, not `npm install` |
| Embedded HTML string | Keeps the binary self-contained; no static file directory to ship alongside the binary |
| ES5 JavaScript | Avoids any transpilation requirement; works in all modern browsers without a bundler |
| CDN libraries | Offloads library distribution without adding to binary size; marked, DOMPurify, and html2pdf are too large to embed comfortably |
| No CSS framework | Tailwind/Bootstrap would require a build step; hand-written CSS variables give full control with ~250 lines |
| `net/http` only | No Gin/Echo/Fiber dependency; the dashboard is simple enough that the stdlib router is sufficient |
