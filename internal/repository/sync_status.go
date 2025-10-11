package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// SyncStatusRepository handles sync status operations
type SyncStatusRepository struct {
	db *database.Database
}

// NewSyncStatusRepository creates a new sync status repository
func NewSyncStatusRepository(db *database.Database) *SyncStatusRepository {
	return &SyncStatusRepository{db: db}
}

// GetSyncStatus retrieves the sync status for a symbol and data type
func (r *SyncStatusRepository) GetSyncStatus(
	ctx context.Context,
	symbol, dataType string,
	interval *string,
) (*models.SyncStatus, error) {
	query := `
		SELECT symbol, data_type, interval, last_sync_time, last_data_time,
			   status, error_message, updated_at
		FROM sync_status
		WHERE symbol = $1 AND data_type = $2 AND COALESCE(interval, '') = COALESCE($3, '')
	`

	var status models.SyncStatus
	err := r.db.Pool.QueryRow(ctx, query, symbol, dataType, interval).Scan(
		&status.Symbol, &status.DataType, &status.Interval,
		&status.LastSyncTime, &status.LastDataTime,
		&status.Status, &status.ErrorMessage, &status.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No sync status found
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return &status, nil
}

// UpsertSyncStatus inserts or updates sync status
func (r *SyncStatusRepository) UpsertSyncStatus(ctx context.Context, status *models.SyncStatus) error {
	query := `
		INSERT INTO sync_status (
			symbol, data_type, interval, last_sync_time, last_data_time, status, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (symbol, data_type, COALESCE(interval, '')) DO UPDATE SET
			last_sync_time = EXCLUDED.last_sync_time,
			last_data_time = EXCLUDED.last_data_time,
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Pool.Exec(ctx, query,
		status.Symbol, status.DataType, status.Interval,
		status.LastSyncTime, status.LastDataTime,
		status.Status, status.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert sync status: %w", err)
	}

	return nil
}

// UpdateLastDataTime updates the last data time for a sync status
func (r *SyncStatusRepository) UpdateLastDataTime(
	ctx context.Context,
	symbol, dataType string,
	interval *string,
	lastDataTime time.Time,
) error {
	query := `
		UPDATE sync_status
		SET last_data_time = $1,
			last_sync_time = CURRENT_TIMESTAMP,
			status = 'active',
			error_message = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE symbol = $2 AND data_type = $3 AND COALESCE(interval, '') = COALESCE($4, '')
	`

	_, err := r.db.Pool.Exec(ctx, query, lastDataTime, symbol, dataType, interval)
	if err != nil {
		return fmt.Errorf("failed to update last data time: %w", err)
	}

	return nil
}

// GetAllSyncStatuses retrieves all sync statuses for active symbols
func (r *SyncStatusRepository) GetAllSyncStatuses(ctx context.Context) ([]models.SyncStatus, error) {
	query := `
		SELECT s.symbol, s.data_type, s.interval, s.last_sync_time, s.last_data_time,
			   s.status, s.error_message, s.updated_at
		FROM sync_status s
		INNER JOIN symbols sym ON s.symbol = sym.symbol
		WHERE sym.is_active = true
		ORDER BY s.symbol, s.data_type, s.interval
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync statuses: %w", err)
	}
	defer rows.Close()

	var statuses []models.SyncStatus
	for rows.Next() {
		var status models.SyncStatus
		err := rows.Scan(
			&status.Symbol, &status.DataType, &status.Interval,
			&status.LastSyncTime, &status.LastDataTime,
			&status.Status, &status.ErrorMessage, &status.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}
