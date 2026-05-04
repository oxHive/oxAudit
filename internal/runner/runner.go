package runner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/graditya/oxaudit/internal/types"
)

// Runner executes AWS CLI commands and saves output to disk.
type Runner struct {
	Profile    string
	AuditRunID string
	Logger     *CommandLogger
	// ExecCommand is replaceable for testing.
	ExecCommand func(ctx context.Context, name string, args ...string) *exec.Cmd
}

// New creates a Runner with the real exec.CommandContext.
func New(profile, auditRunID string, logger *CommandLogger) *Runner {
	return &Runner{
		Profile:     profile,
		AuditRunID:  auditRunID,
		Logger:      logger,
		ExecCommand: exec.CommandContext,
	}
}

// Run executes spec, streams stdout to outPath, computes a SHA-256 checksum, and logs the result.
// For optional commands, failures are logged but no error is returned.
func (r *Runner) Run(ctx context.Context, spec types.CommandSpec, outPath string) (types.RawFile, error) {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return types.RawFile{}, fmt.Errorf("creating output dir for %s: %w", spec.Name, err)
	}

	args := make([]string, 0, len(spec.Args)+4)
	if r.Profile != "" {
		args = append(args, "--profile", r.Profile)
	}
	args = append(args, "--output", "json")
	args = append(args, spec.Args...)

	cmd := r.ExecCommand(ctx, "aws", args...)

	outFile, err := os.Create(outPath)
	if err != nil {
		return types.RawFile{}, fmt.Errorf("creating output file %s: %w", outPath, err)
	}

	var stderr bytes.Buffer
	cmd.Stdout = outFile
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)
	outFile.Close()

	// Compute checksum and size
	checksum, size := checksumFile(outPath)

	rf := types.RawFile{
		AuditRunID:   r.AuditRunID,
		CommandName:  spec.Name,
		Service:      spec.Service,
		Region:       spec.Region,
		FilePath:     outPath,
		Checksum:     checksum,
		BytesWritten: size,
		CollectedAt:  time.Now().UTC(),
		DurationMs:   elapsed.Milliseconds(),
		Status:       "ok",
		Required:     spec.Required,
	}

	if runErr != nil {
		rf.Status = "error"
		rf.ErrorMsg = stderr.String()
	}

	if r.Logger != nil {
		r.Logger.Log(rf)
	}

	if runErr != nil && spec.Required {
		return rf, fmt.Errorf("required command %q failed: %s", spec.Name, rf.ErrorMsg)
	}

	return rf, nil
}

func checksumFile(path string) (string, int64) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0
	}
	defer f.Close()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", n
	}
	return fmt.Sprintf("%x", h.Sum(nil)), n
}
