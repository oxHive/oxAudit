# MCP Integration Design

**Date:** 2026-05-27
**Status:** Approved

---

## Goal

Expose oxaudit's audit data through an MCP server so Claude can query findings, costs, and resources conversationally — and trigger the full audit pipeline on behalf of the user.

---

## Architecture

`oxaudit mcp` is a new Cobra subcommand. It opens the SQLite database, resolves the latest audit run, and enters an MCP read loop over stdio.

**Transport:** JSON-RPC 2.0 over stdio. Claude Desktop and Claude Code launch the server as a subprocess. No networking required.

**Protocol:** Implemented natively in Go. No external MCP SDK. The protocol surface is small enough (3 methods) that a dependency adds more complexity than it saves.

**Package layout:**
```
cmd/mcp.go                  — cobra subcommand, wires server together
internal/mcpserver/
    server.go               — MCP protocol loop (stdio read/write, JSON-RPC dispatch)
    tools.go                — tool definitions (name, description, input JSON schema)
    handlers.go             — one handler per tool, all queries scoped to latest run
```

---

## Config Default Change

`internal/config/config.go` — `defaults()` changes `Output.Directory` from `"./aws-cost-audit"` to `~/.config/oxaudit` (resolved at runtime via `os.UserHomeDir()`).

A `resolvePath()` helper expands `~` before any path is used. All downstream paths (DB, run folders, exports) derive from `Output.Directory` and follow automatically.

**Result — all data lives under:**
```
~/.config/oxaudit/
    db/aws_cost_audit.sqlite
    runs/<YYYYMMDD-HHMMSS>/raw/
    runs/<YYYYMMDD-HHMMSS>/exports/
    config.yaml
```

---

## Startup Sequence

1. Resolve DB path: `~/.config/oxaudit/db/aws_cost_audit.sqlite` or `--db <path>` override
2. If DB file does not exist → print `"No database found at <path>. Run 'oxaudit all' first."` and exit
3. Query latest `audit_run_id` by `executed_at` — if none → same message and exit
4. Register 6 tools and enter MCP stdio loop

---

## MCP Methods Handled

| Method | Behaviour |
|---|---|
| `initialize` | Returns server name (`oxaudit`), version, capabilities |
| `tools/list` | Returns all 6 tool definitions with descriptions and input schemas |
| `tools/call` | Dispatches to matching handler; returns result or MCP error |
| *(unknown)* | Returns standard JSON-RPC error, does not crash |

Panics inside handlers are recovered and returned as MCP errors.

---

## Tools

All read tools are scoped to the latest `audit_run_id` resolved at startup.

### `run_audit`
- **Type:** Action
- **Inputs:** none
- **Behaviour:** Execs `oxaudit all` (with same `--db` path), captures stdout+stderr, returns combined output and exit status on completion
- **Errors:** Returns stderr text on failure so Claude can surface auth/config errors in plain language. Returns `"Audit already in progress"` if called concurrently.

### `get_summary`
- **Type:** Read
- **Inputs:** none
- **Returns:** audit period (start/end), accounts count, regions count, total findings, P0/P1/P2/P3 counts, total estimated monthly savings (USD), run timestamp

### `list_findings`
- **Type:** Read
- **Inputs (all optional):**
  - `priority` — `"P0"` | `"P1"` | `"P2"` | `"P3"`
  - `service` — e.g. `"EC2"`, `"RDS"`
  - `category` — e.g. `"Waste"`, `"Tagging"`, `"Anomaly"`
  - `min_savings_usd` — float, filter findings below this threshold
  - `status` — `"open"` | `"dismissed"` | `"resolved"` (default: `"open"`)
  - `limit` — integer, default 50
- **Returns:** array of findings — ID, title, priority, service, region, account, savings, confidence, risk

### `get_finding`
- **Type:** Read
- **Inputs:** `finding_id` (required)
- **Returns:** full finding record — all fields including evidence, recommended action, resource IDs, source files
- **Error:** `"Finding <id> not found in latest run"`

### `get_cost_breakdown`
- **Type:** Read
- **Inputs (all optional):**
  - `group_by` — `"service"` | `"account"` (default `"service"`)
  - `months` — integer, how many months of history to include (default 3)
- **Returns:** ranked list of groups with monthly cost totals and month-over-month change

### `query_resources`
- **Type:** Read
- **Inputs (all optional):**
  - `resource_type` — e.g. `"aws:ec2:instance"`, `"aws:ec2:volume"`
  - `state` — e.g. `"stopped"`, `"available"`
  - `region` — e.g. `"us-east-1"`
  - `account_id`
  - `limit` — integer, default 50
- **Returns:** matching resources — type, state, region, name, estimated monthly cost, tags

---

## Error Handling

| Scenario | Behaviour |
|---|---|
| DB file missing at startup | Exit with clear message before entering MCP loop |
| No audit runs in DB | Exit with clear message suggesting `oxaudit all` |
| `get_finding` unknown ID | MCP error response, not crash |
| `run_audit` while running | MCP error: `"Audit already in progress"` |
| `run_audit` non-zero exit | Return stderr in result so Claude can explain it |
| Empty `list_findings` / `query_resources` | Empty array — not an error |
| DB query error | Log to stderr, return as MCP error response |

---

## Typical Claude Workflow

```
User:   "Run my AWS audit and tell me what to fix first."

Claude: [run_audit]           → audit runs, returns summary
        [get_summary]         → orient: totals, savings, scope
        [list_findings P0]    → urgent issues
        [get_finding FND-xxx] → full detail on top finding
        → Responds with prioritised action plan grounded in real data
```

---

## Out of Scope

- Writing or mutating findings via MCP
- Streaming `run_audit` progress (returns on completion only)
- Multi-run comparison via MCP
- Authentication / per-user access control (local tool, single user)
