package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/progress"
	"github.com/graditya/oxaudit/internal/runner"
	"github.com/graditya/oxaudit/internal/types"
)

// discoverRegions runs ec2 describe-regions and returns the list of enabled region names.
func discoverRegions(ctx context.Context, r *runner.Runner, cfg *config.Config, runFolder string, prog *progress.Progress) ([]string, error) {
	if cfg.AWS.Regions.Mode == "manual" && len(cfg.AWS.Regions.Include) > 0 {
		prog.SubItem("regions: using manual list (%d regions)", len(cfg.AWS.Regions.Include))
		return filterExcluded(cfg.AWS.Regions.Include, cfg.AWS.Regions.Exclude), nil
	}

	specs := runner.RegionSpecs(cfg.AWS.BillingRegion)
	spec := specs[0]
	outPath := filepath.Join(runFolder, "raw", spec.OutputDir, spec.Name+".json")

	rf, err := r.Run(ctx, spec, outPath)
	prog.SubItem("regions: discovered via %s (%dms)", spec.Name, rf.DurationMs)
	if err != nil {
		return nil, fmt.Errorf("describe-regions failed: %w", err)
	}

	regions, err := parseRegions(outPath)
	if err != nil {
		return nil, fmt.Errorf("parsing regions: %w", err)
	}
	regions = filterExcluded(regions, cfg.AWS.Regions.Exclude)
	prog.SubItem("regions: %d enabled", len(regions))
	return regions, nil
}

func parseRegions(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v struct {
		Regions []struct {
			RegionName string `json:"RegionName"`
		} `json:"Regions"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(v.Regions))
	for _, r := range v.Regions {
		out = append(out, r.RegionName)
	}
	return out, nil
}

func filterExcluded(regions, exclude []string) []string {
	excl := make(map[string]bool, len(exclude))
	for _, e := range exclude {
		excl[e] = true
	}
	out := make([]string, 0, len(regions))
	for _, r := range regions {
		if !excl[r] {
			out = append(out, r)
		}
	}
	return out
}

// CollectInventory collects all regional inventory specs with a bounded worker pool.
func CollectInventory(ctx context.Context, r *runner.Runner, cfg *config.Config, regions []string, runFolder string, prog *progress.Progress) ([]types.RawFile, int, error) {
	workers := cfg.Concurrency.Workers
	if workers < 1 {
		workers = 3
	}

	type work struct {
		spec    types.CommandSpec
		outPath string
	}

	jobs := make(chan work)
	var (
		mu       sync.Mutex
		allFiles []types.RawFile
		errCount int64
	)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				rf, err := r.Run(ctx, job.spec, job.outPath)
				status := "done"
				if err != nil || rf.Status == "error" {
					status = "SKIPPED (error)"
					atomic.AddInt64(&errCount, 1)
				}
				prog.SubItem("inventory: %s [%s] ... %s", job.spec.Name, job.spec.Region, status)

				mu.Lock()
				allFiles = append(allFiles, rf)
				mu.Unlock()
			}
		}()
	}

	for _, region := range regions {
		for _, spec := range runner.InventorySpecs(region) {
			outPath := filepath.Join(runFolder, "raw", spec.OutputDir,
				fmt.Sprintf("%s_%s.json", spec.Name, spec.Region))
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				return allFiles, int(atomic.LoadInt64(&errCount)), ctx.Err()
			case jobs <- work{spec: spec, outPath: outPath}:
			}
		}
	}
	close(jobs)
	wg.Wait()

	return allFiles, int(atomic.LoadInt64(&errCount)), nil
}
