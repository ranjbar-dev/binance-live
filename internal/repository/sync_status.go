package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/db"
	"github.com/binance-live/internal/models"
	"github.com/jackc/pgx/v5"
)

// SyncStatusRepository handles sync status operations
type SyncStatusRepository struct {
	database *database.Database
	queries  *db.Queries
}

// NewSyncStatusRepository creates a new sync status repository
func NewSyncStatusRepository(database *database.Database) *SyncStatusRepository {
	return &SyncStatusRepository{
		database: database,
		queries:  db.New(database.Pool),
	}
}

// GetSyncStatus retrieves the sync status for a symbol and data type
func (r *SyncStatusRepository) GetSyncStatus(
	ctx context.Context,
	symbol, dataType string,
	interval *string,
) (*models.SyncStatus, error) {
	var intervalParam sql.NullString
	if interval != nil {
		intervalParam = sql.NullString{String: *interval, Valid: true}
	}

	dbStatus, err := r.queries.GetSyncStatus(ctx, db.GetSyncStatusParams{
		Symbol:   symbol,
		DataType: dataType,
		Interval: intervalParam,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No sync status found
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	status := &models.SyncStatus{
		Symbol:       dbStatus.Symbol,
		DataType:     dbStatus.DataType,
		LastSyncTime: dbStatus.LastSyncTime,
		LastDataTime: dbStatus.LastDataTime,
		Status:       dbStatus.Status,
		UpdatedAt:    dbStatus.UpdatedAt,
	}

	if dbStatus.Interval.Valid {
		status.Interval = &dbStatus.Interval.String
	}
	if dbStatus.ErrorMessage.Valid {
		status.ErrorMessage = &dbStatus.ErrorMessage.String
	}

	return status, nil
}

// UpsertSyncStatus inserts or updates sync status
func (r *SyncStatusRepository) UpsertSyncStatus(ctx context.Context, status *models.SyncStatus) error {
	var intervalParam sql.NullString
	if status.Interval != nil {
		intervalParam = sql.NullString{String: *status.Interval, Valid: true}
	}

	var errorMessageParam sql.NullString
	if status.ErrorMessage != nil {
		errorMessageParam = sql.NullString{String: *status.ErrorMessage, Valid: true}
	}

	err := r.queries.UpsertSyncStatus(ctx, db.UpsertSyncStatusParams{
		Symbol:       status.Symbol,
		DataType:     status.DataType,
		Interval:     intervalParam,
		LastSyncTime: status.LastSyncTime,
		LastDataTime: status.LastDataTime,
		Status:       status.Status,
		ErrorMessage: errorMessageParam,
	})

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
	lastDataTime int64,
) error {
	var intervalParam sql.NullString
	if interval != nil {
		intervalParam = sql.NullString{String: *interval, Valid: true}
	}

	err := r.queries.UpdateLastDataTime(ctx, db.UpdateLastDataTimeParams{
		Symbol:       symbol,
		DataType:     dataType,
		Interval:     intervalParam,
		LastDataTime: lastDataTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update last data time: %w", err)
	}

	return nil
}

// GetAllSyncStatuses retrieves all sync statuses for active symbols
func (r *SyncStatusRepository) GetAllSyncStatuses(ctx context.Context) ([]models.SyncStatus, error) {
	dbStatuses, err := r.queries.GetAllSyncStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync statuses: %w", err)
	}

	statuses := make([]models.SyncStatus, 0, len(dbStatuses))
	for _, dbStatus := range dbStatuses {
		status := models.SyncStatus{
			Symbol:       dbStatus.Symbol,
			DataType:     dbStatus.DataType,
			LastSyncTime: dbStatus.LastSyncTime,
			LastDataTime: dbStatus.LastDataTime,
			Status:       dbStatus.Status,
			UpdatedAt:    dbStatus.UpdatedAt,
		}

		if dbStatus.Interval.Valid {
			status.Interval = &dbStatus.Interval.String
		}
		if dbStatus.ErrorMessage.Valid {
			status.ErrorMessage = &dbStatus.ErrorMessage.String
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
