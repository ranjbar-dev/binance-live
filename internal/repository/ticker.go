package repository

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
)

// TickerRepository handles ticker data operations
type TickerRepository struct {
	db *database.Database
}

// NewTickerRepository creates a new ticker repository
func NewTickerRepository(db *database.Database) *TickerRepository {
	return &TickerRepository{db: db}
}

// Insert inserts a ticker record
func (r *TickerRepository) Insert(ctx context.Context, ticker *models.Ticker) error {
	query := `
		INSERT INTO tickers (
			symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
			volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
			high_24h, low_24h, trades_count_24h
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (symbol, timestamp) DO UPDATE SET
			price = EXCLUDED.price,
			bid_price = EXCLUDED.bid_price,
			bid_qty = EXCLUDED.bid_qty,
			ask_price = EXCLUDED.ask_price,
			ask_qty = EXCLUDED.ask_qty,
			volume_24h = EXCLUDED.volume_24h,
			quote_volume_24h = EXCLUDED.quote_volume_24h,
			price_change_24h = EXCLUDED.price_change_24h,
			price_change_percent_24h = EXCLUDED.price_change_percent_24h,
			high_24h = EXCLUDED.high_24h,
			low_24h = EXCLUDED.low_24h,
			trades_count_24h = EXCLUDED.trades_count_24h
	`

	_, err := r.db.Pool.Exec(ctx, query,
		ticker.Symbol, ticker.Timestamp, ticker.Price,
		ticker.BidPrice, ticker.BidQty, ticker.AskPrice, ticker.AskQty,
		ticker.Volume24h, ticker.QuoteVolume24h, ticker.PriceChange24h,
		ticker.PriceChangePercent24h, ticker.High24h, ticker.Low24h,
		ticker.TradesCount24h,
	)

	if err != nil {
		return fmt.Errorf("failed to insert ticker: %w", err)
	}

	return nil
}

// BatchInsert inserts multiple ticker records
func (r *TickerRepository) BatchInsert(ctx context.Context, tickers []models.Ticker) error {
	if len(tickers) == 0 {
		return nil
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO tickers (
			symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
			volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
			high_24h, low_24h, trades_count_24h
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (symbol, timestamp) DO UPDATE SET
			price = EXCLUDED.price,
			bid_price = EXCLUDED.bid_price,
			bid_qty = EXCLUDED.bid_qty,
			ask_price = EXCLUDED.ask_price,
			ask_qty = EXCLUDED.ask_qty,
			volume_24h = EXCLUDED.volume_24h,
			quote_volume_24h = EXCLUDED.quote_volume_24h,
			price_change_24h = EXCLUDED.price_change_24h,
			price_change_percent_24h = EXCLUDED.price_change_percent_24h,
			high_24h = EXCLUDED.high_24h,
			low_24h = EXCLUDED.low_24h,
			trades_count_24h = EXCLUDED.trades_count_24h
	`

	for _, ticker := range tickers {
		_, err := tx.Exec(ctx, query,
			ticker.Symbol, ticker.Timestamp, ticker.Price,
			ticker.BidPrice, ticker.BidQty, ticker.AskPrice, ticker.AskQty,
			ticker.Volume24h, ticker.QuoteVolume24h, ticker.PriceChange24h,
			ticker.PriceChangePercent24h, ticker.High24h, ticker.Low24h,
			ticker.TradesCount24h,
		)
		if err != nil {
			return fmt.Errorf("failed to insert ticker: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
