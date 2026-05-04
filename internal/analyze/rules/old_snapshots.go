package rules

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/types"
)

type OldSnapshotsRule struct{ cfg *config.Config }

func NewOldSnapshotsRule(cfg *config.Config) *OldSnapshotsRule {
	return &OldSnapshotsRule{cfg: cfg}
}
func (r *OldSnapshotsRule) ID() string   { return "AWS-WASTE-002" }
func (r *OldSnapshotsRule) Name() string { return "OldSnapshots" }

func (r *OldSnapshotsRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	threshold := cfg.Analysis.OldSnapshotThresholdDays
	if threshold == 0 {
		threshold = 90
	}

	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, account_id, account_name, region, size_gib, created_at, tags_json, source_file
		FROM resources
		WHERE audit_run_id = ?
		  AND resource_type = 'aws:ec2:snapshot'
		  AND created_at IS NOT NULL
		  AND julianday('now') - julianday(created_at) > ?`, auditRunID, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var (
			resourceID, accountID, accountName, region string
			sizeGiB                                    sql.NullFloat64
			createdAt, tagsJSON, sourceFile            string
		)
		if err := rows.Scan(&resourceID, &accountID, &accountName, &region,
			&sizeGiB, &createdAt, &tagsJSON, &sourceFile); err != nil {
			continue
		}

		size := sizeGiB.Float64
		// S3 snapshot storage ~$0.05/GiB/month
		estCost := size * 0.05

		if estCost < cfg.Analysis.MinMonthlySavingsUSD {
			continue
		}

		ageDays := 0
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			ageDays = int(time.Since(t).Hours() / 24)
		}

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{resourceID}),
			AuditRunID:           auditRunID,
			Priority:             "P2",
			Category:             "Waste",
			Service:              "Amazon EC2",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("Old EBS Snapshot (%d days, %.0f GiB)", ageDays, size),
			Summary:              fmt.Sprintf("Snapshot %s is %d days old and may no longer be needed.", resourceID, ageDays),
			Evidence:             fmt.Sprintf("Created: %s. Age: %d days. Size: %.0f GiB. Threshold: %d days.", createdAt, ageDays, size, threshold),
			RecommendedAction:    "Validate the snapshot is still needed. If not, delete with: aws ec2 delete-snapshot --snapshot-id " + resourceID,
			EstMonthlySavingsUSD: estCost,
			Confidence:           "Medium",
			Risk:                 "Medium",
			Status:               "open",
			ResourceIDsJSON:      findutil.JSONArray([]string{resourceID}),
			TagsJSON:             tagsJSON,
			SourceFilesJSON:      findutil.JSONArray([]string{sourceFile}),
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}
