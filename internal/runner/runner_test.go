package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/graditya/oxaudit/internal/runner"
	"github.com/graditya/oxaudit/internal/types"
)

// ---------------------------------------------------------------------------
// Spec generators — pure functions, no external calls
// ---------------------------------------------------------------------------

func TestCostExplorerSpecs_returnsSpecs(t *testing.T) {
	specs := runner.CostExplorerSpecs("2024-01-01", "2024-01-31", "us-east-1")
	if len(specs) == 0 {
		t.Fatal("expected at least one cost explorer spec")
	}
}

func TestCostExplorerSpecs_firstIsRequired(t *testing.T) {
	specs := runner.CostExplorerSpecs("2024-01-01", "2024-01-31", "us-east-1")
	if !specs[0].Required {
		t.Error("expected first cost explorer spec to be required")
	}
}

func TestCostExplorerSpecs_containsDates(t *testing.T) {
	const start, end = "2024-03-01", "2024-03-31"
	specs := runner.CostExplorerSpecs(start, end, "us-east-1")
	found := false
	for _, s := range specs {
		for _, arg := range s.Args {
			if arg == "Start="+start+",End="+end {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected date range in spec args")
	}
}

func TestAccountSpecs_returnsOne(t *testing.T) {
	specs := runner.AccountSpecs()
	if len(specs) != 1 {
		t.Fatalf("expected 1 account spec, got %d", len(specs))
	}
	if specs[0].Name != "org_list-accounts" {
		t.Errorf("unexpected spec name: %q", specs[0].Name)
	}
}

func TestRegionSpecs_returnsOne(t *testing.T) {
	specs := runner.RegionSpecs("us-east-1")
	if len(specs) != 1 {
		t.Fatalf("expected 1 region spec, got %d", len(specs))
	}
	if !specs[0].Required {
		t.Error("expected region spec to be required")
	}
}

func TestRegionSpecs_usesProvidedRegion(t *testing.T) {
	specs := runner.RegionSpecs("eu-west-1")
	if specs[0].Region != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %q", specs[0].Region)
	}
}

func TestInventorySpecs_returnsMultiple(t *testing.T) {
	specs := runner.InventorySpecs("us-east-1")
	if len(specs) < 3 {
		t.Fatalf("expected multiple inventory specs, got %d", len(specs))
	}
}

func TestInventorySpecs_allHaveRegion(t *testing.T) {
	const region = "ap-southeast-1"
	specs := runner.InventorySpecs(region)
	for _, s := range specs {
		if s.Region != region {
			t.Errorf("spec %q has region %q, want %q", s.Name, s.Region, region)
		}
	}
}

func TestInventorySpecs_allHaveNames(t *testing.T) {
	specs := runner.InventorySpecs("us-east-1")
	for i, s := range specs {
		if s.Name == "" {
			t.Errorf("spec[%d] has empty name", i)
		}
	}
}

// ---------------------------------------------------------------------------
// CommandLogger
// ---------------------------------------------------------------------------

func TestNewCommandLogger_createsFiles(t *testing.T) {
	dir := t.TempDir()
	logger, err := runner.NewCommandLogger(dir)
	if err != nil {
		t.Fatalf("NewCommandLogger: %v", err)
	}
	defer logger.Close()

	if _, err := os.Stat(filepath.Join(dir, "commands.jsonl")); os.IsNotExist(err) {
		t.Error("expected commands.jsonl to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, "errors.jsonl")); os.IsNotExist(err) {
		t.Error("expected errors.jsonl to be created")
	}
}

func TestNewCommandLogger_missingDir(t *testing.T) {
	_, err := runner.NewCommandLogger("/nonexistent/dir/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestCommandLogger_logWritesEntry(t *testing.T) {
	dir := t.TempDir()
	logger, err := runner.NewCommandLogger(dir)
	if err != nil {
		t.Fatalf("NewCommandLogger: %v", err)
	}
	defer logger.Close()

	logger.Log(types.RawFile{
		CommandName: "ec2_describe-volumes",
		Status:      "ok",
	})

	data, err := os.ReadFile(filepath.Join(dir, "commands.jsonl"))
	if err != nil {
		t.Fatalf("read commands.jsonl: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected commands.jsonl to have content")
	}
}

func TestCommandLogger_errorWrittenToBothFiles(t *testing.T) {
	dir := t.TempDir()
	logger, err := runner.NewCommandLogger(dir)
	if err != nil {
		t.Fatalf("NewCommandLogger: %v", err)
	}
	defer logger.Close()

	logger.Log(types.RawFile{
		CommandName: "ec2_describe-volumes",
		Status:      "error",
		ErrorMsg:    "access denied",
	})

	errData, _ := os.ReadFile(filepath.Join(dir, "errors.jsonl"))
	if len(errData) == 0 {
		t.Error("expected errors.jsonl to have content for failed commands")
	}
}

func TestCommandLogger_close(t *testing.T) {
	dir := t.TempDir()
	logger, err := runner.NewCommandLogger(dir)
	if err != nil {
		t.Fatalf("NewCommandLogger: %v", err)
	}
	if err := logger.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
