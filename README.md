# oxAudit

Local AWS cost audit CLI that runs a deterministic pipeline from your workstation to a prioritized findings backlog with estimated monthly savings. No CI, no SaaS. Integrates with Claude via MCP.

## What It Does

`oxaudit all` takes you from an authenticated AWS profile to a prioritized remediation backlog in one command. It collects raw AWS CLI output as immutable evidence, normalizes it into a local SQLite database, runs deterministic finding rules, and exports human-readable and LLM-optimized summaries.

## What It Finds

| Rule | Category |
|---|---|
| Unattached EBS volumes | Waste |
| Old EBS snapshots past retention threshold | Waste |
| Unassociated Elastic IPs | Waste |
| Long-stopped EC2 instances | Waste |
| High NAT Gateway spend | Architecture |
| High CloudWatch Logs cost | Observability |
| Daily cost anomalies (stddev-based) | Anomaly |
| Resources missing required tags | Tagging |

Each finding includes priority (P0–P3), confidence, risk level, estimated monthly savings, linked resource IDs, evidence, and a recommended action.

## Installation

```bash
go install github.com/graditya/oxaudit@latest
```

Or build from source:

```bash
git clone https://github.com/oxhive/oxaudit
cd oxaudit
go build -o oxaudit .
```

## Usage

```bash
# Initialize config
oxaudit init

# Edit ~/.config/oxaudit/config.yaml — set start_date, end_date, aws.profile

# Run the full audit pipeline
oxaudit all

# Or run stages individually
oxaudit collect
oxaudit ingest
oxaudit analyze
oxaudit export

# Browse results in the local dashboard
oxaudit dashboard
```

Output is written to `~/.config/oxaudit/`.

## Configuration

`oxaudit init` writes a default config to `~/.config/oxaudit/config.yaml`. Required fields:

```yaml
aws:
  profile: default
  billing_region: us-east-1

audit:
  start_date: "2026-01-01"
  end_date: "2026-02-01"
```

Override the config path with `--config`.

## Claude MCP Integration

oxAudit exposes an MCP server so Claude can query your audit data and trigger new audits conversationally.

Add to your Claude Desktop or Claude Code config:

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

See [docs/mcp-integration.md](docs/mcp-integration.md) for full setup instructions.

## Tech Stack

- **Language:** Go
- **CLI:** Cobra
- **Database:** SQLite (`modernc.org/sqlite` — pure Go, no CGo)
- **Config:** YAML
- **Dashboard:** embedded HTTP server

## License

MIT