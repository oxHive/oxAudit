package mcpserver

// Tool is one MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is the JSON Schema for a tool's input.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property is one field in an InputSchema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

func allTools() []Tool {
	return []Tool{
		{
			Name:        "run_audit",
			Description: "Run the full oxaudit pipeline (collect → ingest → analyze → export). Use when the user wants to run or refresh the AWS cost audit. Returns pipeline output and any errors.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_summary",
			Description: "Get an overview of the latest audit run: accounts scanned, regions, finding counts by priority (P0–P3), and total estimated monthly savings in USD.",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "list_findings",
			Description: "List audit findings from the latest run. All filters are optional. Returns ID, title, priority, service, region, account, estimated savings, confidence, and risk for each finding.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"priority":        {Type: "string", Enum: []string{"P0", "P1", "P2", "P3"}, Description: "Filter by priority level"},
					"service":         {Type: "string", Description: "Filter by AWS service name, e.g. EC2, RDS, S3"},
					"category":        {Type: "string", Description: "Filter by category, e.g. Waste, Tagging, Anomaly"},
					"min_savings_usd": {Type: "number", Description: "Only return findings with estimated monthly savings >= this value"},
					"status":          {Type: "string", Enum: []string{"open", "dismissed", "resolved"}, Description: "Finding status filter (default: open)"},
					"limit":           {Type: "integer", Description: "Maximum results to return (default: 50)"},
				},
			},
		},
		{
			Name:        "get_finding",
			Description: "Get full details for a single finding by ID, including evidence, recommended action, linked resource IDs, and source files.",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{"finding_id"},
				Properties: map[string]Property{
					"finding_id": {Type: "string", Description: "The finding ID, e.g. FND-a1b2c3d4"},
				},
			},
		},
		{
			Name:        "get_cost_breakdown",
			Description: "Get monthly AWS cost totals grouped by service or account. Useful for identifying top cost drivers and month-over-month trends.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"group_by": {Type: "string", Enum: []string{"service", "account"}, Description: "Group costs by service or account (default: service)"},
					"months":   {Type: "integer", Description: "Number of months of history to include (default: 3)"},
				},
			},
		},
		{
			Name:        "query_resources",
			Description: "Search the AWS resource inventory. All filters are optional. Useful for answering inventory questions like 'how many stopped EC2 instances are in us-east-1?'",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"resource_type": {Type: "string", Description: "Resource type, e.g. aws:ec2:instance, aws:ec2:volume, aws:rds:instance"},
					"state":         {Type: "string", Description: "Resource state, e.g. stopped, available, running"},
					"region":        {Type: "string", Description: "AWS region, e.g. us-east-1"},
					"account_id":    {Type: "string", Description: "AWS account ID"},
					"limit":         {Type: "integer", Description: "Maximum results to return (default: 50)"},
				},
			},
		},
	}
}
