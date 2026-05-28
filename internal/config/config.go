package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	configPath string // set by Load; DBPath/ConfigDir derive from this

	AWS struct {
		Profile       string `yaml:"profile"`
		BillingRegion string `yaml:"billing_region"`
		Regions       struct {
			Mode    string   `yaml:"mode"`    // "auto" | "manual"
			Include []string `yaml:"include"` // used when mode=manual
			Exclude []string `yaml:"exclude"` // always applied
		} `yaml:"regions"`
	} `yaml:"aws"`

	Audit struct {
		StartDate string `yaml:"start_date"`
		EndDate   string `yaml:"end_date"`
		Name      string `yaml:"name"`
		Notes     string `yaml:"notes"`
	} `yaml:"audit"`

	Output struct {
		Directory string `yaml:"directory"`
	} `yaml:"output"`

	Analysis struct {
		OldSnapshotThresholdDays   int      `yaml:"old_snapshot_threshold_days"`
		StoppedEC2ThresholdDays    int      `yaml:"stopped_ec2_threshold_days"`
		UnattachedEBSThresholdDays int      `yaml:"unattached_ebs_threshold_days"`
		MinMonthlySavingsUSD       float64  `yaml:"min_monthly_savings_usd"`
		RequiredTags               []string `yaml:"required_tags"`
		NATCostThresholdUSD        float64  `yaml:"nat_cost_threshold_usd"`
		CWLogsCostThresholdUSD     float64  `yaml:"cw_logs_cost_threshold_usd"`
		AnomalyStdDevMultiplier    float64  `yaml:"anomaly_stddev_multiplier"`
	} `yaml:"analysis"`

	Concurrency struct {
		Workers int `yaml:"workers"`
	} `yaml:"concurrency"`

}

// ResolvePath expands a leading ~ to the user home directory.
func ResolvePath(p string) string {
	if !strings.HasPrefix(p, "~/") && p != "~" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}

func defaults() Config {
	var cfg Config
	cfg.AWS.Profile = "default"
	cfg.AWS.BillingRegion = "us-east-1"
	cfg.AWS.Regions.Mode = "auto"
	cfg.Audit.StartDate = ""
	cfg.Audit.EndDate = ""
	cfg.Output.Directory = "~/.config/oxaudit"
	cfg.Analysis.OldSnapshotThresholdDays = 90
	cfg.Analysis.StoppedEC2ThresholdDays = 30
	cfg.Analysis.UnattachedEBSThresholdDays = 7
	cfg.Analysis.MinMonthlySavingsUSD = 5.0
	cfg.Analysis.NATCostThresholdUSD = 100.0
	cfg.Analysis.CWLogsCostThresholdUSD = 50.0
	cfg.Analysis.AnomalyStdDevMultiplier = 2.0
	cfg.Concurrency.Workers = 3
	return cfg
}

// Load reads a YAML config file and applies defaults for unset fields.
func Load(path string) (*Config, error) {
	cfg := defaults()
	cfg.configPath = path

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	d := defaults()
	if cfg.AWS.Profile == "" {
		cfg.AWS.Profile = d.AWS.Profile
	}
	if cfg.AWS.BillingRegion == "" {
		cfg.AWS.BillingRegion = d.AWS.BillingRegion
	}
	if cfg.AWS.Regions.Mode == "" {
		cfg.AWS.Regions.Mode = d.AWS.Regions.Mode
	}
	if cfg.Output.Directory == "" {
		cfg.Output.Directory = d.Output.Directory
	}
	if cfg.Analysis.OldSnapshotThresholdDays == 0 {
		cfg.Analysis.OldSnapshotThresholdDays = d.Analysis.OldSnapshotThresholdDays
	}
	if cfg.Analysis.StoppedEC2ThresholdDays == 0 {
		cfg.Analysis.StoppedEC2ThresholdDays = d.Analysis.StoppedEC2ThresholdDays
	}
	if cfg.Analysis.UnattachedEBSThresholdDays == 0 {
		cfg.Analysis.UnattachedEBSThresholdDays = d.Analysis.UnattachedEBSThresholdDays
	}
	if cfg.Analysis.NATCostThresholdUSD == 0 {
		cfg.Analysis.NATCostThresholdUSD = d.Analysis.NATCostThresholdUSD
	}
	if cfg.Analysis.CWLogsCostThresholdUSD == 0 {
		cfg.Analysis.CWLogsCostThresholdUSD = d.Analysis.CWLogsCostThresholdUSD
	}
	if cfg.Analysis.AnomalyStdDevMultiplier == 0 {
		cfg.Analysis.AnomalyStdDevMultiplier = d.Analysis.AnomalyStdDevMultiplier
	}
	if cfg.Concurrency.Workers == 0 {
		cfg.Concurrency.Workers = d.Concurrency.Workers
	}
}

// Validate checks that required fields are set.
func (c *Config) Validate() error {
	if c.Audit.StartDate == "" {
		return fmt.Errorf("audit.start_date is required")
	}
	if c.Audit.EndDate == "" {
		return fmt.Errorf("audit.end_date is required")
	}
	if c.AWS.Profile == "" {
		return fmt.Errorf("aws.profile is required")
	}
	return nil
}

// ConfigDir returns the directory containing the config file.
// Falls back to the resolved output directory if the config was not loaded from a file.
func (c *Config) ConfigDir() string {
	if c.configPath != "" {
		return filepath.Dir(c.configPath)
	}
	return ResolvePath(c.Output.Directory)
}

// DBPath returns the path to the shared SQLite database.
// The database always lives next to the config file, independent of output.directory.
func (c *Config) DBPath() string {
	return filepath.Join(c.ConfigDir(), "db", "aws_cost_audit.sqlite")
}

// WriteDefault writes a default config.yaml to path (does not overwrite).
func WriteDefault(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	cfg := defaults()
	cfg.Audit.StartDate = "YYYY-MM-DD"
	cfg.Audit.EndDate = "YYYY-MM-DD"

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	header := "# oxAudit configuration\n# Edit start_date and end_date before running.\n\n"
	return os.WriteFile(path, append([]byte(header), data...), 0644)
}
