package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/db"
)

// runContext holds the resolved runtime dependencies for a pipeline stage.
type runContext struct {
	Cfg        *config.Config
	DB         *sql.DB
	RunFolder  string
	AuditRunID string
}

// loadRunContext resolves config, opens the database, creates the date-stamped run folder,
// and inserts an audit_run record. The caller is responsible for closing DB.
func loadRunContext() (*config.Config, *sql.DB, string, string, error) {
	// Determine config path
	outDir := "./aws-cost-audit"
	configPath := cfgFile
	if configPath == "" {
		configPath = filepath.Join(outDir, "config.yaml")
	} else {
		outDir = filepath.Dir(configPath)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("loading config: %w\n\nRun 'oxaudit init' first.", err)
	}
	outDir = cfg.Output.Directory

	if err := cfg.Validate(); err != nil {
		return nil, nil, "", "", fmt.Errorf("invalid config: %w", err)
	}

	// Date-stamped run folder
	now := time.Now().UTC()
	runFolderName := now.Format("2006-01-02T150405Z")
	runFolder := filepath.Join(outDir, runFolderName)

	// Create run-scoped directories
	for _, sub := range []string{
		filepath.Join("raw", "cost-explorer"),
		filepath.Join("raw", "inventory"),
		filepath.Join("raw", "recommendations"),
		filepath.Join("raw", "budgets"),
		filepath.Join("raw", "anomalies"),
		"exports",
		"logs",
	} {
		if err := os.MkdirAll(filepath.Join(runFolder, sub), 0755); err != nil {
			return nil, nil, "", "", fmt.Errorf("creating run directory: %w", err)
		}
	}

	// Open shared database
	dbPath := cfg.DBPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, nil, "", "", err
	}
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("opening database: %w", err)
	}

	// Build audit run ID
	auditRunID := fmt.Sprintf("aws-cost-audit-%s-%s-%s-%s",
		cfg.Audit.StartDate,
		cfg.Audit.EndDate,
		now.Format("20060102T150405Z"),
		sanitizeID(cfg.AWS.Profile),
	)

	// Marshal config snapshot
	cfgJSON, _ := json.Marshal(cfg)

	// Insert audit_run record
	_, err = database.Exec(`
		INSERT OR IGNORE INTO audit_run
			(id, executed_at, generated_at, period_start, period_end,
			 aws_profile, billing_region, run_folder, config_json, notes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'running')`,
		auditRunID,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		cfg.Audit.StartDate,
		cfg.Audit.EndDate,
		cfg.AWS.Profile,
		cfg.AWS.BillingRegion,
		runFolder,
		string(cfgJSON),
		cfg.Audit.Notes,
	)
	if err != nil {
		database.Close()
		return nil, nil, "", "", fmt.Errorf("inserting audit run record: %w", err)
	}

	return cfg, database, runFolder, auditRunID, nil
}

// markRunComplete updates the audit_run status to 'complete'.
func markRunComplete(database *sql.DB, auditRunID string) {
	database.Exec(`UPDATE audit_run SET status = 'complete' WHERE id = ?`, auditRunID)
}

// markRunFailed updates the audit_run status to 'failed'.
func markRunFailed(database *sql.DB, auditRunID string) {
	database.Exec(`UPDATE audit_run SET status = 'failed' WHERE id = ?`, auditRunID)
}

func sanitizeID(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			out = append(out, byte(c))
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}
