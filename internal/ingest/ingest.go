package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// Registry maps file patterns to Ingesters and orchestrates ingestion of a raw directory.
type Registry struct {
	ingesters []Ingester
}

// NewRegistry builds the default registry with all MVP ingesters.
func NewRegistry() *Registry {
	return &Registry{
		ingesters: []Ingester{
			&CostMonthlyIngester{},
			&CostDailyIngester{},
			&AccountsIngester{},
			&EC2InstancesIngester{},
			&EBSVolumesIngester{},
			&EBSSnapshotsIngester{},
			&ElasticIPsIngester{},
			&NATGatewaysIngester{},
			&RDSInstancesIngester{},
			&CWLogGroupsIngester{},
		},
	}
}

// IngestDir walks rawDir, dispatches each .json file to its ingester, and returns
// the number of files processed and how many failed.
func (reg *Registry) IngestDir(ctx context.Context, db *sql.DB, rawDir, auditRunID string, onFile func(name string, rows int, err error)) error {
	return filepath.WalkDir(rawDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		if err := reg.dispatch(ctx, db, path, auditRunID, onFile); err != nil {
			// Log but don't abort — partial ingestion is acceptable
			if onFile != nil {
				onFile(filepath.Base(path), 0, err)
			}
		}
		return nil
	})
}

func (reg *Registry) dispatch(ctx context.Context, db *sql.DB, path, auditRunID string, onFile func(string, int, error)) error {
	for _, ing := range reg.ingesters {
		if !ing.Matches(path) {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning tx for %s: %w", path, err)
		}

		// Count rows before to compute delta
		if err := ing.Ingest(ctx, tx, path, auditRunID); err != nil {
			tx.Rollback()
			if onFile != nil {
				onFile(filepath.Base(path), 0, err)
			}
			return nil // don't abort the walk for one bad file
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing tx for %s: %w", path, err)
		}
		if onFile != nil {
			onFile(filepath.Base(path), 0, nil)
		}
		return nil
	}
	// No ingester matched — silently skip
	return nil
}
