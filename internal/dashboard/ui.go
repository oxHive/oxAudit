package dashboard

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>oxAudit Dashboard</title>
  <script src="https://cdn.jsdelivr.net/npm/marked@9/marked.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/dompurify@3/dist/purify.min.js"></script>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    :root {
      --bg: #0f1117;
      --sidebar-bg: #161b27;
      --header-bg: #1a2035;
      --border: #252d3d;
      --text: #e2e8f0;
      --text-muted: #7a8499;
      --text-dim: #4a5568;
      --accent: #6366f1;
      --accent-soft: rgba(99,102,241,0.15);
      --accent-hover: #818cf8;
      --code-bg: #1a2035;
    }
    html, body { height: 100%; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: var(--bg);
      color: var(--text);
      display: flex;
      height: 100vh;
      overflow: hidden;
      font-size: 14px;
    }
    /* Sidebar */
    #sidebar {
      width: 248px; flex-shrink: 0;
      background: var(--sidebar-bg);
      border-right: 1px solid var(--border);
      display: flex; flex-direction: column; overflow: hidden;
    }
    #sidebar-header {
      padding: 18px 16px 14px;
      border-bottom: 1px solid var(--border);
    }
    #sidebar-brand {
      display: flex; align-items: center; gap: 8px; margin-bottom: 6px;
    }
    #sidebar-brand span { font-size: 15px; font-weight: 700; letter-spacing: .03em; color: #f1f5f9; }
    #sidebar-tz { font-size: 11px; color: var(--text-muted); }
    #run-list { flex: 1; overflow-y: auto; padding: 6px 0; }
    .run-section-label {
      font-size: 10px; font-weight: 600;
      letter-spacing: .08em; text-transform: uppercase;
      color: var(--text-dim); padding: 8px 16px 4px;
    }
    .run-item {
      padding: 9px 16px; cursor: pointer;
      border-left: 3px solid transparent;
      transition: background .12s;
    }
    .run-item:hover { background: rgba(99,102,241,.07); }
    .run-item.active { border-left-color: var(--accent); background: var(--accent-soft); }
    .run-item .ri-date { font-size: 13px; font-weight: 600; color: var(--text); }
    .run-item .ri-time { font-size: 11px; color: var(--text-muted); margin-top: 1px; }
    .run-item .ri-tz  { font-size: 10px; color: var(--text-dim); }
    /* Main */
    #main { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; }
    #tab-bar {
      display: flex; align-items: center; gap: 2px;
      padding: 0 16px; min-height: 42px;
      background: var(--header-bg); border-bottom: 1px solid var(--border);
      overflow-x: auto; flex-shrink: 0;
    }
    #tab-bar::-webkit-scrollbar { height: 0; }
    .tab {
      padding: 8px 14px; font-size: 12.5px; font-weight: 500;
      cursor: pointer; border-radius: 6px 6px 0 0;
      color: var(--text-muted); white-space: nowrap;
      transition: color .12s, background .12s;
      user-select: none;
      border: 1px solid transparent; border-bottom: none; margin-bottom: -1px;
    }
    .tab:hover { color: var(--text); }
    .tab.active {
      background: var(--bg); color: var(--text);
      border-color: var(--border); border-bottom-color: var(--bg);
    }
    #content { flex: 1; overflow-y: auto; padding: 36px 48px; }
    /* State overlay */
    #state-overlay {
      display: flex; flex-direction: column;
      align-items: center; justify-content: center;
      height: 100%; gap: 14px;
      color: var(--text-muted); text-align: center;
    }
    #state-overlay h2 { font-size: 18px; font-weight: 600; color: var(--text); }
    #state-overlay p  { font-size: 13px; max-width: 380px; line-height: 1.6; }
    /* Markdown */
    #md { max-width: 900px; }
    #md h1 { font-size: 22px; font-weight: 700; color: #f1f5f9; margin-bottom: 20px; padding-bottom: 10px; border-bottom: 1px solid var(--border); }
    #md h2 { font-size: 17px; font-weight: 600; color: #e2e8f0; margin: 32px 0 12px; padding-bottom: 6px; border-bottom: 1px solid var(--border); }
    #md h3 { font-size: 14px; font-weight: 600; color: #cbd5e1; margin: 22px 0 8px; }
    #md h4 { font-size: 13px; font-weight: 600; color: #94a3b8; margin: 16px 0 6px; }
    #md p  { font-size: 14px; line-height: 1.75; color: #b8c5d6; margin-bottom: 12px; }
    #md ul, #md ol { font-size: 14px; line-height: 1.75; color: #b8c5d6; margin: 8px 0 12px 22px; }
    #md li { margin-bottom: 4px; }
    #md strong { color: #e2e8f0; font-weight: 600; }
    #md em { color: #94a3b8; }
    #md a { color: var(--accent-hover); text-decoration: none; }
    #md a:hover { text-decoration: underline; }
    #md blockquote {
      border-left: 3px solid var(--accent);
      padding: 10px 16px; margin: 14px 0;
      color: var(--text-muted); background: var(--accent-soft); border-radius: 0 6px 6px 0;
    }
    #md blockquote p { margin: 0; color: var(--text-muted); }
    #md hr { border: none; border-top: 1px solid var(--border); margin: 28px 0; }
    #md code {
      background: var(--code-bg); padding: 2px 6px; border-radius: 4px;
      font-family: 'JetBrains Mono','Fira Code','Cascadia Code',monospace;
      font-size: 12.5px; color: #a5b4fc;
    }
    #md pre {
      background: var(--code-bg); border: 1px solid var(--border);
      border-radius: 8px; padding: 16px 20px; overflow-x: auto; margin: 14px 0;
    }
    #md pre code { background: none; padding: 0; font-size: 13px; color: #e2e8f0; }
    #md table { width: 100%; border-collapse: collapse; font-size: 13px; margin: 18px 0; }
    #md th {
      background: var(--code-bg); padding: 9px 14px;
      text-align: left; font-weight: 600; color: var(--text); border: 1px solid var(--border);
    }
    #md td { padding: 8px 14px; border: 1px solid var(--border); color: #b8c5d6; vertical-align: top; }
    #md tr:nth-child(even) td { background: rgba(255,255,255,.02); }
    ::-webkit-scrollbar { width: 5px; height: 5px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: var(--border); border-radius: 3px; }
    ::-webkit-scrollbar-thumb:hover { background: #3a4055; }
  </style>
</head>
<body>

<div id="sidebar">
  <div id="sidebar-header">
    <div id="sidebar-brand">
      <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
        <rect width="20" height="20" rx="5" fill="#6366f1" opacity="0.9"/>
        <path d="M5 14l3-4 2 2 3-5 2 3" stroke="white" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <span>oxAudit</span>
    </div>
    <div id="sidebar-tz">Detecting timezone...</div>
  </div>
  <div id="run-list">
    <div class="run-section-label">Audit Runs</div>
  </div>
</div>

<div id="main">
  <div id="tab-bar"></div>
  <div id="content">
    <div id="state-overlay">
      <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
        <circle cx="24" cy="24" r="22" stroke="#252d3d" stroke-width="2"/>
        <path d="M16 32l4-6 4 4 5-8 3 4" stroke="#6366f1" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <h2>Loading audit runs...</h2>
      <p>Scanning the output directory for results.</p>
    </div>
    <div id="md" style="display:none"></div>
  </div>
</div>

<script>
(function () {
  'use strict';

  var tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
  document.getElementById('sidebar-tz').textContent = tz;

  var dateFmt = new Intl.DateTimeFormat(undefined, {
    timeZone: tz, weekday: 'short', month: 'short', day: 'numeric', year: 'numeric'
  });
  var timeFmt = new Intl.DateTimeFormat(undefined, {
    timeZone: tz, hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false
  });

  var runs = [];
  var activeRun = null;

  var tabBar  = document.getElementById('tab-bar');
  var runList = document.getElementById('run-list');
  var mdEl    = document.getElementById('md');
  var overlay = document.getElementById('state-overlay');

  var TAB_LABELS = {
    'ACTION-PLAN.md':         'Action Plan',
    'llm-digest.md':          'Overview',
    'findings.md':            'Findings',
    'remediation-backlog.md': 'Remediation',
    'findings.csv':           'CSV Export',
    'findings.jsonl':         'JSONL Export'
  };

  function tabLabel(file) {
    var base = file.split('/').pop();
    return TAB_LABELS[base] || base;
  }

  function renderSidebar() {
    document.querySelectorAll('.run-item').forEach(function (el) { el.remove(); });
    runs.forEach(function (run, i) {
      var d = new Date(run.executed_at);
      var el = document.createElement('div');
      el.className = 'run-item';

      var date = document.createElement('div');
      date.className = 'ri-date';
      date.textContent = dateFmt.format(d);

      var time = document.createElement('div');
      time.className = 'ri-time';
      time.textContent = timeFmt.format(d);

      var tzEl = document.createElement('div');
      tzEl.className = 'ri-tz';
      tzEl.textContent = tz;

      el.appendChild(date);
      el.appendChild(time);
      el.appendChild(tzEl);
      el.addEventListener('click', function () { selectRun(i); });
      runList.appendChild(el);
    });
  }

  function selectRun(idx) {
    activeRun = runs[idx];
    document.querySelectorAll('.run-item').forEach(function (el, i) {
      el.classList.toggle('active', i === idx);
    });
    renderTabs(activeRun.files || []);
    var first = (activeRun.files || []).find(function (f) { return f.endsWith('.md'); });
    if (first) { selectFile(first); }
  }

  function renderTabs(files) {
    tabBar.textContent = '';
    files.forEach(function (file) {
      var el = document.createElement('div');
      el.className = 'tab';
      el.textContent = tabLabel(file);
      el.dataset.file = file;
      el.addEventListener('click', function () { selectFile(file); });
      tabBar.appendChild(el);
    });
  }

  function selectFile(file) {
    document.querySelectorAll('.tab').forEach(function (el) {
      el.classList.toggle('active', el.dataset.file === file);
    });
    document.getElementById('content').scrollTop = 0;

    fetch('/api/content/' + activeRun.folder_name + '/' + file)
      .then(function (res) {
        if (!res.ok) { throw new Error('HTTP ' + res.status); }
        return res.text();
      })
      .then(function (text) { showContent(file, text); })
      .catch(function (e) { showMsg('Failed to load ' + file + ': ' + e.message); });
  }

  function showContent(file, text) {
    overlay.style.display = 'none';
    mdEl.style.display = '';
    if (file.endsWith('.md')) {
      // Parse markdown then sanitize before setting as HTML
      var raw = marked.parse(text);
      var clean = DOMPurify.sanitize(raw);
      mdEl.innerHTML = clean;
    } else {
      // Plain text: use textContent to avoid any HTML interpretation
      var pre = document.createElement('pre');
      pre.style.cssText = 'color:#e2e8f0;font-size:13px;line-height:1.6';
      pre.textContent = text;
      mdEl.textContent = '';
      mdEl.appendChild(pre);
    }
  }

  function showMsg(msg) {
    overlay.style.display = 'none';
    mdEl.style.display = '';
    var p = document.createElement('p');
    p.style.cssText = 'color:#f87171;font-size:14px';
    p.textContent = msg;
    mdEl.textContent = '';
    mdEl.appendChild(p);
  }

  function boot() {
    fetch('/api/runs')
      .then(function (res) { return res.json(); })
      .then(function (data) {
        runs = data || [];
        if (runs.length === 0) {
          overlay.querySelector('h2').textContent = 'No audit runs found';
          overlay.querySelector('p').textContent = 'Run "oxaudit all" to generate your first report.';
          return;
        }
        renderSidebar();
        selectRun(0);
      })
      .catch(function (e) {
        overlay.querySelector('h2').textContent = 'Could not load runs';
        overlay.querySelector('p').textContent = e.message;
      });
  }

  boot();
})();
</script>
</body>
</html>`
