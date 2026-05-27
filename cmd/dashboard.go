package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/dashboard"
	"github.com/spf13/cobra"
)

var (
	dashboardPort    int
	dashboardNetwork bool
)

// runDashboard is also used as rootCmd.RunE so `oxaudit` (no args) opens the dashboard.
func runDashboard(cmd *cobra.Command, args []string) error {
	outDir := "~/.config/oxaudit"
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "~"
		}
		configPath = filepath.Join(home, ".config", "oxaudit", "config.yaml")
	}
	// Best-effort config load to find the configured output directory.
	if cfg, err := config.Load(configPath); err == nil {
		outDir = config.ResolvePath(cfg.Output.Directory)
	}

	host := "127.0.0.1"
	if dashboardNetwork {
		host = "0.0.0.0"
	}
	srv := dashboard.New(outDir, fmt.Sprintf("%s:%d", host, dashboardPort))
	return srv.Serve()
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the web dashboard to browse audit results",
	RunE:  runDashboard,
}

func init() {
	for _, f := range []*cobra.Command{dashboardCmd, rootCmd} {
		f.Flags().IntVar(&dashboardPort, "port", 7842, "port to listen on")
		f.Flags().BoolVar(&dashboardNetwork, "network", false, "bind to all interfaces so others on the LAN can connect")
	}
	rootCmd.AddCommand(dashboardCmd)
}
