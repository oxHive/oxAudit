package ingest

import (
	"path/filepath"
	"strings"
)

// resourceInsertSQL is shared by all resource ingesters.
const resourceInsertSQL = `
INSERT OR REPLACE INTO resources
	(audit_run_id, resource_id, resource_type, account_id, account_name, region, service,
	 state, name, arn, created_at, size_gib, instance_type, volume_type, iops,
	 est_monthly_cost, tags_json, raw_json, source_file, discovered_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

type awsTag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func tagsToMap(tags []awsTag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[t.Key] = t.Value
	}
	return m
}

func tagValue(tags []awsTag, key string) string {
	for _, t := range tags {
		if t.Key == key {
			return t.Value
		}
	}
	return ""
}

// regionFromPath extracts the AWS region from a file path like "ec2_describe-volumes_us-east-1.json".
func regionFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".json")
	parts := strings.Split(base, "_")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return ""
}

// nullStr returns nil if s is empty, otherwise s — for nullable SQL columns.
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
