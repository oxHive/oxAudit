package types

import (
	"database/sql"
	"time"
)

// AuditRun represents one execution of the audit pipeline.
type AuditRun struct {
	ID            string
	ExecutedAt    time.Time
	GeneratedAt   time.Time
	PeriodStart   string
	PeriodEnd     string
	AWSProfile    string
	BillingRegion string
	RunFolder     string
	ConfigJSON    string
	Notes         string
	Status        string // "running" | "complete" | "failed"
}

// RawFile represents one collected AWS CLI output file.
type RawFile struct {
	ID           int64
	AuditRunID   string
	CommandName  string
	Service      string
	Region       string
	FilePath     string
	Checksum     string // SHA-256 hex
	BytesWritten int64
	CollectedAt  time.Time
	DurationMs   int64
	Status       string // "ok" | "error" | "skipped"
	ErrorMsg     string
	Required     bool
}

// Resource is a normalized inventory row.
type Resource struct {
	PK             int64
	AuditRunID     string
	ResourceID     string
	ResourceType   string // "aws:ec2:volume", "aws:ec2:instance", etc.
	AccountID      string
	AccountName    string
	Region         string
	Service        string
	State          string
	Name           string
	ARN            string
	CreatedAt      sql.NullTime
	SizeGiB        sql.NullFloat64
	InstanceType   string
	VolumeType     string
	IOPS           sql.NullInt64
	EstMonthlyCost sql.NullFloat64
	TagsJSON       string // JSON object
	RawJSON        string // full original AWS JSON for the resource
	SourceFile     string
	DiscoveredAt   time.Time
}

// Finding represents one deterministic audit finding.
type Finding struct {
	ID                   string // FND-{hex8}
	AuditRunID           string
	Priority             string // "P0"|"P1"|"P2"|"P3"
	Category             string // "Waste"|"Tagging"|"Anomaly"|etc.
	Service              string
	AccountID            string
	AccountName          string
	Region               string
	Title                string
	Summary              string
	Evidence             string
	RecommendedAction    string
	EstMonthlySavingsUSD float64
	Confidence           string // "High"|"Medium"|"Low"
	Risk                 string // "Low"|"Medium"|"High"
	Owner                string
	Status               string // "open"|"dismissed"|"resolved"
	ResourceIDsJSON      string // JSON array
	TagsJSON             string
	SourceFilesJSON      string // JSON array
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CostMonthly is a normalized monthly cost record.
type CostMonthly struct {
	AuditRunID    string
	Month         string // "YYYY-MM-DD" (first of month)
	AccountID     string
	AccountName   string
	Service       string
	Region        string
	UsageType     string
	Operation     string
	TagKey        string
	TagValue      string
	UnblendedCost float64
	AmortizedCost float64
	UsageQuantity float64
	Unit          string
	SourceFile    string
}

// CostDaily is a normalized daily cost record.
type CostDaily struct {
	AuditRunID    string
	Date          string // "YYYY-MM-DD"
	AccountID     string
	Service       string
	Region        string
	UnblendedCost float64
	AmortizedCost float64
	SourceFile    string
}

// Account stores AWS account metadata.
type Account struct {
	AccountID   string
	AccountName string
	Email       string
	Environment string
	Owner       string
	Status      string
	AuditRunID  string
}

// Recommendation stores a rightsizing or optimization recommendation.
type Recommendation struct {
	ID                     string
	AuditRunID             string
	RecSource              string // "compute-optimizer" | "ce-rightsizing"
	Service                string
	AccountID              string
	Region                 string
	ResourceID             string
	CurrentResourceType    string
	RecommendedResourceType string
	FindingText            string
	EstMonthlySavingsUSD   float64
	Confidence             string
	RawJSON                string
	SourceFile             string
}

// CommandSpec defines one AWS CLI command to run during collection.
type CommandSpec struct {
	Name      string   // used in file naming and log
	Service   string   // "ec2", "ce", "rds", etc.
	Region    string   // "" for global commands
	Args      []string // everything after "aws --profile P --output json"
	OutputDir string   // subdirectory under raw/ (e.g. "cost-explorer", "inventory")
	Required  bool     // if true, failure aborts the run
}
