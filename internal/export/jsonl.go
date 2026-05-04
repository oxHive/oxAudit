package export

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
)

type JSONLExporter struct{}

func (e *JSONLExporter) Name() string { return "findings.jsonl" }

func (e *JSONLExporter) Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error {
	findings, err := loadFindings(ctx, db, auditRunID)
	if err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	for _, finding := range findings {
		if err := enc.Encode(finding); err != nil {
			return err
		}
	}
	return nil
}
