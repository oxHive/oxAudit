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

type RDSInstancesIngester struct{}

func (r *RDSInstancesIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "rds_describe-db-instances")
}

func (r *RDSInstancesIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		DBInstances []json.RawMessage `json:"DBInstances"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing RDS instances JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.DBInstances {
		var inst struct {
			DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
			DBInstanceClass      string `json:"DBInstanceClass"`
			DBInstanceStatus     string `json:"DBInstanceStatus"`
			Engine               string `json:"Engine"`
			AllocatedStorage     int    `json:"AllocatedStorage"`
			InstanceCreateTime   string `json:"InstanceCreateTime"`
			TagList              []struct {
				Key   string `json:"Key"`
				Value string `json:"Value"`
			} `json:"TagList"`
		}
		if err := json.Unmarshal(raw, &inst); err != nil {
			continue
		}

		tags := make([]awsTag, len(inst.TagList))
		for i, t := range inst.TagList {
			tags[i] = awsTag{Key: t.Key, Value: t.Value}
		}
		tagsJSON, _ := json.Marshal(tagsToMap(tags))

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			inst.DBInstanceIdentifier,
			"aws:rds:instance",
			"", "",
			region,
			"Amazon Relational Database Service",
			inst.DBInstanceStatus,
			inst.DBInstanceIdentifier,
			"",
			nullStr(inst.InstanceCreateTime),
			float64(inst.AllocatedStorage),
			inst.DBInstanceClass, "", nil, nil,
			string(tagsJSON),
			string(raw),
			filePath,
			now,
		); err != nil {
			return err
		}
	}
	return nil
}
