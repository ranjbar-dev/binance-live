-- name: GetActiveSymbols :many
SELECT id, symbol, base_asset, quote_asset, status, is_active, created_at, updated_at
FROM symbols
WHERE is_active = true
ORDER BY symbol;

-- name: GetSymbolByName :one
SELECT id, symbol, base_asset, quote_asset, status, is_active, created_at, updated_at
FROM symbols
WHERE symbol = $1;

-- name: UpsertSymbol :one
INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (symbol) DO UPDATE SET
    base_asset = EXCLUDED.base_asset,
    quote_asset = EXCLUDED.quote_asset,
    status = EXCLUDED.status,
    is_active = EXCLUDED.is_active,
    updated_at = EXTRACT(EPOCH FROM NOW()) * 1000
RETURNING id, created_at, updated_at;

-- name: UpdateSymbolStatus :exec
UPDATE symbols
SET is_active = $2, updated_at = EXTRACT(EPOCH FROM NOW()) * 1000
WHERE symbol = $1;

-- name: GetAllSymbols :many
SELECT id, symbol, base_asset, quote_asset, status, is_active, created_at, updated_at
FROM symbols
ORDER BY symbol;

-- name: DeleteSymbol :exec
DELETE FROM symbols WHERE symbol = $1;
