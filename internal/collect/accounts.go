package collect

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/graditya/oxaudit/internal/runner"
	"github.com/graditya/oxaudit/internal/types"
)

func collectAccounts(ctx context.Context, r *runner.Runner, cfg *config.Config, runFolder string, prog *progress.Progress) ([]types.RawFile, error) {
	specs := runner.AccountSpecs()
	var files []types.RawFile

	for _, spec := range specs {
		outPath := filepath.Join(runFolder, "raw", spec.OutputDir, spec.Name+".json")
		rf, err := r.Run(ctx, spec, outPath)
		files = append(files, rf)
		if err != nil {
			return files, fmt.Errorf("collecting accounts: %w", err)
		}
		prog.SubItem("accounts: %s ... done (%dms)", spec.Name, rf.DurationMs)
	}
	return files, nil
}
