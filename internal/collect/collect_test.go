package collect

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/db"
	"github.com/graditya/oxaudit/internal/types"
)

func setupDB(t *testing.T, runID string) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	now := time.Now().UTC().Format(time.RFC3339)
	d.Exec(`INSERT INTO audit_run (id, executed_at, generated_at, period_start, period_end, aws_profile, status) VALUES (?,?,?,?,?,?,?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "running")
	return d
}

func TestInsertRawFiles_insertsRows(t *testing.T) {
	const runID = "run-rf-001"
	d := setupDB(t, runID)

	now := time.Now().UTC()
	files := []types.RawFile{
		{AuditRunID: runID, CommandName: "ec2_describe-volumes", Service: "ec2", Region: "us-east-1", FilePath: "/tmp/f.json", Status: "ok", CollectedAt: now},
		{AuditRunID: runID, CommandName: "ec2_describe-instances", Service: "ec2", Region: "us-west-2", FilePath: "/tmp/g.json", Status: "ok", CollectedAt: now},
	}

	if err := insertRawFiles(context.Background(), d, files); err != nil {
		t.Fatalf("insertRawFiles: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM raw_file WHERE audit_run_id = ?", runID).Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 raw_file rows, got %d", count)
	}
}

func TestInsertRawFiles_emptySlice(t *testing.T) {
	const runID = "run-rf-002"
	d := setupDB(t, runID)

	if err := insertRawFiles(context.Background(), d, nil); err != nil {
		t.Fatalf("insertRawFiles with empty slice: %v", err)
	}
}

// ---------------------------------------------------------------------------
// filterExcluded
// ---------------------------------------------------------------------------

func TestFilterExcluded_removesExcluded(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	exclude := []string{"us-west-2"}
	got := filterExcluded(regions, exclude)
	if len(got) != 2 {
		t.Fatalf("expected 2 regions, got %d: %v", len(got), got)
	}
	for _, r := range got {
		if r == "us-west-2" {
			t.Error("expected us-west-2 to be excluded")
		}
	}
}

func TestFilterExcluded_nothingExcluded(t *testing.T) {
	regions := []string{"us-east-1", "eu-west-1"}
	got := filterExcluded(regions, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 regions, got %v", got)
	}
}

func TestFilterExcluded_allExcluded(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2"}
	got := filterExcluded(regions, []string{"us-east-1", "us-west-2"})
	if len(got) != 0 {
		t.Fatalf("expected 0 regions, got %v", got)
	}
}

func TestFilterExcluded_emptyInput(t *testing.T) {
	got := filterExcluded(nil, []string{"us-east-1"})
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// parseRegions
// ---------------------------------------------------------------------------

func writeRegionsJSON(t *testing.T, regions []string) string {
	t.Helper()
	type regEntry struct {
		RegionName string `json:"RegionName"`
	}
	type payload struct {
		Regions []regEntry `json:"Regions"`
	}
	p := payload{}
	for _, r := range regions {
		p.Regions = append(p.Regions, regEntry{RegionName: r})
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(t.TempDir(), "ec2_describe-regions.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestParseRegions_returnsRegions(t *testing.T) {
	path := writeRegionsJSON(t, []string{"us-east-1", "eu-west-1", "ap-southeast-1"})
	got, err := parseRegions(path)
	if err != nil {
		t.Fatalf("parseRegions: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 regions, got %d: %v", len(got), got)
	}
}

func TestParseRegions_emptyList(t *testing.T) {
	path := writeRegionsJSON(t, nil)
	got, err := parseRegions(path)
	if err != nil {
		t.Fatalf("parseRegions: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %v", got)
	}
}

func TestParseRegions_missingFile(t *testing.T) {
	_, err := parseRegions("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseRegions_invalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "regions.json")
	os.WriteFile(path, []byte("{invalid"), 0644)
	_, err := parseRegions(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
