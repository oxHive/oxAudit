package findutil_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/db"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

func TestGenerateFindingID_deterministic(t *testing.T) {
	id1 := findutil.GenerateFindingID("AWS-WASTE-001", "123456789", "us-east-1", []string{"vol-abc"})
	id2 := findutil.GenerateFindingID("AWS-WASTE-001", "123456789", "us-east-1", []string{"vol-abc"})
	if id1 != id2 {
		t.Fatalf("expected same ID, got %q and %q", id1, id2)
	}
}

func TestGenerateFindingID_prefix(t *testing.T) {
	id := findutil.GenerateFindingID("AWS-WASTE-001", "123456789", "us-east-1", []string{"vol-abc"})
	if len(id) < 4 || id[:4] != "FND-" {
		t.Fatalf("expected FND- prefix, got %q", id)
	}
}

func TestGenerateFindingID_differentInputs(t *testing.T) {
	id1 := findutil.GenerateFindingID("AWS-WASTE-001", "111", "us-east-1", []string{"vol-aaa"})
	id2 := findutil.GenerateFindingID("AWS-WASTE-001", "222", "us-east-1", []string{"vol-bbb"})
	if id1 == id2 {
		t.Fatal("expected different IDs for different inputs")
	}
}

func TestGenerateFindingID_resourceOrderIndependent(t *testing.T) {
	id1 := findutil.GenerateFindingID("rule", "acct", "us-east-1", []string{"r1", "r2", "r3"})
	id2 := findutil.GenerateFindingID("rule", "acct", "us-east-1", []string{"r3", "r1", "r2"})
	if id1 != id2 {
		t.Fatalf("expected same ID regardless of resource order, got %q and %q", id1, id2)
	}
}

func TestJSONArray_empty(t *testing.T) {
	got := findutil.JSONArray([]string{})
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestJSONArray_single(t *testing.T) {
	got := findutil.JSONArray([]string{"vol-abc"})
	if got != `["vol-abc"]` {
		t.Fatalf("expected [\"vol-abc\"], got %q", got)
	}
}

func TestJSONArray_multiple(t *testing.T) {
	got := findutil.JSONArray([]string{"a", "b", "c"})
	if got != `["a","b","c"]` {
		t.Fatalf("expected [\"a\",\"b\",\"c\"], got %q", got)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func insertTestAuditRun(t *testing.T, d *sql.DB, runID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.Exec(`INSERT INTO audit_run (id, executed_at, generated_at, period_start, period_end, aws_profile, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "complete")
	if err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
}

func TestUpsertFinding_insertsAndReads(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-001"
	insertTestAuditRun(t, d, runID)

	f := types.Finding{
		ID:                   "FND-test0001",
		AuditRunID:           runID,
		Priority:             "P1",
		Category:             "Waste",
		Service:              "Amazon EC2",
		AccountID:            "123456789",
		AccountName:          "test-account",
		Region:               "us-east-1",
		Title:                "Test Finding",
		Summary:              "Test summary",
		Evidence:             "Test evidence",
		RecommendedAction:    "Do something",
		EstMonthlySavingsUSD: 42.0,
		Confidence:           "High",
		Risk:                 "Low",
		Status:               "open",
		ResourceIDsJSON:      `["vol-abc"]`,
		TagsJSON:             `{"Env":"prod"}`,
		SourceFilesJSON:      `["file.json"]`,
	}

	if err := findutil.UpsertFinding(context.Background(), d, f); err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM findings WHERE id = ?", f.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 finding, got %d", count)
	}
}

func TestUpsertFinding_upsertReplace(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-001"
	insertTestAuditRun(t, d, runID)

	f := types.Finding{
		ID:         "FND-dupe",
		AuditRunID: runID,
		Priority:   "P2",
		Status:     "open",
	}

	if err := findutil.UpsertFinding(context.Background(), d, f); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	f.Priority = "P1"
	if err := findutil.UpsertFinding(context.Background(), d, f); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var priority string
	d.QueryRow("SELECT priority FROM findings WHERE id = ?", f.ID).Scan(&priority)
	if priority != "P1" {
		t.Fatalf("expected P1 after upsert, got %q", priority)
	}
}

func TestUpsertFinding_emptyJSONDefaults(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-002"
	insertTestAuditRun(t, d, runID)

	f := types.Finding{
		ID:         "FND-empty",
		AuditRunID: runID,
		Priority:   "P3",
		Status:     "open",
		// no JSON fields set
	}

	if err := findutil.UpsertFinding(context.Background(), d, f); err != nil {
		t.Fatalf("UpsertFinding with empty JSON fields: %v", err)
	}
}
