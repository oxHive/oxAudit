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

type EBSSnapshotsIngester struct{}

func (e *EBSSnapshotsIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ec2_describe-snapshots")
}

func (e *EBSSnapshotsIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		Snapshots []json.RawMessage `json:"Snapshots"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing snapshots JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.Snapshots {
		var snap struct {
			SnapshotId  string   `json:"SnapshotId"`
			VolumeId    string   `json:"VolumeId"`
			State       string   `json:"State"`
			VolumeSize  int      `json:"VolumeSize"`
			StartTime   string   `json:"StartTime"`
			Description string   `json:"Description"`
			Tags        []awsTag `json:"Tags"`
		}
		if err := json.Unmarshal(raw, &snap); err != nil {
			continue
		}

		tagsJSON, _ := json.Marshal(tagsToMap(snap.Tags))
		name := tagValue(snap.Tags, "Name")
		if name == "" {
			name = snap.Description
		}

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			snap.SnapshotId,
			"aws:ec2:snapshot",
			"", "",
			region,
			"Amazon EC2",
			snap.State,
			name,
			"",
			nullStr(snap.StartTime),
			float64(snap.VolumeSize),
			"", "", nil, nil,
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
