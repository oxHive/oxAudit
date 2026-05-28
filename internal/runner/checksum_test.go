package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumFile_returnsHashAndSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := []byte(`{"key":"value"}`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hash, size := checksumFile(path)
	if hash == "" {
		t.Error("expected non-empty checksum")
	}
	if size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), size)
	}
}

func TestChecksumFile_deterministicForSameContent(t *testing.T) {
	dir := t.TempDir()
	content := []byte(`{"key":"value"}`)

	p1 := filepath.Join(dir, "a.json")
	p2 := filepath.Join(dir, "b.json")
	os.WriteFile(p1, content, 0644)
	os.WriteFile(p2, content, 0644)

	h1, _ := checksumFile(p1)
	h2, _ := checksumFile(p2)
	if h1 != h2 {
		t.Errorf("expected same checksum for same content, got %q and %q", h1, h2)
	}
}

func TestChecksumFile_differentForDifferentContent(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.json")
	p2 := filepath.Join(dir, "b.json")
	os.WriteFile(p1, []byte("content-a"), 0644)
	os.WriteFile(p2, []byte("content-b"), 0644)

	h1, _ := checksumFile(p1)
	h2, _ := checksumFile(p2)
	if h1 == h2 {
		t.Error("expected different checksums for different content")
	}
}

func TestChecksumFile_missingFile(t *testing.T) {
	hash, size := checksumFile("/nonexistent/file.json")
	if hash != "" {
		t.Errorf("expected empty hash for missing file, got %q", hash)
	}
	if size != 0 {
		t.Errorf("expected 0 size for missing file, got %d", size)
	}
}

func TestNewRunner_doesNotPanic(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewCommandLogger(dir)
	if err != nil {
		t.Fatalf("NewCommandLogger: %v", err)
	}
	defer logger.Close()

	r := New("default", "run-001", logger)
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}
