package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/export"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Generate JSONL, Markdown, CSV, and LLM digest exports",
	RunE:  runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	_, database, runFolder, auditRunID, err := loadRunContext()
	if err != nil {
		return err
	}
	defer database.Close()

	prog := progress.New(1)
	start := time.Now()
	prog.StepStart(1, "export")

	exportsDir := filepath.Join(runFolder, "exports")
	err = export.Run(context.Background(), database, auditRunID, exportsDir, func(name, path string, fileErr error) {
		if fileErr != nil {
			prog.SubItem("%s ... ERROR: %v", name, fileErr)
		} else {
			prog.SubItem("%s ... done", name)
		}
	})
	if err != nil {
		prog.StepFailed(1, "export", err)
		return err
	}

	prog.StepDone(1, "export", time.Since(start), fmt.Sprintf("%s", exportsDir))
	return nil
}
