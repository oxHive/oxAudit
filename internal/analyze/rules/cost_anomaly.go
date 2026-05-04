package rules

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type CostAnomalyRule struct{ cfg *config.Config }

type costDay struct {
	Date      string
	AccountID string
	Service   string
	Cost      float64
}

func NewCostAnomalyRule(cfg *config.Config) *CostAnomalyRule {
	return &CostAnomalyRule{cfg: cfg}
}
func (r *CostAnomalyRule) ID() string   { return "AWS-ANOMALY-001" }
func (r *CostAnomalyRule) Name() string { return "DailyCostAnomaly" }

func (r *CostAnomalyRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	multiplier := cfg.Analysis.AnomalyStdDevMultiplier
	if multiplier == 0 {
		multiplier = 2.0
	}

	// Load all daily cost rows for this run
	rows, err := db.QueryContext(ctx, `
		SELECT date, account_id, service, SUM(unblended_cost) as daily_cost
		FROM cost_daily
		WHERE audit_run_id = ?
		GROUP BY date, account_id, service
		ORDER BY service, account_id, date`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []costDay
	for rows.Next() {
		var dr costDay
		if err := rows.Scan(&dr.Date, &dr.AccountID, &dr.Service, &dr.Cost); err != nil {
			continue
		}
		all = append(all, dr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Group by service+account, compute mean and stddev, detect spikes
	type key struct{ service, accountID string }
	groups := make(map[key][]costDay)
	for _, dr := range all {
		k := key{dr.Service, dr.AccountID}
		groups[k] = append(groups[k], dr)
	}

	var findings []types.Finding
	for k, days := range groups {
		if len(days) < 7 {
			continue // not enough history
		}

		mean, stddev := computeStats(days)
		if stddev == 0 {
			continue
		}
		threshold := mean + multiplier*stddev

		for _, dr := range days {
			if dr.Cost <= threshold || dr.Cost < cfg.Analysis.MinMonthlySavingsUSD {
				continue
			}
			spike := dr.Cost - mean

			f := types.Finding{
				ID:                   findutil.GenerateFindingID(r.ID(), k.accountID, "", []string{k.service, dr.Date}),
				AuditRunID:           auditRunID,
				Priority:             "P1",
				Category:             "Anomaly",
				Service:              k.service,
				AccountID:            k.accountID,
				Region:               "",
				Title:                fmt.Sprintf("Daily Cost Spike: %s on %s ($%.0f vs avg $%.0f)", k.service, dr.Date, dr.Cost, mean),
				Summary:              fmt.Sprintf("Daily cost for %s spiked to $%.2f on %s, which is %.1f standard deviations above the mean ($%.2f).", k.service, dr.Cost, dr.Date, (dr.Cost-mean)/stddev, mean),
				Evidence:             fmt.Sprintf("Date: %s. Cost: $%.2f. Mean: $%.2f. Stddev: $%.2f. Threshold (%.1f×σ): $%.2f.", dr.Date, dr.Cost, mean, stddev, multiplier, threshold),
				RecommendedAction:    "Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before " + dr.Date + ".",
				EstMonthlySavingsUSD: spike * 30, // annualized if this were recurring
				Confidence:           "Medium",
				Risk:                 "Low",
				Status:               "open",
				ResourceIDsJSON:      findutil.JSONArray([]string{}),
				SourceFilesJSON:      findutil.JSONArray([]string{}),
			}
			findings = append(findings, f)
		}
	}
	return findings, nil
}

func computeStats(days []costDay) (mean, stddev float64) {
	if len(days) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, d := range days {
		sum += d.Cost
	}
	mean = sum / float64(len(days))

	variance := 0.0
	for _, d := range days {
		diff := d.Cost - mean
		variance += diff * diff
	}
	variance /= float64(len(days))
	stddev = math.Sqrt(variance)
	return mean, stddev
}
