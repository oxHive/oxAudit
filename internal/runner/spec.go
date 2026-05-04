package runner

import "github.com/graditya/oxaudit/internal/types"

// CostExplorerSpecs returns the Cost Explorer commands for the given billing period.
func CostExplorerSpecs(startDate, endDate, billingRegion string) []types.CommandSpec {
	return []types.CommandSpec{
		{
			Name:      "ce_monthly-by-service",
			Service:   "ce",
			Region:    billingRegion,
			OutputDir: "cost-explorer",
			Required:  true,
			Args: []string{
				"--region", billingRegion,
				"ce", "get-cost-and-usage",
				"--time-period", "Start=" + startDate + ",End=" + endDate,
				"--granularity", "MONTHLY",
				"--metrics", "UnblendedCost", "AmortizedCost", "UsageQuantity",
				"--group-by", "Type=DIMENSION,Key=SERVICE",
			},
		},
		{
			Name:      "ce_monthly-by-account",
			Service:   "ce",
			Region:    billingRegion,
			OutputDir: "cost-explorer",
			Required:  false,
			Args: []string{
				"--region", billingRegion,
				"ce", "get-cost-and-usage",
				"--time-period", "Start=" + startDate + ",End=" + endDate,
				"--granularity", "MONTHLY",
				"--metrics", "UnblendedCost", "AmortizedCost",
				"--group-by", "Type=DIMENSION,Key=LINKED_ACCOUNT",
			},
		},
		{
			Name:      "ce_monthly-by-region",
			Service:   "ce",
			Region:    billingRegion,
			OutputDir: "cost-explorer",
			Required:  false,
			Args: []string{
				"--region", billingRegion,
				"ce", "get-cost-and-usage",
				"--time-period", "Start=" + startDate + ",End=" + endDate,
				"--granularity", "MONTHLY",
				"--metrics", "UnblendedCost", "AmortizedCost",
				"--group-by", "Type=DIMENSION,Key=REGION",
			},
		},
		{
			Name:      "ce_daily-by-service",
			Service:   "ce",
			Region:    billingRegion,
			OutputDir: "cost-explorer",
			Required:  false,
			Args: []string{
				"--region", billingRegion,
				"ce", "get-cost-and-usage",
				"--time-period", "Start=" + startDate + ",End=" + endDate,
				"--granularity", "DAILY",
				"--metrics", "UnblendedCost", "AmortizedCost",
				"--group-by", "Type=DIMENSION,Key=SERVICE",
			},
		},
	}
}

// AccountSpecs returns the Organizations account listing command.
func AccountSpecs() []types.CommandSpec {
	return []types.CommandSpec{
		{
			Name:      "org_list-accounts",
			Service:   "organizations",
			Region:    "",
			OutputDir: "cost-explorer",
			Required:  false,
			Args:      []string{"organizations", "list-accounts"},
		},
	}
}

// RegionSpecs returns the command to discover enabled regions.
func RegionSpecs(billingRegion string) []types.CommandSpec {
	return []types.CommandSpec{
		{
			Name:      "ec2_describe-regions",
			Service:   "ec2",
			Region:    billingRegion,
			OutputDir: "inventory",
			Required:  true,
			Args: []string{
				"--region", billingRegion,
				"ec2", "describe-regions",
				"--filters", "Name=opt-in-status,Values=opt-in-not-required,opted-in",
			},
		},
	}
}

// InventorySpecs returns all per-region inventory collection commands.
func InventorySpecs(region string) []types.CommandSpec {
	return []types.CommandSpec{
		{
			Name:      "ec2_describe-instances",
			Service:   "ec2",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args:      []string{"--region", region, "ec2", "describe-instances"},
		},
		{
			Name:      "ec2_describe-volumes",
			Service:   "ec2",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args:      []string{"--region", region, "ec2", "describe-volumes"},
		},
		{
			Name:      "ec2_describe-snapshots",
			Service:   "ec2",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args: []string{
				"--region", region,
				"ec2", "describe-snapshots",
				"--owner-ids", "self",
			},
		},
		{
			Name:      "ec2_describe-addresses",
			Service:   "ec2",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args:      []string{"--region", region, "ec2", "describe-addresses"},
		},
		{
			Name:      "ec2_describe-nat-gateways",
			Service:   "ec2",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args:      []string{"--region", region, "ec2", "describe-nat-gateways"},
		},
		{
			Name:      "rds_describe-db-instances",
			Service:   "rds",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args:      []string{"--region", region, "rds", "describe-db-instances"},
		},
		{
			Name:      "logs_describe-log-groups",
			Service:   "logs",
			Region:    region,
			OutputDir: "inventory",
			Required:  false,
			Args: []string{
				"--region", region,
				"logs", "describe-log-groups",
				"--limit", "50",
			},
		},
	}
}
