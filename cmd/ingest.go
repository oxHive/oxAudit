package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/ingest"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Parse raw JSON files and load normalized records into SQLite",
	RunE:  runIngest,
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}

func runIngest(cmd *cobra.Command, args []string) error {
	_, database, runFolder, auditRunID, err := loadRunContext()
	if err != nil {
		return err
	}
	defer database.Close()

	prog := progress.New(1)
	start := time.Now()
	prog.StepStart(1, "ingest")

	rawDir := filepath.Join(runFolder, "raw")
	reg := ingest.NewRegistry()

	var processed, failed int
	err = reg.IngestDir(context.Background(), database, rawDir, auditRunID, func(name string, rows int, fileErr error) {
		if fileErr != nil {
			failed++
			prog.SubItem("%s ... ERROR: %v", name, fileErr)
		} else {
			processed++
			prog.SubItem("%s ... done", name)
		}
	})
	if err != nil {
		prog.StepFailed(1, "ingest", err)
		return err
	}

	detail := fmt.Sprintf("%d files, %d errors", processed, failed)
	prog.StepDone(1, "ingest", time.Since(start), detail)
	return nil
}
