package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/binance-live/internal/binance"
	"github.com/binance-live/internal/config"
	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/logger"
	"github.com/binance-live/internal/publisher"
	"github.com/binance-live/internal/redis"
	"github.com/binance-live/internal/repository"
	"github.com/binance-live/internal/service"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config/config.yaml", "path to configuration file")
)

func main() {

	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {

		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.App.LogLevel, cfg.App.Environment)
	if err != nil {

		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Binance Live Data Collector",
		zap.String("app", cfg.App.Name),
		zap.String("environment", cfg.App.Environment),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run application
	if err := run(ctx, cfg, log); err != nil {

		log.Fatal("Application error", zap.Error(err))
	}

	log.Info("Application shutdown complete")
}

func run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {

	// Create a cancellable context for this run
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize database connection
	log.Info("Connecting to database...")
	db, err := database.New(&cfg.Database, log)
	if err != nil {

		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Run migrations
	migrationSQL, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {

		log.Warn("Failed to read migration file", zap.Error(err))
	} else {

		if err := db.RunMigrations(ctx, string(migrationSQL)); err != nil {

			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	// Initialize Redis client
	log.Info("Connecting to Redis...")
	redisClient, err := redis.New(&cfg.Redis, log)
	if err != nil {

		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	symbolRepo := repository.NewSymbolRepository(db)
	klineRepo := repository.NewKlineRepository(db)
	tickerRepo := repository.NewTickerRepository(db)
	syncStatusRepo := repository.NewSyncStatusRepository(db)

	// Initialize publisher
	pub := publisher.New(redisClient, log)

	// Initialize Binance client
	binanceClient := binance.NewClient(cfg, log)

	// Test Binance API connectivity
	log.Info("Testing Binance API connectivity...")
	if err := binanceClient.REST.Ping(ctx); err != nil {

		return fmt.Errorf("failed to connect to Binance API: %w", err)
	}
	log.Info("Binance API connection established")

	// Get active symbols
	symbols, err := symbolRepo.GetActiveSymbols(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active symbols: %w", err)
	}

	if len(symbols) == 0 {

		log.Warn("No active symbols found in database")
		return fmt.Errorf("no active symbols configured")
	}

	log.Info("Active symbols loaded", zap.Int("count", len(symbols)))

	// Publish active symbols to Redis
	if err := pub.PublishAllSymbols(ctx, symbols); err != nil {

		log.Warn("Failed to publish symbols to Redis", zap.Error(err))
	}

	// Initialize services
	// syncService := service.NewDataSyncService(
	// 	binanceClient,
	// 	symbolRepo,
	// 	klineRepo,
	// 	tickerRepo,
	// 	syncStatusRepo,
	// 	&cfg.Sync,
	// 	&cfg.Binance,
	// 	log,
	// )

	streamService := service.NewStreamService(
		binanceClient,
		klineRepo,
		tickerRepo,
		syncStatusRepo,
		pub,
		log,
	)

	// Synchronize missing data
	// if cfg.Sync.Enabled {
	// 	log.Info("Starting data synchronization...")
	// 	if err := syncService.SyncMissingData(ctx); err != nil {
	// 		log.Error("Data synchronization failed", zap.Error(err))
	// 		// Continue anyway - we can still stream live data
	// 	}
	// }

	// Start live data streaming
	log.Info("Starting live data streaming...")
	if err := streamService.Start(ctx, symbols); err != nil {

		return fmt.Errorf("failed to start streaming: %w", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Info("Application is running. Press Ctrl+C to stop")

	<-sigChan
	log.Info("Shutdown signal received, stopping services...")

	// Graceful shutdown
	cancel()

	// Stop streaming service
	if err := streamService.Stop(); err != nil {

		log.Error("Error stopping stream service", zap.Error(err))
	}

	log.Info("Services stopped successfully")
	return nil
}
