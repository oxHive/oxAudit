package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CostMonthlyIngester struct{}

func (c *CostMonthlyIngester) Matches(path string) bool {
	base := strings.ToLower(path)
	return strings.Contains(base, "ce_monthly-by-service") ||
		strings.Contains(base, "ce_monthly-by-account") ||
		strings.Contains(base, "ce_monthly-by-region")
}

func (c *CostMonthlyIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
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
					Unit   string `json:"Unit"`
				} `json:"Metrics"`
			} `json:"Groups"`
			Total map[string]struct {
				Amount string `json:"Amount"`
				Unit   string `json:"Unit"`
			} `json:"Total"`
		} `json:"ResultsByTime"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing cost monthly JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO cost_monthly
			(audit_run_id, month, account_id, account_name, service, region,
			 usage_type, operation, tag_key, tag_value,
			 unblended_cost, amortized_cost, usage_quantity, unit, source_file)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Detect grouping dimension from filename
	base := strings.ToLower(filePath)
	dimension := "SERVICE"
	if strings.Contains(base, "by-account") {
		dimension = "LINKED_ACCOUNT"
	} else if strings.Contains(base, "by-region") {
		dimension = "REGION"
	}

	for _, rbt := range v.ResultsByTime {
		month := rbt.TimePeriod.Start

		// Handle grouped results
		for _, g := range rbt.Groups {
			keyVal := ""
			if len(g.Keys) > 0 {
				keyVal = g.Keys[0]
			}

			accountID, accountName, service, region := "", "", "", ""
			switch dimension {
			case "LINKED_ACCOUNT":
				accountID = keyVal
			case "REGION":
				region = keyVal
			default:
				service = keyVal
			}

			unblended := parseAmount(g.Metrics["UnblendedCost"].Amount)
			amortized := parseAmount(g.Metrics["AmortizedCost"].Amount)
			usage := parseAmount(g.Metrics["UsageQuantity"].Amount)
			unit := g.Metrics["UnblendedCost"].Unit

			if _, err := stmt.ExecContext(ctx,
				auditRunID, month, accountID, accountName, service, region,
				"", "", "", "",
				unblended, amortized, usage, unit, filePath,
			); err != nil {
				return err
			}
		}

		// Handle ungrouped total
		if len(rbt.Groups) == 0 && rbt.Total != nil {
			unblended := parseAmount(rbt.Total["UnblendedCost"].Amount)
			amortized := parseAmount(rbt.Total["AmortizedCost"].Amount)
			if _, err := stmt.ExecContext(ctx,
				auditRunID, month, "", "", "", "", "", "", "", "",
				unblended, amortized, 0, "USD", filePath,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseAmount(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
