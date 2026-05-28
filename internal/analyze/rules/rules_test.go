package rules_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/db"
	. "github.com/graditya/oxaudit/internal/analyze/rules"
)

// openTestDB returns an in-memory SQLite database with all migrations applied.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func insertAuditRun(t *testing.T, d *sql.DB, runID string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.Exec(`INSERT INTO audit_run (id, executed_at, generated_at, period_start, period_end, aws_profile, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID, now, now, "2024-01-01", "2024-01-31", "default", "complete")
	if err != nil {
		t.Fatalf("insert audit_run: %v", err)
	}
}

func insertResource(t *testing.T, d *sql.DB, runID, resourceID, resourceType, state, region, accountID, accountName string, extras map[string]any) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	sizeGiB := 0.0
	volumeType := ""
	iops := 0
	instanceType := ""
	createdAt := now
	tagsJSON := "{}"
	sourceFile := "test.json"

	if v, ok := extras["size_gib"]; ok {
		sizeGiB = v.(float64)
	}
	if v, ok := extras["volume_type"]; ok {
		volumeType = v.(string)
	}
	if v, ok := extras["iops"]; ok {
		iops = v.(int)
	}
	if v, ok := extras["instance_type"]; ok {
		instanceType = v.(string)
	}
	if v, ok := extras["created_at"]; ok {
		createdAt = v.(string)
	}
	if v, ok := extras["tags_json"]; ok {
		tagsJSON = v.(string)
	}
	if v, ok := extras["source_file"]; ok {
		sourceFile = v.(string)
	}

	_, err := d.Exec(`INSERT INTO resources
		(audit_run_id, resource_id, resource_type, account_id, account_name, region,
		 state, size_gib, volume_type, iops, instance_type, created_at, tags_json, source_file, discovered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, resourceID, resourceType, accountID, accountName, region,
		state, sizeGiB, volumeType, iops, instanceType, createdAt, tagsJSON, sourceFile, now)
	if err != nil {
		t.Fatalf("insert resource %s: %v", resourceID, err)
	}
}

func defaultCfg() *config.Config {
	return &config.Config{}
}

// ---------------------------------------------------------------------------
// UnattachedEBS
// ---------------------------------------------------------------------------

func TestUnattachedEBSRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-001"
	insertAuditRun(t, d, runID)

	rule := NewUnattachedEBSRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestUnattachedEBSRule_findsAvailableVolume(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-002"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "vol-aaa", "aws:ec2:volume", "available", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":    100.0,
		"volume_type": "gp2",
	})

	rule := NewUnattachedEBSRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Category != "Waste" {
		t.Errorf("expected Waste category, got %q", f.Category)
	}
	if len(f.ID) == 0 {
		t.Error("expected non-empty finding ID")
	}
}

func TestUnattachedEBSRule_skipsInUseVolume(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-003"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "vol-bbb", "aws:ec2:volume", "in-use", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":    200.0,
		"volume_type": "gp3",
	})

	rule := NewUnattachedEBSRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for in-use volume, got %d", len(findings))
	}
}

func TestUnattachedEBSRule_p1ForLargeVolume(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-004"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "vol-big", "aws:ec2:volume", "available", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":    600.0,
		"volume_type": "gp2",
	})

	rule := NewUnattachedEBSRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Priority != "P1" {
		t.Errorf("expected P1 for large volume, got %q", findings[0].Priority)
	}
}

func TestUnattachedEBSRule_filtersBelowMinSavings(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-005"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "vol-tiny", "aws:ec2:volume", "available", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":    1.0,
		"volume_type": "sc1",
	})

	cfg := defaultCfg()
	cfg.Analysis.MinMonthlySavingsUSD = 1000 // far above the tiny volume's cost
	rule := NewUnattachedEBSRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings below min savings, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// OldSnapshots
// ---------------------------------------------------------------------------

func TestOldSnapshotsRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-010"
	insertAuditRun(t, d, runID)

	rule := NewOldSnapshotsRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestOldSnapshotsRule_findsOldSnapshot(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-011"
	insertAuditRun(t, d, runID)

	oldDate := time.Now().AddDate(0, 0, -120).UTC().Format(time.RFC3339)
	insertResource(t, d, runID, "snap-old", "aws:ec2:snapshot", "completed", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":   200.0,
		"created_at": oldDate,
	})

	rule := NewOldSnapshotsRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Category != "Waste" {
		t.Errorf("expected Waste, got %q", findings[0].Category)
	}
}

func TestOldSnapshotsRule_skipsRecentSnapshot(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-012"
	insertAuditRun(t, d, runID)

	recentDate := time.Now().AddDate(0, 0, -10).UTC().Format(time.RFC3339)
	insertResource(t, d, runID, "snap-new", "aws:ec2:snapshot", "completed", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":   200.0,
		"created_at": recentDate,
	})

	rule := NewOldSnapshotsRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for recent snapshot, got %d", len(findings))
	}
}

func TestOldSnapshotsRule_usesDefaultThreshold(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-013"
	insertAuditRun(t, d, runID)

	// Threshold=0 should default to 90 days; insert a 95-day-old snapshot
	oldDate := time.Now().AddDate(0, 0, -95).UTC().Format(time.RFC3339)
	insertResource(t, d, runID, "snap-95d", "aws:ec2:snapshot", "completed", "us-east-1", "111", "acct-a", map[string]any{
		"size_gib":   100.0,
		"created_at": oldDate,
	})

	cfg := defaultCfg()
	cfg.Analysis.OldSnapshotThresholdDays = 0 // should use default of 90
	rule := NewOldSnapshotsRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding with default threshold, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// UnassociatedEIP
// ---------------------------------------------------------------------------

func TestUnassociatedEIPRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-020"
	insertAuditRun(t, d, runID)

	rule := NewUnassociatedEIPRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestUnassociatedEIPRule_findsUnassociatedEIP(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-021"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "eipalloc-abc", "aws:ec2:eip", "unassociated", "us-east-1", "111", "acct-a", map[string]any{})

	rule := NewUnassociatedEIPRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Priority != "P1" {
		t.Errorf("expected P1, got %q", f.Priority)
	}
	if f.EstMonthlySavingsUSD != 3.65 {
		t.Errorf("expected $3.65 savings, got %f", f.EstMonthlySavingsUSD)
	}
}

func TestUnassociatedEIPRule_skipsAssociated(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-022"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "eipalloc-xyz", "aws:ec2:eip", "associated", "us-east-1", "111", "acct-a", map[string]any{})

	rule := NewUnassociatedEIPRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for associated EIP, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// StoppedEC2
// ---------------------------------------------------------------------------

func TestStoppedEC2Rule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-030"
	insertAuditRun(t, d, runID)

	rule := NewStoppedEC2Rule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestStoppedEC2Rule_findsStoppedInstance(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-031"
	insertAuditRun(t, d, runID)

	launchTime := time.Now().AddDate(0, 0, -45).UTC().Format(time.RFC3339)
	insertResource(t, d, runID, "i-abc123", "aws:ec2:instance", "stopped", "us-east-1", "111", "acct-a", map[string]any{
		"instance_type": "t3.medium",
		"created_at":    launchTime,
	})

	rule := NewStoppedEC2Rule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Category != "Waste" {
		t.Errorf("expected Waste, got %q", findings[0].Category)
	}
}

func TestStoppedEC2Rule_skipsRunningInstance(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-032"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "i-running", "aws:ec2:instance", "running", "us-east-1", "111", "acct-a", map[string]any{
		"instance_type": "t3.large",
	})

	rule := NewStoppedEC2Rule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for running instance, got %d", len(findings))
	}
}

func TestStoppedEC2Rule_defaultThreshold(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-033"
	insertAuditRun(t, d, runID)

	launchTime := time.Now().AddDate(0, 0, -60).UTC().Format(time.RFC3339)
	insertResource(t, d, runID, "i-old", "aws:ec2:instance", "stopped", "us-east-1", "111", "acct-a", map[string]any{
		"instance_type": "m5.xlarge",
		"created_at":    launchTime,
	})

	cfg := defaultCfg()
	cfg.Analysis.StoppedEC2ThresholdDays = 0 // defaults to 30
	rule := NewStoppedEC2Rule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// MissingTags
// ---------------------------------------------------------------------------

func TestMissingTagsRule_noRequiredTags(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-040"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "i-notag", "aws:ec2:instance", "running", "us-east-1", "111", "acct-a", map[string]any{})

	cfg := defaultCfg()
	// RequiredTags is empty — rule should return nil
	rule := NewMissingTagsRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when RequiredTags empty, got %d", len(findings))
	}
}

func TestMissingTagsRule_findsMissingTag(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-041"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "i-notag", "aws:ec2:instance", "running", "us-east-1", "111", "acct-a", map[string]any{
		"tags_json": `{"Env":"prod"}`,
	})

	cfg := defaultCfg()
	cfg.Analysis.RequiredTags = []string{"Env", "Owner", "CostCenter"}
	rule := NewMissingTagsRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Category != "Tagging" {
		t.Errorf("expected Tagging category, got %q", findings[0].Category)
	}
}

func TestMissingTagsRule_skipsCompliantResource(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-042"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "i-tagged", "aws:ec2:instance", "running", "us-east-1", "111", "acct-a", map[string]any{
		"tags_json": `{"Env":"prod","Owner":"team-a","CostCenter":"eng"}`,
	})

	cfg := defaultCfg()
	cfg.Analysis.RequiredTags = []string{"Env", "Owner", "CostCenter"}
	rule := NewMissingTagsRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for compliant resource, got %d", len(findings))
	}
}

func TestMissingTagsRule_rdsServiceLabel(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-043"
	insertAuditRun(t, d, runID)

	insertResource(t, d, runID, "db-notag", "aws:rds:instance", "available", "us-east-1", "111", "acct-a", map[string]any{
		"tags_json": `{}`,
	})

	cfg := defaultCfg()
	cfg.Analysis.RequiredTags = []string{"Owner"}
	rule := NewMissingTagsRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Service != "Amazon Relational Database Service" {
		t.Errorf("expected RDS service label, got %q", findings[0].Service)
	}
}

// ---------------------------------------------------------------------------
// HighNATCost
// ---------------------------------------------------------------------------

func insertCostMonthly(t *testing.T, d *sql.DB, runID, month, accountID, accountName, region, service string, cost float64) {
	t.Helper()
	_, err := d.Exec(`INSERT INTO cost_monthly
		(audit_run_id, month, account_id, account_name, region, service, unblended_cost)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID, month, accountID, accountName, region, service, cost)
	if err != nil {
		t.Fatalf("insert cost_monthly: %v", err)
	}
}

func TestHighNATCostRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-050"
	insertAuditRun(t, d, runID)

	rule := NewHighNATCostRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestHighNATCostRule_findsHighCost(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-051"
	insertAuditRun(t, d, runID)

	insertCostMonthly(t, d, runID, "2024-01", "111", "acct-a", "us-east-1", "Amazon Virtual Private Cloud", 250.0)

	cfg := defaultCfg()
	cfg.Analysis.NATCostThresholdUSD = 100
	rule := NewHighNATCostRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Category != "Architecture" {
		t.Errorf("expected Architecture, got %q", findings[0].Category)
	}
}

func TestHighNATCostRule_belowThreshold(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-052"
	insertAuditRun(t, d, runID)

	insertCostMonthly(t, d, runID, "2024-01", "111", "acct-a", "us-east-1", "Amazon Virtual Private Cloud", 50.0)

	cfg := defaultCfg()
	cfg.Analysis.NATCostThresholdUSD = 100
	rule := NewHighNATCostRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings below threshold, got %d", len(findings))
	}
}

func TestHighNATCostRule_defaultThreshold(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-053"
	insertAuditRun(t, d, runID)

	insertCostMonthly(t, d, runID, "2024-01", "111", "acct-a", "us-east-1", "Amazon Virtual Private Cloud", 150.0)

	cfg := defaultCfg()
	cfg.Analysis.NATCostThresholdUSD = 0 // defaults to 100
	rule := NewHighNATCostRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding with default threshold, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// HighCWCost
// ---------------------------------------------------------------------------

func TestHighCWCostRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-060"
	insertAuditRun(t, d, runID)

	rule := NewHighCWCostRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestHighCWCostRule_findsHighCost(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-061"
	insertAuditRun(t, d, runID)

	insertCostMonthly(t, d, runID, "2024-01", "111", "acct-a", "us-east-1", "Amazon CloudWatch", 200.0)

	cfg := defaultCfg()
	cfg.Analysis.CWLogsCostThresholdUSD = 50
	rule := NewHighCWCostRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Service != "Amazon CloudWatch" {
		t.Errorf("expected Amazon CloudWatch, got %q", findings[0].Service)
	}
}

func TestHighCWCostRule_defaultThreshold(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-062"
	insertAuditRun(t, d, runID)

	insertCostMonthly(t, d, runID, "2024-01", "111", "acct-a", "us-east-1", "Amazon CloudWatch", 75.0)

	cfg := defaultCfg()
	cfg.Analysis.CWLogsCostThresholdUSD = 0 // defaults to 50
	rule := NewHighCWCostRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding with default threshold, got %d", len(findings))
	}
}

// ---------------------------------------------------------------------------
// CostAnomaly
// ---------------------------------------------------------------------------

func insertCostDaily(t *testing.T, d *sql.DB, runID, date, accountID, service string, cost float64) {
	t.Helper()
	_, err := d.Exec(`INSERT INTO cost_daily (audit_run_id, date, account_id, service, unblended_cost)
		VALUES (?, ?, ?, ?, ?)`, runID, date, accountID, service, cost)
	if err != nil {
		t.Fatalf("insert cost_daily: %v", err)
	}
}

func TestCostAnomalyRule_noRows(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-070"
	insertAuditRun(t, d, runID)

	rule := NewCostAnomalyRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestCostAnomalyRule_detectsSpike(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-071"
	insertAuditRun(t, d, runID)

	// 14 days of baseline at $10, then one spike at $100
	base := time.Now().AddDate(0, 0, -15)
	for i := 0; i < 14; i++ {
		date := base.AddDate(0, 0, i).Format("2006-01-02")
		insertCostDaily(t, d, runID, date, "111", "Amazon EC2", 10.0)
	}
	spikeDate := base.AddDate(0, 0, 14).Format("2006-01-02")
	insertCostDaily(t, d, runID, spikeDate, "111", "Amazon EC2", 100.0)

	cfg := defaultCfg()
	cfg.Analysis.AnomalyStdDevMultiplier = 2.0
	rule := NewCostAnomalyRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected at least 1 anomaly finding for spike day")
	}
	if findings[0].Category != "Anomaly" {
		t.Errorf("expected Anomaly category, got %q", findings[0].Category)
	}
}

func TestCostAnomalyRule_requiresMinHistory(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-072"
	insertAuditRun(t, d, runID)

	// Only 5 days — less than the 7-day minimum
	base := time.Now().AddDate(0, 0, -5)
	for i := 0; i < 5; i++ {
		date := base.AddDate(0, 0, i).Format("2006-01-02")
		insertCostDaily(t, d, runID, date, "111", "Amazon EC2", 10.0)
	}
	insertCostDaily(t, d, runID, base.AddDate(0, 0, 5).Format("2006-01-02"), "111", "Amazon EC2", 999.0)

	rule := NewCostAnomalyRule(defaultCfg())
	findings, err := rule.Run(context.Background(), d, defaultCfg(), runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with < 7 days history, got %d", len(findings))
	}
}

func TestCostAnomalyRule_defaultMultiplier(t *testing.T) {
	d := openTestDB(t)
	const runID = "run-073"
	insertAuditRun(t, d, runID)

	base := time.Now().AddDate(0, 0, -15)
	for i := 0; i < 14; i++ {
		date := base.AddDate(0, 0, i).Format("2006-01-02")
		insertCostDaily(t, d, runID, date, "111", "Amazon S3", 5.0)
	}
	insertCostDaily(t, d, runID, base.AddDate(0, 0, 14).Format("2006-01-02"), "111", "Amazon S3", 200.0)

	cfg := defaultCfg()
	cfg.Analysis.AnomalyStdDevMultiplier = 0 // uses default 2.0
	rule := NewCostAnomalyRule(cfg)
	findings, err := rule.Run(context.Background(), d, cfg, runID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected anomaly finding with default multiplier")
	}
}

// ---------------------------------------------------------------------------
// Rule metadata
// ---------------------------------------------------------------------------

func TestRuleIDs(t *testing.T) {
	cfg := defaultCfg()
	cases := []struct {
		name     string
		id       string
		ruleName string
	}{
		{"UnattachedEBS", "AWS-WASTE-001", "UnattachedEBS"},
		{"OldSnapshots", "AWS-WASTE-002", "OldSnapshots"},
		{"UnassociatedEIP", "AWS-WASTE-003", "UnassociatedEIP"},
		{"StoppedEC2", "AWS-WASTE-004", "StoppedEC2"},
		{"HighNATCost", "AWS-COST-001", "HighNATCost"},
		{"HighCWCost", "AWS-COST-002", "HighCWLogsCost"},
		{"MissingTags", "AWS-GOV-001", "MissingTags"},
		{"CostAnomaly", "AWS-ANOMALY-001", "DailyCostAnomaly"},
	}

	rules := []interface {
		ID() string
		Name() string
	}{
		NewUnattachedEBSRule(cfg),
		NewOldSnapshotsRule(cfg),
		NewUnassociatedEIPRule(cfg),
		NewStoppedEC2Rule(cfg),
		NewHighNATCostRule(cfg),
		NewHighCWCostRule(cfg),
		NewMissingTagsRule(cfg),
		NewCostAnomalyRule(cfg),
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rules[i].ID(); got != tc.id {
				t.Errorf("ID() = %q, want %q", got, tc.id)
			}
			if got := rules[i].Name(); got != tc.ruleName {
				t.Errorf("Name() = %q, want %q", got, tc.ruleName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ebsMonthlyCost (tested indirectly via UnattachedEBSRule)
// ---------------------------------------------------------------------------

func TestEBSMonthlyCostViaRule(t *testing.T) {
	cases := []struct {
		volumeType  string
		sizeGiB     float64
		iops        int
		minExpected float64
	}{
		{"gp2", 100, 0, 9.0},   // 100 * 0.10
		{"gp3", 100, 3000, 7.0}, // 100 * 0.08 (no extra IOPS)
		{"io1", 100, 1000, 77.0}, // 100*0.125 + 1000*0.065
		{"st1", 100, 0, 4.0},   // 100 * 0.045
		{"sc1", 100, 0, 2.0},   // 100 * 0.025
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s_%dgib", tc.volumeType, int(tc.sizeGiB)), func(t *testing.T) {
			d := openTestDB(t)
			runID := fmt.Sprintf("run-ebs-%s", tc.volumeType)
			insertAuditRun(t, d, runID)

			insertResource(t, d, runID, "vol-test", "aws:ec2:volume", "available", "us-east-1", "111", "acct-a", map[string]any{
				"size_gib":    tc.sizeGiB,
				"volume_type": tc.volumeType,
				"iops":        tc.iops,
			})

			cfg := defaultCfg()
			rule := NewUnattachedEBSRule(cfg)
			findings, err := rule.Run(context.Background(), d, cfg, runID)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(findings) == 0 {
				t.Fatal("expected a finding")
			}
			if findings[0].EstMonthlySavingsUSD < tc.minExpected {
				t.Errorf("expected savings >= %.2f, got %.2f", tc.minExpected, findings[0].EstMonthlySavingsUSD)
			}
		})
	}
}
