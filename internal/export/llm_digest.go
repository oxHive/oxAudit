package export

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/graditya/oxaudit/internal/types"
)

type LLMDigestExporter struct{}

func (e *LLMDigestExporter) Name() string { return "llm-digest.md" }

func (e *LLMDigestExporter) Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error {
	findings, err := loadFindings(ctx, db, auditRunID)
	if err != nil {
		return err
	}
	run, err := loadAuditRun(ctx, db, auditRunID)
	if err != nil {
		return err
	}

	// Aggregate stats
	var totalSavings float64
	counts := map[string]int{}
	savingsByService := map[string]float64{}
	savingsByAccount := map[string]float64{}

	for _, f := range findings {
		totalSavings += f.EstMonthlySavingsUSD
		counts[f.Priority]++
		savingsByService[f.Service] += f.EstMonthlySavingsUSD
		acct := f.AccountName
		if acct == "" {
			acct = f.AccountID
		}
		savingsByAccount[acct] += f.EstMonthlySavingsUSD
	}

	// Get total spend
	var totalSpend float64
	db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(unblended_cost), 0)
		FROM cost_monthly WHERE audit_run_id = ?`, auditRunID).Scan(&totalSpend)

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# AWS Cost Audit — LLM Digest\n\n")
	fmt.Fprintf(f, "> This is the first file to read. It summarizes all findings from this audit run.\n\n")
	fmt.Fprintf(f, "**Audit Run:** `%s`  \n", run.ID)
	fmt.Fprintf(f, "**Period:** %s to %s  \n", run.PeriodStart, run.PeriodEnd)
	fmt.Fprintf(f, "**Generated:** %s  \n", run.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(f, "**AWS Profile:** %s  \n\n", run.AWSProfile)

	fmt.Fprintf(f, "## Executive Summary\n\n")
	fmt.Fprintf(f, "| Metric | Value |\n|--------|-------|\n")
	if totalSpend > 0 {
		fmt.Fprintf(f, "| Total Monthly Spend | $%.2f |\n", totalSpend)
	}
	fmt.Fprintf(f, "| Total Estimated Savings | $%.2f/month |\n", totalSavings)
	fmt.Fprintf(f, "| Total Findings | %d |\n", len(findings))
	fmt.Fprintf(f, "| P0 (Urgent) | %d |\n", counts["P0"])
	fmt.Fprintf(f, "| P1 (High Confidence Waste) | %d |\n", counts["P1"])
	fmt.Fprintf(f, "| P2 (Optimization) | %d |\n", counts["P2"])
	fmt.Fprintf(f, "| P3 (Governance) | %d |\n\n", counts["P3"])

	if len(findings) > 0 {
		fmt.Fprintf(f, "## Top Findings by Savings\n\n")
		fmt.Fprintf(f, "| Priority | Title | Service | Account | Savings/mo | Confidence |\n")
		fmt.Fprintf(f, "|----------|-------|---------|---------|------------|------------|\n")
		top := findings
		if len(top) > 10 {
			top = top[:10]
		}
		for _, finding := range top {
			acct := finding.AccountName
			if acct == "" {
				acct = finding.AccountID
			}
			title := finding.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			fmt.Fprintf(f, "| %s | %s | %s | %s | $%.0f | %s |\n",
				finding.Priority, title, finding.Service, acct,
				finding.EstMonthlySavingsUSD, finding.Confidence)
		}
		fmt.Fprintf(f, "\n")
	}

	// P0 and P1 detail
	for _, priority := range []string{"P0", "P1"} {
		label := map[string]string{"P0": "Urgent — Immediate Action Required", "P1": "High Confidence Waste"}[priority]
		var group []types.Finding
		for _, ff := range findings {
			if ff.Priority == priority {
				group = append(group, ff)
			}
		}
		if len(group) == 0 {
			continue
		}
		fmt.Fprintf(f, "## %s — %s\n\n", priority, label)
		for _, ff := range group {
			fmt.Fprintf(f, "### %s: %s\n\n", ff.ID, ff.Title)
			fmt.Fprintf(f, "%s\n\n", ff.Summary)
			fmt.Fprintf(f, "- **Evidence:** %s\n", ff.Evidence)
			fmt.Fprintf(f, "- **Savings:** $%.2f/month\n", ff.EstMonthlySavingsUSD)
			fmt.Fprintf(f, "- **Action:** %s\n\n", ff.RecommendedAction)
		}
	}

	// Savings by service
	if len(savingsByService) > 0 {
		fmt.Fprintf(f, "## Savings by Service\n\n")
		fmt.Fprintf(f, "| Service | Est. Savings/mo |\n|---------|----------------|\n")
		for svc, amt := range savingsByService {
			if amt > 0 {
				fmt.Fprintf(f, "| %s | $%.2f |\n", svc, amt)
			}
		}
		fmt.Fprintf(f, "\n")
	}

	if len(savingsByAccount) > 0 {
		fmt.Fprintf(f, "## Savings by Account\n\n")
		fmt.Fprintf(f, "| Account | Est. Savings/mo |\n|---------|----------------|\n")
		for acct, amt := range savingsByAccount {
			if amt > 0 {
				fmt.Fprintf(f, "| %s | $%.2f |\n", acct, amt)
			}
		}
		fmt.Fprintf(f, "\n")
	}

	fmt.Fprintf(f, "## Known Limitations\n\n")
	fmt.Fprintf(f, "- Savings estimates use static pricing tables and may not reflect your exact pricing tier.\n")
	fmt.Fprintf(f, "- CloudWatch and NAT Gateway cost findings are based on aggregated service spend, not per-resource attribution.\n")
	fmt.Fprintf(f, "- Stopped EC2 compute savings are not estimated (already stopped); attached EBS costs are captured separately.\n")
	fmt.Fprintf(f, "- Rightsizing recommendations require Compute Optimizer data (not yet collected in MVP).\n")

	return nil
}
