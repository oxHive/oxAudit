package rules

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
	"github.com/graditya/oxaudit/internal/types"
)

type MissingTagsRule struct{ cfg *config.Config }

func NewMissingTagsRule(cfg *config.Config) *MissingTagsRule {
	return &MissingTagsRule{cfg: cfg}
}
func (r *MissingTagsRule) ID() string   { return "AWS-GOV-001" }
func (r *MissingTagsRule) Name() string { return "MissingTags" }

func (r *MissingTagsRule) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error) {
	if len(cfg.Analysis.RequiredTags) == 0 {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, resource_type, account_id, account_name, region, tags_json, source_file
		FROM resources
		WHERE audit_run_id = ?
		  AND resource_type IN (
			  'aws:ec2:instance', 'aws:ec2:volume', 'aws:rds:instance', 'aws:ec2:nat-gateway'
		  )`, auditRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []types.Finding
	for rows.Next() {
		var resourceID, resourceType, accountID, accountName, region, tagsJSON, sourceFile string
		if err := rows.Scan(&resourceID, &resourceType, &accountID, &accountName, &region, &tagsJSON, &sourceFile); err != nil {
			continue
		}

		var tags map[string]string
		_ = json.Unmarshal([]byte(tagsJSON), &tags)

		var missing []string
		for _, key := range cfg.Analysis.RequiredTags {
			if _, ok := tags[key]; !ok {
				missing = append(missing, key)
			}
		}
		if len(missing) == 0 {
			continue
		}

		shortType := strings.TrimPrefix(resourceType, "aws:")

		f := types.Finding{
			ID:                   findutil.GenerateFindingID(r.ID(), accountID, region, []string{resourceID}),
			AuditRunID:           auditRunID,
			Priority:             "P3",
			Category:             "Tagging",
			Service:              serviceFromType(resourceType),
			AccountID:            accountID,
			AccountName:          accountName,
			Region:               region,
			Title:                fmt.Sprintf("Missing Required Tags on %s (%s)", shortType, resourceID),
			Summary:              fmt.Sprintf("Resource %s is missing required tags: %s", resourceID, strings.Join(missing, ", ")),
			Evidence:             fmt.Sprintf("Required tags: %s. Present tags: %s. Missing: %s.", strings.Join(cfg.Analysis.RequiredTags, ", "), tagsJSON, strings.Join(missing, ", ")),
			RecommendedAction:    fmt.Sprintf("Apply the missing tags using aws resourcegroupstaggingapi tag-resources or the AWS Console."),
			EstMonthlySavingsUSD: 0,
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

func serviceFromType(rt string) string {
	switch {
	case strings.HasPrefix(rt, "aws:ec2"):
		return "Amazon EC2"
	case strings.HasPrefix(rt, "aws:rds"):
		return "Amazon Relational Database Service"
	case strings.HasPrefix(rt, "aws:logs"):
		return "Amazon CloudWatch"
	default:
		return "AWS"
	}
}
