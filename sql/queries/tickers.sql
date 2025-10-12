-- name: InsertTicker :exec
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
    trades_count_24h = EXCLUDED.trades_count_24h;

-- name: GetLatestTicker :one
SELECT symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
       volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
       high_24h, low_24h, trades_count_24h, created_at
FROM tickers
WHERE symbol = $1
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetTickersByTimeRange :many
SELECT symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
       volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
       high_24h, low_24h, trades_count_24h, created_at
FROM tickers
WHERE symbol = $1
  AND timestamp >= $2 AND timestamp < $3
ORDER BY timestamp ASC;

-- name: GetAllLatestTickers :many
SELECT DISTINCT ON (symbol) symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
       volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
       high_24h, low_24h, trades_count_24h, created_at
FROM tickers
ORDER BY symbol, timestamp DESC;

-- name: DeleteOldTickers :exec
DELETE FROM tickers 
WHERE timestamp < $1;
