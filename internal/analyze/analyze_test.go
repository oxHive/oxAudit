package analyze_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/analyze"
	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/db"
)

func openTestDB(t *testing.T, runID string) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := d.Exec(`INSERT INTO audit_run (id, executed_at, generated_at, period_start, period_end, aws_profile, status) VALUES (?,?,?,?,?,?,?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "complete"); err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
	return d
}

func insertVolume(t *testing.T, d *sql.DB, runID, volID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := d.Exec(`INSERT INTO resources
		(audit_run_id, resource_id, resource_type, state, size_gib, volume_type, region, account_id, account_name, tags_json, source_file, created_at, discovered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, volID, "aws:ec2:volume", "available", 100.0, "gp2", "us-east-1", "111", "acct", "{}", "test.json", now, now); err != nil {
		t.Fatalf("insert volume: %v", err)
	}
}

func TestNew_returnsEngine(t *testing.T) {
	cfg := &config.Config{}
	e := analyze.New(cfg)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(e.Rules) == 0 {
		t.Fatal("expected at least one rule")
	}
}

func TestEngine_run_emptyDB(t *testing.T) {
	const runID = "run-analyze-001"
	d := openTestDB(t, runID)

	cfg := &config.Config{}
	e := analyze.New(cfg)

	called := 0
	if err := e.Run(context.Background(), d, cfg, runID, func(ruleID, _ string, _ int, err error) {
		called++
		if err != nil {
			t.Errorf("rule %s returned error: %v", ruleID, err)
		}
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if called == 0 {
		t.Error("expected onRule callback to be called at least once")
	}
}

func TestEngine_run_callbackNil(t *testing.T) {
	const runID = "run-analyze-002"
	d := openTestDB(t, runID)

	cfg := &config.Config{}
	e := analyze.New(cfg)

	if err := e.Run(context.Background(), d, cfg, runID, nil); err != nil {
		t.Fatalf("Run with nil callback: %v", err)
	}
}

func TestEngine_run_persistsFindings(t *testing.T) {
	const runID = "run-analyze-003"
	d := openTestDB(t, runID)
	insertVolume(t, d, runID, "vol-persist")

	cfg := &config.Config{}
	e := analyze.New(cfg)
	if err := e.Run(context.Background(), d, cfg, runID, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM findings WHERE audit_run_id = ?", runID).Scan(&count)
	if count == 0 {
		t.Error("expected at least one finding persisted after Run")
	}
}

func TestEngine_run_allRulesExecuted(t *testing.T) {
	const runID = "run-analyze-004"
	d := openTestDB(t, runID)

	cfg := &config.Config{}
	e := analyze.New(cfg)

	var ruleIDs []string
	e.Run(context.Background(), d, cfg, runID, func(ruleID, _ string, _ int, _ error) {
		ruleIDs = append(ruleIDs, ruleID)
	})

	expected := []string{
		"AWS-WASTE-001", "AWS-WASTE-002", "AWS-WASTE-003", "AWS-WASTE-004",
		"AWS-GOV-001", "AWS-COST-001", "AWS-COST-002", "AWS-ANOMALY-001",
	}
	if len(ruleIDs) != len(expected) {
		t.Errorf("expected %d rules, got %d: %v", len(expected), len(ruleIDs), ruleIDs)
	}
}

func TestEngine_run_multipleFindings(t *testing.T) {
	const runID = "run-analyze-005"
	d := openTestDB(t, runID)
	insertVolume(t, d, runID, "vol-a")
	insertVolume(t, d, runID, "vol-b")

	cfg := &config.Config{}
	e := analyze.New(cfg)
	if err := e.Run(context.Background(), d, cfg, runID, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM findings WHERE audit_run_id = ?", runID).Scan(&count)
	if count < 2 {
		t.Errorf("expected at least 2 findings, got %d", count)
	}
}
