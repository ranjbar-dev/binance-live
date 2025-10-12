-- name: InsertDepthSnapshot :one
INSERT INTO depth_snapshots (
    symbol, timestamp, last_update_id, bids, asks
) VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at;

-- name: GetLatestDepthSnapshot :one
SELECT id, symbol, timestamp, last_update_id, bids, asks, created_at
FROM depth_snapshots
WHERE symbol = $1
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetDepthSnapshotsByTimeRange :many
SELECT id, symbol, timestamp, last_update_id, bids, asks, created_at
FROM depth_snapshots
WHERE symbol = $1
  AND timestamp >= $2 AND timestamp < $3
ORDER BY timestamp ASC;

-- name: DeleteOldDepthSnapshots :exec
DELETE FROM depth_snapshots 
WHERE timestamp < $1;
