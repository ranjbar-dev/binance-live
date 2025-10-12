package repository

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/db"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// SymbolRepository handles symbol data operations
type SymbolRepository struct {
	database *database.Database
	queries  *db.Queries
}

// NewSymbolRepository creates a new symbol repository
func NewSymbolRepository(database *database.Database) *SymbolRepository {
	return &SymbolRepository{
		database: database,
		queries:  db.New(database.Pool),
	}
}

// GetActiveSymbols retrieves all active trading symbols
func (r *SymbolRepository) GetActiveSymbols(ctx context.Context) ([]models.Symbol, error) {
	dbSymbols, err := r.queries.GetActiveSymbols(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query active symbols: %w", err)
	}

	symbols := make([]models.Symbol, 0, len(dbSymbols))
	for _, dbSymbol := range dbSymbols {
		symbols = append(symbols, models.Symbol{
			ID:         int(dbSymbol.ID),
			Symbol:     dbSymbol.Symbol,
			BaseAsset:  dbSymbol.BaseAsset,
			QuoteAsset: dbSymbol.QuoteAsset,
			Status:     dbSymbol.Status,
			IsActive:   dbSymbol.IsActive,
			CreatedAt:  dbSymbol.CreatedAt,
			UpdatedAt:  dbSymbol.UpdatedAt,
		})
	}

	return symbols, nil
}

// GetSymbolByName retrieves a symbol by its name
func (r *SymbolRepository) GetSymbolByName(ctx context.Context, symbol string) (*models.Symbol, error) {
	dbSymbol, err := r.queries.GetSymbolByName(ctx, symbol)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("symbol %s not found", symbol)
		}
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	return &models.Symbol{
		ID:         int(dbSymbol.ID),
		Symbol:     dbSymbol.Symbol,
		BaseAsset:  dbSymbol.BaseAsset,
		QuoteAsset: dbSymbol.QuoteAsset,
		Status:     dbSymbol.Status,
		IsActive:   dbSymbol.IsActive,
		CreatedAt:  dbSymbol.CreatedAt,
		UpdatedAt:  dbSymbol.UpdatedAt,
	}, nil
}

// UpsertSymbol inserts or updates a symbol
func (r *SymbolRepository) UpsertSymbol(ctx context.Context, symbol *models.Symbol) error {
	result, err := r.queries.UpsertSymbol(ctx, db.UpsertSymbolParams{
		Symbol:     symbol.Symbol,
		BaseAsset:  symbol.BaseAsset,
		QuoteAsset: symbol.QuoteAsset,
		Status:     symbol.Status,
		IsActive:   symbol.IsActive,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert symbol: %w", err)
	}

	symbol.ID = int(result.ID)
	symbol.CreatedAt = result.CreatedAt
	symbol.UpdatedAt = result.UpdatedAt

	return nil
}

// UpdateSymbolStatus updates the active status of a symbol
func (r *SymbolRepository) UpdateSymbolStatus(ctx context.Context, symbol string, isActive bool) error {
	err := r.queries.UpdateSymbolStatus(ctx, db.UpdateSymbolStatusParams{
		Symbol:   symbol,
		IsActive: isActive,
	})
	if err != nil {
		return fmt.Errorf("failed to update symbol status: %w", err)
	}

	return nil
}
