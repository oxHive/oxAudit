package rules

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type HighCWCostRule struct{ cfg *config.Config }

func NewHighCWCostRule(cfg *config.Config) *HighCWCostRule {
	return &HighCWCostRule{cfg: cfg}
}
func (r *HighCWCostRule) ID() string   { return "AWS-COST-002" }
func (r *HighCWCostRule) Name() string { return "HighCWLogsCost" }

func (r *HighCWCostRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	threshold := cfg.Analysis.CWLogsCostThresholdUSD
	if threshold == 0 {
		threshold = 50
	}

	rows, err := db.QueryContext(ctx, `
		SELECT account_id, account_name, region, SUM(unblended_cost) as total
		FROM cost_monthly
		WHERE audit_run_id = ?
		  AND (service = 'Amazon CloudWatch' OR service LIKE '%CloudWatch%')
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
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{"CloudWatch", "monthly"}),
			AuditRunID:           auditRunID,
			Priority:             "P2",
			Category:             "Observability",
			Service:              "Amazon CloudWatch",
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("High CloudWatch Logs Cost ($%.0f/month) in %s", total, region),
			Summary:              fmt.Sprintf("CloudWatch costs in %s total $%.2f/month, exceeding the $%.0f threshold.", region, total, threshold),
			Evidence:             fmt.Sprintf("Aggregated CloudWatch service cost from cost_monthly: $%.2f. Region: %s.", total, region),
			RecommendedAction:    "Review log groups without retention policies. Set retention (e.g. 30-90 days) to reduce storage costs. Reduce verbose logging where appropriate.",
			EstMonthlySavingsUSD: total * 0.4, // conservative 40% reduction via retention
			Confidence:           "Medium",
			Risk:                 "Low",
			Status:               "open",
			ResourceIDsJSON:      findutil.JSONArray([]string{}),
			SourceFilesJSON:      findutil.JSONArray([]string{}),
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}
