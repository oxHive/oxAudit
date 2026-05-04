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

type ElasticIPsIngester struct{}

func (e *ElasticIPsIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ec2_describe-addresses")
}

func (e *ElasticIPsIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		Addresses []json.RawMessage `json:"Addresses"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing elastic IPs JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.Addresses {
		var addr struct {
			AllocationId        string   `json:"AllocationId"`
			PublicIp            string   `json:"PublicIp"`
			AssociationId       string   `json:"AssociationId"`
			InstanceId          string   `json:"InstanceId"`
			NetworkInterfaceId  string   `json:"NetworkInterfaceId"`
			Tags                []awsTag `json:"Tags"`
		}
		if err := json.Unmarshal(raw, &addr); err != nil {
			continue
		}

		state := "associated"
		if addr.AssociationId == "" && addr.InstanceId == "" && addr.NetworkInterfaceId == "" {
			state = "unassociated"
		}

		tagsJSON, _ := json.Marshal(tagsToMap(addr.Tags))
		name := tagValue(addr.Tags, "Name")
		if name == "" {
			name = addr.PublicIp
		}

		id := addr.AllocationId
		if id == "" {
			id = addr.PublicIp
		}

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			id,
			"aws:ec2:eip",
			"", "",
			region,
			"Amazon EC2",
			state,
			name,
			"",
			nil, nil, "", "", nil, nil,
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
