package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type AccountsIngester struct{}

func (a *AccountsIngester) Matches(path string) bool {
	return strings.Contains(strings.ToLower(path), "org_list-accounts")
}

func (a *AccountsIngester) Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var v struct {
		Accounts []struct {
			Id     string `json:"Id"`
			Name   string `json:"Name"`
			Email  string `json:"Email"`
			Status string `json:"Status"`
		} `json:"Accounts"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing accounts JSON: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO accounts
			(account_id, account_name, email, environment, owner, status, audit_run_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, acc := range v.Accounts {
		if _, err := stmt.ExecContext(ctx,
			acc.Id, acc.Name, acc.Email, "", "", acc.Status, auditRunID,
		); err != nil {
			return err
		}
	}
	return nil
}
