package rules

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type HighNATCostRule struct{ cfg *config.Config }

func NewHighNATCostRule(cfg *config.Config) *HighNATCostRule {
	return &HighNATCostRule{cfg: cfg}
}
func (r *HighNATCostRule) ID() string   { return "AWS-COST-001" }
func (r *HighNATCostRule) Name() string { return "HighNATCost" }

func (r *HighNATCostRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	threshold := cfg.Analysis.NATCostThresholdUSD
	if threshold == 0 {
		threshold = 100
	}

	rows, err := db.QueryContext(ctx, `
		SELECT account_id, account_name, region, SUM(unblended_cost) as total
		FROM cost_monthly
		WHERE audit_run_id = ?
		  AND (service = 'Amazon Virtual Private Cloud' OR service LIKE '%NatGateway%')
		GROUP BY account_id, region
		HAVING total > ?
		ORDER BY total DESC`, auditRunID, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var accountID, accountName, region string
		var total float64
		if err := rows.Scan(&accountID, &accountName, &region, &total); err != nil {
			continue
		}

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{"NatGateway", "monthly"}),
			AuditRunID:           auditRunID,
			Priority:             "P2",
			Category:             "Architecture",
			Service:              "Amazon Virtual Private Cloud",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("High NAT Gateway Cost ($%.0f/month) in %s", total, region),
			Summary:              fmt.Sprintf("NAT Gateway charges in %s total $%.2f/month, exceeding the $%.0f threshold.", region, total, threshold),
			Evidence:             fmt.Sprintf("Aggregated VPC service cost from cost_monthly: $%.2f. Region: %s. Account: %s.", total, region, accountID),
			RecommendedAction:    "Review traffic routing. Consider VPC endpoints for S3/DynamoDB to eliminate NAT data processing charges. Review cross-AZ NAT Gateway usage.",
			EstMonthlySavingsUSD: total * 0.3, // conservative 30% potential reduction
			Confidence:           "Medium",
			Risk:                 "Medium",
			Status:               "open",
			ResourceIDsJSON:      findutil.JSONArray([]string{}),
			SourceFilesJSON:      findutil.JSONArray([]string{}),
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}
