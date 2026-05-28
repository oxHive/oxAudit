package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/graditya/oxaudit/internal/db"
)

func TestOpen_inMemory(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:): %v", err)
	}
	defer d.Close()

	if err := d.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpen_createsSchema(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	tables := []string{"audit_run", "resources", "findings", "cost_monthly", "cost_daily"}
	for _, table := range tables {
		var count int
		err := d.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil || count == 0 {
			t.Errorf("expected table %q to exist", table)
		}
	}
}

func TestOpen_file(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	d, err := db.Open(path)
	if err != nil {
		t.Fatalf("Open(file): %v", err)
	}
	d.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected database file to be created")
	}
}

func TestMigrate_idempotent(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	// Running Migrate a second time should be safe.
	if err := db.Migrate(d); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestOpen_setsUserVersion(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var version int
	if err := d.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version == 0 {
		t.Fatal("expected user_version > 0 after migrations")
	}
}
