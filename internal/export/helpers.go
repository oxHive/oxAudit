package export

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/graditya/oxaudit/internal/types"
)

// loadFindings queries all findings for an audit run ordered by priority and savings.
func loadFindings(ctx context.Context, db *sql.DB, auditRunID string) ([]types.Finding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, audit_run_id, priority, category, service, account_id, account_name, region,
		       title, summary, evidence, recommended_action,
		       est_monthly_savings_usd, confidence, risk, owner, status,
		       resource_ids_json, tags_json, source_files_json, created_at, updated_at
		FROM findings
		WHERE audit_run_id = ?
		ORDER BY
			CASE priority WHEN 'P0' THEN 0 WHEN 'P1' THEN 1 WHEN 'P2' THEN 2 ELSE 3 END,
			est_monthly_savings_usd DESC`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var f types.Finding
		var createdAt, updatedAt string
		if err := rows.Scan(
			&f.ID, &f.AuditRunID, &f.Priority, &f.Category, &f.Service,
			&f.AccountID, &f.AccountName, &f.Region,
			&f.Title, &f.Summary, &f.Evidence, &f.RecommendedAction,
			&f.EstMonthlySavingsUSD, &f.Confidence, &f.Risk, &f.Owner, &f.Status,
			&f.ResourceIDsJSON, &f.TagsJSON, &f.SourceFilesJSON,
			&createdAt, &updatedAt,
		); err != nil {
			continue
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		f.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

// loadAuditRun fetches the audit_run record for the given ID.
func loadAuditRun(ctx context.Context, db *sql.DB, auditRunID string) (*types.AuditRun, error) {
	var run types.AuditRun
	var executedAt, generatedAt string
	err := db.QueryRowContext(ctx, `
		SELECT id, executed_at, generated_at, period_start, period_end, aws_profile,
		       billing_region, run_folder, notes, status
		FROM audit_run WHERE id = ?`, auditRunID).Scan(
		&run.ID, &executedAt, &generatedAt,
		&run.PeriodStart, &run.PeriodEnd, &run.AWSProfile,
		&run.BillingRegion, &run.RunFolder, &run.Notes, &run.Status,
	)
	if err != nil {
		return nil, err
	}
	run.ExecutedAt, _ = time.Parse(time.RFC3339, executedAt)
	run.GeneratedAt, _ = time.Parse(time.RFC3339, generatedAt)
	return &run, nil
}

// resourceIDs extracts the resource ID list from a finding's JSON field.
func resourceIDs(f types.Finding) []string {
	var ids []string
	_ = json.Unmarshal([]byte(f.ResourceIDsJSON), &ids)
	return ids
}

// sourceFiles extracts the source files list from a finding's JSON field.
func sourceFiles(f types.Finding) []string {
	var files []string
	_ = json.Unmarshal([]byte(f.SourceFilesJSON), &files)
	return files
}
