package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	internaldb "github.com/graditya/oxaudit/internal/db"
	"github.com/graditya/oxaudit/internal/mcpserver"
	"github.com/spf13/cobra"
)

var (
	mcpDBPath string
	mcpCmd    = &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server for Claude integration (stdio transport)",
		Long:  "Serves oxaudit data over the MCP protocol via stdin/stdout. Configure Claude Desktop or Claude Code to launch this command as an MCP server.",
		RunE:  runMCP,
	}
)

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpDBPath, "db", "", "path to SQLite database (default: ~/.config/oxaudit/db/aws_cost_audit.sqlite)")
}

func runMCP(_ *cobra.Command, _ []string) error {
	dbPath := mcpDBPath
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".config", "oxaudit", "db", "aws_cost_audit.sqlite")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("no database found at %s\nRun 'oxaudit all' first to generate audit data", dbPath)
	}

	db, err := internaldb.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	var auditRunID string
	err = db.QueryRowContext(
		context.Background(),
		`SELECT id FROM audit_run WHERE status = 'complete' ORDER BY executed_at DESC LIMIT 1`,
	).Scan(&auditRunID)
	if err != nil {
		return fmt.Errorf("no completed audit runs found\nRun 'oxaudit all' first to generate audit data")
	}

	srv := mcpserver.New(db, auditRunID)
	return srv.Serve(context.Background(), os.Stdin, os.Stdout)
}
