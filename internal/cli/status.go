package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/repository"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Status and monitoring commands",
		Long:  `Commands for checking synchronization status and system health`,
	}

	statusCmd.AddCommand(NewSyncStatusCmd())
	statusCmd.AddCommand(NewHealthCheckCmd())

	return statusCmd
}

func NewSyncStatusCmd() *cobra.Command {
	var symbol string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Show synchronization status",
		Long:  `Show synchronization status for all symbols or a specific symbol`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncStatus(symbol)
		},
	}

	cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Show status for specific symbol")

	return cmd
}

func NewHealthCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Perform health check",
		Long:  `Check connectivity to database, Redis, and Binance API`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealthCheck()
		},
	}

	return cmd
}

func runSyncStatus(symbol string) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	syncStatusRepo := repository.NewSyncStatusRepository(db)

	if symbol != "" {
		// Show status for specific symbol
		symbol = strings.ToUpper(symbol)
		// Note: GetSyncStatusesBySymbol method needs to be added to repository
		// For now, get all statuses and filter
		allStatuses, err := syncStatusRepo.GetAllSyncStatuses(ctx)
		if err != nil {
			return fmt.Errorf("failed to get sync status for %s: %w", symbol, err)
		}

		// Filter statuses for the specific symbol
		var statuses []models.SyncStatus
		for _, status := range allStatuses {
			if status.Symbol == symbol {
				statuses = append(statuses, status)
			}
		}

		if len(statuses) == 0 {
			fmt.Printf("No sync status found for symbol %s\n", symbol)
			return nil
		}

		fmt.Printf("Sync status for %s:\n\n", symbol)
		printSyncStatusTable(statuses)
	} else {
		// Show status for all active symbols
		statuses, err := syncStatusRepo.GetAllSyncStatuses(ctx)
		if err != nil {
			return fmt.Errorf("failed to get sync statuses: %w", err)
		}

		if len(statuses) == 0 {
			fmt.Println("No sync status records found")
			return nil
		}

		fmt.Printf("Sync status for all active symbols (%d records):\n\n", len(statuses))
		printSyncStatusTable(statuses)
	}

	return nil
}

func runHealthCheck() error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	fmt.Println("Performing health check...\n")

	// Check database connectivity
	fmt.Print("Database connection: ")
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		fmt.Printf("❌ FAILED - %v\n", err)
	} else {
		if err := db.HealthCheck(ctx); err != nil {
			fmt.Printf("❌ FAILED - %v\n", err)
		} else {
			fmt.Println("✅ OK")
		}
		db.Close()
	}

	// Check Redis connectivity
	fmt.Print("Redis connection: ")
	// Note: You'd need to implement Redis health check
	fmt.Println("⏭️  SKIPPED (not implemented)")

	// Check Binance API connectivity
	fmt.Print("Binance API: ")
	// Note: You'd need to implement this check
	fmt.Println("⏭️  SKIPPED (not implemented)")

	// Check configuration
	fmt.Print("Configuration: ")
	if cfg.App.Name == "" {
		fmt.Println("❌ FAILED - Invalid configuration")
	} else {
		fmt.Println("✅ OK")
	}

	fmt.Println("\nHealth check completed.")
	return nil
}

func printSyncStatusTable(statuses []models.SyncStatus) {
	// Print header
	fmt.Printf("%-15s %-10s %-10s %-15s %-15s %-10s %-20s\n",
		"SYMBOL", "DATA_TYPE", "INTERVAL", "LAST_SYNC", "LAST_DATA", "STATUS", "ERROR")
	fmt.Println(strings.Repeat("-", 110))

	// Print actual status data
	for _, status := range statuses {
		interval := ""
		if status.Interval != nil {
			interval = *status.Interval
		}

		errorMsg := ""
		if status.ErrorMessage != nil {
			errorMsg = *status.ErrorMessage
			if len(errorMsg) > 20 {
				errorMsg = errorMsg[:17] + "..."
			}
		}

		lastSync := formatTimestamp(status.LastSyncTime)
		lastData := formatTimestamp(status.LastDataTime)

		fmt.Printf("%-15s %-10s %-10s %-15s %-15s %-10s %-20s\n",
			status.Symbol, status.DataType, interval, lastSync, lastData, status.Status, errorMsg)
	}
}

func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "Never"
	}
	t := time.UnixMilli(timestamp)
	now := time.Now()

	diff := now.Sub(t)
	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return t.Format("2006-01-02 15:04")
	}
}
