package collect

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/graditya/oxaudit/internal/runner"
	"github.com/graditya/oxaudit/internal/types"
)

// Result summarizes the outcome of a collection run.
type Result struct {
	Files      []types.RawFile
	ErrorCount int
	Regions    []string
}

// Collect runs the full collection pipeline: accounts → cost-explorer → regional inventory.
func Collect(ctx context.Context, cfg *config.Config, db *sql.DB, runFolder, auditRunID string, prog *progress.Progress) (*Result, error) {
	logsDir := filepath.Join(runFolder, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating logs dir: %w", err)
	}

	logger, err := runner.NewCommandLogger(logsDir)
	if err != nil {
		return nil, fmt.Errorf("creating command logger: %w", err)
	}
	defer logger.Close()

	r := runner.New(cfg.AWS.Profile, auditRunID, logger)
	result := &Result{}

	// 1. Accounts (optional — not all accounts have Organizations access)
	accountFiles, err := collectAccounts(ctx, r, cfg, runFolder, prog)
	result.Files = append(result.Files, accountFiles...)
	if err != nil {
		result.ErrorCount++
		prog.SubItem("accounts: %v (skipping)", err)
	}

	// 2. Cost Explorer (sequential to respect rate limits)
	ceFiles, err := CollectCostExplorer(ctx, r, cfg, runFolder, prog)
	result.Files = append(result.Files, ceFiles...)
	if err != nil {
		return result, fmt.Errorf("cost explorer collection failed: %w", err)
	}

	// 3. Discover regions
	regions, err := discoverRegions(ctx, r, cfg, runFolder, prog)
	if err != nil {
		return result, fmt.Errorf("region discovery failed: %w", err)
	}
	result.Regions = regions

	// 4. Regional inventory (concurrent)
	invFiles, errCount, err := CollectInventory(ctx, r, cfg, regions, runFolder, prog)
	result.Files = append(result.Files, invFiles...)
	result.ErrorCount += errCount
	if err != nil {
		return result, err
	}

	// Persist raw_file rows to DB
	if err := insertRawFiles(ctx, db, result.Files); err != nil {
		return result, fmt.Errorf("persisting raw file records: %w", err)
	}

	return result, nil
}

func insertRawFiles(ctx context.Context, db *sql.DB, files []types.RawFile) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO raw_file
			(audit_run_id, command_name, service, region, file_path, checksum, bytes_written,
			 collected_at, duration_ms, status, error_msg, required)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, rf := range files {
		req := 0
		if rf.Required {
			req = 1
		}
		_, err := stmt.ExecContext(ctx,
			rf.AuditRunID, rf.CommandName, rf.Service, rf.Region,
			rf.FilePath, rf.Checksum, rf.BytesWritten,
			rf.CollectedAt.Format(time.RFC3339), rf.DurationMs,
			rf.Status, rf.ErrorMsg, req,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
