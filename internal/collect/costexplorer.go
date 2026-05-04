package collect

import (
	"context"
	"path/filepath"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/graditya/oxaudit/internal/runner"
	"github.com/graditya/oxaudit/internal/types"
)

// CollectCostExplorer runs all Cost Explorer commands sequentially.
// Returns on the first required command failure.
func CollectCostExplorer(ctx context.Context, r *runner.Runner, cfg *config.Config, runFolder string, prog *progress.Progress) ([]types.RawFile, error) {
	specs := runner.CostExplorerSpecs(cfg.Audit.StartDate, cfg.Audit.EndDate, cfg.AWS.BillingRegion)
	var files []types.RawFile

	for _, spec := range specs {
		outPath := filepath.Join(runFolder, "raw", spec.OutputDir, spec.Name+".json")
		rf, err := r.Run(ctx, spec, outPath)
		files = append(files, rf)

		status := "done"
		if rf.Status == "error" {
			status = "ERROR"
		}
		prog.SubItem("cost-explorer: %s ... %s (%dms)", spec.Name, status, rf.DurationMs)

		if err != nil {
			return files, err // required command failed
		}
	}
	return files, nil
}
