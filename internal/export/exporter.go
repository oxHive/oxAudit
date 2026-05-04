package export

import (
	"context"
	"database/sql"
)

// Exporter writes one output file from the findings in SQLite.
type Exporter interface {
	Name() string
	Export(ctx context.Context, db *sql.DB, auditRunID, outPath string) error
}
