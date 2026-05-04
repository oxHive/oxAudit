package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "oxaudit",
		Short: "Local AWS cost audit CLI",
		Long:  "oxaudit collects AWS cost and inventory data, generates deterministic findings, and exports LLM-ready summaries.",
		RunE:  runDashboard,
	}
)

// Execute is the entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: <output-dir>/config.yaml)")
}
