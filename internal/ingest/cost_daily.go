package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type CostDailyIngester struct{}

func (c *CostDailyIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "ce_daily-by-service")
}

func (c *CostDailyIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		ResultsByTime []struct {
			TimePeriod struct {
				Start string `json:"Start"`
			} `json:"TimePeriod"`
			Groups []struct {
				Keys    []string `json:"Keys"`
				Metrics map[string]struct {
					Amount string `json:"Amount"`
				} `json:"Metrics"`
			} `json:"Groups"`
		} `json:"ResultsByTime"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing cost daily JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO cost_daily
			(audit_run_id, date, account_id, service, region, unblended_cost, amortized_cost, source_file)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rbt := range v.ResultsByTime {
		date := rbt.TimePeriod.Start
		for _, g := range rbt.Groups {
			service := ""
			if len(g.Keys) > 0 {
				service = g.Keys[0]
			}
			unblended := parseAmount(g.Metrics["UnblendedCost"].Amount)
			amortized := parseAmount(g.Metrics["AmortizedCost"].Amount)

			if _, err := stmt.ExecContext(ctx, auditRunID, date, "", service, "", unblended, amortized, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}
