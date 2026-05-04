package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type CWLogGroupsIngester struct{}

func (c *CWLogGroupsIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "logs_describe-log-groups")
}

func (c *CWLogGroupsIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		LogGroups []json.RawMessage `json:"logGroups"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing log groups JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.LogGroups {
		var lg struct {
			LogGroupName    string  `json:"logGroupName"`
			CreationTime    int64   `json:"creationTime"`
			RetentionInDays int     `json:"retentionInDays"`
			StoredBytes     float64 `json:"storedBytes"`
		}
		if err := json.Unmarshal(raw, &lg); err != nil {
			continue
		}

		state := "no-retention"
		if lg.RetentionInDays > 0 {
			state = fmt.Sprintf("retention-%dd", lg.RetentionInDays)
		}

		sizeGiB := lg.StoredBytes / (1024 * 1024 * 1024)

		// Estimate $0.03/GB/month storage cost
		estCost := sizeGiB * 0.03

		var createdAt interface{}
		if lg.CreationTime > 0 {
			t := time.Unix(lg.CreationTime/1000, 0).UTC()
			createdAt = t.Format(time.RFC3339)
		}

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			lg.LogGroupName,
			"aws:logs:log-group",
			"", "",
			region,
			"Amazon CloudWatch",
			state,
			lg.LogGroupName,
			"",
			createdAt,
			sizeGiB,
			"", "", nil,
			estCost,
			"{}",
			string(raw),
			filePath,
			now,
		); err != nil {
			return err
		}
	}
	return nil
}
