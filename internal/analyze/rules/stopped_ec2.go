package rules

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type StoppedEC2Rule struct{ cfg *config.Config }

func NewStoppedEC2Rule(cfg *config.Config) *StoppedEC2Rule {
	return &StoppedEC2Rule{cfg: cfg}
}
func (r *StoppedEC2Rule) ID() string   { return "AWS-WASTE-004" }
func (r *StoppedEC2Rule) Name() string { return "StoppedEC2" }

func (r *StoppedEC2Rule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	threshold := cfg.Analysis.StoppedEC2ThresholdDays
	if threshold == 0 {
		threshold = 30
	}

	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, account_id, account_name, region, instance_type, created_at, tags_json, source_file
		FROM resources
		WHERE audit_run_id = ?
		  AND resource_type = 'aws:ec2:instance'
		  AND state = 'stopped'`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var (
			resourceID, accountID, accountName, region string
			instanceType                               string
			launchTime, tagsJSON, sourceFile           string
		)
		if err := rows.Scan(&resourceID, &accountID, &accountName, &region,
			&instanceType, &launchTime, &tagsJSON, &sourceFile); err != nil {
			continue
		}

		ageDays := 0
		if t, err := time.Parse(time.RFC3339, launchTime); err == nil {
			ageDays = int(time.Since(t).Hours() / 24)
		}

		// Stopped instances don't accrue compute cost, but attached EBS volumes do.
		// We report $0 savings here — the EBS rule will capture the storage cost.
		evidence := fmt.Sprintf("Instance state: stopped. Type: %s. Launch time: %s. Age: %d days. Threshold: %d days.",
			instanceType, launchTime, ageDays, threshold)

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{resourceID}),
			AuditRunID:           auditRunID,
			Priority:             "P2",
			Category:             "Waste",
			Service:              "Amazon EC2",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("Stopped EC2 Instance (%s, %s)", resourceID, instanceType),
			Summary:              fmt.Sprintf("Instance %s has been stopped. Attached EBS volumes continue to incur charges.", resourceID),
			Evidence:             evidence,
			RecommendedAction:    fmt.Sprintf("Confirm the instance is no longer needed. Terminate with: aws ec2 terminate-instances --instance-ids %s", resourceID),
			EstMonthlySavingsUSD: 0, // compute is already stopped; EBS rule accounts for storage
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
