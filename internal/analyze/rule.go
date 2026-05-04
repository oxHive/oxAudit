package analyze

import (
	"context"
	"database/sql"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/types"
)

// Rule is a deterministic finding rule that queries SQLite and returns findings.
type Rule interface {
	ID() string   // e.g. "AWS-WASTE-001"
	Name() string // human-readable name
	Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string) ([]types.Finding, error)
}
