package ingest

import "testing"

func TestRegionFromPath_withRegion(t *testing.T) {
	got := regionFromPath("inventory/ec2_describe-volumes_us-east-1.json")
	if got != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", got)
	}
}

func TestRegionFromPath_noRegion(t *testing.T) {
	got := regionFromPath("org_list-accounts.json")
	if got != "" {
		t.Errorf("expected empty string for short name, got %q", got)
	}
}

func TestRegionFromPath_withSubdir(t *testing.T) {
	got := regionFromPath("/some/path/ec2_describe-instances_eu-west-1.json")
	if got != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %q", got)
	}
}

func TestTagsToMap_empty(t *testing.T) {
	m := tagsToMap(nil)
	if len(m) != 0 {
		t.Errorf("expected empty map for nil tags, got %v", m)
	}
}

func TestTagsToMap_multiple(t *testing.T) {
	tags := []awsTag{
		{Key: "Name", Value: "my-instance"},
		{Key: "Env", Value: "prod"},
	}
	m := tagsToMap(tags)
	if m["Name"] != "my-instance" {
		t.Errorf("expected my-instance, got %q", m["Name"])
	}
	if m["Env"] != "prod" {
		t.Errorf("expected prod, got %q", m["Env"])
	}
}

func TestTagValue_found(t *testing.T) {
	tags := []awsTag{{Key: "Name", Value: "test"}}
	if got := tagValue(tags, "Name"); got != "test" {
		t.Errorf("expected test, got %q", got)
	}
}

func TestTagValue_notFound(t *testing.T) {
	tags := []awsTag{{Key: "Env", Value: "prod"}}
	if got := tagValue(tags, "Name"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestNullStr_empty(t *testing.T) {
	if nullStr("") != nil {
		t.Error("expected nil for empty string")
	}
}

func TestNullStr_nonEmpty(t *testing.T) {
	if nullStr("hello") != "hello" {
		t.Error("expected hello for non-empty string")
	}
}
