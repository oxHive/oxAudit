# MCP Integration Guide

oxAudit exposes an MCP (Model Context Protocol) server so Claude Desktop and Claude Code can query your AWS audit data directly.

## Prerequisites

- oxAudit installed (`go install github.com/graditya/oxaudit@latest` or built locally)
- At least one completed audit run (`oxaudit all`)
- Claude Desktop or Claude Code

## Quick Start

```bash
oxaudit all
```

Data is stored at `~/.config/oxaudit/db/aws_cost_audit.sqlite` by default.

## Claude Desktop Setup

Config file location:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux:** `~/.config/claude/claude_desktop_config.json`

Add the following to your config:

```json
{
  "mcpServers": {
    "oxaudit": {
      "command": "oxaudit",
      "args": ["mcp"]
    }
  }
}
```

If `oxaudit` is not on PATH, use the full path (e.g., `/home/yourname/go/bin/oxaudit`).

Restart Claude Desktop. The oxaudit tools appear under the hammer (🔨) icon.

## Claude Code Setup

Add to `~/.claude/settings.json` (user-wide) or `.claude/settings.json` (project-scoped):

```json
{
  "mcpServers": {
    "oxaudit": {
      "command": "oxaudit",
      "args": ["mcp"]
    }
  }
}
```

Or via CLI:

```bash
claude mcp add oxaudit oxaudit mcp
```

## Custom Database Path

If your audit data is not in the default location:

```json
{
  "mcpServers": {
    "oxaudit": {
      "command": "oxaudit",
      "args": ["mcp", "--db", "/path/to/aws_cost_audit.sqlite"]
    }
  }
}
```

## Available Tools

| Tool | What it does |
|---|---|
| `run_audit` | Runs `oxaudit all` — collect, ingest, analyze, export |
| `get_summary` | Audit overview: accounts, regions, finding counts, total savings |
| `list_findings` | List findings — filterable by priority, service, category, savings, status |
| `get_finding` | Full detail on one finding: evidence, recommended action, resources |
| `get_cost_breakdown` | Monthly cost by service or account |
| `query_resources` | Search inventory by type, state, region, account |

## Example Conversations

**"Run my AWS audit and tell me what to fix first."**
Claude calls `run_audit` → `get_summary` → `list_findings(P0)` → `get_finding` on the top result → responds with an action plan.

**"What's driving our EC2 costs in us-east-1 this month?"**
Claude calls `get_cost_breakdown` + `query_resources(region=us-east-1)`.

**"What can we safely delete this week?"**
Claude calls `list_findings(category=Waste, min_savings_usd=10)` → `get_finding` on top results to check risk and confidence.

## Troubleshooting

| Symptom | Fix |
|---|---|
| "No database found" | Run `oxaudit all` first |
| "No completed audit runs found" | Previous run may have failed — check `oxaudit all` output |
| "oxaudit not found" | Check PATH with `which oxaudit` or use the full path in the config |
| MCP server doesn't appear in Claude | Restart after editing config; verify the config file is valid JSON |
