package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// KlineRepository handles kline data operations
type KlineRepository struct {
	db *database.Database
}

// NewKlineRepository creates a new kline repository
func NewKlineRepository(db *database.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// Insert inserts a single kline record
func (r *KlineRepository) Insert(ctx context.Context, kline *models.Kline) error {
	query := `
		INSERT INTO klines (
			symbol, interval, open_time, close_time, open_price, high_price,
			low_price, close_price, volume, quote_volume, trades_count,
			taker_buy_volume, taker_buy_quote_volume
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (symbol, interval, open_time) DO UPDATE SET
			close_time = EXCLUDED.close_time,
			open_price = EXCLUDED.open_price,
			high_price = EXCLUDED.high_price,
			low_price = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume = EXCLUDED.volume,
			quote_volume = EXCLUDED.quote_volume,
			trades_count = EXCLUDED.trades_count,
			taker_buy_volume = EXCLUDED.taker_buy_volume,
			taker_buy_quote_volume = EXCLUDED.taker_buy_quote_volume
	`

	_, err := r.db.Pool.Exec(ctx, query,
		kline.Symbol, kline.Interval, kline.OpenTime, kline.CloseTime,
		kline.OpenPrice, kline.HighPrice, kline.LowPrice, kline.ClosePrice,
		kline.Volume, kline.QuoteVolume, kline.TradesCount,
		kline.TakerBuyVolume, kline.TakerBuyQuoteVolume,
	)

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

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	query := `
		INSERT INTO klines (
			symbol, interval, open_time, close_time, open_price, high_price,
			low_price, close_price, volume, quote_volume, trades_count,
			taker_buy_volume, taker_buy_quote_volume
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (symbol, interval, open_time) DO UPDATE SET
			close_time = EXCLUDED.close_time,
			open_price = EXCLUDED.open_price,
			high_price = EXCLUDED.high_price,
			low_price = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume = EXCLUDED.volume,
			quote_volume = EXCLUDED.quote_volume,
			trades_count = EXCLUDED.trades_count,
			taker_buy_volume = EXCLUDED.taker_buy_volume,
			taker_buy_quote_volume = EXCLUDED.taker_buy_quote_volume
	`

	for _, kline := range klines {
		batch.Queue(query,
			kline.Symbol, kline.Interval, kline.OpenTime, kline.CloseTime,
			kline.OpenPrice, kline.HighPrice, kline.LowPrice, kline.ClosePrice,
			kline.Volume, kline.QuoteVolume, kline.TradesCount,
			kline.TakerBuyVolume, kline.TakerBuyQuoteVolume,
		)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range klines {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLastKline retrieves the most recent kline for a symbol and interval
func (r *KlineRepository) GetLastKline(ctx context.Context, symbol, interval string) (*models.Kline, error) {
	query := `
		SELECT symbol, interval, open_time, close_time, open_price, high_price,
			   low_price, close_price, volume, quote_volume, trades_count,
			   taker_buy_volume, taker_buy_quote_volume, created_at
		FROM klines
		WHERE symbol = $1 AND interval = $2
		ORDER BY open_time DESC
		LIMIT 1
	`

	var kline models.Kline
	err := r.db.Pool.QueryRow(ctx, query, symbol, interval).Scan(
		&kline.Symbol, &kline.Interval, &kline.OpenTime, &kline.CloseTime,
		&kline.OpenPrice, &kline.HighPrice, &kline.LowPrice, &kline.ClosePrice,
		&kline.Volume, &kline.QuoteVolume, &kline.TradesCount,
		&kline.TakerBuyVolume, &kline.TakerBuyQuoteVolume, &kline.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No data found
		}
		return nil, fmt.Errorf("failed to get last kline: %w", err)
	}

	return &kline, nil
}

// GetKlinesByTimeRange retrieves klines within a time range
func (r *KlineRepository) GetKlinesByTimeRange(
	ctx context.Context,
	symbol, interval string,
	startTime, endTime time.Time,
) ([]models.Kline, error) {
	query := `
		SELECT symbol, interval, open_time, close_time, open_price, high_price,
			   low_price, close_price, volume, quote_volume, trades_count,
			   taker_buy_volume, taker_buy_quote_volume, created_at
		FROM klines
		WHERE symbol = $1 AND interval = $2
		  AND open_time >= $3 AND open_time < $4
		ORDER BY open_time ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, symbol, interval, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query klines: %w", err)
	}
	defer rows.Close()

	var klines []models.Kline
	for rows.Next() {
		var kline models.Kline
		err := rows.Scan(
			&kline.Symbol, &kline.Interval, &kline.OpenTime, &kline.CloseTime,
			&kline.OpenPrice, &kline.HighPrice, &kline.LowPrice, &kline.ClosePrice,
			&kline.Volume, &kline.QuoteVolume, &kline.TradesCount,
			&kline.TakerBuyVolume, &kline.TakerBuyQuoteVolume, &kline.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan kline: %w", err)
		}
		klines = append(klines, kline)
	}

	return klines, nil
}
