package rules

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type UnassociatedEIPRule struct{ cfg *config.Config }

func NewUnassociatedEIPRule(cfg *config.Config) *UnassociatedEIPRule {
	return &UnassociatedEIPRule{cfg: cfg}
}
func (r *UnassociatedEIPRule) ID() string   { return "AWS-WASTE-003" }
func (r *UnassociatedEIPRule) Name() string { return "UnassociatedEIP" }

func (r *UnassociatedEIPRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, account_id, account_name, region, tags_json, source_file
		FROM resources
		WHERE audit_run_id = ?
		  AND resource_type = 'aws:ec2:eip'
		  AND state = 'unassociated'`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var resourceID, accountID, accountName, region, tagsJSON, sourceFile string
		if err := rows.Scan(&resourceID, &accountID, &accountName, &region, &tagsJSON, &sourceFile); err != nil {
			continue
		}

		// AWS charges ~$3.60/month for unassociated EIP
		const eipMonthlyCost = 3.65

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{resourceID}),
			AuditRunID:           auditRunID,
			Priority:             "P1",
			Category:             "Waste",
			Service:              "Amazon EC2",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("Unassociated Elastic IP (%s)", resourceID),
			Summary:              fmt.Sprintf("Elastic IP %s has no instance or network interface association and is accruing charges.", resourceID),
			Evidence:             "EIP state: unassociated. No AssociationId, InstanceId, or NetworkInterfaceId present.",
			RecommendedAction:    "Confirm the IP is no longer needed, then release it: aws ec2 release-address --allocation-id " + resourceID,
			EstMonthlySavingsUSD: eipMonthlyCost,
			Confidence:           "High",
			Risk:                 "Low",
			Status:               "open",
			ResourceIDsJSON:      findutil.JSONArray([]string{resourceID}),
			TagsJSON:             tagsJSON,
			SourceFilesJSON:      findutil.JSONArray([]string{sourceFile}),
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}
