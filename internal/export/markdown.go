package export

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/graditya/oxaudit/internal/types"
)

type MarkdownExporter struct{}

func (e *MarkdownExporter) Name() string { return "findings.md" }

func (e *MarkdownExporter) Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error {
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

	fmt.Fprintf(f, "# AWS Cost Audit Findings\n\n")
	fmt.Fprintf(f, "**Audit Run:** %s  \n", run.ID)
	fmt.Fprintf(f, "**Period:** %s to %s  \n", run.PeriodStart, run.PeriodEnd)
	fmt.Fprintf(f, "**Generated:** %s  \n\n", run.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))

	// Group by priority
	byPriority := map[string][]types.Finding{}
	for _, finding := range findings {
		byPriority[finding.Priority] = append(byPriority[finding.Priority], finding)
	}

	for _, priority := range []string{"P0", "P1", "P2", "P3"} {
		group := byPriority[priority]
		if len(group) == 0 {
			continue
		}

		label := map[string]string{
			"P0": "P0 — Urgent",
			"P1": "P1 — High Confidence Waste",
			"P2": "P2 — Optimization Required",
			"P3": "P3 — Governance",
		}[priority]

		fmt.Fprintf(f, "## %s (%d findings)\n\n", label, len(group))

		for _, finding := range group {
			fmt.Fprintf(f, "### %s: %s\n\n", finding.ID, finding.Title)
			fmt.Fprintf(f, "| Field | Value |\n|-------|-------|\n")
			fmt.Fprintf(f, "| Priority | %s |\n", finding.Priority)
			fmt.Fprintf(f, "| Category | %s |\n", finding.Category)
			fmt.Fprintf(f, "| Service | %s |\n", finding.Service)
			fmt.Fprintf(f, "| Account | %s (%s) |\n", finding.AccountName, finding.AccountID)
			fmt.Fprintf(f, "| Region | %s |\n", finding.Region)
			fmt.Fprintf(f, "| Resources | %s |\n", strings.Join(resourceIDs(finding), ", "))
			fmt.Fprintf(f, "| Est. Savings | $%.2f/month |\n", finding.EstMonthlySavingsUSD)
			fmt.Fprintf(f, "| Confidence | %s |\n", finding.Confidence)
			fmt.Fprintf(f, "| Risk | %s |\n", finding.Risk)
			fmt.Fprintf(f, "| Status | %s |\n\n", finding.Status)
			fmt.Fprintf(f, "**Summary:** %s\n\n", finding.Summary)
			fmt.Fprintf(f, "**Evidence:** %s\n\n", finding.Evidence)
			fmt.Fprintf(f, "**Recommended Action:** %s\n\n", finding.RecommendedAction)
			if len(sourceFiles(finding)) > 0 {
				fmt.Fprintf(f, "**Source Files:** %s\n\n", strings.Join(sourceFiles(finding), ", "))
			}
			fmt.Fprintf(f, "---\n\n")
		}
	}
	return nil
}
