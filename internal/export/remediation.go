package export

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/graditya/oxaudit/internal/types"
)

type RemediationExporter struct{}

func (e *RemediationExporter) Name() string { return "remediation-backlog.md" }

func (e *RemediationExporter) Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error {
	findings, err := loadFindings(ctx, db, auditRunID)
	if err != nil {
		return err
	}
	run, err := loadAuditRun(ctx, db, auditRunID)
	if err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# AWS Cost Audit — Remediation Backlog\n\n")
	fmt.Fprintf(f, "**Audit Run:** %s  \n", run.ID)
	fmt.Fprintf(f, "**Period:** %s to %s  \n\n", run.PeriodStart, run.PeriodEnd)

	sections := []struct {
		Priority string
		Label    string
		Desc     string
	}{
		{"P0", "Immediate Action", "Urgent issues requiring same-day review."},
		{"P1", "Safe Cleanup", "High-confidence waste that can be addressed this week."},
		{"P2", "Engineering Validation", "Optimization requiring owner validation before action."},
		{"P3", "Governance Improvements", "Low-savings or compliance improvements."},
	}

	for _, sec := range sections {
		var group []types.Finding
		for _, ff := range findings {
			if ff.Priority == sec.Priority {
				group = append(group, ff)
			}
		}
		if len(group) == 0 {
			continue
		}

		totalSavings := 0.0
		for _, ff := range group {
			totalSavings += ff.EstMonthlySavingsUSD
		}

		fmt.Fprintf(f, "## %s — %s\n\n", sec.Priority, sec.Label)
		fmt.Fprintf(f, "> %s Total potential savings: **$%.0f/month**.\n\n", sec.Desc, totalSavings)

		for _, ff := range group {
			acct := ff.AccountName
			if acct == "" {
				acct = ff.AccountID
			}
			approval := "None required"
			if ff.Risk == "High" {
				approval = "Requires engineering + manager approval"
			} else if ff.Risk == "Medium" {
				approval = "Requires owner confirmation"
			}

			fmt.Fprintf(f, "### %s — %s\n\n", ff.ID, ff.Title)
			fmt.Fprintf(f, "| Field | Value |\n|-------|-------|\n")
			fmt.Fprintf(f, "| Service | %s |\n", ff.Service)
			fmt.Fprintf(f, "| Account | %s |\n", acct)
			fmt.Fprintf(f, "| Region | %s |\n", ff.Region)
			fmt.Fprintf(f, "| Savings | $%.2f/month |\n", ff.EstMonthlySavingsUSD)
			fmt.Fprintf(f, "| Confidence | %s |\n", ff.Confidence)
			fmt.Fprintf(f, "| Risk | %s |\n", ff.Risk)
			fmt.Fprintf(f, "| Approval | %s |\n\n", approval)
			fmt.Fprintf(f, "**Action:** %s\n\n", ff.RecommendedAction)
			fmt.Fprintf(f, "**Validation:** Confirm the resource is no longer in use before deleting. Check with the owning team if tags are missing.\n\n")
			fmt.Fprintf(f, "---\n\n")
		}
	}
	return nil
}
