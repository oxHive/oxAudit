# MCP Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `oxaudit mcp` subcommand that exposes audit data to Claude via the MCP protocol over stdio, and change the default data directory to `~/.config/oxaudit`.

**Architecture:** JSON-RPC 2.0 over stdio, implemented natively in Go. On startup the server resolves the latest `audit_run_id` from SQLite and scopes all queries to it. Six tools cover reading findings/costs/resources and triggering a new audit run.

**Tech Stack:** Go, `modernc.org/sqlite`, `github.com/spf13/cobra`, standard library `encoding/json`, `bufio`, `os/exec`

---

## File Map

| Action | Path | Responsibility |
|---|---|---|
| Modify | `internal/config/config.go` | Change default output dir to `~/.config/oxaudit`, add `resolvePath()` |
| Modify | `cmd/context.go` | Update default config path to `~/.config/oxaudit/config.yaml` |
| Modify | `internal/config/config_test.go` | Test new default DB path |
| Create | `internal/mcpserver/server.go` | JSON-RPC types, protocol loop, dispatch |
| Create | `internal/mcpserver/tools.go` | All 6 tool definitions with JSON schemas |
| Create | `internal/mcpserver/handlers.go` | All 6 tool handlers + helpers |
| Create | `internal/mcpserver/handlers_test.go` | Handler tests using real SQLite |
| Create | `internal/mcpserver/server_test.go` | Protocol loop tests |
| Create | `cmd/mcp.go` | Cobra subcommand wiring |
| Create | `docs/mcp-integration.md` | User-facing integration guide |

---

## Task 1: Change default output directory to ~/.config/oxaudit

**Files:**
- Modify: `internal/config/config.go`
- Modify: `cmd/context.go`
- Modify (create if missing): `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultDBPathUsesConfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	cfg := defaults()
	// defaults() returns tilde string; resolve it
	got := filepath.Join(resolvePath(cfg.Output.Directory), "db", "aws_cost_audit.sqlite")
	want := filepath.Join(home, ".config", "oxaudit", "db", "aws_cost_audit.sqlite")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestResolvePathExpandsTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := resolvePath("~/.config/oxaudit")
	if !strings.HasPrefix(got, home) {
		t.Errorf("resolvePath did not expand tilde: got %s", got)
	}
}

func TestResolvePathPassthroughAbsolute(t *testing.T) {
	p := "/absolute/path"
	if resolvePath(p) != p {
		t.Errorf("resolvePath changed absolute path")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/graditya/projects/oxhive/oxaudit && go test ./internal/config/... -v
```

Expected: FAIL — `resolvePath` undefined

- [ ] **Step 3: Add `resolvePath` and update default in `internal/config/config.go`**

Add this function (before `defaults()`):

```go
// resolvePath expands a leading ~ to the user home directory.
func resolvePath(p string) string {
	if !strings.HasPrefix(p, "~/") && p != "~" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}
```

Add these imports to `internal/config/config.go` (they may already be partially present):
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)
```

In `defaults()`, change:
```go
cfg.Output.Directory = "./aws-cost-audit"
```
to:
```go
cfg.Output.Directory = "~/.config/oxaudit"
```

Update `DBPath()` to resolve the tilde:
```go
func (c *Config) DBPath() string {
	return filepath.Join(resolvePath(c.Output.Directory), "db", "aws_cost_audit.sqlite")
}
```

- [ ] **Step 4: Update `cmd/context.go` default config path**

Replace these lines at the top of `loadRunContext()`:
```go
outDir := "./aws-cost-audit"
configPath := cfgFile
if configPath == "" {
    configPath = filepath.Join(outDir, "config.yaml")
} else {
    outDir = filepath.Dir(configPath)
}
```

With:
```go
home, err := os.UserHomeDir()
if err != nil {
    return nil, nil, "", "", fmt.Errorf("resolving home directory: %w", err)
}
configPath := cfgFile
if configPath == "" {
    configPath = filepath.Join(home, ".config", "oxaudit", "config.yaml")
}
```

Also remove the now-unused `outDir` variable — the value comes from `cfg.Output.Directory` later:
```go
// line that reads: outDir = cfg.Output.Directory  → keep this one
```

Make sure `resolvePath` is called on `cfg.Output.Directory` when used as a file path. In `loadRunContext()`, find the line:
```go
outDir = cfg.Output.Directory
```
and change it to:
```go
outDir = resolvePath(cfg.Output.Directory)
```

Add the import of the config package's resolvePath — since it's unexported, duplicate the helper inline in `cmd/context.go` or export it. **Export it** by renaming to `ResolvePath` in `config.go` and updating the test.

Updated `config.go`:
```go
// ResolvePath expands a leading ~ to the user home directory.
func ResolvePath(p string) string {
	if !strings.HasPrefix(p, "~/") && p != "~" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}
```

Update `DBPath()`:
```go
func (c *Config) DBPath() string {
	return filepath.Join(ResolvePath(c.Output.Directory), "db", "aws_cost_audit.sqlite")
}
```

Update `config_test.go` to use `ResolvePath` (capital R).

In `cmd/context.go`, add after `outDir = cfg.Output.Directory`:
```go
outDir = config.ResolvePath(outDir)
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/config/... -v
```

Expected: PASS

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go cmd/context.go
git commit -m "feat(config): change default output directory to ~/.config/oxaudit"
```

---

## Task 2: MCP server types, protocol loop, and tool definitions

**Files:**
- Create: `internal/mcpserver/server.go`
- Create: `internal/mcpserver/tools.go`
- Create: `internal/mcpserver/server_test.go`

- [ ] **Step 1: Write the failing server test**

Create `internal/mcpserver/server_test.go`:

```go
package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	internaldb "github.com/graditya/oxaudit/internal/db"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := internaldb.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return New(d, "test-run-001")
}

func serverRoundTrip(t *testing.T, srv *Server, req map[string]interface{}) map[string]interface{} {
	t.Helper()
	b, _ := json.Marshal(req)
	r := bytes.NewReader(append(b, '\n'))
	var w bytes.Buffer
	if err := srv.Serve(context.Background(), r, &w); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, w.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	})
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result, got: %v", resp)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocolVersion: %v", result["protocolVersion"])
	}
}

func TestToolsList(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result, got: %v", resp)
	}
	tools, ok := result["tools"].([]interface{})
	if !ok || len(tools) != 6 {
		t.Errorf("expected 6 tools, got: %v", result["tools"])
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "unknown/method",
	})
	if resp["error"] == nil {
		t.Errorf("expected error for unknown method, got: %v", resp)
	}
}

func TestNotificationIgnored(t *testing.T) {
	srv := newTestServer(t)
	// Notifications have no "id" — server must not respond
	notif, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	req, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})
	input := append(notif, '\n')
	input = append(input, req...)
	input = append(input, '\n')

	var w bytes.Buffer
	srv.Serve(context.Background(), bytes.NewReader(input), &w)

	// Should have exactly one JSON object in output (the tools/list response)
	dec := json.NewDecoder(&w)
	var count int
	for dec.More() {
		var v interface{}
		dec.Decode(&v)
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 response, got %d", count)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/mcpserver/... -v
```

Expected: FAIL — package does not exist

- [ ] **Step 3: Create `internal/mcpserver/server.go`**

```go
package mcpserver

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Server is the oxaudit MCP server.
type Server struct {
	db         *sql.DB
	mu         sync.RWMutex // protects auditRunID
	auditRunID string
	runMu      sync.Mutex // held while run_audit executes
}

// New creates a Server backed by the given database and initial audit run ID.
func New(db *sql.DB, auditRunID string) *Server {
	return &Server{db: db, auditRunID: auditRunID}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CallResult is the standard MCP tool result envelope.
type CallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content is a single content block in a CallResult.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Serve reads newline-delimited JSON-RPC requests from r and writes responses to w.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue // malformed line — skip
		}

		// Notifications have no "id" field — do not respond.
		if len(req.ID) == 0 {
			continue
		}

		result, rpcErr := s.dispatch(ctx, req)
		resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encoding response: %w", err)
		}
	}
	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, req rpcRequest) (interface{}, *rpcError) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize()
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(ctx, req.Params)
	case "ping":
		return map[string]interface{}{}, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found: " + req.Method}
	}
}

func (s *Server) handleInitialize() (interface{}, *rpcError) {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
		"serverInfo":      map[string]interface{}{"name": "oxaudit", "version": "0.1.0"},
	}, nil
}

func (s *Server) handleToolsList() (interface{}, *rpcError) {
	return map[string]interface{}{"tools": allTools()}, nil
}

type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, *rpcError) {
	var p toolsCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}
	if p.Arguments == nil {
		p.Arguments = map[string]interface{}{}
	}

	var result CallResult
	var rpcErr *rpcError

	switch p.Name {
	case "run_audit":
		result, rpcErr = s.handleRunAudit(ctx, p.Arguments)
	case "get_summary":
		result, rpcErr = s.handleGetSummary(ctx, p.Arguments)
	case "list_findings":
		result, rpcErr = s.handleListFindings(ctx, p.Arguments)
	case "get_finding":
		result, rpcErr = s.handleGetFinding(ctx, p.Arguments)
	case "get_cost_breakdown":
		result, rpcErr = s.handleGetCostBreakdown(ctx, p.Arguments)
	case "query_resources":
		result, rpcErr = s.handleQueryResources(ctx, p.Arguments)
	default:
		return nil, &rpcError{Code: -32602, Message: "unknown tool: " + p.Name}
	}

	if rpcErr != nil {
		return nil, rpcErr
	}
	return result, nil
}
```

- [ ] **Step 4: Create `internal/mcpserver/tools.go`**

```go
package mcpserver

// Tool is one MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is the JSON Schema for a tool's input.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property is one field in an InputSchema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

func allTools() []Tool {
	return []Tool{
		{
			Name:        "run_audit",
			Description: "Run the full oxaudit pipeline (collect → ingest → analyze → export). Use when the user wants to run or refresh the AWS cost audit. Returns pipeline output and any errors.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_summary",
			Description: "Get an overview of the latest audit run: accounts scanned, regions, finding counts by priority (P0–P3), and total estimated monthly savings in USD.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "list_findings",
			Description: "List audit findings from the latest run. All filters are optional. Returns ID, title, priority, service, region, account, estimated savings, confidence, and risk for each finding.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"priority":        {Type: "string", Enum: []string{"P0", "P1", "P2", "P3"}, Description: "Filter by priority level"},
					"service":         {Type: "string", Description: "Filter by AWS service name, e.g. EC2, RDS, S3"},
					"category":        {Type: "string", Description: "Filter by category, e.g. Waste, Tagging, Anomaly"},
					"min_savings_usd": {Type: "number", Description: "Only return findings with estimated monthly savings >= this value"},
					"status":          {Type: "string", Enum: []string{"open", "dismissed", "resolved"}, Description: "Finding status filter (default: open)"},
					"limit":           {Type: "integer", Description: "Maximum results to return (default: 50)"},
				},
			},
		},
		{
			Name:        "get_finding",
			Description: "Get full details for a single finding by ID, including evidence, recommended action, linked resource IDs, and source files.",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{"finding_id"},
				Properties: map[string]Property{
					"finding_id": {Type: "string", Description: "The finding ID, e.g. FND-a1b2c3d4"},
				},
			},
		},
		{
			Name:        "get_cost_breakdown",
			Description: "Get monthly AWS cost totals grouped by service or account. Useful for identifying top cost drivers and month-over-month trends.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"group_by": {Type: "string", Enum: []string{"service", "account"}, Description: "Group costs by service or account (default: service)"},
					"months":   {Type: "integer", Description: "Number of months of history to include (default: 3)"},
				},
			},
		},
		{
			Name:        "query_resources",
			Description: "Search the AWS resource inventory. All filters are optional. Useful for answering inventory questions like 'how many stopped EC2 instances are in us-east-1?'",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"resource_type": {Type: "string", Description: "Resource type, e.g. aws:ec2:instance, aws:ec2:volume, aws:rds:instance"},
					"state":         {Type: "string", Description: "Resource state, e.g. stopped, available, running"},
					"region":        {Type: "string", Description: "AWS region, e.g. us-east-1"},
					"account_id":    {Type: "string", Description: "AWS account ID"},
					"limit":         {Type: "integer", Description: "Maximum results to return (default: 50)"},
				},
			},
		},
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/mcpserver/... -v -run "TestInitialize|TestToolsList|TestUnknownMethod|TestNotificationIgnored"
```

Expected: PASS (handlers not yet implemented — tools/call tests come in Task 3)

- [ ] **Step 6: Commit**

```bash
git add internal/mcpserver/
git commit -m "feat(mcp): add MCP server protocol loop and tool definitions"
```

---

## Task 3: Handlers — get_summary, list_findings, get_finding

**Files:**
- Create: `internal/mcpserver/handlers.go`
- Create: `internal/mcpserver/handlers_test.go`

- [ ] **Step 1: Write the failing handler tests**

Create `internal/mcpserver/handlers_test.go`:

```go
package mcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	internaldb "github.com/graditya/oxaudit/internal/db"
)

const testRunID = "test-run-001"

func setupTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	d, err := internaldb.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	_, err = d.Exec(`
		INSERT INTO audit_run
			(id, executed_at, generated_at, period_start, period_end,
			 aws_profile, billing_region, run_folder, status)
		VALUES (?, datetime('now'), datetime('now'), '2026-01-01', '2026-02-01',
		        'default', 'us-east-1', '/tmp', 'complete')`, testRunID)
	if err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
	return d, testRunID
}

func insertFindings(t *testing.T, d *sql.DB) {
	t.Helper()
	_, err := d.Exec(`
		INSERT INTO findings
			(id, audit_run_id, priority, category, service, account_id, account_name,
			 region, title, summary, evidence, recommended_action,
			 est_monthly_savings_usd, confidence, risk, status,
			 resource_ids_json, tags_json, source_files_json, created_at, updated_at)
		VALUES
		('FND-00000001', ?, 'P1', 'Waste', 'EC2', '123456789012', 'prod',
		 'us-east-1', 'Unattached EBS volume', 'vol-abc is unattached',
		 '100 GiB gp2 unattached 30 days', 'Delete or snapshot',
		 8.0, 'High', 'Low', 'open', '["vol-abc"]', '{}', '[]',
		 datetime('now'), datetime('now')),
		('FND-00000002', ?, 'P0', 'Anomaly', 'EC2', '123456789012', 'prod',
		 'us-east-1', 'Cost spike detected', 'EC2 spend spiked 3x',
		 'Daily cost $10 → $30', 'Investigate EC2 usage',
		 0.0, 'High', 'High', 'open', '[]', '{}', '[]',
		 datetime('now'), datetime('now'))`,
		testRunID, testRunID)
	if err != nil {
		t.Fatalf("insert findings: %v", err)
	}
}

func insertAccounts(t *testing.T, d *sql.DB) {
	t.Helper()
	_, err := d.Exec(`
		INSERT INTO accounts (account_id, account_name, audit_run_id)
		VALUES ('123456789012', 'prod', ?)`, testRunID)
	if err != nil {
		t.Fatalf("insert accounts: %v", err)
	}
}

func callTool(t *testing.T, srv *Server, name string, args map[string]interface{}) map[string]interface{} {
	t.Helper()
	if args == nil {
		args = map[string]interface{}{}
	}
	result, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      name,
		"arguments": args,
	}))
	if rpcErr != nil {
		t.Fatalf("rpc error: %+v", rpcErr)
	}
	cr, ok := result.(CallResult)
	if !ok || len(cr.Content) == 0 {
		t.Fatalf("empty or wrong result type: %v", result)
	}
	var out map[string]interface{}
	json.Unmarshal([]byte(cr.Content[0].Text), &out)
	return out
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestGetSummary(t *testing.T) {
	d, runID := setupTestDB(t)
	insertFindings(t, d)
	insertAccounts(t, d)
	srv := New(d, runID)

	out := callTool(t, srv, "get_summary", nil)
	if out["total_findings"].(float64) != 2 {
		t.Errorf("expected 2 findings, got %v", out["total_findings"])
	}
	if out["p0"].(float64) != 1 {
		t.Errorf("expected p0=1, got %v", out["p0"])
	}
	if out["p1"].(float64) != 1 {
		t.Errorf("expected p1=1, got %v", out["p1"])
	}
}

func TestListFindings_NoFilter(t *testing.T) {
	d, runID := setupTestDB(t)
	insertFindings(t, d)
	srv := New(d, runID)

	_, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "list_findings",
		"arguments": map[string]interface{}{},
	}))
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
}

func TestListFindings_PriorityFilter(t *testing.T) {
	d, runID := setupTestDB(t)
	insertFindings(t, d)
	srv := New(d, runID)

	result, _ := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "list_findings",
		"arguments": map[string]interface{}{"priority": "P0"},
	}))
	cr := result.(CallResult)
	var findings []interface{}
	json.Unmarshal([]byte(cr.Content[0].Text), &findings)
	if len(findings) != 1 {
		t.Errorf("expected 1 P0 finding, got %d", len(findings))
	}
}

func TestGetFinding_Found(t *testing.T) {
	d, runID := setupTestDB(t)
	insertFindings(t, d)
	srv := New(d, runID)

	out := callTool(t, srv, "get_finding", map[string]interface{}{"finding_id": "FND-00000001"})
	if out["id"] != "FND-00000001" {
		t.Errorf("expected FND-00000001, got %v", out["id"])
	}
	if out["evidence"] == "" {
		t.Errorf("expected evidence to be populated")
	}
}

func TestGetFinding_NotFound(t *testing.T) {
	d, runID := setupTestDB(t)
	srv := New(d, runID)

	result, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "get_finding",
		"arguments": map[string]interface{}{"finding_id": "FND-nonexistent"},
	}))
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	cr := result.(CallResult)
	if !cr.IsError {
		t.Errorf("expected IsError=true for missing finding")
	}
}

func TestGetFinding_MissingID(t *testing.T) {
	d, runID := setupTestDB(t)
	srv := New(d, runID)

	result, _ := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "get_finding",
		"arguments": map[string]interface{}{},
	}))
	cr := result.(CallResult)
	if !cr.IsError {
		t.Errorf("expected IsError=true when finding_id missing")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/mcpserver/... -v -run "TestGetSummary|TestListFindings|TestGetFinding"
```

Expected: FAIL — handlers not defined

- [ ] **Step 3: Create `internal/mcpserver/handlers.go`** (partial — summary, list, get)

```go
package mcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// helpers

func textOK(v interface{}) (CallResult, *rpcError) {
	b, _ := json.MarshalIndent(v, "", "  ")
	return CallResult{Content: []Content{{Type: "text", Text: string(b)}}}, nil
}

func textErr(msg string) (CallResult, *rpcError) {
	return CallResult{Content: []Content{{Type: "text", Text: msg}}, IsError: true}, nil
}

func strArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

func floatArg(args map[string]interface{}, key string, def float64) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

// handleGetSummary returns audit-level statistics for the latest run.
func (s *Server) handleGetSummary(ctx context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	var periodStart, periodEnd, executedAt string
	s.db.QueryRowContext(ctx,
		`SELECT period_start, period_end, executed_at FROM audit_run WHERE id = ?`, runID,
	).Scan(&periodStart, &periodEnd, &executedAt)

	var total, p0, p1, p2, p3 int
	var totalSavings float64
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN priority='P0' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P1' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P2' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P3' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(est_monthly_savings_usd), 0)
		FROM findings WHERE audit_run_id = ? AND status = 'open'`, runID,
	).Scan(&total, &p0, &p1, &p2, &p3, &totalSavings)

	var accounts, regions int
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM accounts WHERE audit_run_id = ?`, runID).Scan(&accounts)
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT region) FROM resources WHERE audit_run_id = ?`, runID).Scan(&regions)

	return textOK(map[string]interface{}{
		"audit_run_id":              runID,
		"period_start":              periodStart,
		"period_end":                periodEnd,
		"executed_at":               executedAt,
		"accounts":                  accounts,
		"regions":                   regions,
		"total_findings":            total,
		"p0":                        p0,
		"p1":                        p1,
		"p2":                        p2,
		"p3":                        p3,
		"estimated_monthly_savings": totalSavings,
	})
}

// handleListFindings lists findings with optional filters.
func (s *Server) handleListFindings(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	priority := strArg(args, "priority")
	service := strArg(args, "service")
	category := strArg(args, "category")
	status := strArg(args, "status")
	if status == "" {
		status = "open"
	}
	minSavings := floatArg(args, "min_savings_usd", 0)
	limit := intArg(args, "limit", 50)

	query := `SELECT id, title, priority, category, service, region, account_name,
	                 est_monthly_savings_usd, confidence, risk, status
	          FROM findings WHERE audit_run_id = ? AND status = ?`
	queryArgs := []interface{}{runID, status}

	if priority != "" {
		query += " AND priority = ?"
		queryArgs = append(queryArgs, priority)
	}
	if service != "" {
		query += " AND service = ?"
		queryArgs = append(queryArgs, service)
	}
	if category != "" {
		query += " AND category = ?"
		queryArgs = append(queryArgs, category)
	}
	if minSavings > 0 {
		query += " AND est_monthly_savings_usd >= ?"
		queryArgs = append(queryArgs, minSavings)
	}
	query += " ORDER BY priority, est_monthly_savings_usd DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)

	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	defer rows.Close()

	type finding struct {
		ID         string  `json:"id"`
		Title      string  `json:"title"`
		Priority   string  `json:"priority"`
		Category   string  `json:"category"`
		Service    string  `json:"service"`
		Region     string  `json:"region"`
		Account    string  `json:"account_name"`
		Savings    float64 `json:"est_monthly_savings_usd"`
		Confidence string  `json:"confidence"`
		Risk       string  `json:"risk"`
		Status     string  `json:"status"`
	}

	var findings []finding
	for rows.Next() {
		var f finding
		rows.Scan(&f.ID, &f.Title, &f.Priority, &f.Category, &f.Service,
			&f.Region, &f.Account, &f.Savings, &f.Confidence, &f.Risk, &f.Status)
		findings = append(findings, f)
	}
	if findings == nil {
		findings = []finding{}
	}
	return textOK(findings)
}

// handleGetFinding returns full detail for one finding.
func (s *Server) handleGetFinding(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	id := strArg(args, "finding_id")
	if id == "" {
		return textErr("finding_id is required")
	}

	type fullFinding struct {
		ID                string  `json:"id"`
		Priority          string  `json:"priority"`
		Category          string  `json:"category"`
		Service           string  `json:"service"`
		AccountID         string  `json:"account_id"`
		AccountName       string  `json:"account_name"`
		Region            string  `json:"region"`
		Title             string  `json:"title"`
		Summary           string  `json:"summary"`
		Evidence          string  `json:"evidence"`
		RecommendedAction string  `json:"recommended_action"`
		Savings           float64 `json:"est_monthly_savings_usd"`
		Confidence        string  `json:"confidence"`
		Risk              string  `json:"risk"`
		Owner             string  `json:"owner"`
		Status            string  `json:"status"`
		ResourceIDs       string  `json:"resource_ids"`
		SourceFiles       string  `json:"source_files"`
		CreatedAt         string  `json:"created_at"`
	}

	var f fullFinding
	err := s.db.QueryRowContext(ctx, `
		SELECT id, priority, category, service, account_id, account_name, region,
		       title, summary, evidence, recommended_action, est_monthly_savings_usd,
		       confidence, risk, owner, status, resource_ids_json, source_files_json, created_at
		FROM findings WHERE id = ? AND audit_run_id = ?`, id, runID,
	).Scan(&f.ID, &f.Priority, &f.Category, &f.Service, &f.AccountID, &f.AccountName,
		&f.Region, &f.Title, &f.Summary, &f.Evidence, &f.RecommendedAction,
		&f.Savings, &f.Confidence, &f.Risk, &f.Owner, &f.Status,
		&f.ResourceIDs, &f.SourceFiles, &f.CreatedAt)

	if err == sql.ErrNoRows {
		return textErr(fmt.Sprintf("finding %s not found in latest run", id))
	}
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	return textOK(f)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/mcpserver/... -v -run "TestGetSummary|TestListFindings|TestGetFinding"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/mcpserver/handlers.go internal/mcpserver/handlers_test.go
git commit -m "feat(mcp): add get_summary, list_findings, get_finding handlers"
```

---

## Task 4: Handlers — get_cost_breakdown, query_resources

**Files:**
- Modify: `internal/mcpserver/handlers.go`
- Modify: `internal/mcpserver/handlers_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/mcpserver/handlers_test.go`:

```go
func insertCosts(t *testing.T, d *sql.DB) {
	t.Helper()
	_, err := d.Exec(`
		INSERT INTO cost_monthly
			(audit_run_id, month, account_id, account_name, service, unblended_cost, amortized_cost)
		VALUES
		(?, date('now', 'start of month'), '123456789012', 'prod', 'EC2', 150.0, 145.0),
		(?, date('now', 'start of month'), '123456789012', 'prod', 'RDS', 80.0, 78.0)`,
		testRunID, testRunID)
	if err != nil {
		t.Fatalf("insert costs: %v", err)
	}
}

func insertResources(t *testing.T, d *sql.DB) {
	t.Helper()
	_, err := d.Exec(`
		INSERT INTO resources
			(audit_run_id, resource_id, resource_type, account_id, account_name,
			 region, service, state, name, tags_json, raw_json, discovered_at)
		VALUES
		(?, 'vol-abc', 'aws:ec2:volume', '123456789012', 'prod',
		 'us-east-1', 'EC2', 'available', 'my-volume', '{}', '{}', datetime('now')),
		(?, 'i-12345', 'aws:ec2:instance', '123456789012', 'prod',
		 'us-east-1', 'EC2', 'stopped', 'web-server', '{}', '{}', datetime('now'))`,
		testRunID, testRunID)
	if err != nil {
		t.Fatalf("insert resources: %v", err)
	}
}

func TestGetCostBreakdown_ByService(t *testing.T) {
	d, runID := setupTestDB(t)
	insertCosts(t, d)
	srv := New(d, runID)

	result, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "get_cost_breakdown",
		"arguments": map[string]interface{}{"group_by": "service"},
	}))
	if rpcErr != nil {
		t.Fatalf("rpc error: %v", rpcErr)
	}
	cr := result.(CallResult)
	if cr.IsError {
		t.Fatalf("unexpected tool error: %s", cr.Content[0].Text)
	}
	var out map[string]interface{}
	json.Unmarshal([]byte(cr.Content[0].Text), &out)
	data, ok := out["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Errorf("expected cost data, got: %v", out)
	}
}

func TestQueryResources_NoFilter(t *testing.T) {
	d, runID := setupTestDB(t)
	insertResources(t, d)
	srv := New(d, runID)

	result, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "query_resources",
		"arguments": map[string]interface{}{},
	}))
	if rpcErr != nil {
		t.Fatalf("rpc error: %v", rpcErr)
	}
	cr := result.(CallResult)
	var resources []interface{}
	json.Unmarshal([]byte(cr.Content[0].Text), &resources)
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestQueryResources_StateFilter(t *testing.T) {
	d, runID := setupTestDB(t)
	insertResources(t, d)
	srv := New(d, runID)

	result, _ := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "query_resources",
		"arguments": map[string]interface{}{"state": "stopped"},
	}))
	cr := result.(CallResult)
	var resources []interface{}
	json.Unmarshal([]byte(cr.Content[0].Text), &resources)
	if len(resources) != 1 {
		t.Errorf("expected 1 stopped resource, got %d", len(resources))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/mcpserver/... -v -run "TestGetCostBreakdown|TestQueryResources"
```

Expected: FAIL — handlers not defined

- [ ] **Step 3: Append handlers to `internal/mcpserver/handlers.go`**

```go
// handleGetCostBreakdown returns monthly cost totals grouped by service or account.
func (s *Server) handleGetCostBreakdown(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	groupBy := strArg(args, "group_by")
	if groupBy != "account" {
		groupBy = "service"
	}
	months := intArg(args, "months", 3)

	groupCol := groupBy
	if groupBy == "account" {
		groupCol = "account_id"
	}

	query := fmt.Sprintf(`
		SELECT %s, month, SUM(unblended_cost) as total
		FROM cost_monthly
		WHERE audit_run_id = ?
		  AND month >= date('now', '-%d months')
		GROUP BY %s, month
		ORDER BY total DESC`, groupCol, months, groupCol)

	rows, err := s.db.QueryContext(ctx, query, runID)
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	defer rows.Close()

	type row struct {
		Group string  `json:"group"`
		Month string  `json:"month"`
		Total float64 `json:"unblended_cost_usd"`
	}

	var results []row
	for rows.Next() {
		var r row
		rows.Scan(&r.Group, &r.Month, &r.Total)
		results = append(results, r)
	}
	if results == nil {
		results = []row{}
	}
	return textOK(map[string]interface{}{
		"group_by": groupBy,
		"months":   months,
		"data":     results,
	})
}

// handleQueryResources searches the resource inventory with optional filters.
func (s *Server) handleQueryResources(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	resourceType := strArg(args, "resource_type")
	state := strArg(args, "state")
	region := strArg(args, "region")
	accountID := strArg(args, "account_id")
	limit := intArg(args, "limit", 50)

	query := `SELECT resource_id, resource_type, account_id, account_name,
	                 region, service, state, name, est_monthly_cost, tags_json
	          FROM resources WHERE audit_run_id = ?`
	queryArgs := []interface{}{runID}

	if resourceType != "" {
		query += " AND resource_type = ?"
		queryArgs = append(queryArgs, resourceType)
	}
	if state != "" {
		query += " AND state = ?"
		queryArgs = append(queryArgs, state)
	}
	if region != "" {
		query += " AND region = ?"
		queryArgs = append(queryArgs, region)
	}
	if accountID != "" {
		query += " AND account_id = ?"
		queryArgs = append(queryArgs, accountID)
	}
	query += " ORDER BY COALESCE(est_monthly_cost, 0) DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)

	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	defer rows.Close()

	type resource struct {
		ResourceID   string   `json:"resource_id"`
		ResourceType string   `json:"resource_type"`
		AccountID    string   `json:"account_id"`
		AccountName  string   `json:"account_name"`
		Region       string   `json:"region"`
		Service      string   `json:"service"`
		State        string   `json:"state"`
		Name         string   `json:"name"`
		EstCost      *float64 `json:"est_monthly_cost_usd"`
		Tags         string   `json:"tags"`
	}

	var resources []resource
	for rows.Next() {
		var r resource
		var estCost sql.NullFloat64
		rows.Scan(&r.ResourceID, &r.ResourceType, &r.AccountID, &r.AccountName,
			&r.Region, &r.Service, &r.State, &r.Name, &estCost, &r.Tags)
		if estCost.Valid {
			r.EstCost = &estCost.Float64
		}
		resources = append(resources, r)
	}
	if resources == nil {
		resources = []resource{}
	}
	return textOK(resources)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/mcpserver/... -v -run "TestGetCostBreakdown|TestQueryResources"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/mcpserver/handlers.go internal/mcpserver/handlers_test.go
git commit -m "feat(mcp): add get_cost_breakdown and query_resources handlers"
```

---

## Task 5: Handler — run_audit + cobra subcommand

**Files:**
- Modify: `internal/mcpserver/handlers.go`
- Modify: `internal/mcpserver/handlers_test.go`
- Create: `cmd/mcp.go`

- [ ] **Step 1: Write the failing run_audit test**

Append to `internal/mcpserver/handlers_test.go`:

```go
func TestRunAudit_ConcurrentGuard(t *testing.T) {
	d, runID := setupTestDB(t)
	srv := New(d, runID)

	// Acquire the run lock to simulate an in-progress audit.
	srv.runMu.Lock()
	defer srv.runMu.Unlock()

	result, rpcErr := srv.handleToolsCall(context.Background(), mustJSON(t, map[string]interface{}{
		"name":      "run_audit",
		"arguments": map[string]interface{}{},
	}))
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	cr := result.(CallResult)
	if !cr.IsError {
		t.Errorf("expected IsError=true when audit already running")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/mcpserver/... -v -run "TestRunAudit"
```

Expected: FAIL — handler not defined

- [ ] **Step 3: Append `handleRunAudit` to `internal/mcpserver/handlers.go`**

```go
// handleRunAudit executes the full oxaudit pipeline as a subprocess.
func (s *Server) handleRunAudit(ctx context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	if !s.runMu.TryLock() {
		return textErr("audit already in progress")
	}
	defer s.runMu.Unlock()

	exe, err := os.Executable()
	if err != nil {
		return textErr("could not resolve oxaudit binary: " + err.Error())
	}

	cmd := exec.CommandContext(ctx, exe, "all")
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return textErr(fmt.Sprintf("audit failed:\n%s", output))
	}

	// Refresh auditRunID to pick up the newly completed run.
	s.mu.Lock()
	var newRunID string
	s.db.QueryRowContext(ctx,
		`SELECT id FROM audit_run WHERE status = 'complete' ORDER BY executed_at DESC LIMIT 1`,
	).Scan(&newRunID)
	if newRunID != "" {
		s.auditRunID = newRunID
	}
	s.mu.Unlock()

	return textOK(map[string]interface{}{
		"status": "complete",
		"output": output,
	})
}
```

Add these imports to `handlers.go`:
```go
import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/mcpserver/... -v -run "TestRunAudit"
```

Expected: PASS

- [ ] **Step 5: Create `cmd/mcp.go`**

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	internaldb "github.com/graditya/oxaudit/internal/db"
	"github.com/graditya/oxaudit/internal/mcpserver"
	"github.com/spf13/cobra"
)

var (
	mcpDBPath string
	mcpCmd    = &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server for Claude integration (stdio transport)",
		Long:  "Serves oxaudit data over the MCP protocol via stdin/stdout. Configure Claude Desktop or Claude Code to launch this command as an MCP server.",
		RunE:  runMCP,
	}
)

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpDBPath, "db", "", "path to SQLite database (default: ~/.config/oxaudit/db/aws_cost_audit.sqlite)")
}

func runMCP(_ *cobra.Command, _ []string) error {
	dbPath := mcpDBPath
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".config", "oxaudit", "db", "aws_cost_audit.sqlite")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("no database found at %s\nRun 'oxaudit all' first to generate audit data", dbPath)
	}

	db, err := internaldb.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	var auditRunID string
	err = db.QueryRowContext(
		context.Background(),
		`SELECT id FROM audit_run WHERE status = 'complete' ORDER BY executed_at DESC LIMIT 1`,
	).Scan(&auditRunID)
	if err != nil {
		return fmt.Errorf("no completed audit runs found\nRun 'oxaudit all' first to generate audit data")
	}

	srv := mcpserver.New(db, auditRunID)
	return srv.Serve(context.Background(), os.Stdin, os.Stdout)
}
```

- [ ] **Step 6: Build and verify compilation**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 7: Run all tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/mcpserver/handlers.go internal/mcpserver/handlers_test.go cmd/mcp.go
git commit -m "feat(mcp): add run_audit handler and mcp cobra subcommand"
```

---

## Task 6: Integration documentation

**Files:**
- Create: `docs/mcp-integration.md`

- [ ] **Step 1: Create `docs/mcp-integration.md`**

```markdown
# oxAudit MCP Integration

oxAudit exposes an MCP server so Claude can query your AWS audit data and trigger new audits conversationally.

## Prerequisites

- oxAudit installed (`go install github.com/graditya/oxaudit@latest` or built locally)
- At least one completed audit run (`oxaudit all`)
- Claude Desktop or Claude Code

## Quick Start

Run an audit first:

```bash
oxaudit all
```

Data is stored at `~/.config/oxaudit/db/aws_cost_audit.sqlite` by default.

---

## Claude Desktop

Add this to your `claude_desktop_config.json`
(macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
 Linux: `~/.config/claude/claude_desktop_config.json`):

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

If `oxaudit` is not on your PATH, use the full binary path:

```json
{
  "mcpServers": {
    "oxaudit": {
      "command": "/home/yourname/go/bin/oxaudit",
      "args": ["mcp"]
    }
  }
}
```

Restart Claude Desktop. You should see **oxaudit** listed under the MCP tools icon (hammer).

---

## Claude Code

Add to your project's `.claude/settings.json` or `~/.claude/settings.json`:

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

Or register it from the Claude Code CLI:

```bash
claude mcp add oxaudit oxaudit mcp
```

---

## Custom Database Path

If your audit data is not in the default location, pass `--db`:

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

---

## Available Tools

Once connected, Claude can call these tools:

| Tool | What it does |
|---|---|
| `run_audit` | Runs `oxaudit all` — collect, ingest, analyze, export |
| `get_summary` | Audit overview: accounts, regions, finding counts, total savings |
| `list_findings` | List findings — filterable by priority, service, category, savings, status |
| `get_finding` | Full detail on one finding: evidence, recommended action, resources |
| `get_cost_breakdown` | Monthly cost by service or account |
| `query_resources` | Search inventory by type, state, region, account |

---

## Example Conversations

**Run an audit and triage:**
> "Run my AWS audit and tell me what to fix first."

Claude calls `run_audit`, then `get_summary`, then `list_findings` with `priority=P0` to give you a prioritized action plan.

**Investigate a cost spike:**
> "What's driving our EC2 costs in us-east-1 this month?"

Claude calls `get_cost_breakdown` and `query_resources` with `region=us-east-1` to identify the culprits.

**Safe cleanup:**
> "What can we safely delete this week?"

Claude calls `list_findings` with `category=Waste` and `min_savings_usd=10`, then uses `get_finding` on the top results to check risk and confidence before recommending actions.

---

## Troubleshooting

**"No database found"**
Run `oxaudit all` first. The database must exist at `~/.config/oxaudit/db/aws_cost_audit.sqlite`.

**"No completed audit runs found"**
A previous run may have failed. Check `oxaudit all` output for errors, or run it again.

**oxaudit not found**
Make sure the binary is on your PATH (`which oxaudit`) or use the full path in the MCP config.

**MCP server doesn't appear in Claude**
Restart Claude Desktop/Code after editing the config file. Check the config file is valid JSON.
```

- [ ] **Step 2: Commit**

```bash
git add docs/mcp-integration.md
git commit -m "docs: add MCP integration guide for Claude Desktop and Claude Code"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Covered by |
|---|---|
| `oxaudit mcp` subcommand | Task 5 — `cmd/mcp.go` |
| stdio transport | Task 2 — `server.go` Serve method |
| Default DB `~/.config/oxaudit` | Task 1 — config.go + context.go |
| `--db` override flag | Task 5 — `cmd/mcp.go` flag |
| Startup validation (no DB / no runs) | Task 5 — `runMCP()` |
| `initialize` / `tools/list` / `tools/call` | Task 2 — dispatch |
| 6 tool definitions with schemas | Task 2 — `tools.go` |
| `get_summary` | Task 3 |
| `list_findings` with all filters | Task 3 |
| `get_finding` with not-found error | Task 3 |
| `get_cost_breakdown` | Task 4 |
| `query_resources` with all filters | Task 4 |
| `run_audit` with concurrent guard | Task 5 |
| `run_audit` refreshes auditRunID | Task 5 |
| Notifications silently ignored | Task 2 — test + impl |
| Unknown tool → MCP error | Task 2 — dispatch |
| Documentation for Claude Desktop | Task 6 |
| Documentation for Claude Code | Task 6 |

No gaps found.
