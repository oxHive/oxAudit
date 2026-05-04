package ingest

import (
	"context"
	"database/sql"
)

// Ingester parses one type of raw AWS CLI JSON file and writes normalized rows into SQLite.
type Ingester interface {
	// Matches returns true if this ingester should handle the given file path.
	Matches(filePath string) bool
	// Ingest parses the file and inserts rows within the provided transaction.
	Ingest(ctx context.Context, tx *sql.Tx, filePath, auditRunID string) error
}
