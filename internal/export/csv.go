package export

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type CSVExporter struct{}

func (e *CSVExporter) Name() string { return "findings.csv" }

func (e *CSVExporter) Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error {
	findings, err := loadFindings(ctx, db, auditRunID)
	if err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"finding_id", "priority", "category", "service", "account_id", "account_name",
		"region", "title", "summary", "est_monthly_savings_usd", "confidence", "risk",
		"status", "resource_ids", "recommended_action",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, finding := range findings {
		row := []string{
			finding.ID,
			finding.Priority,
			finding.Category,
			finding.Service,
			finding.AccountID,
			finding.AccountName,
			finding.Region,
			finding.Title,
			finding.Summary,
			fmt.Sprintf("%.2f", finding.EstMonthlySavingsUSD),
			finding.Confidence,
			finding.Risk,
			finding.Status,
			strings.Join(resourceIDs(finding), ";"),
			finding.RecommendedAction,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}
