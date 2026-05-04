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

type UnattachedEBSRule struct{ cfg *config.Config }

func NewUnattachedEBSRule(cfg *config.Config) *UnattachedEBSRule {
	return &UnattachedEBSRule{cfg: cfg}
}
func (r *UnattachedEBSRule) ID() string   { return "AWS-WASTE-001" }
func (r *UnattachedEBSRule) Name() string { return "UnattachedEBS" }

func (r *UnattachedEBSRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, account_id, account_name, region, size_gib, volume_type, iops,
		       created_at, tags_json, source_file
		FROM resources
		WHERE audit_run_id = ?
		  AND resource_type = 'aws:ec2:volume'
		  AND state = 'available'`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var (
			resourceID, accountID, accountName, region string
			sizeGiB                                    sql.NullFloat64
			volumeType                                 string
			iops                                       sql.NullInt64
			createdAt, tagsJSON, sourceFile            string
		)
		if err := rows.Scan(&resourceID, &accountID, &accountName, &region,
			&sizeGiB, &volumeType, &iops, &createdAt, &tagsJSON, &sourceFile); err != nil {
			continue
		}

		size := sizeGiB.Float64
		estCost := ebsMonthlyCost(volumeType, size, int(iops.Int64))

		if estCost < cfg.Analysis.MinMonthlySavingsUSD {
			continue
		}

		priority := "P2"
		if size > 500 || iops.Int64 > 3000 {
			priority = "P1"
		}

		age := ""
		if createdAt != "" {
			if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				days := int(time.Since(t).Hours() / 24)
				age = fmt.Sprintf("%d days", days)
			}
		}

		evidence := fmt.Sprintf("Volume state: available. Size: %.0f GiB %s. Age: %s. IOPS: %d.",
			size, volumeType, age, iops.Int64)

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{resourceID}),
			AuditRunID:           auditRunID,
			Priority:             priority,
			Category:             "Waste",
			Service:              "Amazon EC2",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("Unattached EBS Volume (%.0f GiB %s)", size, volumeType),
			Summary:              fmt.Sprintf("EBS volume %s is unattached (state=available) and accruing storage costs.", resourceID),
			Evidence:             evidence,
			RecommendedAction:    "Snapshot the volume then delete it after owner confirmation. Use: aws ec2 delete-volume --volume-id " + resourceID,
			EstMonthlySavingsUSD: estCost,
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

// ebsMonthlyCost estimates monthly cost using static pricing (USD/GiB/month).
func ebsMonthlyCost(volumeType string, sizeGiB float64, iops int) float64 {
	switch volumeType {
	case "gp3":
		cost := sizeGiB * 0.08
		if iops > 3000 {
			cost += float64(iops-3000) * 0.005
		}
		return cost
	case "gp2":
		return sizeGiB * 0.10
	case "io1", "io2":
		return sizeGiB*0.125 + float64(iops)*0.065
	case "st1":
		return sizeGiB * 0.045
	case "sc1":
		return sizeGiB * 0.025
	default:
		return sizeGiB * 0.08
	}
}
