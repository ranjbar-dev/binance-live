package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/db"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// KlineRepository handles kline data operations
type KlineRepository struct {
	database *database.Database
	queries  *db.Queries
}

// NewKlineRepository creates a new kline repository
func NewKlineRepository(database *database.Database) *KlineRepository {
	return &KlineRepository{
		database: database,
		queries:  db.New(database.Pool),
	}
}

// Insert inserts a single kline record
func (r *KlineRepository) Insert(ctx context.Context, kline *models.Kline) error {
	err := r.queries.InsertKline(ctx, db.InsertKlineParams{
		Symbol:              kline.Symbol,
		Interval:            kline.Interval,
		OpenTime:            kline.OpenTime,
		CloseTime:           kline.CloseTime,
		OpenPrice:           kline.OpenPrice,
		HighPrice:           kline.HighPrice,
		LowPrice:            kline.LowPrice,
		ClosePrice:          kline.ClosePrice,
		Volume:              kline.Volume,
		QuoteVolume:         kline.QuoteVolume,
		TradesCount:         int32(kline.TradesCount),
		TakerBuyVolume:      kline.TakerBuyVolume,
		TakerBuyQuoteVolume: kline.TakerBuyQuoteVolume,
	})

	if err != nil {
		return fmt.Errorf("failed to insert kline: %w", err)
	}

	return nil
}

// BatchInsert inserts multiple kline records in a single transaction
func (r *KlineRepository) BatchInsert(ctx context.Context, klines []models.Kline) error {

	if len(klines) == 0 {
		return nil
	}

	// For large batches, process in smaller chunks to avoid long-running transactions
	const maxBatchSize = 100 // Further reduced for better connection management
	if len(klines) > maxBatchSize {

		return r.batchInsertChunked(ctx, klines, maxBatchSize)
	}

	return r.executeBatchInsert(ctx, klines)
}

// batchInsertChunked processes large batches in smaller chunks with delays
func (r *KlineRepository) batchInsertChunked(ctx context.Context, klines []models.Kline, chunkSize int) error {
	for i := 0; i < len(klines); i += chunkSize {
		end := i + chunkSize
		if end > len(klines) {
			end = len(klines)
		}

		// Add delay between chunks to prevent connection pool exhaustion
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}

		chunk := klines[i:end]
		if err := r.executeBatchInsert(ctx, chunk); err != nil {
			return err
		}
	}
	return nil
}

// executeBatchInsert executes a batch insert with proper transaction management
func (r *KlineRepository) executeBatchInsert(ctx context.Context, klines []models.Kline) error {
	// Add timeout context to prevent long-running transactions
	txCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.database.Pool.Begin(txCtx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Use explicit rollback handling
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				// Log rollback error but don't overwrite the original error
			}
		}
	}()

	// Use sqlc queries with transaction
	txQueries := r.queries.WithTx(tx)

	for _, kline := range klines {
		err := txQueries.InsertKline(txCtx, db.InsertKlineParams{
			Symbol:              kline.Symbol,
			Interval:            kline.Interval,
			OpenTime:            kline.OpenTime,
			CloseTime:           kline.CloseTime,
			OpenPrice:           kline.OpenPrice,
			HighPrice:           kline.HighPrice,
			LowPrice:            kline.LowPrice,
			ClosePrice:          kline.ClosePrice,
			Volume:              kline.Volume,
			QuoteVolume:         kline.QuoteVolume,
			TradesCount:         int32(kline.TradesCount),
			TakerBuyVolume:      kline.TakerBuyVolume,
			TakerBuyQuoteVolume: kline.TakerBuyQuoteVolume,
		})
		if err != nil {
			return fmt.Errorf("failed to insert kline: %w", err)
		}
	}

	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

// GetLastKline retrieves the most recent kline for a symbol and interval
func (r *KlineRepository) GetLastKline(ctx context.Context, symbol, interval string) (*models.Kline, error) {
	dbKline, err := r.queries.GetLastKline(ctx, db.GetLastKlineParams{
		Symbol:   symbol,
		Interval: interval,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No data found
		}
		return nil, fmt.Errorf("failed to get last kline: %w", err)
	}

	return &models.Kline{
		Symbol:              dbKline.Symbol,
		Interval:            dbKline.Interval,
		OpenTime:            dbKline.OpenTime,
		CloseTime:           dbKline.CloseTime,
		OpenPrice:           dbKline.OpenPrice,
		HighPrice:           dbKline.HighPrice,
		LowPrice:            dbKline.LowPrice,
		ClosePrice:          dbKline.ClosePrice,
		Volume:              dbKline.Volume,
		QuoteVolume:         dbKline.QuoteVolume,
		TradesCount:         int(dbKline.TradesCount),
		TakerBuyVolume:      dbKline.TakerBuyVolume,
		TakerBuyQuoteVolume: dbKline.TakerBuyQuoteVolume,
		CreatedAt:           dbKline.CreatedAt,
	}, nil
}

// GetKlinesByTimeRange retrieves klines within a time range
func (r *KlineRepository) GetKlinesByTimeRange(
	ctx context.Context,
	symbol, interval string,
	startTime, endTime int64,
) ([]models.Kline, error) {
	dbKlines, err := r.queries.GetKlinesByTimeRange(ctx, db.GetKlinesByTimeRangeParams{
		Symbol:     symbol,
		Interval:   interval,
		OpenTime:   startTime,
		OpenTime_2: endTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query klines: %w", err)
	}

	klines := make([]models.Kline, 0, len(dbKlines))
	for _, dbKline := range dbKlines {
		klines = append(klines, models.Kline{
			Symbol:              dbKline.Symbol,
			Interval:            dbKline.Interval,
			OpenTime:            dbKline.OpenTime,
			CloseTime:           dbKline.CloseTime,
			OpenPrice:           dbKline.OpenPrice,
			HighPrice:           dbKline.HighPrice,
			LowPrice:            dbKline.LowPrice,
			ClosePrice:          dbKline.ClosePrice,
			Volume:              dbKline.Volume,
			QuoteVolume:         dbKline.QuoteVolume,
			TradesCount:         int(dbKline.TradesCount),
			TakerBuyVolume:      dbKline.TakerBuyVolume,
			TakerBuyQuoteVolume: dbKline.TakerBuyQuoteVolume,
			CreatedAt:           dbKline.CreatedAt,
		})
	}

	return klines, nil
}
