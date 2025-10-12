-- name: InsertTrade :one
INSERT INTO trades (
    symbol, trade_id, timestamp, price, quantity, quote_quantity, is_buyer_maker
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (symbol, trade_id, timestamp) DO UPDATE SET
    price = EXCLUDED.price,
    quantity = EXCLUDED.quantity,
    quote_quantity = EXCLUDED.quote_quantity,
    is_buyer_maker = EXCLUDED.is_buyer_maker
RETURNING id, created_at;

-- name: GetTradesByTimeRange :many
SELECT id, symbol, trade_id, timestamp, price, quantity, quote_quantity, is_buyer_maker, created_at
FROM trades
WHERE symbol = $1
  AND timestamp >= $2 AND timestamp < $3
ORDER BY timestamp ASC;

-- name: GetLatestTrades :many
SELECT id, symbol, trade_id, timestamp, price, quantity, quote_quantity, is_buyer_maker, created_at
FROM trades
WHERE symbol = $1
ORDER BY timestamp DESC
LIMIT $2;

-- name: DeleteOldTrades :exec
DELETE FROM trades 
WHERE timestamp < $1;
