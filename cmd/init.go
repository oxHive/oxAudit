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
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var (
	initOutputDir string
	initForce     bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize audit directory structure and SQLite database",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initOutputDir, "output-dir", "~/.config/oxaudit", "directory for config file and database")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config.yaml")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	prog := progress.New(1)
	start := time.Now()
	prog.StepStart(1, "init")

	outDir := config.ResolvePath(initOutputDir)
	if cfgFile != "" {
		// derive output dir from the config location if --config is set
		outDir = filepath.Dir(cfgFile)
	}

	// Create config root layout — only the db dir lives here.
	// Raw data, exports, and logs are created per-run under output.directory.
	dirs := []string{
		filepath.Join(outDir, "db"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
		prog.SubItem("mkdir %s", d)
	}

	// Write default config.yaml
	cfgPath := filepath.Join(outDir, "config.yaml")
	if initForce {
		_ = os.Remove(cfgPath)
	}
	wrote := false
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := config.WriteDefault(cfgPath); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		wrote = true
	}
	if wrote {
		prog.SubItem("wrote %s", cfgPath)
	} else {
		prog.SubItem("config already exists: %s", cfgPath)
	}

	// Write .gitignore
	giPath := filepath.Join(outDir, ".gitignore")
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		gitignore := "# oxAudit output — do not commit\n*\n!config.yaml\n!.gitignore\n"
		if err := os.WriteFile(giPath, []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("writing .gitignore: %w", err)
		}
		prog.SubItem("wrote %s", giPath)
	}

	// Open / create SQLite database
	dbPath := filepath.Join(outDir, "db", "aws_cost_audit.sqlite")
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}
	defer database.Close()
	prog.SubItem("database ready: %s", dbPath)

	// Seed an init audit_run record so operators can verify the DB is functional
	if err := seedInitRecord(database, outDir); err != nil {
		return fmt.Errorf("seeding init record: %w", err)
	}

	prog.StepDone(1, "init", time.Since(start), "")
	fmt.Printf("\nConfig & database: %s\n", outDir)
	fmt.Printf("Config file      : %s\n", cfgPath)
	fmt.Printf("Database         : %s\n", dbPath)
	fmt.Println("\nRun data (raw files, exports) will be written to output.directory in config.yaml.")
	fmt.Println("Edit config.yaml to set your AWS profile, audit period, and output directory, then run: oxaudit all")
	return nil
}

func seedInitRecord(database *sql.DB, outDir string) error {
	now := time.Now().UTC()
	cfg := map[string]string{"initialized": now.Format(time.RFC3339)}
	cfgJSON, _ := json.Marshal(cfg)

	_, err := database.Exec(`
		INSERT OR IGNORE INTO audit_run
			(id, executed_at, generated_at, period_start, period_end, aws_profile, billing_region, run_folder, config_json, notes, status)
		VALUES (?, ?, ?, '', '', '', '', ?, ?, 'init record', 'init')`,
		"oxaudit-init-"+now.Format("20060102T150405Z"),
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		outDir,
		string(cfgJSON),
	)
	return err
}
