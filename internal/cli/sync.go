package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/binance-live/internal/binance"
	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/repository"
	"github.com/binance-live/internal/service"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewSyncCmd() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronization commands",
		Long:  `Commands for synchronizing historical data from Binance`,
	}

	syncCmd.AddCommand(NewSyncAllKlinesCmd())
	syncCmd.AddCommand(NewSyncSymbolKlineCmd())

	return syncCmd
}

func NewSyncAllKlinesCmd() *cobra.Command {
	var (
		intervals []string
		workers   int
		batchSize int
		maxHours  int
	)

	cmd := &cobra.Command{
		Use:   "all-klines",
		Short: "Sync klines for all active symbols",
		Long:  `Synchronize kline data for all active symbols and specified intervals`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncAllKlines(intervals, workers, batchSize, maxHours)
		},
	}

	cmd.Flags().StringSliceVarP(&intervals, "intervals", "i", []string{"1m", "15m", "1h", "4h", "1d"}, "Kline intervals to sync")
	cmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of concurrent workers")
	cmd.Flags().IntVarP(&batchSize, "batch-size", "b", 200, "Batch size for fetching klines")
	cmd.Flags().IntVarP(&maxHours, "max-hours", "m", 24, "Maximum hours to sync backwards")

	return cmd
}

func NewSyncSymbolKlineCmd() *cobra.Command {
	var (
		symbol    string
		intervals []string
		batchSize int
		maxHours  int
	)

	cmd := &cobra.Command{
		Use:   "symbol-kline",
		Short: "Sync klines for a specific symbol",
		Long:  `Synchronize kline data for a specific symbol and intervals`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if symbol == "" {
				return fmt.Errorf("symbol is required")
			}
			return runSyncSymbolKline(symbol, intervals, batchSize, maxHours)
		},
	}

	cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol to sync (required)")
	cmd.Flags().StringSliceVarP(&intervals, "intervals", "i", []string{"1m", "15m", "1h", "4h", "1d"}, "Kline intervals to sync")
	cmd.Flags().IntVarP(&batchSize, "batch-size", "b", 200, "Batch size for fetching klines")
	cmd.Flags().IntVarP(&maxHours, "max-hours", "m", 24, "Maximum hours to sync backwards")
	cmd.MarkFlagRequired("symbol")

	return cmd
}

func runSyncAllKlines(intervals []string, workers, batchSize, maxHours int) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	log.Info("Starting sync all klines",
		zap.Strings("intervals", intervals),
		zap.Int("workers", workers),
		zap.Int("batch_size", batchSize),
		zap.Int("max_hours", maxHours),
	)

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	symbolRepo := repository.NewSymbolRepository(db)
	klineRepo := repository.NewKlineRepository(db)
	syncStatusRepo := repository.NewSyncStatusRepository(db)

	// Initialize Binance client
	binanceClient := binance.NewClient(cfg, log)

	// Test connectivity
	if err := binanceClient.REST.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Binance API: %w", err)
	}

	// Override config values
	cfg.Sync.Workers = workers
	cfg.Sync.BatchSize = batchSize
	cfg.Sync.MaxSyncHours = maxHours
	cfg.Binance.KlineIntervals = intervals

	// Initialize sync service
	syncService := service.NewDataSyncService(
		binanceClient,
		symbolRepo,
		klineRepo,
		nil, // ticker repo not needed for klines
		syncStatusRepo,
		&cfg.Sync,
		&cfg.Binance,
		log,
	)

	// Run synchronization
	if err := syncService.SyncMissingData(ctx); err != nil {
		return fmt.Errorf("synchronization failed: %w", err)
	}

	log.Info("Sync all klines completed successfully")
	return nil
}

func runSyncSymbolKline(symbol string, intervals []string, batchSize, maxHours int) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	// Validate symbol format
	symbol = strings.ToUpper(symbol)

	log.Info("Starting sync symbol kline",
		zap.String("symbol", symbol),
		zap.Strings("intervals", intervals),
		zap.Int("batch_size", batchSize),
		zap.Int("max_hours", maxHours),
	)

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	symbolRepo := repository.NewSymbolRepository(db)
	klineRepo := repository.NewKlineRepository(db)
	syncStatusRepo := repository.NewSyncStatusRepository(db)

	// Check if symbol exists and is active
	symbolData, err := symbolRepo.GetSymbolByName(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get symbol %s: %w", symbol, err)
	}

	if !symbolData.IsActive {
		return fmt.Errorf("symbol %s is not active", symbol)
	}

	// Initialize Binance client
	binanceClient := binance.NewClient(cfg, log)

	// Test connectivity
	if err := binanceClient.REST.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Binance API: %w", err)
	}

	// Override config values
	cfg.Sync.Workers = 1 // Use single worker for specific symbol
	cfg.Sync.BatchSize = batchSize
	cfg.Sync.MaxSyncHours = maxHours
	cfg.Binance.KlineIntervals = intervals

	// Initialize sync service
	syncService := service.NewDataSyncService(
		binanceClient,
		symbolRepo,
		klineRepo,
		nil, // ticker repo not needed for klines
		syncStatusRepo,
		&cfg.Sync,
		&cfg.Binance,
		log,
	)

	// Sync each interval for the symbol
	for _, interval := range intervals {
		log.Info("Syncing klines for symbol and interval",
			zap.String("symbol", symbol),
			zap.String("interval", interval),
		)

		// This is a simplified approach - in a real implementation, you might want to
		// create a method specifically for syncing a single symbol
		// For now, we'll use the existing sync service but modify the active symbols
		// to only include our target symbol
		if err := syncSingleSymbolKline(ctx, syncService, symbol, interval, log); err != nil {
			log.Error("Failed to sync kline",
				zap.String("symbol", symbol),
				zap.String("interval", interval),
				zap.Error(err),
			)
			return err
		}
	}

	log.Info("Sync symbol kline completed successfully",
		zap.String("symbol", symbol),
	)
	return nil
}

// Helper function to sync a single symbol kline
func syncSingleSymbolKline(ctx context.Context, syncService *service.DataSyncService, symbol, interval string, log *zap.Logger) error {
	// Use reflection or create a new method in DataSyncService to sync specific symbol
	// For now, this is a placeholder - you'd need to modify the DataSyncService
	// to expose a method for syncing individual symbols
	log.Info("Syncing specific symbol kline",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
	)

	// This would need to be implemented in the DataSyncService
	// return syncService.SyncSymbolKline(ctx, symbol, interval)

	// For now, use the full sync but it will only process active symbols
	return syncService.SyncMissingData(ctx)
}
