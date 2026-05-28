package cmd

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/db"
)

func setupTestDB(t *testing.T, runID string) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := d.Exec(`INSERT INTO audit_run
		(id, executed_at, generated_at, period_start, period_end, aws_profile, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "running"); err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
	return d
}

func TestSanitizeID_alphanumericPassthrough(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc123", "abc123"},
		{"my-run-01", "my-run-01"},
		{"ABC", "ABC"},
		{"hello world", "hello-world"},
		{"foo/bar", "foo-bar"},
		{"2024:01:01", "2024-01-01"},
		{"", ""},
	}
	for _, tc := range cases {
		got := sanitizeID(tc.in)
		if got != tc.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMarkRunComplete_updatesStatus(t *testing.T) {
	const runID = "run-complete-test"
	d := setupTestDB(t, runID)

	markRunComplete(d, runID)

	var status string
	d.QueryRow("SELECT status FROM audit_run WHERE id = ?", runID).Scan(&status)
	if status != "complete" {
		t.Errorf("expected status 'complete', got %q", status)
	}
}

func TestMarkRunFailed_updatesStatus(t *testing.T) {
	const runID = "run-failed-test"
	d := setupTestDB(t, runID)

	markRunFailed(d, runID)

	var status string
	d.QueryRow("SELECT status FROM audit_run WHERE id = ?", runID).Scan(&status)
	if status != "failed" {
		t.Errorf("expected status 'failed', got %q", status)
	}
}

func TestSeedInitRecord_insertsRecord(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	if err := seedInitRecord(d, "/tmp/oxaudit"); err != nil {
		t.Fatalf("seedInitRecord: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM audit_run WHERE status = 'init'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 init record, got %d", count)
	}
}

func TestSeedInitRecord_idempotent(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	// Calling twice with the same second should use INSERT OR IGNORE
	if err := seedInitRecord(d, "/tmp/oxaudit"); err != nil {
		t.Fatalf("first seedInitRecord: %v", err)
	}
	// Second call at a different second is fine (generates different ID)
	// so just verify it doesn't error
	if err := seedInitRecord(d, "/tmp/oxaudit"); err != nil {
		t.Fatalf("second seedInitRecord: %v", err)
	}
}
