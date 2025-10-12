-- name: InsertKline :exec
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
    taker_buy_quote_volume = EXCLUDED.taker_buy_quote_volume;

-- name: GetLastKline :one
SELECT symbol, interval, open_time, close_time, open_price, high_price,
       low_price, close_price, volume, quote_volume, trades_count,
       taker_buy_volume, taker_buy_quote_volume, created_at
FROM klines
WHERE symbol = $1 AND interval = $2
ORDER BY open_time DESC
LIMIT 1;

-- name: GetKlinesByTimeRange :many
SELECT symbol, interval, open_time, close_time, open_price, high_price,
       low_price, close_price, volume, quote_volume, trades_count,
       taker_buy_volume, taker_buy_quote_volume, created_at
FROM klines
WHERE symbol = $1 AND interval = $2
  AND open_time >= $3 AND open_time < $4
ORDER BY open_time ASC;

-- name: GetLatestKlines :many
SELECT symbol, interval, open_time, close_time, open_price, high_price,
       low_price, close_price, volume, quote_volume, trades_count,
       taker_buy_volume, taker_buy_quote_volume, created_at
FROM klines
WHERE symbol = $1 AND interval = $2
ORDER BY open_time DESC
LIMIT $3;

-- name: DeleteOldKlines :exec
DELETE FROM klines 
WHERE open_time < $1;
