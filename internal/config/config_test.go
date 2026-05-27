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
