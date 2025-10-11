package repository

import (
	"context"
	"fmt"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// SymbolRepository handles symbol data operations
type SymbolRepository struct {
	db *database.Database
}

// NewSymbolRepository creates a new symbol repository
func NewSymbolRepository(db *database.Database) *SymbolRepository {
	return &SymbolRepository{db: db}
}

// GetActiveSymbols retrieves all active trading symbols
func (r *SymbolRepository) GetActiveSymbols(ctx context.Context) ([]models.Symbol, error) {
	query := `
		SELECT id, symbol, base_asset, quote_asset, status, is_active, created_at, updated_at
		FROM symbols
		WHERE is_active = true
		ORDER BY symbol
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active symbols: %w", err)
	}
	defer rows.Close()

	var symbols []models.Symbol
	for rows.Next() {
		var s models.Symbol
		err := rows.Scan(
			&s.ID, &s.Symbol, &s.BaseAsset, &s.QuoteAsset,
			&s.Status, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}

// GetSymbolByName retrieves a symbol by its name
func (r *SymbolRepository) GetSymbolByName(ctx context.Context, symbol string) (*models.Symbol, error) {
	query := `
		SELECT id, symbol, base_asset, quote_asset, status, is_active, created_at, updated_at
		FROM symbols
		WHERE symbol = $1
	`

	var s models.Symbol
	err := r.db.Pool.QueryRow(ctx, query, symbol).Scan(
		&s.ID, &s.Symbol, &s.BaseAsset, &s.QuoteAsset,
		&s.Status, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("symbol %s not found", symbol)
		}
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	return &s, nil
}

// UpsertSymbol inserts or updates a symbol
func (r *SymbolRepository) UpsertSymbol(ctx context.Context, symbol *models.Symbol) error {
	query := `
		INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (symbol) DO UPDATE SET
			base_asset = EXCLUDED.base_asset,
			quote_asset = EXCLUDED.quote_asset,
			status = EXCLUDED.status,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(ctx, query,
		symbol.Symbol, symbol.BaseAsset, symbol.QuoteAsset, symbol.Status, symbol.IsActive,
	).Scan(&symbol.ID, &symbol.CreatedAt, &symbol.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert symbol: %w", err)
	}

	return nil
}

// UpdateSymbolStatus updates the active status of a symbol
func (r *SymbolRepository) UpdateSymbolStatus(ctx context.Context, symbol string, isActive bool) error {
	query := `
		UPDATE symbols
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE symbol = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, isActive, symbol)
	if err != nil {
		return fmt.Errorf("failed to update symbol status: %w", err)
	}

	return nil
}
