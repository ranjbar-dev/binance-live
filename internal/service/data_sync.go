package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/binance-live/internal/binance"
	"github.com/binance-live/internal/config"
	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/repository"
	"go.uber.org/zap"
)

// DataSyncService handles historical data synchronization
type DataSyncService struct {
	binanceClient  *binance.Client
	symbolRepo     *repository.SymbolRepository
	klineRepo      *repository.KlineRepository
	tickerRepo     *repository.TickerRepository
	syncStatusRepo *repository.SyncStatusRepository
	config         *config.SyncConfig
	binanceConfig  *config.BinanceConfig
	logger         *zap.Logger
}

// NewDataSyncService creates a new data sync service
func NewDataSyncService(
	binanceClient *binance.Client,
	symbolRepo *repository.SymbolRepository,
	klineRepo *repository.KlineRepository,
	tickerRepo *repository.TickerRepository,
	syncStatusRepo *repository.SyncStatusRepository,
	cfg *config.SyncConfig,
	binanceCfg *config.BinanceConfig,
	logger *zap.Logger,
) *DataSyncService {
	return &DataSyncService{
		binanceClient:  binanceClient,
		symbolRepo:     symbolRepo,
		klineRepo:      klineRepo,
		tickerRepo:     tickerRepo,
		syncStatusRepo: syncStatusRepo,
		config:         cfg,
		binanceConfig:  binanceCfg,
		logger:         logger,
	}
}

// SyncMissingData synchronizes missing data for all active symbols
func (s *DataSyncService) SyncMissingData(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Data synchronization is disabled")
		return nil
	}

	s.logger.Info("Starting data synchronization")

	// Get all active symbols
	symbols, err := s.symbolRepo.GetActiveSymbols(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active symbols: %w", err)
	}

	if len(symbols) == 0 {
		s.logger.Warn("No active symbols found")
		return nil
	}

	s.logger.Info("Found active symbols", zap.Int("count", len(symbols)))

	// Create worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.Workers)
	errChan := make(chan error, len(symbols)*len(s.binanceConfig.KlineIntervals))

	// Sync klines for each symbol and interval with sequential processing
	// Process symbols sequentially to minimize database connection pressure
	for _, symbol := range symbols {
		for _, interval := range s.binanceConfig.KlineIntervals {
			wg.Add(1)

			go func(sym models.Symbol, intv string) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Add small delay before starting each operation
				select {
				case <-ctx.Done():
					return
				case <-time.After(50 * time.Millisecond):
				}

				if err := s.syncKlinesForSymbol(ctx, sym.Symbol, intv); err != nil {
					s.logger.Error("Failed to sync klines",
						zap.String("symbol", sym.Symbol),
						zap.String("interval", intv),
						zap.Error(err),
					)
					errChan <- err
				}
			}(symbol, interval)
		}
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var syncErrors []error
	for err := range errChan {
		syncErrors = append(syncErrors, err)
	}

	if len(syncErrors) > 0 {
		s.logger.Warn("Data synchronization completed with errors",
			zap.Int("error_count", len(syncErrors)),
		)
	} else {
		s.logger.Info("Data synchronization completed successfully")
	}

	return nil
}

// syncKlinesForSymbol synchronizes kline data for a specific symbol and interval
func (s *DataSyncService) syncKlinesForSymbol(ctx context.Context, symbol, interval string) error {

	s.logger.Info("Syncing klines",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
	)

	// Get sync status
	syncStatus, err := s.syncStatusRepo.GetSyncStatus(ctx, symbol, "kline", &interval)
	if err != nil {

		return fmt.Errorf("failed to get sync status: %w", err)
	}

	// Determine start time for sync
	var startTime time.Time
	if syncStatus != nil && syncStatus.LastDataTime != 0 {

		startTime = time.UnixMilli(syncStatus.LastDataTime)
	} else {

		// Start from max sync hours ago
		startTime = time.Now().Add(-time.Duration(s.config.MaxSyncHours) * time.Hour)
	}

	endTime := time.Now()

	// Fetch and store klines in batches
	currentTime := startTime
	totalKlines := 0

	for currentTime.Before(endTime) {

		select {
		case <-ctx.Done():

			return ctx.Err()
		default:
		}

		// Calculate batch end time
		batchEndTime := currentTime.Add(time.Duration(s.config.BatchSize) * getIntervalDuration(interval))
		if batchEndTime.After(endTime) {

			batchEndTime = endTime
		}

		// Fetch klines from Binance
		klines, err := s.binanceClient.REST.GetKlines(ctx, symbol, interval, &currentTime, &batchEndTime, s.config.BatchSize)
		if err != nil {

			return fmt.Errorf("failed to fetch klines: %w", err)
		}

		if len(klines) == 0 {

			break
		}

		// Convert and store klines
		modelKlines := make([]models.Kline, 0, len(klines))
		for _, k := range klines {

			klineData, err := binance.ParseKlineResponse(k)
			if err != nil {

				s.logger.Warn("Failed to parse kline", zap.Error(err))
				continue
			}

			modelKline, err := s.convertToModelKline(symbol, interval, klineData)
			if err != nil {

				s.logger.Warn("Failed to convert kline", zap.Error(err))
				continue
			}

			modelKlines = append(modelKlines, *modelKline)
		}

		// Batch insert klines with retry logic and rate limiting
		if len(modelKlines) > 0 {

			if err := s.klineRepo.BatchInsert(ctx, modelKlines); err != nil {

				return fmt.Errorf("failed to insert klines: %w", err)
			}

			totalKlines += len(modelKlines)

			// Update sync status with additional delay
			lastKline := modelKlines[len(modelKlines)-1]
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(50 * time.Millisecond):
			}

			if err := s.syncStatusRepo.UpsertSyncStatus(ctx, &models.SyncStatus{
				Symbol:       symbol,
				DataType:     "kline",
				Interval:     &interval,
				LastSyncTime: time.Now().UnixMilli(),
				LastDataTime: lastKline.OpenTime,
				Status:       "active",
				ErrorMessage: nil,
				UpdatedAt:    time.Now().UnixMilli(),
			}); err != nil {
				
				s.logger.Warn("Failed to update sync status", zap.Error(err))
			}
		}

		// Move to next batch
		currentTime = batchEndTime
	}

	s.logger.Info("Klines synced successfully",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.Int("total_klines", totalKlines),
	)

	return nil
}

// convertToModelKline converts Binance kline data to model
func (s *DataSyncService) convertToModelKline(symbol, interval string, data *binance.KlineData) (*models.Kline, error) {
	openPrice, _ := strconv.ParseFloat(data.Open, 64)
	highPrice, _ := strconv.ParseFloat(data.High, 64)
	lowPrice, _ := strconv.ParseFloat(data.Low, 64)
	closePrice, _ := strconv.ParseFloat(data.Close, 64)
	volume, _ := strconv.ParseFloat(data.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(data.QuoteAssetVolume, 64)
	takerBuyVolume, _ := strconv.ParseFloat(data.TakerBuyBaseAssetVolume, 64)
	takerBuyQuoteVolume, _ := strconv.ParseFloat(data.TakerBuyQuoteAssetVolume, 64)

	return &models.Kline{
		Symbol:              symbol,
		Interval:            interval,
		OpenTime:            data.OpenTime,
		CloseTime:           data.CloseTime,
		OpenPrice:           openPrice,
		HighPrice:           highPrice,
		LowPrice:            lowPrice,
		ClosePrice:          closePrice,
		Volume:              volume,
		QuoteVolume:         quoteVolume,
		TradesCount:         data.NumberOfTrades,
		TakerBuyVolume:      takerBuyVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
		CreatedAt:           time.Now().UnixMilli(),
	}, nil
}

// getIntervalDuration converts interval string to duration
func getIntervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "8h":
		return 8 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "3d":
		return 3 * 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	case "1M":
		return 30 * 24 * time.Hour
	default:
		return time.Hour
	}
}
