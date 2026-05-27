package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultDBPathUsesConfigDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	cfg := defaults()
	// defaults() returns tilde string; resolve it
	got := filepath.Join(ResolvePath(cfg.Output.Directory), "db", "aws_cost_audit.sqlite")
	want := filepath.Join(home, ".config", "oxaudit", "db", "aws_cost_audit.sqlite")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestResolvePathExpandsTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := ResolvePath("~/.config/oxaudit")
	if !strings.HasPrefix(got, home) {
		t.Errorf("ResolvePath did not expand tilde: got %s", got)
	}
}

func TestResolvePathPassthroughAbsolute(t *testing.T) {
	p := "/absolute/path"
	if ResolvePath(p) != p {
		t.Errorf("ResolvePath changed absolute path")
	}
}
