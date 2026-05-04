package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/graditya/oxaudit/internal/analyze"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Run deterministic finding rules against the SQLite database",
	RunE:  runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg, database, _, auditRunID, err := loadRunContext()
	if err != nil {
		return err
	}
	defer database.Close()

	prog := progress.New(1)
	start := time.Now()
	prog.StepStart(1, "analyze")

	engine := analyze.New(cfg)
	totalFindings := 0

	err = engine.Run(context.Background(), database, cfg, auditRunID, func(ruleID, ruleName string, count int, ruleErr error) {
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
	})
	if err != nil {
		prog.StepFailed(1, "analyze", err)
		return err
	}

	prog.StepDone(1, "analyze", time.Since(start), fmt.Sprintf("%d findings", totalFindings))
	return nil
}
