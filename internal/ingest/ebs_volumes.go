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

type EBSVolumesIngester struct{}

func (e *EBSVolumesIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ec2_describe-volumes")
}

func (e *EBSVolumesIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		Volumes []json.RawMessage `json:"Volumes"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing ebs volumes JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.Volumes {
		var vol struct {
			VolumeId   string   `json:"VolumeId"`
			State      string   `json:"State"`
			Size       int      `json:"Size"`
			VolumeType string   `json:"VolumeType"`
			Iops       int      `json:"Iops"`
			CreateTime string   `json:"CreateTime"`
			Tags       []awsTag `json:"Tags"`
		}
		if err := json.Unmarshal(raw, &vol); err != nil {
			continue
		}

		tagsJSON, _ := json.Marshal(tagsToMap(vol.Tags))
		name := tagValue(vol.Tags, "Name")

		var iops interface{}
		if vol.Iops > 0 {
			iops = vol.Iops
		}

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			vol.VolumeId,
			"aws:ec2:volume",
			"", "",
			region,
			"Amazon EC2",
			vol.State,
			name,
			"",
			nullStr(vol.CreateTime),
			float64(vol.Size),
			"", vol.VolumeType, iops,
			nil,
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
