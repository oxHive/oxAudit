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

type NATGatewaysIngester struct{}

func (n *NATGatewaysIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ec2_describe-nat-gateways")
}

func (n *NATGatewaysIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		NatGateways []json.RawMessage `json:"NatGateways"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing NAT gateways JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, resourceInsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	region := regionFromPath(filePath)

	for _, raw := range v.NatGateways {
		var gw struct {
			NatGatewayId string   `json:"NatGatewayId"`
			State        string   `json:"State"`
			CreateTime   string   `json:"CreateTime"`
			VpcId        string   `json:"VpcId"`
			Tags         []awsTag `json:"Tags"`
		}
		if err := json.Unmarshal(raw, &gw); err != nil {
			continue
		}

		tagsJSON, _ := json.Marshal(tagsToMap(gw.Tags))
		name := tagValue(gw.Tags, "Name")

		if _, err := stmt.ExecContext(ctx,
			auditRunID,
			gw.NatGatewayId,
			"aws:ec2:nat-gateway",
			"", "",
			region,
			"Amazon Virtual Private Cloud",
			gw.State,
			name,
			"",
			nullStr(gw.CreateTime),
			nil, "", "", nil, nil,
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
