package export_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/db"
	"github.com/graditya/oxaudit/internal/export"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

func setupExportDB(t *testing.T, runID string) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := d.Exec(`INSERT INTO audit_run
		(id, executed_at, generated_at, period_start, period_end, aws_profile, billing_region, run_folder, notes, status)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "us-east-1", "", "", "complete"); err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
	return d
}

func insertFinding(t *testing.T, d *sql.DB, runID, priority string, savings float64) {
	t.Helper()
	f := types.Finding{
		ID:                   findutil.GenerateFindingID("TEST-001", "111", "us-east-1", []string{"r-" + priority}),
		AuditRunID:           runID,
		Priority:             priority,
		Category:             "Waste",
		Service:              "Amazon EC2",
		AccountID:            "111",
		AccountName:          "test-acct",
		Region:               "us-east-1",
		Title:                "Test Finding " + priority,
		Summary:              "Test summary",
		Evidence:             "Test evidence",
		RecommendedAction:    "Do something",
		EstMonthlySavingsUSD: savings,
		Confidence:           "High",
		Risk:                 "Low",
		Status:               "open",
		ResourceIDsJSON:      `["r-` + priority + `"]`,
		TagsJSON:             `{}`,
		SourceFilesJSON:      `["test.json"]`,
	}
	if err := findutil.UpsertFinding(context.Background(), d, f); err != nil {
		t.Fatalf("insert finding: %v", err)
	}
}

func TestRun_createsAllOutputFiles(t *testing.T) {
	const runID = "run-export-001"
	d := setupExportDB(t, runID)
	insertFinding(t, d, runID, "P1", 100.0)
	insertFinding(t, d, runID, "P2", 50.0)

	dir := t.TempDir()
	err := export.Run(context.Background(), d, runID, dir, nil)
	if err != nil {
		t.Fatalf("export.Run: %v", err)
	}

	expected := []string{
		"findings.jsonl",
		"findings.csv",
		"findings.md",
		"llm-digest.md",
		"remediation-backlog.md",
	}
	for _, name := range expected {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected output file %q to be created", name)
		}
	}
}

func TestRun_callbackInvokedForEachExporter(t *testing.T) {
	const runID = "run-export-002"
	d := setupExportDB(t, runID)

	dir := t.TempDir()
	var called []string
	err := export.Run(context.Background(), d, runID, dir, func(name, path string, err error) {
		called = append(called, name)
		if err != nil {
			t.Errorf("exporter %s: %v", name, err)
		}
	})
	if err != nil {
		t.Fatalf("export.Run: %v", err)
	}
	if len(called) == 0 {
		t.Error("expected callback to be called at least once")
	}
}

func TestRun_emptyFindings(t *testing.T) {
	const runID = "run-export-003"
	d := setupExportDB(t, runID)

	dir := t.TempDir()
	if err := export.Run(context.Background(), d, runID, dir, nil); err != nil {
		t.Fatalf("export.Run with no findings: %v", err)
	}
}

func TestRun_jsonlContainsFindings(t *testing.T) {
	const runID = "run-export-004"
	d := setupExportDB(t, runID)
	insertFinding(t, d, runID, "P0", 999.0)

	dir := t.TempDir()
	if err := export.Run(context.Background(), d, runID, dir, nil); err != nil {
		t.Fatalf("export.Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "findings.jsonl"))
	if err != nil {
		t.Fatalf("read findings.jsonl: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected findings.jsonl to be non-empty")
	}
}

func TestRun_csvContainsHeader(t *testing.T) {
	const runID = "run-export-005"
	d := setupExportDB(t, runID)

	dir := t.TempDir()
	if err := export.Run(context.Background(), d, runID, dir, nil); err != nil {
		t.Fatalf("export.Run: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "findings.csv"))
	if len(data) == 0 {
		t.Error("expected findings.csv to have at least a header")
	}
}

func TestRun_multiplePriorities(t *testing.T) {
	const runID = "run-export-006"
	d := setupExportDB(t, runID)
	insertFinding(t, d, runID, "P0", 500.0)
	insertFinding(t, d, runID, "P1", 200.0)
	insertFinding(t, d, runID, "P3", 10.0)

	dir := t.TempDir()
	if err := export.Run(context.Background(), d, runID, dir, nil); err != nil {
		t.Fatalf("export.Run: %v", err)
	}

	// Remediation backlog should have content for multiple priorities
	data, _ := os.ReadFile(filepath.Join(dir, "remediation-backlog.md"))
	if len(data) == 0 {
		t.Error("expected non-empty remediation-backlog.md")
	}
}
