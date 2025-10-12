package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/db"
	"github.com/binance-live/internal/models"
)

// TickerRepository handles ticker data operations
type TickerRepository struct {
	database *database.Database
	queries  *db.Queries
}

// NewTickerRepository creates a new ticker repository
func NewTickerRepository(database *database.Database) *TickerRepository {
	return &TickerRepository{
		database: database,
		queries:  db.New(database.Pool),
	}
}

// Insert inserts a ticker record
func (r *TickerRepository) Insert(ctx context.Context, ticker *models.Ticker) error {
	// Convert nullable pointers to sql.NullFloat64 and sql.NullInt32
	var bidPrice, bidQty, askPrice, askQty sql.NullFloat64
	var volume24h, quoteVolume24h, priceChange24h, priceChangePercent24h sql.NullFloat64
	var high24h, low24h sql.NullFloat64
	var tradesCount24h sql.NullInt32

	if ticker.BidPrice != nil {
		bidPrice = sql.NullFloat64{Float64: *ticker.BidPrice, Valid: true}
	}
	if ticker.BidQty != nil {
		bidQty = sql.NullFloat64{Float64: *ticker.BidQty, Valid: true}
	}
	if ticker.AskPrice != nil {
		askPrice = sql.NullFloat64{Float64: *ticker.AskPrice, Valid: true}
	}
	if ticker.AskQty != nil {
		askQty = sql.NullFloat64{Float64: *ticker.AskQty, Valid: true}
	}
	if ticker.Volume24h != nil {
		volume24h = sql.NullFloat64{Float64: *ticker.Volume24h, Valid: true}
	}
	if ticker.QuoteVolume24h != nil {
		quoteVolume24h = sql.NullFloat64{Float64: *ticker.QuoteVolume24h, Valid: true}
	}
	if ticker.PriceChange24h != nil {
		priceChange24h = sql.NullFloat64{Float64: *ticker.PriceChange24h, Valid: true}
	}
	if ticker.PriceChangePercent24h != nil {
		priceChangePercent24h = sql.NullFloat64{Float64: *ticker.PriceChangePercent24h, Valid: true}
	}
	if ticker.High24h != nil {
		high24h = sql.NullFloat64{Float64: *ticker.High24h, Valid: true}
	}
	if ticker.Low24h != nil {
		low24h = sql.NullFloat64{Float64: *ticker.Low24h, Valid: true}
	}
	if ticker.TradesCount24h != nil {
		tradesCount24h = sql.NullInt32{Int32: int32(*ticker.TradesCount24h), Valid: true}
	}

	err := r.queries.InsertTicker(ctx, db.InsertTickerParams{
		Symbol:                ticker.Symbol,
		Timestamp:             ticker.Timestamp,
		Price:                 ticker.Price,
		BidPrice:              bidPrice,
		BidQty:                bidQty,
		AskPrice:              askPrice,
		AskQty:                askQty,
		Volume24h:             volume24h,
		QuoteVolume24h:        quoteVolume24h,
		PriceChange24h:        priceChange24h,
		PriceChangePercent24h: priceChangePercent24h,
		High24h:               high24h,
		Low24h:                low24h,
		TradesCount24h:        tradesCount24h,
	})

	if err != nil {
		return fmt.Errorf("failed to insert ticker: %w", err)
	}

	return nil
}

// BatchInsert inserts multiple ticker records with improved transaction management
func (r *TickerRepository) BatchInsert(ctx context.Context, tickers []models.Ticker) error {
	if len(tickers) == 0 {
		return nil
	}

	// For large batches, process in smaller chunks to avoid long-running transactions
	const maxBatchSize = 100 // Further reduced for better connection management
	if len(tickers) > maxBatchSize {
		return r.batchInsertChunked(ctx, tickers, maxBatchSize)
	}

	return r.executeBatchInsert(ctx, tickers)
}

// batchInsertChunked processes large batches in smaller chunks with delays
func (r *TickerRepository) batchInsertChunked(ctx context.Context, tickers []models.Ticker, chunkSize int) error {
	for i := 0; i < len(tickers); i += chunkSize {
		end := i + chunkSize
		if end > len(tickers) {
			end = len(tickers)
		}

		// Add delay between chunks to prevent connection pool exhaustion
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}

		chunk := tickers[i:end]
		if err := r.executeBatchInsert(ctx, chunk); err != nil {
			return err
		}
	}
	return nil
}

// executeBatchInsert executes a batch insert with proper transaction management
func (r *TickerRepository) executeBatchInsert(ctx context.Context, tickers []models.Ticker) error {
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

	for _, ticker := range tickers {
		// Convert nullable pointers to sql.NullFloat64 and sql.NullInt32
		var bidPrice, bidQty, askPrice, askQty sql.NullFloat64
		var volume24h, quoteVolume24h, priceChange24h, priceChangePercent24h sql.NullFloat64
		var high24h, low24h sql.NullFloat64
		var tradesCount24h sql.NullInt32

		if ticker.BidPrice != nil {
			bidPrice = sql.NullFloat64{Float64: *ticker.BidPrice, Valid: true}
		}
		if ticker.BidQty != nil {
			bidQty = sql.NullFloat64{Float64: *ticker.BidQty, Valid: true}
		}
		if ticker.AskPrice != nil {
			askPrice = sql.NullFloat64{Float64: *ticker.AskPrice, Valid: true}
		}
		if ticker.AskQty != nil {
			askQty = sql.NullFloat64{Float64: *ticker.AskQty, Valid: true}
		}
		if ticker.Volume24h != nil {
			volume24h = sql.NullFloat64{Float64: *ticker.Volume24h, Valid: true}
		}
		if ticker.QuoteVolume24h != nil {
			quoteVolume24h = sql.NullFloat64{Float64: *ticker.QuoteVolume24h, Valid: true}
		}
		if ticker.PriceChange24h != nil {
			priceChange24h = sql.NullFloat64{Float64: *ticker.PriceChange24h, Valid: true}
		}
		if ticker.PriceChangePercent24h != nil {
			priceChangePercent24h = sql.NullFloat64{Float64: *ticker.PriceChangePercent24h, Valid: true}
		}
		if ticker.High24h != nil {
			high24h = sql.NullFloat64{Float64: *ticker.High24h, Valid: true}
		}
		if ticker.Low24h != nil {
			low24h = sql.NullFloat64{Float64: *ticker.Low24h, Valid: true}
		}
		if ticker.TradesCount24h != nil {
			tradesCount24h = sql.NullInt32{Int32: int32(*ticker.TradesCount24h), Valid: true}
		}

		err := txQueries.InsertTicker(txCtx, db.InsertTickerParams{
			Symbol:                ticker.Symbol,
			Timestamp:             ticker.Timestamp,
			Price:                 ticker.Price,
			BidPrice:              bidPrice,
			BidQty:                bidQty,
			AskPrice:              askPrice,
			AskQty:                askQty,
			Volume24h:             volume24h,
			QuoteVolume24h:        quoteVolume24h,
			PriceChange24h:        priceChange24h,
			PriceChangePercent24h: priceChangePercent24h,
			High24h:               high24h,
			Low24h:                low24h,
			TradesCount24h:        tradesCount24h,
		})
		if err != nil {
			return fmt.Errorf("failed to insert ticker: %w", err)
		}
	}

	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}
