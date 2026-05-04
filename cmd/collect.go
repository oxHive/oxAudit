package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/graditya/oxaudit/internal/collect"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Run AWS CLI commands and save raw JSON evidence",
	RunE:  runCollect,
}

func init() {
	rootCmd.AddCommand(collectCmd)
}

func runCollect(cmd *cobra.Command, args []string) error {
	cfg, database, runFolder, auditRunID, err := loadRunContext()
	if err != nil {
		return err
	}
	defer database.Close()

	prog := progress.New(1)
	start := time.Now()
	prog.StepStart(1, "collect")

	result, err := collect.Collect(context.Background(), cfg, database, runFolder, auditRunID, prog)
	if err != nil {
		prog.StepFailed(1, "collect", err)
		return err
	}

	detail := fmt.Sprintf("%d files, %d errors", len(result.Files), result.ErrorCount)
	prog.StepDone(1, "collect", time.Since(start), detail)
	return nil
}
