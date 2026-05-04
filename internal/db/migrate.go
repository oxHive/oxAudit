package db

import (
	"database/sql"
	"fmt"
)

// Migrate applies any pending schema migrations using PRAGMA user_version for tracking.
func Migrate(db *sql.DB) error {
	var current int
	if err := db.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("reading user_version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		version := i + 1
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("beginning migration %d: %w", version, err)
		}

		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %d: %w", version, err)
		}

		// SQLite doesn't allow PRAGMA inside a transaction for user_version via Exec
		// so we commit the DDL first, then update the version.
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", version, err)
		}

		if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
			return fmt.Errorf("updating user_version to %d: %w", version, err)
		}
	}

	return nil
}
