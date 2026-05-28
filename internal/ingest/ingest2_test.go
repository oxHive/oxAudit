package ingest_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/graditya/oxaudit/internal/ingest"
)

// ---------------------------------------------------------------------------
// RDSInstancesIngester
// ---------------------------------------------------------------------------

func TestRDSInstancesIngester_ingestsInstance(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"DBInstances": []any{
			map[string]any{
				"DBInstanceIdentifier": "mydb",
				"DBInstanceClass":      "db.t3.medium",
				"DBInstanceStatus":     "available",
				"Engine":               "mysql",
				"AllocatedStorage":     100,
				"InstanceCreateTime":   "2024-01-01T00:00:00Z",
				"TagList":              []any{},
			},
		},
	}
	path := writeTempJSON(t, dir, "rds_describe-db-instances_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &RDSInstancesIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:rds:instance'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 RDS instance, got %d", count)
	}
}

func TestRDSInstancesIngester_emptyList(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	path := writeTempJSON(t, dir, "rds_describe-db-instances_us-east-1.json", map[string]any{"DBInstances": []any{}})
	if err := ingestFile(t, d, runID, &RDSInstancesIngester{}, path); err != nil {
		t.Fatalf("Ingest empty: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CWLogGroupsIngester
// ---------------------------------------------------------------------------

func TestCWLogGroupsIngester_ingestsLogGroup(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"logGroups": []any{
			map[string]any{
				"logGroupName":    "/aws/lambda/my-function",
				"creationTime":    1704067200000,
				"retentionInDays": 30,
				"storedBytes":     1024.0,
			},
		},
	}
	path := writeTempJSON(t, dir, "logs_describe-log-groups_us-east-1.json", payload)

	if err := ingestFile(t, d, runID, &CWLogGroupsIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM resources WHERE resource_type = 'aws:logs:log-group'").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 log group, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// CostDailyIngester
// ---------------------------------------------------------------------------

func TestCostDailyIngester_ingestsRows(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"ResultsByTime": []any{
			map[string]any{
				"TimePeriod": map[string]any{"Start": "2024-01-15", "End": "2024-01-16"},
				"Groups": []any{
					map[string]any{
						"Keys": []any{"Amazon EC2", "us-east-1"},
						"Metrics": map[string]any{
							"UnblendedCost": map[string]any{"Amount": "12.50", "Unit": "USD"},
							"AmortizedCost": map[string]any{"Amount": "12.50", "Unit": "USD"},
						},
					},
				},
			},
		},
	}
	path := writeTempJSON(t, dir, "ce_daily-by-service.json", payload)

	if err := ingestFile(t, d, runID, &CostDailyIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM cost_daily WHERE audit_run_id = ?", runID).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 cost_daily row, got %d", count)
	}
}

func TestCostDailyIngester_emptyResults(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	path := writeTempJSON(t, dir, "ce_daily-by-service.json", map[string]any{"ResultsByTime": []any{}})
	if err := ingestFile(t, d, runID, &CostDailyIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CostMonthlyIngester
// ---------------------------------------------------------------------------

func TestCostMonthlyIngester_ingestsRows(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	payload := map[string]any{
		"ResultsByTime": []any{
			map[string]any{
				"TimePeriod": map[string]any{"Start": "2024-01-01", "End": "2024-01-31"},
				"Groups": []any{
					map[string]any{
						"Keys": []any{"Amazon EC2"},
						"Metrics": map[string]any{
							"UnblendedCost": map[string]any{"Amount": "350.00", "Unit": "USD"},
							"AmortizedCost": map[string]any{"Amount": "350.00", "Unit": "USD"},
							"UsageQuantity": map[string]any{"Amount": "720", "Unit": "Hrs"},
						},
					},
				},
			},
		},
	}
	path := writeTempJSON(t, dir, "ce_monthly-by-service.json", payload)

	if err := ingestFile(t, d, runID, &CostMonthlyIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	var count int
	d.QueryRow("SELECT COUNT(*) FROM cost_monthly WHERE audit_run_id = ?", runID).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 cost_monthly row, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// NATGatewaysIngester (additional)
// ---------------------------------------------------------------------------

func TestNATGatewaysIngester_emptyList(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	path := writeTempJSON(t, dir, "ec2_describe-nat-gateways_us-east-1.json", map[string]any{"NatGateways": []any{}})
	if err := ingestFile(t, d, runID, &NATGatewaysIngester{}, path); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Registry error handling
// ---------------------------------------------------------------------------

func TestRegistry_ingestDir_invalidJSONFile(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	// Write a snapshots file with invalid JSON
	badPath := filepath.Join(dir, "ec2_describe-snapshots_us-east-1.json")
	if err := os.WriteFile(badPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	// Also write a valid volumes file to confirm walk continues
	writeTempJSON(t, dir, "ec2_describe-volumes_us-east-1.json", map[string]any{"Volumes": []any{}})

	reg := NewRegistry()
	var errCount int
	err := reg.IngestDir(context.Background(), d, dir, runID, func(name string, rows int, err error) {
		if err != nil {
			errCount++
		}
	})
	if err != nil {
		t.Fatalf("IngestDir should not return error on bad file: %v", err)
	}
	// The bad file should have triggered the error callback
	if errCount == 0 {
		t.Error("expected at least one error callback for invalid JSON file")
	}
}

func TestRegistry_ingestDir_nonJSONFilesIgnored(t *testing.T) {
	d, runID := setupDB(t)
	dir := t.TempDir()

	// Write non-JSON files — should all be ignored
	for _, name := range []string{"file.txt", "file.csv", "file.yaml"} {
		os.WriteFile(filepath.Join(dir, name), []byte("content"), 0644)
	}

	reg := NewRegistry()
	var processed []string
	err := reg.IngestDir(context.Background(), d, dir, runID, func(name string, rows int, err error) {
		processed = append(processed, name)
	})
	if err != nil {
		t.Fatalf("IngestDir: %v", err)
	}
	if len(processed) != 0 {
		t.Errorf("expected no files processed, got %v", processed)
	}
}
