// Package findutil provides shared utilities for building findings.
// It is a leaf package imported by both analyze and analyze/rules.
package findutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/graditya/oxaudit/internal/types"
)

// GenerateFindingID produces a deterministic finding ID.
func GenerateFindingID(ruleID, accountID, region string, resourceIDs []string) string {
	ids := make([]string, len(resourceIDs))
	copy(ids, resourceIDs)
	sort.Strings(ids)
	parts := append([]string{ruleID, accountID, region}, ids...)
	raw := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(raw))
	return "FND-" + hex.EncodeToString(h[:8])
}

// JSONArray marshals a string slice to a JSON array string.
func JSONArray(ss []string) string {
	b, _ := json.Marshal(ss)
	return string(b)
}

// UpsertFinding inserts or replaces a finding in the database.
func UpsertFinding(ctx context.Context, db *sql.DB, f types.Finding) error {
	ridJSON := f.ResourceIDsJSON
	if ridJSON == "" {
		ridJSON = "[]"
	}
	sfJSON := f.SourceFilesJSON
	if sfJSON == "" {
		sfJSON = "[]"
	}
	tagsJSON := f.TagsJSON
	if tagsJSON == "" {
		tagsJSON = "{}"
	}

	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.ExecContext(ctx, `
		INSERT OR REPLACE INTO findings
			(id, audit_run_id, priority, category, service, account_id, account_name, region,
			 title, summary, evidence, recommended_action,
			 est_monthly_savings_usd, confidence, risk, owner, status,
			 resource_ids_json, tags_json, source_files_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.AuditRunID, f.Priority, f.Category, f.Service,
		f.AccountID, f.AccountName, f.Region,
		f.Title, f.Summary, f.Evidence, f.RecommendedAction,
		f.EstMonthlySavingsUSD, f.Confidence, f.Risk, f.Owner, f.Status,
		ridJSON, tagsJSON, sfJSON,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("upserting finding %s: %w", f.ID, err)
	}
	return nil
}
