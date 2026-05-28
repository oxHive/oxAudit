package ingest_test

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
	. "github.com/graditya/oxaudit/internal/ingest"
)

// setupDB opens an in-memory SQLite database and inserts a test audit_run.
func setupDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	const runID = "test-run"
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = d.Exec(`INSERT INTO audit_run (id, executed_at, generated_at, period_start, period_end, aws_profile, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, runID, now, now, "2024-01-01", "2024-01-31", "default", "complete")
	if err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
	return d, runID
}

func ingestFile(t *testing.T, d *sql.DB, runID string, ing Ingester, filePath string) error {
	t.Helper()
	tx, err := d.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := ing.Ingest(context.Background(), tx, filePath, runID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func writeTempJSON(t *testing.T, dir, name string, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write JSON file: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Helpers (via Matches)
// ---------------------------------------------------------------------------

func TestMatches(t *testing.T) {
	cases := []struct {
		ingester Ingester
		match    string
		nomatch  string
	}{
		{&EBSVolumesIngester{}, "ec2_describe-volumes_us-east-1.json", "ec2_describe-instances.json"},
		{&EBSSnapshotsIngester{}, "ec2_describe-snapshots_us-east-1.json", "ec2_describe-volumes.json"},
		{&ElasticIPsIngester{}, "ec2_describe-addresses_us-east-1.json", "ec2_describe-volumes.json"},
		{&EC2InstancesIngester{}, "ec2_describe-instances_us-east-1.json", "ec2_describe-volumes.json"},
		{&NATGatewaysIngester{}, "ec2_describe-nat-gateways_us-east-1.json", "ec2_describe-volumes.json"},
		{&RDSInstancesIngester{}, "rds_describe-db-instances_us-east-1.json", "ec2_describe-volumes.json"},
		{&CWLogGroupsIngester{}, "logs_describe-log-groups_us-east-1.json", "ec2_describe-volumes.json"},
		{&AccountsIngester{}, "org_list-accounts.json", "ec2_describe-volumes.json"},
		{&CostMonthlyIngester{}, "ce_monthly-by-service.json", "ec2_describe-volumes.json"},
		{&CostDailyIngester{}, "ce_daily-by-service.json", "ec2_describe-volumes.json"},
	}
	for _, tc := range cases {
		if !tc.ingester.Matches(tc.match) {
			t.Errorf("%T.Matches(%q) = false, want true", tc.ingester, tc.match)
		}
		if tc.ingester.Matches(tc.nomatch) {
			t.Errorf("%T.Matches(%q) = true, want false", tc.ingester, tc.nomatch)
		}
	}
}

// ---------------------------------------------------------------------------
// EBSVolumesIngester
// ---------------------------------------------------------------------------

func TestEBSVolumesIngester_ingestsVolumes(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"Volumes": []any{
			map[string]any{
				"VolumeId":   "vol-abc123",
				"State":      "available",
				"Size":       100,
				"VolumeType": "gp2",
				"Iops":       0,
				"CreateTime": "2024-01-01T00:00:00Z",
				"Tags":       []any{map[string]any{"Key": "Name", "Value": "my-vol"}},
			},
		},
	}
	path := writeTempJSON(t, dir, "ec2_describe-volumes_us-east-1.json", payload)

	ing := &EBSVolumesIngester{}
	if err := ingestFile(t, d, runID, ing, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:volume'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 volume resource, got %d", count)
	}
}

func TestEBSVolumesIngester_emptyList(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	path := writeTempJSON(t, dir, "ec2_describe-volumes_us-east-1.json", map[string]any{"Volumes": []any{}})
	if err := ingestFile(t, d, runID, &EBSVolumesIngester{}, path); err != nil {
		t.Fatalf("Ingest empty list: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ElasticIPsIngester
// ---------------------------------------------------------------------------

func TestElasticIPsIngester_ingestsUnassociated(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"Addresses": []any{
			map[string]any{
				"AllocationId": "eipalloc-abc",
				"PublicIp":     "1.2.3.4",
				"Tags":         []any{},
			},
		},
	}
	path := writeTempJSON(t, dir, "ec2_describe-addresses_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &ElasticIPsIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:eip'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 EIP resource, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// EC2InstancesIngester
// ---------------------------------------------------------------------------

func TestEC2InstancesIngester_ingestsInstance(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"Reservations": []any{
			map[string]any{
				"Instances": []any{
					map[string]any{
						"InstanceId":   "i-abc123",
						"InstanceType": "t3.medium",
						"State":        map[string]any{"Name": "stopped"},
						"LaunchTime":   "2024-01-01T00:00:00Z",
						"Tags":         []any{map[string]any{"Key": "Name", "Value": "test"}},
					},
				},
			},
		},
	}
	path := writeTempJSON(t, dir, "ec2_describe-instances_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &EC2InstancesIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:instance'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 EC2 instance resource, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// EBSSnapshotsIngester
// ---------------------------------------------------------------------------

func TestEBSSnapshotsIngester_ingestsSnapshot(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"Snapshots": []any{
			map[string]any{
				"SnapshotId":  "snap-abc123",
				"State":       "completed",
				"VolumeSize":  200,
				"StartTime":   "2023-01-01T00:00:00Z",
				"Tags":        []any{},
			},
		},
	}
	path := writeTempJSON(t, dir, "ec2_describe-snapshots_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &EBSSnapshotsIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:snapshot'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 snapshot resource, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// AccountsIngester
// ---------------------------------------------------------------------------

func TestAccountsIngester_ingestsAccounts(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"Accounts": []any{
			map[string]any{"Id": "111111111111", "Name": "prod", "Email": "prod@example.com", "Status": "ACTIVE"},
			map[string]any{"Id": "222222222222", "Name": "dev", "Email": "dev@example.com", "Status": "ACTIVE"},
		},
	}
	path := writeTempJSON(t, dir, "org_list-accounts.json", payload)

	if err := ingestFile(t, d, runID, &AccountsIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM accounts WHERE audit_run_id = ?", runID).Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 accounts, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// NATGatewaysIngester
// ---------------------------------------------------------------------------

func TestNATGatewaysIngester_ingestsGateway(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"NatGateways": []any{
			map[string]any{
				"NatGatewayId": "nat-abc123",
				"State":        "available",
				"CreateTime":   "2024-01-01T00:00:00Z",
				"Tags":         []any{},
			},
		},
	}
	path := writeTempJSON(t, dir, "ec2_describe-nat-gateways_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &NATGatewaysIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:nat-gateway'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 NAT gateway resource, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// CostMonthlyIngester
// ---------------------------------------------------------------------------

func TestCostMonthlyIngester_ingests(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"ResultsByTime": []any{
			map[string]any{
				"TimePeriod": map[string]any{"Start": "2024-01-01", "End": "2024-01-31"},
				"Groups": []any{
					map[string]any{
						"Keys": []any{"Amazon EC2", "us-east-1"},
						"Metrics": map[string]any{
							"UnblendedCost": map[string]any{"Amount": "42.50", "Unit": "USD"},
						},
					},
				},
			},
		},
	}
	path := writeTempJSON(t, dir, "ce_cost-and-usage-monthly.json", payload)

	if err := ingestFile(t, d, runID, &CostMonthlyIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

func TestRegistry_ingestDir(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	// Write a volumes file that the registry should dispatch
	payload := map[string]any{
		"Volumes": []any{
			map[string]any{
				"VolumeId":   "vol-reg1",
				"State":      "available",
				"Size":       50,
				"VolumeType": "gp3",
				"Tags":       []any{},
			},
		},
	}
	writeTempJSON(t, dir, "ec2_describe-volumes_us-east-1.json", payload)

	reg := NewRegistry()
	var processed []string
	err := reg.IngestDir(context.Background(), d, dir, runID, func(name string, rows int, err error) {
		processed = append(processed, name)
	})
	if err != nil {
		t.Fatalf("IngestDir: %v", err)
	}
	if len(processed) == 0 {
		t.Error("expected at least one file processed")
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:ec2:volume'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 volume in DB, got %d", count)
	}
}

func TestRegistry_ingestDir_emptyDir(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	reg := NewRegistry()
	if err := reg.IngestDir(context.Background(), d, dir, runID, nil); err != nil {
		t.Fatalf("IngestDir empty dir: %v", err)
	}
}

func TestRegistry_ingestDir_unknownFile(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	// Write a JSON file that no ingester matches — should be ignored silently
	writeTempJSON(t, dir, "unknown_service_command.json", map[string]any{"foo": "bar"})

	reg := NewRegistry()
	if err := reg.IngestDir(context.Background(), d, dir, runID, nil); err != nil {
		t.Fatalf("IngestDir with unknown file: %v", err)
	}
}
