-- name: GetSyncStatus :one
SELECT symbol, data_type, interval, last_sync_time, last_data_time,
       status, error_message, updated_at
FROM sync_status
WHERE symbol = $1 AND data_type = $2 AND COALESCE(interval, '') = COALESCE($3, '');

-- name: UpsertSyncStatus :exec
INSERT INTO sync_status (
    symbol, data_type, interval, last_sync_time, last_data_time, status, error_message
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (symbol, data_type, interval) DO UPDATE SET
    last_sync_time = EXCLUDED.last_sync_time,
    last_data_time = EXCLUDED.last_data_time,
    status = EXCLUDED.status,
    error_message = EXCLUDED.error_message,
    updated_at = EXTRACT(EPOCH FROM NOW()) * 1000;

-- name: UpdateLastDataTime :exec
UPDATE sync_status
SET last_data_time = $4,
    last_sync_time = EXTRACT(EPOCH FROM NOW()) * 1000,
    status = 'active',
    error_message = NULL,
    updated_at = EXTRACT(EPOCH FROM NOW()) * 1000
WHERE symbol = $1 AND data_type = $2 AND COALESCE(interval, '') = COALESCE($3, '');

-- name: GetAllSyncStatuses :many
SELECT s.symbol, s.data_type, s.interval, s.last_sync_time, s.last_data_time,
       s.status, s.error_message, s.updated_at
FROM sync_status s
INNER JOIN symbols sym ON s.symbol = sym.symbol
WHERE sym.is_active = true
ORDER BY s.symbol, s.data_type, s.interval;

-- name: GetSyncStatusesBySymbol :many
SELECT symbol, data_type, interval, last_sync_time, last_data_time,
       status, error_message, updated_at
FROM sync_status
WHERE symbol = $1
ORDER BY data_type, interval;

-- name: DeleteSyncStatus :exec
DELETE FROM sync_status 
WHERE symbol = $1 AND data_type = $2 AND COALESCE(interval, '') = COALESCE($3, '');
