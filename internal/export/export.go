package export

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
)

// Run executes all exporters and writes output files to exportsDir.
// Returns a map of exporter name → output path for summary display.
func Run(ctx context.Context, db *sql.DB, auditRunID, exportsDir string, onFile func(name, path string, err error)) error {
	exporters := []Exporter{
		&JSONLExporter{},
		&CSVExporter{},
		&MarkdownExporter{},
		&LLMDigestExporter{},
		&RemediationExporter{},
	}

	for _, exp := range exporters {
		outPath := filepath.Join(exportsDir, exp.Name())
		err := exp.Export(ctx, db, auditRunID, outPath)
		if onFile != nil {
			onFile(exp.Name(), outPath, err)
		}
		if err != nil {
			return fmt.Errorf("exporter %s: %w", exp.Name(), err)
		}
	}
	return nil
}
