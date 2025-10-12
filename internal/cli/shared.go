package cli

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/config"
	"github.com/binance-live/internal/logger"
	"go.uber.org/zap"
)

// getSharedResources loads config and creates shared resources used by commands
// This is the same function as in main.go but accessible to subcommands
func getSharedResources() (*config.Config, *zap.Logger, context.Context, error) {
	
	// Load configuration from the global configPath variable
	// Note: This assumes the configPath is set by the root command
	configPath := "config/config.yaml" // Default path

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.App.LogLevel, cfg.App.Environment)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	ctx := context.Background()
	return cfg, log, ctx, nil
}
