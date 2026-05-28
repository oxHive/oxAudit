package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDBPathUsesConfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	cfg := defaults()
	got := cfg.DBPath()
	want := filepath.Join(home, ".config", "oxaudit", "db", "aws_cost_audit.sqlite")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLoad_DBPathDerivedFromConfigPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yaml := `
audit:
  start_date: "2024-01-01"
  end_date: "2024-01-31"
output:
  directory: /some/other/place
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// DB must live next to the config file, not under output.directory
	want := filepath.Join(dir, "db", "aws_cost_audit.sqlite")
	got := cfg.DBPath()
	if got != want {
		t.Errorf("DBPath = %s, want %s", got, want)
	}
}

func TestConfigDir_fromLoadedConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("audit:\n  start_date: x\n  end_date: x\n"), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ConfigDir() != dir {
		t.Errorf("ConfigDir = %s, want %s", cfg.ConfigDir(), dir)
	}
}

func TestResolvePathExpandsTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "oxaudit")
	got := ResolvePath("~/.config/oxaudit")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestResolvePathPassthroughAbsolute(t *testing.T) {
	p := "/absolute/path"
	if ResolvePath(p) != p {
		t.Errorf("ResolvePath changed absolute path")
	}
}

func TestResolvePathPassthroughRelative(t *testing.T) {
	p := "relative/path"
	if ResolvePath(p) != p {
		t.Errorf("ResolvePath changed relative path")
	}
}

func TestResolvePathTildeOnly(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := ResolvePath("~")
	if got != home {
		t.Errorf("expected home dir for ~, got %q", got)
	}
}

func TestLoad_readsYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yaml := `
aws:
  profile: myprofile
  billing_region: us-west-2
audit:
  start_date: "2024-01-01"
  end_date: "2024-01-31"
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AWS.Profile != "myprofile" {
		t.Errorf("expected myprofile, got %q", cfg.AWS.Profile)
	}
	if cfg.AWS.BillingRegion != "us-west-2" {
		t.Errorf("expected us-west-2, got %q", cfg.AWS.BillingRegion)
	}
	if cfg.Audit.StartDate != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %q", cfg.Audit.StartDate)
	}
}

func TestLoad_appliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Minimal config — no thresholds set
	yaml := `
audit:
  start_date: "2024-01-01"
  end_date: "2024-01-31"
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Analysis.OldSnapshotThresholdDays == 0 {
		t.Error("expected OldSnapshotThresholdDays to have a default")
	}
	if cfg.Concurrency.Workers == 0 {
		t.Error("expected Workers to have a default")
	}
}

func TestLoad_missingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("{ invalid yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidate_missingStartDate(t *testing.T) {
	cfg := defaults()
	cfg.Audit.StartDate = ""
	cfg.Audit.EndDate = "2024-01-31"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing start_date")
	}
}

func TestValidate_missingEndDate(t *testing.T) {
	cfg := defaults()
	cfg.Audit.StartDate = "2024-01-01"
	cfg.Audit.EndDate = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing end_date")
	}
}

func TestValidate_missingProfile(t *testing.T) {
	cfg := defaults()
	cfg.AWS.Profile = ""
	cfg.Audit.StartDate = "2024-01-01"
	cfg.Audit.EndDate = "2024-01-31"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing aws.profile")
	}
}

func TestValidate_valid(t *testing.T) {
	cfg := defaults()
	cfg.Audit.StartDate = "2024-01-01"
	cfg.Audit.EndDate = "2024-01-31"
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWriteDefault_doesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := []byte("original content")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteDefault(path); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Error("WriteDefault should not overwrite existing file")
	}
}

func TestWriteDefault_createsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}
