package main

import (
	"context"
	"fmt"
	"os"

	"github.com/binance-live/internal/binance"
	"github.com/binance-live/internal/config"
	"github.com/binance-live/internal/logger"
)

var (
	configPath string = "config/config.yaml"
)

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

	binanceClient := binance.NewClient(cfg, log)

	// startTime := time.Now().Add(-time.Hour * 24)
	// endTime := time.Now()

	data, err := binanceClient.REST.GetTicker24hr(context.Background(), "BTCUSDT")
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to get exchange info: %v", err)
	}

	fmt.Println(data)

}
