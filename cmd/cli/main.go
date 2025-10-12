package main

import (
	"fmt"
	"os"

	"github.com/binance-live/internal/cli"
	"github.com/binance-live/internal/config"
	"github.com/binance-live/internal/logger"
	"github.com/spf13/cobra"
)

var (
	configPath string
	rootCmd    = &cobra.Command{
		Use:   "binance-cli",
		Short: "Binance Live Data CLI utilities",
		Long:  `Command line utilities for managing Binance live data collection and synchronization`,
	}
)

func init() {

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config/config.yaml", "path to configuration file")

	// Add subcommands
	rootCmd.AddCommand(cli.NewSyncCmd())
	rootCmd.AddCommand(cli.NewSymbolsCmd())
	rootCmd.AddCommand(cli.NewStatusCmd())
}

func main() {

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to load configuration: %v", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.App.LogLevel, cfg.App.Environment)
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v", err)
	}
	defer log.Sync()

	if err := rootCmd.Execute(); err != nil {

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
