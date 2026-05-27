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
		 'Daily cost $10 to $30', 'Investigate EC2 usage',
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
	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name":      name,
		"arguments": args,
	})
	result, rpcErr := srv.handleToolsCall(context.Background(), paramsBytes)
	if rpcErr != nil {
		t.Fatalf("rpc error: %+v", rpcErr)
	}
	cr, ok := result.(CallResult)
	if !ok || len(cr.Content) == 0 {
		t.Fatalf("empty or wrong result type: %v", result)
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(cr.Content[0].Text), &out); err != nil {
		t.Fatalf("callTool %s: response is not JSON: %s", name, cr.Content[0].Text)
	}
	return out
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

	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name":      "list_findings",
		"arguments": map[string]interface{}{},
	})
	result, rpcErr := srv.handleToolsCall(context.Background(), paramsBytes)
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	cr := result.(CallResult)
	if cr.IsError {
		t.Fatalf("unexpected tool error: %s", cr.Content[0].Text)
	}
	var findings []interface{}
	if err := json.Unmarshal([]byte(cr.Content[0].Text), &findings); err != nil {
		t.Fatalf("response is not a JSON array: %s", cr.Content[0].Text)
	}
	if len(findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(findings))
	}
}

func TestListFindings_PriorityFilter(t *testing.T) {
	d, runID := setupTestDB(t)
	insertFindings(t, d)
	srv := New(d, runID)

	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name":      "list_findings",
		"arguments": map[string]interface{}{"priority": "P0"},
	})
	result, _ := srv.handleToolsCall(context.Background(), paramsBytes)
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

	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name":      "get_finding",
		"arguments": map[string]interface{}{"finding_id": "FND-nonexistent"},
	})
	result, rpcErr := srv.handleToolsCall(context.Background(), paramsBytes)
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

	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name":      "get_finding",
		"arguments": map[string]interface{}{},
	})
	result, _ := srv.handleToolsCall(context.Background(), paramsBytes)
	cr := result.(CallResult)
	if !cr.IsError {
		t.Errorf("expected IsError=true when finding_id missing")
	}
}

func TestGetSummary_Empty(t *testing.T) {
	d, runID := setupTestDB(t)
	// No insertFindings — empty audit run
	srv := New(d, runID)

	out := callTool(t, srv, "get_summary", nil)
	if out["total_findings"].(float64) != 0 {
		t.Errorf("expected 0 findings, got %v", out["total_findings"])
	}
	if out["audit_run_id"] != runID {
		t.Errorf("expected audit_run_id=%s, got %v", runID, out["audit_run_id"])
	}
}
