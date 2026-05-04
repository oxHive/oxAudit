package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/analyze"
	"github.com/graditya/oxaudit/internal/collect"
	"github.com/graditya/oxaudit/internal/export"
	"github.com/graditya/oxaudit/internal/ingest"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run the full audit pipeline: init → collect → ingest → analyze → export",
	RunE:  runAll,
}

func init() {
	rootCmd.AddCommand(allCmd)
}

func runAll(cmd *cobra.Command, args []string) error {
	const totalSteps = 5
	prog := progress.New(totalSteps)
	prog.PipelineStart()
	pipelineStart := time.Now()

	// ── Step 1: init ──────────────────────────────────────────────
	prog.StepStart(1, "init")
	t := time.Now()
	if err := runInit(cmd, args); err != nil {
		prog.StepFailed(1, "init", err)
		return err
	}
	prog.StepDone(1, "init", time.Since(t), "")

	// Load shared run context (config + DB + run folder + audit run ID)
	cfg, database, runFolder, auditRunID, err := loadRunContext()
	if err != nil {
		return err
	}
	defer func() {
		markRunComplete(database, auditRunID)
		database.Close()
	}()

	// ── Step 2: collect ───────────────────────────────────────────
	prog.StepStart(2, "collect")
	t = time.Now()
	collectResult, err := collect.Collect(context.Background(), cfg, database, runFolder, auditRunID, prog)
	if err != nil {
		prog.StepFailed(2, "collect", err)
		markRunFailed(database, auditRunID)
		return err
	}
	prog.StepDone(2, "collect", time.Since(t),
		fmt.Sprintf("%d files, %d errors", len(collectResult.Files), collectResult.ErrorCount))

	// ── Step 3: ingest ────────────────────────────────────────────
	prog.StepStart(3, "ingest")
	t = time.Now()
	rawDir := filepath.Join(runFolder, "raw")
	reg := ingest.NewRegistry()
	var ingestProcessed, ingestFailed int
	if err := reg.IngestDir(context.Background(), database, rawDir, auditRunID, func(name string, rows int, fileErr error) {
		if fileErr != nil {
			ingestFailed++
			prog.SubItem("%s ... ERROR: %v", name, fileErr)
		} else {
			ingestProcessed++
			prog.SubItem("%s ... done", name)
		}
	}); err != nil {
		prog.StepFailed(3, "ingest", err)
		markRunFailed(database, auditRunID)
		return err
	}
	prog.StepDone(3, "ingest", time.Since(t),
		fmt.Sprintf("%d files, %d errors", ingestProcessed, ingestFailed))

	// ── Step 4: analyze ───────────────────────────────────────────
	prog.StepStart(4, "analyze")
	t = time.Now()
	engine := analyze.New(cfg)
	totalFindings := 0
	p0Count, p1Count := 0, 0
	if err := engine.Run(context.Background(), database, cfg, auditRunID, func(ruleID, ruleName string, count int, ruleErr error) {
		if ruleErr != nil {
			prog.SubItem("%s %-20s ERROR: %v", ruleID, ruleName, ruleErr)
			return
		}
		noun := "findings"
		if count == 1 {
			noun = "finding"
		}
		prog.SubItem("%s %-20s %d %s", ruleID, ruleName, count, noun)
		totalFindings += count
	}); err != nil {
		prog.StepFailed(4, "analyze", err)
		markRunFailed(database, auditRunID)
		return err
	}
	// Re-count P0/P1 from DB for summary
	database.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM findings WHERE audit_run_id=? AND priority='P0'`, auditRunID).Scan(&p0Count)
	database.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM findings WHERE audit_run_id=? AND priority='P1'`, auditRunID).Scan(&p1Count)
	prog.StepDone(4, "analyze", time.Since(t), fmt.Sprintf("%d findings", totalFindings))

	// ── Step 5: export ────────────────────────────────────────────
	prog.StepStart(5, "export")
	t = time.Now()
	exportsDir := filepath.Join(runFolder, "exports")
	if err := export.Run(context.Background(), database, auditRunID, exportsDir, func(name, path string, fileErr error) {
		if fileErr != nil {
			prog.SubItem("%s ... ERROR: %v", name, fileErr)
		} else {
			prog.SubItem("%s ... done", name)
		}
	}); err != nil {
		prog.StepFailed(5, "export", err)
		markRunFailed(database, auditRunID)
		return err
	}
	prog.StepDone(5, "export", time.Since(t), "")

	// ── Final summary ─────────────────────────────────────────────
	var totalSavings float64
	database.QueryRowContext(context.Background(),
		`SELECT COALESCE(SUM(est_monthly_savings_usd),0) FROM findings WHERE audit_run_id=?`, auditRunID).Scan(&totalSavings)

	var regionCount int
	database.QueryRowContext(context.Background(),
		`SELECT COUNT(DISTINCT region) FROM resources WHERE audit_run_id=?`, auditRunID).Scan(&regionCount)

	var accountCount int
	database.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM accounts WHERE audit_run_id=?`, auditRunID).Scan(&accountCount)

	totalElapsed := time.Since(pipelineStart)

	prog.Summary([][2]string{
		{"Audit run", auditRunID},
		{"Run folder", runFolder},
		{"Database", cfg.DBPath()},
		{"Total time", fmt.Sprintf("%s", totalElapsed.Round(time.Millisecond))},
		{"Raw files", fmt.Sprintf("%d", len(collectResult.Files))},
		{"Accounts", fmt.Sprintf("%d", accountCount)},
		{"Regions", fmt.Sprintf("%d regions", regionCount)},
		{"Findings", fmt.Sprintf("%d  (%d P0 · %d P1 · %d others)", totalFindings, p0Count, p1Count, totalFindings-p0Count-p1Count)},
		{"Est. savings", fmt.Sprintf("$%.0f/month", totalSavings)},
		{"LLM digest", filepath.Join(exportsDir, "llm-digest.md")},
		{"Findings JSONL", filepath.Join(exportsDir, "findings.jsonl")},
		{"Collection errors", fmt.Sprintf("%d", collectResult.ErrorCount)},
	})

	return nil
}
