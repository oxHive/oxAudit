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

type EC2InstancesIngester struct{}

func (e *EC2InstancesIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ec2_describe-instances")
}

func (e *EC2InstancesIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		Reservations []struct {
			Instances []json.RawMessage `json:"Instances"`
		} `json:"Reservations"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing ec2 instances JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, res := range v.Reservations {
		for _, rawInst := range res.Instances {
			var inst struct {
				InstanceId   string `json:"InstanceId"`
				InstanceType string `json:"InstanceType"`
				State        struct {
					Name string `json:"Name"`
				} `json:"State"`
				LaunchTime string          `json:"LaunchTime"`
				Tags       []awsTag        `json:"Tags"`
				Placement  struct {
					AvailabilityZone string `json:"AvailabilityZone"`
				} `json:"Placement"`
			}
			if err := json.Unmarshal(rawInst, &inst); err != nil {
				continue
			}

			tagsJSON, _ := json.Marshal(tagsToMap(inst.Tags))
			name := tagValue(inst.Tags, "Name")

			if _, err := stmt.ExecContext(ctx,
				auditRunID,
				inst.InstanceId,
				"aws:ec2:instance",
				"", // account_id filled by join later
				"",
				region,
				"Amazon EC2",
				inst.State.Name,
				name,
				"", // arn
				nullStr(inst.LaunchTime),
				nil,  // size_gib
				inst.InstanceType,
				"",   // volume_type
				nil,  // iops
				nil,  // est_monthly_cost
				string(tagsJSON),
				string(rawInst),
				filePath,
				now,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
